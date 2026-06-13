package processors

import (
	"bytes"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorsai "cm_collectors_server/processorsAI"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const aiTagPromptVersion = "v1"

type AiTag struct{}

// AiTagFilesBasesSetting 是前端文件库配置的扁平结构。
// 数据库里 include/exclude 以 JSON 字符串保存，这里转换成数组，避免前端处理序列化细节。
type AiTagFilesBasesSetting struct {
	FilesBasesID       string   `json:"filesBasesId"`
	Enabled            bool     `json:"enabled"`
	IncludeTagClassIDs []string `json:"includeTagClassIds"`
	ExcludeTagClassIDs []string `json:"excludeTagClassIds"`
}

type AiTagStats struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Success    int64 `json:"success"`
	Failed     int64 `json:"failed"`
	Skipped    int64 `json:"skipped"`
}

type AiTagRunResult struct {
	Processed int  `json:"processed"`
	Success   int  `json:"success"`
	Failed    int  `json:"failed"`
	Skipped   int  `json:"skipped"`
	Started   bool `json:"started"`
	Running   bool `json:"running"`
}

type AiTagSettingRecommendation struct {
	Field            string      `json:"field"`
	Label            string      `json:"label"`
	CurrentValue     interface{} `json:"currentValue"`
	RecommendedValue interface{} `json:"recommendedValue"`
	Reason           string      `json:"reason"`
	Impact           string      `json:"impact"`
}

type AiTagModelTestResult struct {
	processorsai.ModelTestReport
	Recommendations []AiTagSettingRecommendation `json:"recommendations"`
}

type AiTagAnalysisRecordView struct {
	models.AiTagAnalysisRecord
	FilesBasesName      string   `json:"filesBasesName"`
	ResourceTitle       string   `json:"resourceTitle"`
	ResourceIssueNumber string   `json:"resourceIssueNumber"`
	ResourceName        string   `json:"resourceName"`
	WrittenTagNames     []string `json:"writtenTagNames"`
	WrittenTagText      string   `json:"writtenTagText"`
}

type aiTagFrameEvidence struct {
	BatchIndex      int     `json:"batchIndex"`
	DramaSeriesID   string  `json:"dramaSeriesId"`
	VideoIndex      int     `json:"videoIndex"`
	PositionPercent float64 `json:"positionPercent"`
	TimeSeconds     float64 `json:"timeSeconds"`
}

type aiTagEvidence struct {
	TotalFrames           int                    `json:"totalFrames"`
	MaxImagesPerAIRequest int                    `json:"maxImagesPerAiRequest"`
	Batches               [][]aiTagFrameEvidence `json:"batches"`
}

type aiTagBatchResult struct {
	BatchIndex int                            `json:"batchIndex"`
	Raw        *processorsai.TagAnalyzeResult `json:"raw"`
	Content    string                         `json:"content"`
	Frames     []aiTagFrameEvidence           `json:"frames"`
}

type aiTagScoredSuggestion struct {
	TagID           string   `json:"tagId"`
	FinalScore      float64  `json:"finalScore"`
	MaxConfidence   float64  `json:"maxConfidence"`
	AvgConfidence   float64  `json:"avgConfidence"`
	EvidenceBatches int      `json:"evidenceBatches"`
	Reasons         []string `json:"reasons"`
}

var aiTagWorkerMu sync.Mutex
var aiTagWorkerRunning bool
var aiTagWorkerFilesBasesID string
var aiTagStopCurrentRun bool
var aiTagRestartAfterStop bool

// Setting 返回脱敏后的全局设置。
// 首次使用时会自动创建默认设置，保证前端设置页无需先跑迁移或初始化接口。
func (AiTag) Setting() (*models.AiTagSetting, error) {
	db := core.DBS()
	info, err := (models.AiTagSetting{}).First(db)
	if err == nil {
		normalizeAiTagSetting(info)
		return maskAiTagSetting(info), nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	def := (models.AiTagSetting{}).Default()
	def.ID = core.GenerateUniqueID()
	if err := (models.AiTagSetting{}).Create(db, &def); err != nil {
		return nil, err
	}
	return maskAiTagSetting(&def), nil
}

// settingRaw 返回包含真实 API Key 的设置，只允许后端内部调用。
// 外部接口必须走 Setting，避免把密钥回传到前端。
func (AiTag) settingRaw() (*models.AiTagSetting, error) {
	db := core.DBS()
	info, err := (models.AiTagSetting{}).First(db)
	if err == gorm.ErrRecordNotFound {
		def := (models.AiTagSetting{}).Default()
		def.ID = core.GenerateUniqueID()
		err = (models.AiTagSetting{}).Create(db, &def)
		info = &def
	}
	if err != nil {
		return nil, err
	}
	normalizeAiTagSetting(info)
	return info, nil
}

// SaveSetting 保存全局配置。
// 前端展示的是脱敏 key，保存时如果传回 ****xxxx 或空字符串，需要保留数据库里的真实 key。
func (AiTag) SaveSetting(par *models.AiTagSetting) (*models.AiTagSetting, error) {
	db := core.DBS()
	current, err := (AiTag{}).settingRaw()
	if err != nil {
		return nil, err
	}
	par.ID = current.ID
	if strings.HasPrefix(par.APIKey, "****") || par.APIKey == "" {
		par.APIKey = current.APIKey
	}
	normalizeAiTagSetting(par)
	if err := aiTagRetrySQLiteBusy(func() error { return (models.AiTagSetting{}).Update(db, par) }); err != nil {
		return nil, err
	}
	return (AiTag{}).Setting()
}

// FilesBasesSettings 合并“所有文件库”和“AI 启用配置”。
// 没有记录的文件库按未启用处理，但仍返回给前端，方便用户直接开启。
func (AiTag) FilesBasesSettings() ([]AiTagFilesBasesSetting, error) {
	db := core.DBS()
	filesBases, err := models.FilesBases{}.DataList(db)
	if err != nil {
		return nil, err
	}
	list, err := (models.AiTagEnabledFilesBases{}).List(db)
	if err != nil {
		return nil, err
	}
	existing := map[string]models.AiTagEnabledFilesBases{}
	for _, item := range *list {
		existing[item.FilesBasesID] = item
	}
	result := make([]AiTagFilesBasesSetting, 0, len(*filesBases))
	for _, fb := range *filesBases {
		item := existing[fb.ID]
		result = append(result, AiTagFilesBasesSetting{
			FilesBasesID:       fb.ID,
			Enabled:            item.Enabled,
			IncludeTagClassIDs: parseJSONStringSlice(item.IncludeTagClassIDs),
			ExcludeTagClassIDs: parseJSONStringSlice(item.ExcludeTagClassIDs),
		})
	}
	return result, nil
}

// SaveFilesBasesSettings 批量保存文件库启用和标签分类筛选。
// 使用事务是为了保证多个文件库配置要么一起保存成功，要么一起失败，避免计划任务读到半套配置。
func (AiTag) SaveFilesBasesSettings(items []AiTagFilesBasesSetting) error {
	db := core.DBS()
	return aiTagRetrySQLiteBusy(func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			for _, item := range items {
				include, _ := json.Marshal(item.IncludeTagClassIDs)
				exclude, _ := json.Marshal(item.ExcludeTagClassIDs)
				record := models.AiTagEnabledFilesBases{
					ID:                 core.GenerateUniqueID(),
					FilesBasesID:       item.FilesBasesID,
					Enabled:            item.Enabled,
					IncludeTagClassIDs: string(include),
					ExcludeTagClassIDs: string(exclude),
				}
				if err := (models.AiTagEnabledFilesBases{}).Upsert(tx, &record); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

// aiTagRetrySQLiteBusy 对 SQLite 短暂锁库做轻量重试。
// AI 自动标签会和前端保存、计划任务同时写库；这里避免偶发 SQLITE_BUSY 导致状态残留。
func aiTagRetrySQLiteBusy(fn func() error) error {
	var err error
	for i := 0; i < 5; i++ {
		err = fn()
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "database is locked") {
			return err
		}
		time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
	}
	return err
}

func (AiTag) Stats(filesBasesID string) (*AiTagStats, error) {
	stats, err := (models.AiTagAnalysisRecord{}).Stats(core.DBS(), filesBasesID)
	if err != nil {
		return nil, err
	}
	return &AiTagStats{
		Pending:    stats[models.AiTagRecordStatusPending],
		Processing: stats[models.AiTagRecordStatusProcessing],
		Success:    stats[models.AiTagRecordStatusSuccess],
		Failed:     stats[models.AiTagRecordStatusFailed],
		Skipped:    stats[models.AiTagRecordStatusSkipped],
	}, nil
}

// Records 返回任务状态列表，并补齐文件库名、资源名和已写入标签名。
// 分析记录里只保存 ID，列表展示时再查询名称，可以避免标签改名后历史记录显示旧名称。
func (AiTag) Records(filesBasesID, status string, page, limit int) (*[]AiTagAnalysisRecordView, int64, error) {
	db := core.DBS()
	list, total, err := (models.AiTagAnalysisRecord{}).List(db, filesBasesID, status, page, limit)
	if err != nil {
		return nil, 0, err
	}
	resourceIDs := make([]string, 0, len(*list))
	filesBasesIDs := make([]string, 0, len(*list))
	tagIDs := []string{}
	for _, item := range *list {
		resourceIDs = append(resourceIDs, item.ResourcesID)
		filesBasesIDs = append(filesBasesIDs, item.FilesBasesID)
		tagIDs = append(tagIDs, parseJSONStringSlice(item.WrittenTagIDs)...)
	}
	resourcesMap := map[string]models.Resources{}
	if len(resourceIDs) > 0 {
		var resources []models.Resources
		if err := db.Select("id", "title", "issueNumber").Where("id in ?", resourceIDs).Find(&resources).Error; err != nil {
			return nil, 0, err
		}
		for _, resource := range resources {
			resourcesMap[resource.ID] = resource
		}
	}
	filesBasesMap := map[string]models.FilesBases{}
	if len(filesBasesIDs) > 0 {
		var filesBases []models.FilesBases
		if err := db.Select("id", "name").Where("id in ?", filesBasesIDs).Find(&filesBases).Error; err != nil {
			return nil, 0, err
		}
		for _, item := range filesBases {
			filesBasesMap[item.ID] = item
		}
	}
	tagMap := map[string]models.Tag{}
	if len(tagIDs) > 0 {
		var tags []models.Tag
		if err := db.Select("id", "name").Where("id in ?", tagIDs).Find(&tags).Error; err != nil {
			return nil, 0, err
		}
		for _, tag := range tags {
			tagMap[tag.ID] = tag
		}
	}
	result := make([]AiTagAnalysisRecordView, 0, len(*list))
	for _, item := range *list {
		resource := resourcesMap[item.ResourcesID]
		resourceName := resource.Title
		if resource.IssueNumber != "" {
			if resourceName != "" {
				resourceName = resource.IssueNumber + " - " + resourceName
			} else {
				resourceName = resource.IssueNumber
			}
		}
		writtenTagNames := []string{}
		for _, tagID := range parseJSONStringSlice(item.WrittenTagIDs) {
			if tag, ok := tagMap[tagID]; ok && tag.Name != "" {
				writtenTagNames = append(writtenTagNames, tag.Name)
			} else {
				writtenTagNames = append(writtenTagNames, tagID)
			}
		}
		result = append(result, AiTagAnalysisRecordView{
			AiTagAnalysisRecord: item,
			FilesBasesName:      filesBasesMap[item.FilesBasesID].Name,
			ResourceTitle:       resource.Title,
			ResourceIssueNumber: resource.IssueNumber,
			ResourceName:        resourceName,
			WrittenTagNames:     writtenTagNames,
			WrittenTagText:      strings.Join(writtenTagNames, ", "),
		})
	}
	return &result, total, nil
}

// ResetFailed 只把 failed 记录放回 pending。
// 运行中的任务不允许重置，避免刚失败/正在写状态的记录被并发改回待处理。
func (AiTag) ResetFailed(filesBasesID string) (int64, error) {
	if isAiTagWorkerRunning() {
		return 0, fmt.Errorf("AI自动标签任务正在执行中，请先暂停并等待当前资源结束")
	}
	var reset int64
	err := aiTagRetrySQLiteBusy(func() error {
		var err error
		reset, err = (models.AiTagAnalysisRecord{}).ResetFailed(core.DBS(), filesBasesID)
		return err
	})
	return reset, err
}

// ResetProcessing 用于服务异常关闭后的人工恢复。
// 正常暂停不会立刻中断当前资源，因此只有确认 worker 停止后才允许把残留 processing 改回 pending。
func (AiTag) ResetProcessing(filesBasesID string) (int64, error) {
	if isAiTagWorkerRunning() {
		return 0, fmt.Errorf("AI自动标签任务正在执行中，请先暂停并等待当前资源结束")
	}
	var reset int64
	err := aiTagRetrySQLiteBusy(func() error {
		var err error
		reset, err = (models.AiTagAnalysisRecord{}).ResetProcessing(core.DBS(), filesBasesID)
		return err
	})
	return reset, err
}

// Pause 的语义是“停止领取下一个资源”，不是强杀当前 AI 请求。
// 当前已进入 analyzeResource 的资源会继续跑完并落到 success/failed，避免重复消耗 AI。
func (AiTag) Pause() (*models.AiTagSetting, error) {
	setting, err := (AiTag{}).settingRaw()
	if err != nil {
		return nil, err
	}
	setAiTagStopCurrentRun(true)
	setAiTagRestartAfterStop(false)
	setting.Paused = true
	if err := aiTagRetrySQLiteBusy(func() error { return (models.AiTagSetting{}).Update(core.DBS(), setting) }); err != nil {
		return nil, err
	}
	return (AiTag{}).Setting()
}

// Resume 恢复暂停状态。
// 如果旧 worker 正在等待当前资源结束，则标记 restartAfterStop，让旧 worker 退出后自动启动新 worker。
func (AiTag) Resume() (*models.AiTagSetting, error) {
	setting, err := (AiTag{}).settingRaw()
	if err != nil {
		return nil, err
	}
	setting.Paused = false
	if err := aiTagRetrySQLiteBusy(func() error { return (models.AiTagSetting{}).Update(core.DBS(), setting) }); err != nil {
		return nil, err
	}
	if isAiTagWorkerRunning() && isAiTagStopCurrentRun() {
		setAiTagRestartAfterStop(true)
	}
	return (AiTag{}).Setting()
}

// Rescan 清空某个文件库的分析记录，让资源重新进入待分析队列。
// 运行中禁止重扫，否则可能把当前资源的 processing 记录删除后再次领取。
func (AiTag) Rescan(filesBasesID string) error {
	if isAiTagWorkerRunning() {
		return fmt.Errorf("AI自动标签任务正在执行中，请先暂停并等待当前资源结束")
	}
	if filesBasesID == "" {
		return fmt.Errorf("filesBasesID不能为空")
	}
	return aiTagRetrySQLiteBusy(func() error {
		return (models.AiTagAnalysisRecord{}).DeleteByFilesBasesID(core.DBS(), filesBasesID)
	})
}

// TestConnection 使用前端当前表单配置执行一次模型输出测试。
// 这个接口不保存配置；如果 API Key 是脱敏值，则沿用数据库里已保存的真实 key。
func (AiTag) TestConnection(par *models.AiTagSetting) (*AiTagModelTestResult, error) {
	current, err := (AiTag{}).settingRaw()
	if err != nil {
		return nil, err
	}
	setting := *current
	if par != nil {
		setting = *par
		if strings.HasPrefix(setting.APIKey, "****") || setting.APIKey == "" {
			setting.APIKey = current.APIKey
		}
	}
	normalizeAiTagSetting(&setting)
	report := processorsai.Client{BaseURL: setting.BaseURL, APIKey: setting.APIKey, Model: setting.Model, RequestTimeoutSeconds: setting.RequestTimeoutSeconds}.TestModelOutput()
	return &AiTagModelTestResult{
		ModelTestReport: *report,
		Recommendations: buildAiTagSettingRecommendations(&setting, report),
	}, nil
}

func (AiTag) TestService() error {
	setting, err := (AiTag{}).settingRaw()
	if err != nil {
		return err
	}
	return processorsai.Client{BaseURL: setting.BaseURL, APIKey: setting.APIKey, Model: setting.Model, RequestTimeoutSeconds: setting.RequestTimeoutSeconds}.TestService()
}

// RunOnce 只负责启动后台 worker，并立即返回启动状态。
// 真正的分析在 goroutine 中持续进行，这样前端关闭或请求超时都不会影响后端任务。
func (AiTag) RunOnce(filesBasesID string) (*AiTagRunResult, error) {
	setting, err := (AiTag{}).settingRaw()
	if err != nil {
		return nil, err
	}
	if !setting.Enabled {
		return &AiTagRunResult{}, nil
	}
	if setting.Paused {
		return &AiTagRunResult{}, nil
	}
	if !startAiTagWorker(filesBasesID) {
		return &AiTagRunResult{Running: true}, nil
	}
	return &AiTagRunResult{Started: true}, nil
}

func (AiTag) RunForCron(filesBasesID string) error {
	_, err := (AiTag{}).RunOnce(filesBasesID)
	return err
}

// startAiTagWorker 保证同一时间只有一个 AI 标签 worker。
// 如果暂停后又继续，restartAfterStop 会在旧 worker 退出的 defer 中触发重启，避免两个 worker 并发抢同一资源。
func startAiTagWorker(filesBasesID string) bool {
	aiTagWorkerMu.Lock()
	defer aiTagWorkerMu.Unlock()
	if aiTagWorkerRunning {
		return false
	}
	aiTagWorkerRunning = true
	aiTagWorkerFilesBasesID = filesBasesID
	aiTagStopCurrentRun = false
	aiTagRestartAfterStop = false
	go func() {
		defer func() {
			restart := false
			restartFilesBasesID := ""
			aiTagWorkerMu.Lock()
			restart = aiTagRestartAfterStop
			restartFilesBasesID = aiTagWorkerFilesBasesID
			aiTagWorkerRunning = false
			aiTagWorkerFilesBasesID = ""
			aiTagStopCurrentRun = false
			aiTagRestartAfterStop = false
			aiTagWorkerMu.Unlock()
			if restart {
				startAiTagWorker(restartFilesBasesID)
			}
		}()
		(AiTag{}).runWorker(filesBasesID)
	}()
	return true
}

// runWorker 是后台主循环：每轮只领取一个资源、分析完成后再读最新设置。
// 这样暂停、禁用、文件库配置变化都能在资源边界生效，当前资源则保持完整落库。
func (AiTag) runWorker(filesBasesID string) {
	for {
		if isAiTagStopCurrentRun() {
			return
		}
		setting, err := (AiTag{}).settingRaw()
		if err != nil || !setting.Enabled || setting.Paused {
			return
		}
		enabled, resource, err := (AiTag{}).nextPendingResource(setting, filesBasesID)
		if err != nil || resource == nil {
			return
		}
		_ = (AiTag{}).analyzeResource(setting, *enabled, resource)
		if isAiTagStopCurrentRun() {
			return
		}
	}
}

func isAiTagWorkerRunning() bool {
	aiTagWorkerMu.Lock()
	defer aiTagWorkerMu.Unlock()
	return aiTagWorkerRunning
}

func setAiTagStopCurrentRun(value bool) {
	aiTagWorkerMu.Lock()
	defer aiTagWorkerMu.Unlock()
	aiTagStopCurrentRun = value
}

func isAiTagStopCurrentRun() bool {
	aiTagWorkerMu.Lock()
	defer aiTagWorkerMu.Unlock()
	return aiTagStopCurrentRun
}

func setAiTagRestartAfterStop(value bool) {
	aiTagWorkerMu.Lock()
	defer aiTagWorkerMu.Unlock()
	aiTagRestartAfterStop = value
}

// nextPendingResource 按启用文件库依次寻找下一个可领取资源。
// filesBasesID 为空表示扫描所有启用文件库；非空则只处理指定库，供计划任务或手动执行使用。
func (AiTag) nextPendingResource(setting *models.AiTagSetting, filesBasesID string) (*models.AiTagEnabledFilesBases, *models.Resources, error) {
	enabledList, err := (models.AiTagEnabledFilesBases{}).ListEnabled(core.DBS())
	if err != nil {
		return nil, nil, err
	}
	for _, enabled := range *enabledList {
		if filesBasesID != "" && enabled.FilesBasesID != filesBasesID {
			continue
		}
		resource, err := (AiTag{}).nextPendingResourceInFilesBases(setting, enabled)
		if err != nil {
			return nil, nil, err
		}
		if resource != nil {
			enabledCopy := enabled
			return &enabledCopy, resource, nil
		}
	}
	return nil, nil, nil
}

// nextPendingResourceInFilesBases 只返回“可以安全进入 processing”的资源。
// success/failed/processing/skipped 都不会被普通扫描再次领取；失败重试必须显式 resetFailed。
func (AiTag) nextPendingResourceInFilesBases(setting *models.AiTagSetting, enabled models.AiTagEnabledFilesBases) (*models.Resources, error) {
	var resources []models.Resources
	q := (models.Resources{}).Preload(core.DBS()).Where("filesBases_id = ? AND status = ? AND mode IN ?", enabled.FilesBasesID, true, []datatype.E_resourceMode{datatype.E_resourceMode_Movies, datatype.E_resourceMode_VideoLink})
	if err := q.Order("addTime desc").Limit(setting.MaxResourcesPerRun * 5).Find(&resources).Error; err != nil {
		return nil, err
	}
	for _, resource := range resources {
		if setting.WriteMode == models.AiTagWriteModeOnlyEmpty && len(resource.Tags) > 0 {
			continue
		}
		record, err := (models.AiTagAnalysisRecord{}).GetByResourceID(core.DBS(), resource.ID)
		claimable, err := isAiTagRecordClaimable(record, err)
		if err != nil {
			return nil, err
		}
		if claimable {
			resourceCopy := resource
			return &resourceCopy, nil
		}
	}
	return nil, nil
}

// analyzeResource 是单个资源的完整状态机：
// 1. 计算资源和标签池 hash；
// 2. 条件更新为 processing，防止 success 被并发改回处理中；
// 3. 抽帧并按批次请求 AI；
// 4. 汇总、校验已有标签并写入 resourcesTags；
// 5. 最终落到 success 或 failed。
func (AiTag) analyzeResource(setting *models.AiTagSetting, enabled models.AiTagEnabledFilesBases, resource *models.Resources) error {
	db := core.DBS()
	srcHash := buildSrcHash(resource.ResourcesDramaSeries)
	tagHash, err := (AiTag{}).tagVersionHash(resource.FilesBasesID, enabled)
	if err != nil {
		return err
	}
	record, err := (models.AiTagAnalysisRecord{}).GetByResourceID(db, resource.ID)
	if err == gorm.ErrRecordNotFound {
		record = &models.AiTagAnalysisRecord{ID: core.GenerateUniqueID(), ResourcesID: resource.ID, FilesBasesID: resource.FilesBasesID}
		if err := aiTagRetrySQLiteBusy(func() error { return (models.AiTagAnalysisRecord{}).Create(db, record) }); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !canAiTagRecordEnterProcessing(record.Status) {
		return nil
	}
	record.Status = models.AiTagRecordStatusProcessing
	record.SrcHash = srcHash
	record.TagVersionHash = tagHash
	record.PromptVersion = aiTagPromptVersion
	record.Model = setting.Model
	record.FailReason = ""
	if err := aiTagRetrySQLiteBusy(func() error {
		// 二次条件更新是防重复分析的最后防线：
		// 即使扫描阶段误拿到了旧对象，数据库也只允许 pending/空状态进入 processing。
		res := db.Model(&models.AiTagAnalysisRecord{}).
			Where("id = ? AND (status = ? OR status = '')", record.ID, models.AiTagRecordStatusPending).
			Updates(map[string]interface{}{
				"status":           record.Status,
				"src_hash":         record.SrcHash,
				"tag_version_hash": record.TagVersionHash,
				"prompt_version":   record.PromptVersion,
				"model":            record.Model,
				"fail_reason":      record.FailReason,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if len(resource.ResourcesDramaSeries) == 0 {
		return (AiTag{}).failRecord(record, "资源没有视频分集")
	}
	tagClasses, validTags, err := (AiTag{}).promptTags(resource.FilesBasesID, enabled)
	if err != nil {
		return (AiTag{}).failRecord(record, err.Error())
	}
	if len(validTags) == 0 {
		return (AiTag{}).failRecord(record, (AiTag{}).emptyTagPoolReason(resource.FilesBasesID, enabled))
	}
	frames, evidence, err := (AiTag{}).extractResourceFrames(setting, resource)
	if err != nil {
		return (AiTag{}).failRecord(record, err.Error())
	}
	if len(frames) == 0 {
		return (AiTag{}).failRecord(record, "未提取到视频截图")
	}

	client := processorsai.Client{BaseURL: setting.BaseURL, APIKey: setting.APIKey, Model: setting.Model, RequestTimeoutSeconds: setting.RequestTimeoutSeconds}
	batches := chunkImages(frames, setting.MaxImagesPerAIRequest)
	evidence.Batches = make([][]aiTagFrameEvidence, len(batches))
	batchResults := []aiTagBatchResult{}
	for i, batch := range batches {
		batchIndex := i + 1
		prompt, _ := processorsai.BuildTagPrompt(buildPromptResource(resource), tagClasses, batchIndex, len(batches))
		images := make([]processorsai.ImageInput, len(batch))
		frameEvidence := make([]aiTagFrameEvidence, len(batch))
		for j, frame := range batch {
			images[j] = processorsai.ImageInput{DataURL: frame.dataURL}
			frame.evidence.BatchIndex = batchIndex
			frameEvidence[j] = frame.evidence
		}
		evidence.Batches[i] = frameEvidence
		result, content, err := client.AnalyzeTags(prompt, images, batchIndex)
		if err != nil {
			return (AiTag{}).failRecord(record, err.Error())
		}
		batchResults = append(batchResults, aiTagBatchResult{BatchIndex: batchIndex, Raw: result, Content: content, Frames: frameEvidence})
	}
	writtenTagIDs, resultJSON, err := (AiTag{}).mergeAndWrite(setting, resource, validTags, batchResults)
	if err != nil {
		return (AiTag{}).failRecord(record, err.Error())
	}
	evidenceJSON, _ := json.Marshal(evidence)
	now := datatype.CustomTime(core.TimeNow())
	record.Status = models.AiTagRecordStatusSuccess
	record.WrittenTagIDs = jsonStringSlice(writtenTagIDs)
	record.RecommendedTagIDs = record.WrittenTagIDs
	record.ResultJSON = resultJSON
	record.EvidenceJSON = string(evidenceJSON)
	record.FailReason = ""
	record.AnalyzedAt = &now
	return aiTagRetrySQLiteBusy(func() error { return (models.AiTagAnalysisRecord{}).Update(db, record) })
}

// isAiTagRecordClaimable 判断记录是否能被 worker 领取。
// 只有“无记录”或“待处理”可以领取；成功记录绝不自动重跑，避免暂停/继续后重复消耗 AI。
func isAiTagRecordClaimable(record *models.AiTagAnalysisRecord, err error) (bool, error) {
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return canAiTagRecordEnterProcessing(record.Status), nil
}

// canAiTagRecordEnterProcessing 是所有领取入口共用的状态白名单。
// 后续如果新增状态，默认不允许进入 processing，必须显式放开。
func canAiTagRecordEnterProcessing(status string) bool {
	return status == "" || status == models.AiTagRecordStatusPending
}

type aiTagFrame struct {
	dataURL  string
	evidence aiTagFrameEvidence
}

// extractResourceFrames 从资源下多个视频中抽取截图。
// 先限制参与视频数，再按视频时长决定抽帧数量；最终仍受每资源最大截图数约束。
func (AiTag) extractResourceFrames(setting *models.AiTagSetting, resource *models.Resources) ([]aiTagFrame, *aiTagEvidence, error) {
	series := resource.ResourcesDramaSeries
	sort.Slice(series, func(i, j int) bool { return series[i].Sort < series[j].Sort })
	if len(series) > setting.MaxVideosPerResource {
		series = selectDramaSeries(series, setting.MaxVideosPerResource)
	}
	frames := []aiTagFrame{}
	for videoIndex, ds := range series {
		if len(frames) >= setting.MaxFramesPerResource {
			break
		}
		duration := float64(ds.DurationSeconds)
		if duration <= 0 {
			formatInfo, err := (processorsffmpeg.VideoInfo{}).GetVideoFormatInfo(ds.Src)
			if err != nil {
				return frames, nil, err
			}
			duration = (processorsffmpeg.VideoInfo{}).GetVideoDuration(formatInfo)
		}
		if duration <= 0 {
			continue
		}
		count := frameCountByDuration(duration)
		if count > setting.MaxFramesPerVideo {
			count = setting.MaxFramesPerVideo
		}
		if remain := setting.MaxFramesPerResource - len(frames); count > remain {
			count = remain
		}
		for _, pos := range framePositions(count) {
			sec := duration * pos
			img, err := (processorsffmpeg.FrameExtract{}).ExtractFrameAt(ds.Src, sec)
			if err != nil {
				return frames, nil, err
			}
			dataURL, err := imageToDataURL(img)
			if err != nil {
				return frames, nil, err
			}
			frames = append(frames, aiTagFrame{
				dataURL: dataURL,
				evidence: aiTagFrameEvidence{
					DramaSeriesID:   ds.ID,
					VideoIndex:      videoIndex + 1,
					PositionPercent: math.Round(pos*10000) / 100,
					TimeSeconds:     math.Round(sec*100) / 100,
				},
			})
		}
	}
	return frames, &aiTagEvidence{TotalFrames: len(frames), MaxImagesPerAIRequest: setting.MaxImagesPerAIRequest}, nil
}

// promptTags 构建喂给 AI 的标签池。
// 只包含启用分类、未被 include/exclude 排除、标签启用且 aiEnabled=true 的已有标签。
// 返回前按 sort/id 稳定排序，确保 tagVersionHash 不会因为 map 遍历顺序变化而误判重扫。
func (AiTag) promptTags(filesBasesID string, enabled models.AiTagEnabledFilesBases) ([]processorsai.PromptTagClass, map[string]models.Tag, error) {
	tagData, err := (Tag{}).TagData(filesBasesID)
	if err != nil {
		return nil, nil, err
	}
	include := stringSet(parseJSONStringSlice(enabled.IncludeTagClassIDs))
	exclude := stringSet(parseJSONStringSlice(enabled.ExcludeTagClassIDs))
	classMap := map[string]models.TagClass{}
	for _, class := range *tagData.TagClass {
		if !class.Status {
			continue
		}
		if len(include) > 0 && !include[class.ID] {
			continue
		}
		if exclude[class.ID] {
			continue
		}
		classMap[class.ID] = class
	}
	grouped := map[string][]models.Tag{}
	valid := map[string]models.Tag{}
	for _, tag := range *tagData.Tag {
		if !tag.Status || !tag.AIEnabled {
			continue
		}
		class, ok := classMap[tag.TagClassID]
		if !ok {
			continue
		}
		grouped[class.ID] = append(grouped[class.ID], tag)
		valid[tag.ID] = tag
	}
	classes := []models.TagClass{}
	for _, class := range classMap {
		classes = append(classes, class)
	}
	sort.Slice(classes, func(i, j int) bool {
		if classes[i].Sort == classes[j].Sort {
			return classes[i].ID < classes[j].ID
		}
		return classes[i].Sort < classes[j].Sort
	})
	result := []processorsai.PromptTagClass{}
	for _, class := range classes {
		tags := grouped[class.ID]
		if len(tags) == 0 {
			continue
		}
		sort.Slice(tags, func(i, j int) bool {
			if tags[i].Sort == tags[j].Sort {
				return tags[i].ID < tags[j].ID
			}
			return tags[i].Sort < tags[j].Sort
		})
		promptTags := make([]processorsai.PromptTag, len(tags))
		for i, tag := range tags {
			promptTags[i] = processorsai.PromptTag{TagID: tag.ID, Name: tag.Name, AIDescription: tag.AIDescription}
		}
		result = append(result, processorsai.PromptTagClass{TagClassID: class.ID, TagClassName: class.Name, Tags: promptTags})
	}
	return result, valid, nil
}

// emptyTagPoolReason 在标签池为空时生成可排查的失败原因。
// 这里拆分统计分类、启用标签、AI 参与标签和筛选后标签，方便用户判断是配置问题还是数据问题。
func (AiTag) emptyTagPoolReason(filesBasesID string, enabled models.AiTagEnabledFilesBases) string {
	tagData, err := (Tag{}).TagData(filesBasesID)
	if err != nil {
		return "AI可用标签为空：读取标签失败：" + err.Error()
	}
	includeIDs := parseJSONStringSlice(enabled.IncludeTagClassIDs)
	excludeIDs := parseJSONStringSlice(enabled.ExcludeTagClassIDs)
	include := stringSet(includeIDs)
	exclude := stringSet(excludeIDs)
	selectedClassIDs := map[string]bool{}
	totalClasses := len(*tagData.TagClass)
	enabledClasses := 0
	selectedClasses := 0
	for _, class := range *tagData.TagClass {
		if !class.Status {
			continue
		}
		enabledClasses++
		if len(include) > 0 && !include[class.ID] {
			continue
		}
		if exclude[class.ID] {
			continue
		}
		selectedClasses++
		selectedClassIDs[class.ID] = true
	}

	totalTags := len(*tagData.Tag)
	enabledTags := 0
	aiEnabledTags := 0
	selectedEnabledTags := 0
	selectedAIEnabledTags := 0
	for _, tag := range *tagData.Tag {
		if !tag.Status {
			continue
		}
		enabledTags++
		if tag.AIEnabled {
			aiEnabledTags++
		}
		if !selectedClassIDs[tag.TagClassID] {
			continue
		}
		selectedEnabledTags++
		if tag.AIEnabled {
			selectedAIEnabledTags++
		}
	}

	parts := []string{
		fmt.Sprintf("AI可用标签为空：分类总数%d，启用分类%d，筛选后分类%d；标签总数%d，启用标签%d，参与AI标签%d，筛选后可参与AI标签%d。",
			totalClasses, enabledClasses, selectedClasses, totalTags, enabledTags, aiEnabledTags, selectedAIEnabledTags),
	}
	if totalClasses == 0 {
		parts = append(parts, "当前文件库没有标签分类。")
	} else if enabledClasses == 0 {
		parts = append(parts, "所有标签分类都未启用。")
	} else if selectedClasses == 0 {
		parts = append(parts, "当前 include/exclude 分类筛选后没有可用分类。")
	}
	if totalTags == 0 {
		parts = append(parts, "当前文件库没有标签。")
	} else if enabledTags == 0 {
		parts = append(parts, "所有标签都未启用。")
	} else if aiEnabledTags == 0 {
		parts = append(parts, "所有标签的“参与 AI 自动标签”都关闭了。")
	} else if selectedEnabledTags == 0 {
		parts = append(parts, "筛选后的分类下没有启用标签。")
	} else if selectedAIEnabledTags == 0 {
		parts = append(parts, "可参与 AI 的标签不在当前筛选分类内，或筛选分类内的标签未参与 AI。")
	}
	if len(includeIDs) > 0 || len(excludeIDs) > 0 {
		parts = append(parts, fmt.Sprintf("当前分类筛选 include=%d，exclude=%d。", len(includeIDs), len(excludeIDs)))
	}
	return strings.Join(parts, " ")
}

// tagVersionHash 标记“当前可参与 AI 的标签池版本”。
// 标签名称、分类、AI 说明、参与开关、include/exclude 变化都会改变 hash，让资源重新分析。
func (AiTag) tagVersionHash(filesBasesID string, enabled models.AiTagEnabledFilesBases) (string, error) {
	tagClasses, _, err := (AiTag{}).promptTags(filesBasesID, enabled)
	if err != nil {
		return "", err
	}
	include := parseJSONStringSlice(enabled.IncludeTagClassIDs)
	exclude := parseJSONStringSlice(enabled.ExcludeTagClassIDs)
	sort.Strings(include)
	sort.Strings(exclude)
	b, _ := json.Marshal(struct {
		TagClasses []processorsai.PromptTagClass `json:"tagClasses"`
		Include    []string                      `json:"include"`
		Exclude    []string                      `json:"exclude"`
	}{tagClasses, include, exclude})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// mergeAndWrite 合并多个 AI 批次结果，并只写入合法已有标签。
// 聚合分数综合最高置信度、平均置信度和出现批次数，避免单批偶然误判直接主导最终结果。
func (AiTag) mergeAndWrite(setting *models.AiTagSetting, resource *models.Resources, validTags map[string]models.Tag, batches []aiTagBatchResult) ([]string, string, error) {
	type agg struct {
		max     float64
		sum     float64
		count   int
		reasons []string
	}
	aggMap := map[string]*agg{}
	for _, batch := range batches {
		if batch.Raw == nil {
			continue
		}
		seenInBatch := map[string]bool{}
		for _, suggestion := range batch.Raw.Tags {
			if _, ok := validTags[suggestion.TagID]; !ok || suggestion.Confidence < setting.MinConfidence {
				continue
			}
			a := aggMap[suggestion.TagID]
			if a == nil {
				a = &agg{}
				aggMap[suggestion.TagID] = a
			}
			if suggestion.Confidence > a.max {
				a.max = suggestion.Confidence
			}
			a.sum += suggestion.Confidence
			if !seenInBatch[suggestion.TagID] {
				a.count++
				seenInBatch[suggestion.TagID] = true
			}
			if suggestion.Reason != "" {
				a.reasons = append(a.reasons, suggestion.Reason)
			}
		}
	}
	scored := []aiTagScoredSuggestion{}
	for tagID, a := range aggMap {
		avg := a.sum / float64(maxInt(a.count, 1))
		ratio := float64(a.count) / float64(maxInt(len(batches), 1))
		score := a.max*0.5 + avg*0.3 + ratio*0.2
		scored = append(scored, aiTagScoredSuggestion{TagID: tagID, FinalScore: score, MaxConfidence: a.max, AvgConfidence: avg, EvidenceBatches: a.count, Reasons: a.reasons})
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].FinalScore > scored[j].FinalScore })
	if len(scored) > setting.MaxTagsPerResource {
		scored = scored[:setting.MaxTagsPerResource]
	}
	tagIDs := make([]string, len(scored))
	for i, item := range scored {
		tagIDs[i] = item.TagID
	}
	if err := (AiTag{}).writeTags(resource.ID, tagIDs, setting.WriteMode); err != nil {
		return nil, "", err
	}
	resultJSON, _ := json.Marshal(scored)
	return tagIDs, string(resultJSON), nil
}

// writeTags 根据写入策略更新资源标签。
// append 模式保留人工标签；replace 模式才会覆盖；only_empty 已在领取阶段跳过已有标签资源。
func (AiTag) writeTags(resourceID string, tagIDs []string, writeMode string) error {
	db := core.DBS()
	err := db.Transaction(func(tx *gorm.DB) error {
		if writeMode == models.AiTagWriteModeReplace {
			return (ResourcesTags{}).SetResourcesTag(tx, resourceID, tagIDs)
		}
		existing, err := (ResourcesTags{}).GetTagIdsByResourceID(tx, resourceID)
		if err != nil {
			return err
		}
		next := append([]string{}, existing...)
		exists := stringSet(existing)
		for _, tagID := range tagIDs {
			if !exists[tagID] {
				next = append(next, tagID)
			}
		}
		return (ResourcesTags{}).SetResourcesTag(tx, resourceID, next)
	})
	if err == nil {
		AutoBackup{}.RecordResourceChanges(1)
	}
	return err
}

// failRecord 统一写 failed 状态。
// 即便返回错误给 worker，也优先保证数据库里能看到失败原因，方便任务状态页排查。
func (AiTag) failRecord(record *models.AiTagAnalysisRecord, reason string) error {
	now := datatype.CustomTime(core.TimeNow())
	record.Status = models.AiTagRecordStatusFailed
	record.FailReason = reason
	record.AnalyzedAt = &now
	_ = aiTagRetrySQLiteBusy(func() error { return (models.AiTagAnalysisRecord{}).Update(core.DBS(), record) })
	return fmt.Errorf("%s", reason)
}

// normalizeAiTagSetting 对空值和异常值做兜底。
// 数据库迁移、旧配置或前端未传字段时，都通过这里回到可运行的默认值。
func normalizeAiTagSetting(s *models.AiTagSetting) {
	def := (models.AiTagSetting{}).Default()
	if s.Provider == "" {
		s.Provider = def.Provider
	}
	if s.MaxResourcesPerRun <= 0 {
		s.MaxResourcesPerRun = def.MaxResourcesPerRun
	}
	if s.RequestTimeoutSeconds <= 0 {
		s.RequestTimeoutSeconds = def.RequestTimeoutSeconds
	}
	if s.MaxFramesPerResource <= 0 {
		s.MaxFramesPerResource = def.MaxFramesPerResource
	}
	if s.MaxFramesPerVideo <= 0 {
		s.MaxFramesPerVideo = def.MaxFramesPerVideo
	}
	if s.MaxVideosPerResource <= 0 {
		s.MaxVideosPerResource = def.MaxVideosPerResource
	}
	if s.MaxImagesPerAIRequest <= 0 {
		s.MaxImagesPerAIRequest = def.MaxImagesPerAIRequest
	}
	if s.FrameStrategy == "" {
		s.FrameStrategy = def.FrameStrategy
	}
	if s.ImageResizeMode == "" {
		s.ImageResizeMode = def.ImageResizeMode
	}
	if s.FallbackImageMaxWidth <= 0 {
		s.FallbackImageMaxWidth = def.FallbackImageMaxWidth
	}
	if s.ImageJPEGQuality <= 0 {
		s.ImageJPEGQuality = def.ImageJPEGQuality
	}
	if s.MinConfidence <= 0 {
		s.MinConfidence = def.MinConfidence
	}
	if s.MaxTagsPerResource <= 0 {
		s.MaxTagsPerResource = def.MaxTagsPerResource
	}
	if s.WriteMode == "" {
		s.WriteMode = def.WriteMode
	}
}

// buildAiTagSettingRecommendations 根据测试模型输出结果给出设置建议。
// 建议只返回给前端应用到表单，不自动保存，避免一次测试直接改变用户配置。
func buildAiTagSettingRecommendations(setting *models.AiTagSetting, report *processorsai.ModelTestReport) []AiTagSettingRecommendation {
	if report == nil {
		return nil
	}
	recommendations := []AiTagSettingRecommendation{}
	add := func(field, label string, currentValue, recommendedValue interface{}, reason, impact string) {
		if fmt.Sprintf("%v", currentValue) == fmt.Sprintf("%v", recommendedValue) {
			return
		}
		recommendations = append(recommendations, AiTagSettingRecommendation{
			Field:            field,
			Label:            label,
			CurrentValue:     currentValue,
			RecommendedValue: recommendedValue,
			Reason:           reason,
			Impact:           impact,
		})
	}

	message := strings.ToLower(report.Error + " " + report.FirstError)
	elapsedSeconds := float64(report.Metrics.ElapsedMs) / 1000
	if strings.Contains(message, "context") || strings.Contains(message, "token") || strings.Contains(message, "exceeds") || strings.Contains(message, "too large") {
		nextImages := maxInt(1, minInt(setting.MaxImagesPerAIRequest-1, 5))
		add("maxImagesPerAiRequest", "每次请求图片数", setting.MaxImagesPerAIRequest, nextImages, "模型提示上下文或 token 超限，单次请求图片数需要降低。", "减少单次请求压力，批次数会增加。")
		if setting.ImageResizeMode == models.AiTagImageResizeOriginal {
			add("imageResizeMode", "图片模式", setting.ImageResizeMode, models.AiTagImageResizeAutoFallback, "模型服务可能无法承受原始截图输入。", "优先保留原图，超限时自动降级压缩。")
		}
	}
	if strings.Contains(message, "timeout") || strings.Contains(message, "deadline") || elapsedSeconds > 60 {
		nextTimeout := maxInt(setting.RequestTimeoutSeconds, 3600)
		if elapsedSeconds > 0 {
			nextTimeout = maxInt(nextTimeout, int(elapsedSeconds*3))
		}
		add("requestTimeoutSeconds", "AI 请求超时", setting.RequestTimeoutSeconds, minInt(nextTimeout, 86400), "本次模型输出耗时较长或发生超时。", "给本地大模型更充足的等待时间。")
		if setting.MaxImagesPerAIRequest > 3 {
			add("maxImagesPerAiRequest", "每次请求图片数", setting.MaxImagesPerAIRequest, 3, "模型响应偏慢，降低每批图片数可以减少单次请求耗时。", "准确度通过更多批次汇总保留，单次请求更稳定。")
		}
	}
	speed := report.Metrics.EstimatedTokensPerSecond
	if speed == 0 {
		speed = report.Metrics.ServiceGeneratedPerSecond
	}
	if speed == 0 {
		speed = report.Metrics.ServiceTokensPerSecond
	}
	if report.Success && speed > 0 && speed < 3 {
		if setting.MaxResourcesPerRun > 5 {
			add("maxResourcesPerRun", "每轮资源数", setting.MaxResourcesPerRun, 5, "模型输出速度较慢，单轮资源数过高会让任务持续很久。", "减少单轮任务耗时，计划任务可分多轮继续处理。")
		}
		if setting.MaxImagesPerAIRequest > 5 {
			add("maxImagesPerAiRequest", "每次请求图片数", setting.MaxImagesPerAIRequest, 5, "模型输出速度较慢，建议控制单批图片数量。", "降低单次请求压力，减少超时概率。")
		}
	}
	if report.FallbackUsed {
		add("maxImagesPerAiRequest", "每次请求图片数", setting.MaxImagesPerAIRequest, minInt(setting.MaxImagesPerAIRequest, 5), "当前服务不完全支持 json_schema，已使用 text 兼容模式成功。", "保持较小批次有助于模型稳定返回 JSON。")
	}
	if len(recommendations) == 0 && report.Success && setting.MaxImagesPerAIRequest > 10 {
		add("maxImagesPerAiRequest", "每次请求图片数", setting.MaxImagesPerAIRequest, 5, "当前设置单批图片数较高，容易增加上下文压力。", "使用默认 5 张更稳，最终结果仍会按批次汇总。")
	}
	return recommendations
}

// maskAiTagSetting 对 API Key 做脱敏。
// 前端需要知道“已保存过 key”，但不能拿到完整密钥。
func maskAiTagSetting(s *models.AiTagSetting) *models.AiTagSetting {
	cp := *s
	if cp.APIKey != "" {
		if len(cp.APIKey) > 4 {
			cp.APIKey = "****" + cp.APIKey[len(cp.APIKey)-4:]
		} else {
			cp.APIKey = "****"
		}
	}
	return &cp
}

// buildSrcHash 标记资源视频源版本。
// 只关心视频路径和排序，资源标题等元数据变化不会导致重复 AI 分析。
func buildSrcHash(series []models.ResourcesDramaSeries) string {
	ordered := append([]models.ResourcesDramaSeries{}, series...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Sort != ordered[j].Sort {
			return ordered[i].Sort < ordered[j].Sort
		}
		if ordered[i].Src != ordered[j].Src {
			return ordered[i].Src < ordered[j].Src
		}
		return ordered[i].ID < ordered[j].ID
	})
	parts := make([]string, len(ordered))
	for i, ds := range ordered {
		parts[i] = fmt.Sprintf("%s|%d", ds.Src, ds.Sort)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

// buildPromptResource 提取 prompt 所需的资源摘要信息。
// Prompt 不直接传完整资源对象，避免把无关字段、内部字段或过长文本喂给模型。
func buildPromptResource(resource *models.Resources) processorsai.PromptResource {
	files := []string{}
	for _, ds := range resource.ResourcesDramaSeries {
		files = append(files, filepath.Base(ds.Src))
	}
	performers := []string{}
	for _, p := range resource.Performers {
		performers = append(performers, p.Name)
	}
	directors := []string{}
	for _, p := range resource.Directors {
		directors = append(directors, p.Name)
	}
	return processorsai.PromptResource{ResourceID: resource.ID, Title: resource.Title, IssueNumber: resource.IssueNumber, Abstract: resource.Abstract, Country: resource.Country, Definition: resource.Definition, Performers: performers, Directors: directors, Files: files}
}

// frameCountByDuration 按视频时长自适应抽帧。
// 长视频内容变化更多，因此给更多截图；最终数量还会受每视频/每资源上限限制。
func frameCountByDuration(duration float64) int {
	minutes := duration / 60
	switch {
	case minutes < 5:
		return 6
	case minutes < 20:
		return 10
	case minutes < 60:
		return 16
	case minutes < 120:
		return 24
	case minutes < 180:
		return 32
	default:
		return 40
	}
}

// framePositions 生成均匀分布的抽帧位置。
// 避开 0% 和 100%，减少黑屏、片头片尾或播放器过渡帧对判断的影响。
func framePositions(count int) []float64 {
	if count <= 1 {
		return []float64{0.5}
	}
	start, end := 0.03, 0.97
	step := (end - start) / float64(count-1)
	positions := make([]float64, count)
	for i := 0; i < count; i++ {
		positions[i] = start + float64(i)*step
	}
	return positions
}

// imageToDataURL 把 FFmpeg 抽出的图片编码为 JPEG data URL。
// 当前默认使用原始分辨率；后续压缩策略可以在这个边界内扩展。
func imageToDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// chunkImages 按“每次请求图片数”把同一资源拆成多批。
// 批次越小，单次上下文压力越低；最终结果仍由 mergeAndWrite 汇总。
func chunkImages(frames []aiTagFrame, size int) [][]aiTagFrame {
	if size <= 0 {
		size = 10
	}
	var chunks [][]aiTagFrame
	for start := 0; start < len(frames); start += size {
		end := start + size
		if end > len(frames) {
			end = len(frames)
		}
		chunks = append(chunks, frames[start:end])
	}
	return chunks
}

// selectDramaSeries 在资源包含大量视频时选择参与分析的视频。
// 当前策略从全资源范围均匀取样，避免只分析开头几个视频导致长资源代表性不足。
func selectDramaSeries(series []models.ResourcesDramaSeries, limit int) []models.ResourcesDramaSeries {
	if len(series) <= limit {
		return series
	}
	selected := []models.ResourcesDramaSeries{series[0], series[len(series)-1]}
	mid := series[len(series)/2]
	selected = append(selected, mid)
	for _, ds := range series {
		if len(selected) >= limit {
			break
		}
		exists := false
		for _, s := range selected {
			if s.ID == ds.ID {
				exists = true
				break
			}
		}
		if !exists {
			selected = append(selected, ds)
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Sort != selected[j].Sort {
			return selected[i].Sort < selected[j].Sort
		}
		if selected[i].Src != selected[j].Src {
			return selected[i].Src < selected[j].Src
		}
		return selected[i].ID < selected[j].ID
	})
	return selected
}

func parseJSONStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func jsonStringSlice(values []string) string {
	b, _ := json.Marshal(values)
	return string(b)
}

func stringSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, v := range values {
		set[v] = true
	}
	return set
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
