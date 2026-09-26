package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

func metadataSourceExtension(src string) string {
	return utils.MediaSourceExtension(src)
}

// 采集运行期间可能被排除、整理或补录；已经失效的采集结果不能写回。
func metadataCanSaveProbe(tx *gorm.DB, ds models.ResourcesDramaSeries) (bool, error) {
	reason, err := models.MetadataExclusionSQL(tx)
	if err != nil {
		return false, err
	}
	var count int64
	err = tx.Table("resourcesDramaSeries ds").Joins("JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id").
		Where("ds.id = ? AND ds.src = ? AND vm.probe_status = ?", ds.ID, ds.Src, models.VideoMetadataStatusProcessing).
		Where("(" + reason + ") = ''").Count(&count).Error
	if err != nil {
		return false, err
	}
	if count == 0 {
		// 若采集中途改变了规则，释放 processing 状态，撤销规则后仍可正常重采。
		err = tx.Model(&models.ResourcesVideoMetadata{}).Where("drama_series_id = ? AND probe_status = ?", ds.ID, models.VideoMetadataStatusProcessing).Updates(map[string]interface{}{"probe_status": models.VideoMetadataStatusStale, "metadata_version": 0}).Error
		return false, err
	}
	err = tx.Model(&models.ResourcesDramaSeries{}).Where("id = ? AND video_metadata_classification IN ?", ds.ID, []string{"system", "custom"}).Update("video_metadata_excluded", false).Error
	return err == nil, err
}

func metadataEligibleScope(db *gorm.DB) *gorm.DB {
	reason, err := models.MetadataExclusionSQL(db.Session(&gorm.Session{NewDB: true}))
	if err != nil {
		db.AddError(err)
		return db
	}
	return db.Where("(" + reason + ") = ''")
}

func metadataExcluded(db *gorm.DB, id string) bool {
	reason, err := models.MetadataExclusionSQL(db)
	if err != nil {
		core.LogErr(err)
		return true
	}
	var count int64
	err = db.Table("resourcesDramaSeries ds").Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id").Where("ds.id = ?", id).Where("(" + reason + ") = ''").Count(&count).Error
	if err != nil {
		core.LogErr(err)
		return true
	}
	return count == 0
}

type metadataCleanupRow struct {
	ID     string
	Src    string
	Reason string
}

func metadataCleanupQuery(db *gorm.DB, scope string, ids []string) (*gorm.DB, error) {
	reason, err := models.MetadataExclusionSQL(db)
	if err != nil {
		return nil, err
	}
	q := db.Table("resourcesDramaSeries ds").Joins("JOIN resources r ON r.id = ds.resources_id").Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id").
		Where("r.mode IN ?", []string{"movies", "videoLink"}).Where("(" + reason + ") IN ('system','custom')").
		Where("COALESCE(vm.probe_status,'') <> 'manual'").Where("vm.drama_series_id IS NOT NULL OR COALESCE(ds.durationProbeStatus,'') <> '' OR COALESCE(ds.durationSeconds,0) <> 0").
		Select("ds.id, ds.src, " + reason + " AS reason")
	if scope == models.VideoMetadataScopeSelected {
		q = q.Where("r.filesBases_id IN ?", ids)
	}
	return q, nil
}

// 按200条事务提交，不删除资源或文件；每批重新判断规则和人工覆盖，避免覆盖刚补录的数据。
func cleanMetadataRules(db *gorm.DB, scope string, ids []string, progress func(int) error) (int, error) {
	total := 0
	for {
		if progress != nil {
			if err := progress(total); err != nil {
				return total, err
			}
		}
		n := 0
		err := db.Transaction(func(tx *gorm.DB) error {
			q, err := metadataCleanupQuery(tx, scope, ids)
			if err != nil {
				return err
			}
			var rows []metadataCleanupRow
			if err := q.Order("ds.id").Limit(200).Scan(&rows).Error; err != nil {
				return err
			}
			for _, reason := range []string{"system", "custom"} {
				selected := []string{}
				for _, row := range rows {
					if row.Reason == reason {
						selected = append(selected, row.ID)
					}
				}
				if len(selected) == 0 {
					continue
				}
				if err := tx.Model(&models.ResourcesDramaSeries{}).Where("id IN ?", selected).Updates(map[string]interface{}{
					"video_metadata_excluded": true, "video_metadata_classification": reason, "durationSeconds": 0, "durationProbeStatus": "", "durationProbeTime": nil,
				}).Error; err != nil {
					return err
				}
				if err := (models.ResourcesVideoMetadata{}).DeleteByDramaSeriesIDs(tx, selected); err != nil {
					return err
				}
			}
			n = len(rows)
			return nil
		})
		if err != nil {
			return total, err
		}
		total += n
		if n == 0 {
			return total, nil
		}
	}
}

type MetadataCleanupPreview struct {
	Total       int            `json:"total"`
	ByExtension map[string]int `json:"byExtension"`
	Token       string         `json:"token"`
}
type MetadataCleanupState struct {
	Status    string `json:"status"`
	Processed int    `json:"processed"`
	Total     int    `json:"total"`
	Error     string `json:"error"`
}

var metadataCleanupMu sync.Mutex
var metadataCleanupState = MetadataCleanupState{Status: "idle"}

func metadataRulesToken(db *gorm.DB) (string, error) {
	setting, err := models.MetadataRules(db)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal([]interface{}{setting.AutoExcludeNonVideo, setting.ExcludedExtensions})
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func (VideoMetadata) CleanupPreview() (*MetadataCleanupPreview, error) {
	db := core.DBS()
	token, err := metadataRulesToken(db)
	if err != nil {
		return nil, err
	}
	q, err := metadataCleanupQuery(db, models.VideoMetadataScopeAll, nil)
	if err != nil {
		return nil, err
	}
	result := &MetadataCleanupPreview{ByExtension: map[string]int{}, Token: token}
	// 使用游标分批读取路径，预览不访问磁盘，也不修改记录。
	last := ""
	for {
		var rows []metadataCleanupRow
		if err := q.Session(&gorm.Session{}).Where("ds.id > ?", last).Order("ds.id").Limit(500).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			result.Total++
			result.ByExtension[metadataSourceExtension(row.Src)]++
		}
		if len(rows) < 500 {
			break
		}
		last = rows[len(rows)-1].ID
	}
	return result, nil
}

func (VideoMetadata) CleanupStatus() MetadataCleanupState {
	metadataCleanupMu.Lock()
	defer metadataCleanupMu.Unlock()
	return metadataCleanupState
}
func (VideoMetadata) StartCleanup(token string) error {
	metadataCleanupMu.Lock()
	defer metadataCleanupMu.Unlock()
	if metadataCleanupState.Status == "running" {
		return errors.New("已有整理任务正在执行")
	}
	preview, err := (VideoMetadata{}).CleanupPreview()
	if err != nil {
		return err
	}
	if token == "" || token != preview.Token {
		return errors.New("排除规则已变化，请重新预览")
	}
	metadataCleanupState = MetadataCleanupState{Status: "running", Total: preview.Total}
	go func() {
		n, err := cleanMetadataRules(core.DBS(), models.VideoMetadataScopeAll, nil, func(n int) error {
			metadataCleanupMu.Lock()
			metadataCleanupState.Processed = n
			metadataCleanupMu.Unlock()
			current, err := metadataRulesToken(core.DBS())
			if err != nil {
				return err
			}
			if current != token {
				return errors.New("排除规则已变化，整理已停止，请重新预览")
			}
			return nil
		})
		metadataCleanupMu.Lock()
		defer metadataCleanupMu.Unlock()
		metadataCleanupState.Processed = n
		metadataCleanupState.Status = "completed"
		if err != nil {
			metadataCleanupState.Status = "failed"
			metadataCleanupState.Error = err.Error()
		}
	}()
	return nil
}

type MetadataBulkRequest struct {
	IDs    []string `json:"ids"`
	Action string   `json:"action"`
}
type MetadataBulkResult struct {
	Succeeded int               `json:"succeeded"`
	Failed    map[string]string `json:"failed"`
}

func (VideoMetadata) Bulk(request MetadataBulkRequest) (*MetadataBulkResult, error) {
	ids := uniqueVideoMetadataStrings(request.IDs)
	if len(ids) == 0 || len(ids) > 200 {
		return nil, errors.New("每次请选择1至200条记录")
	}
	if request.Action != "retry" && request.Action != "nonvideo" && request.Action != "video" {
		return nil, errors.New("不支持的批量操作")
	}
	result := &MetadataBulkResult{Failed: map[string]string{}}
	for _, id := range ids {
		var err error
		if request.Action == "retry" {
			err = (VideoMetadata{}).RetryFailure(id)
		} else {
			err = (VideoMetadata{}).SetClassification(VideoMetadataClassificationRequest{DramaSeriesID: id, IsVideo: request.Action == "video"})
		}
		if err != nil {
			result.Failed[id] = err.Error()
		} else {
			result.Succeeded++
		}
	}
	return result, nil
}
