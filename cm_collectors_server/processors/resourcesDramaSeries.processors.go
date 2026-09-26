package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"fmt"

	"gorm.io/gorm"
)

type ResourcesDramaSeries struct{}

func (ResourcesDramaSeries) DataListByResourcesID(resourcesID string) (*[]models.ResourcesDramaSeries, error) {
	return models.ResourcesDramaSeries{}.DataListByResourcesID(core.DBS(), resourcesID)
}

func (ResourcesDramaSeries) DataLisWithResourcetByFilesBasesIds(filesBasesIds []string) (*[]models.DramaSeriesWithResource, error) {
	return models.ResourcesDramaSeries{}.DataLisWithResourcetByFilesBasesIds(core.DBS(), filesBasesIds)
}

func (t ResourcesDramaSeries) FirstInfoByResourcesID(resourcesID string) (*models.ResourcesDramaSeries, error) {
	dramaSeriesSlc, err := t.DataListByResourcesID(resourcesID)
	if err != nil {
		return nil, err
	}
	if len(*dramaSeriesSlc) == 0 {
		return nil, errorMessage.Err_Resources_Play_DramaSeries_Not_Found
	}
	firstDramaSeries := (*dramaSeriesSlc)[0]
	return &firstDramaSeries, nil

}

func (ResourcesDramaSeries) SearchPath(filesBasesIds []string, searchPath string) (*[]models.DramaSeriesWithResource, error) {
	return models.ResourcesDramaSeries{}.SearchPath(core.DBS(), filesBasesIds, searchPath)
}
func (ResourcesDramaSeries) ReplacePath(filesBasesIds []string, searchPath, replacePath string) (*[]models.DramaSeriesWithResource, error) {
	return models.ResourcesDramaSeries{}.ReplacePath(
		core.DBS(), filesBasesIds, searchPath, replacePath, core.GenerateUniqueID,
	)
}

func (ResourcesDramaSeries) Info(id string) (*models.ResourcesDramaSeries, error) {
	return models.ResourcesDramaSeries{}.Info(core.DBS(), id)
}

func (ResourcesDramaSeries) ListBySrc(src string) (*[]models.ResourcesDramaSeries, error) {
	return models.ResourcesDramaSeries{}.ListBySrc(core.DBS(), src)
}

func (t ResourcesDramaSeries) GetSrc(id string) (string, error) {
	info, err := t.Info(id)
	if err == nil && info.Src == "" {
		return info.Src, errorMessage.Err_Resources_Play_DramaSeries_Not_Found
	}
	return info.Src, err
}

// FindDramaSeriesSlcBySearchPath 根据搜索路径查找同一目录下的剧集资源列表
// 该函数首先根据文件库ID和搜索路径查找相关的剧集资源，然后筛选出与搜索路径在同一目录下的资源，
// 并进一步筛选出具有相同资源ID的项目，最终返回符合条件的剧集资源列表
//
// 参数:
//   - filesBasesId: 文件基础ID，用于限定搜索范围
//   - searchPath: 搜索路径，用于查找相关资源
//
// 返回值:
//   - *[]models.DramaSeriesWithResource: 符合条件的剧集资源列表指针
//   - error: 错误信息，如果查找过程中出现错误则返回相应错误
func (ResourcesDramaSeries) FindDramaSeriesSlcBySearchPath(filesBasesId string, searchPath string) (*[]models.DramaSeriesWithResource, error) {
	// 根据搜索路径查找相关的剧集资源
	list, err := models.ResourcesDramaSeries{}.SearchPath(core.DBS(), []string{filesBasesId}, searchPath)
	if err != nil {
		return nil, err
	}
	dataList := []models.DramaSeriesWithResource{}
	resourcesID := ""
	// 遍历查找到的列表，筛选出与搜索路径在同一目录下的项目
	for _, item := range *list {
		if utils.IsSameDirectory(searchPath, item.Src) {
			// 设置资源ID并筛选相同资源ID的项目
			if resourcesID == "" {
				resourcesID = item.ResourcesID
			}
			if resourcesID == item.ResourcesID {
				dataList = append(dataList, item)
			}

		}
	}
	return &dataList, nil
}
func (t ResourcesDramaSeries) SetResourcesDramaSeries(db *gorm.DB, resourceID string, dramaSeriesSlc []datatype.ReqParam_resourceDramaSeries_Base) error {
	return db.Transaction(func(tx *gorm.DB) error {
		rules, err := models.MetadataRules(tx)
		if err != nil {
			return err
		}
		vfM := models.VideoFingerprint{}
		dsM := models.ResourcesDramaSeries{}

		oldDramaSeries, err := dsM.ListByResourceID(tx, resourceID)
		if err != nil {
			return err
		}
		var manualIDs []string
		if err := tx.Table("resources_video_metadata vm").Joins("JOIN resourcesDramaSeries ds ON ds.id = vm.drama_series_id").Where("ds.resources_id = ? AND vm.probe_status = ?", resourceID, models.VideoMetadataStatusManual).Pluck("vm.drama_series_id", &manualIDs).Error; err != nil {
			return err
		}
		manualMetadata := map[string]bool{}
		for _, id := range manualIDs {
			manualMetadata[id] = true
		}
		oldByID := make(map[string]models.ResourcesDramaSeries, len(*oldDramaSeries))
		oldBySrc := make(map[string][]models.ResourcesDramaSeries, len(*oldDramaSeries))
		for _, oldDS := range *oldDramaSeries {
			oldByID[oldDS.ID] = oldDS
			oldBySrc[oldDS.Src] = append(oldBySrc[oldDS.Src], oldDS)
		}

		var res models.Resources
		err = tx.Select("id, filesBases_id, mode").Where("id = ?", resourceID).First(&res).Error
		if err != nil {
			return err
		}

		matched := make(map[string]struct{}, len(dramaSeriesSlc))
		submittedIDs := make(map[string]struct{}, len(dramaSeriesSlc))
		newDramaSeries := make([]models.ResourcesDramaSeries, 0)
		existingUpdates := make([]models.ResourcesDramaSeries, 0, len(dramaSeriesSlc))
		resetFingerprintIDs := make([]string, 0)
		staleMetadataIDs := make([]string, 0)
		excludedMetadataIDs := make([]string, 0)

		for sort, submitted := range dramaSeriesSlc {
			var current models.ResourcesDramaSeries
			found := false
			if submitted.ID != "" {
				if _, duplicated := submittedIDs[submitted.ID]; duplicated {
					return fmt.Errorf("分集 ID 重复：%s", submitted.ID)
				}
				submittedIDs[submitted.ID] = struct{}{}
				var belongs bool
				current, belongs = oldByID[submitted.ID]
				if !belongs {
					return fmt.Errorf("分集 %s 不属于当前资源", submitted.ID)
				}
				found = true
			} else {
				// 兼容旧客户端：没有 ID 时以完全相同的路径复用尚未匹配的旧分集。
				for _, candidate := range oldBySrc[submitted.Src] {
					if _, used := matched[candidate.ID]; !used {
						current = candidate
						found = true
						break
					}
				}
			}

			if !found {
				reason := models.MetadataSourceRule(rules, submitted.Src)
				current = models.ResourcesDramaSeries{
					ID:                          core.GenerateUniqueID(),
					ResourcesID:                 resourceID,
					Src:                         submitted.Src,
					Sort:                        sort,
					VideoMetadataExcluded:       reason != "",
					VideoMetadataClassification: "system",
				}
				newDramaSeries = append(newDramaSeries, current)
				matched[current.ID] = struct{}{}
				continue
			}

			matched[current.ID] = struct{}{}
			pathChanged := current.Src != submitted.Src
			classificationChanged := false
			current.Src = submitted.Src
			current.Sort = sort
			if !manualMetadata[current.ID] && (current.VideoMetadataClassification == "system" || current.VideoMetadataClassification == "custom" || (!current.VideoMetadataExcluded && current.VideoMetadataClassification == "")) {
				nextExcluded := models.MetadataSourceRule(rules, submitted.Src) != ""
				classificationChanged = current.VideoMetadataExcluded != nextExcluded
				current.VideoMetadataExcluded = nextExcluded
				current.VideoMetadataClassification = "system"
			}
			existingUpdates = append(existingUpdates, current)
			if !pathChanged && !classificationChanged {
				continue
			}
			if pathChanged {
				resetFingerprintIDs = append(resetFingerprintIDs, current.ID)
			}
			if current.VideoMetadataExcluded {
				excludedMetadataIDs = append(excludedMetadataIDs, current.ID)
			} else {
				staleMetadataIDs = append(staleMetadataIDs, current.ID)
			}
		}

		// 资源可能包含大量分集，按块批量更新路径和排序，避免逐条执行 SQL。
		const updateBatchSize = 200
		for start := 0; start < len(existingUpdates); start += updateBatchSize {
			end := min(start+updateBatchSize, len(existingUpdates))
			if err := models.BatchUpdate(
				tx,
				models.ResourcesDramaSeries{}.TableName(),
				"id",
				[]string{"src", "sort", "video_metadata_excluded", "video_metadata_classification"},
				existingUpdates[start:end],
				func(item models.ResourcesDramaSeries) map[string]interface{} {
					return map[string]interface{}{
						"id": item.ID, "src": item.Src, "sort": item.Sort,
						"video_metadata_excluded":       item.VideoMetadataExcluded,
						"video_metadata_classification": item.VideoMetadataClassification,
					}
				},
			); err != nil {
				return err
			}
		}
		if len(staleMetadataIDs) > 0 {
			// 路径变化时保留上一次采集值，仅将可重新生成的视频信息标记为失效。
			if err := tx.Model(&models.ResourcesVideoMetadata{}).
				Where("drama_series_id IN ?", staleMetadataIDs).
				Updates(map[string]interface{}{
					"probe_status":     models.VideoMetadataStatusStale,
					"metadata_version": 0,
					"next_retry_time":  nil,
					"retry_count":      0,
					"error_code":       "",
					"error_message":    "",
				}).Error; err != nil {
				return err
			}
		}
		if len(excludedMetadataIDs) > 0 {
			if err := (models.ResourcesVideoMetadata{}).DeleteByDramaSeriesIDs(tx, excludedMetadataIDs); err != nil {
				return err
			}
			if err := tx.Model(&models.ResourcesDramaSeries{}).Where("id IN ?", excludedMetadataIDs).
				Updates(map[string]interface{}{
					"durationSeconds": 0, "durationProbeStatus": "", "durationProbeTime": nil,
				}).Error; err != nil {
				return err
			}
		}

		deletedIDs := make([]string, 0)
		for _, oldDS := range *oldDramaSeries {
			if _, keep := matched[oldDS.ID]; !keep {
				deletedIDs = append(deletedIDs, oldDS.ID)
			}
		}
		if err := vfM.DeleteByDramaSeriesIDs(tx, deletedIDs); err != nil {
			return err
		}
		if err := dsM.DeleteIDS(tx, deletedIDs); err != nil {
			return err
		}
		if len(newDramaSeries) > 0 {
			if err := dsM.Creates(tx, &newDramaSeries); err != nil {
				return err
			}
		}
		if err := vfM.RebuildByDramaSeriesIDs(tx, resetFingerprintIDs, core.GenerateUniqueID); err != nil {
			return err
		}
		return vfM.SyncForResource(
			tx, resourceID, res.FilesBasesID, res.Mode, core.GenerateUniqueID,
		)
	})
}

func (ResourcesDramaSeries) SortByMode(resourceID string, sortMode datatype.SeriesSortMode) error {
	db := core.DBS()
	list, err := models.ResourcesDramaSeries{}.ListByResourceID(db, resourceID)
	if err != nil {
		return err
	}
	sortedList := append([]models.ResourcesDramaSeries(nil), (*list)...)
	filePaths := make([]string, 0, len(sortedList))
	for _, item := range sortedList {
		filePaths = append(filePaths, item.Src)
	}

	var order utils.FilesSortOrder
	switch sortMode {
	case datatype.SeriesSortModeNameAsc:
		order = utils.FileNameAsc
	case datatype.SeriesSortModeNameDesc:
		order = utils.FileNameDesc
	case datatype.SeriesSortModeSizeAsc:
		order = utils.FileSizeAsc
	case datatype.SeriesSortModeSizeDesc:
		order = utils.FileSizeDesc
	default:
		return nil
	}
	filePaths = utils.SortFilesByOrder(filePaths, order)

	itemsBySrc := make(map[string][]models.ResourcesDramaSeries, len(sortedList))
	for _, item := range sortedList {
		itemsBySrc[item.Src] = append(itemsBySrc[item.Src], item)
	}
	orderedList := make([]models.ResourcesDramaSeries, 0, len(sortedList))
	for _, filePath := range filePaths {
		items := itemsBySrc[filePath]
		if len(items) == 0 {
			continue
		}
		orderedList = append(orderedList, items[0])
		itemsBySrc[filePath] = items[1:]
	}
	for i := range orderedList {
		orderedList[i].Sort = i
	}
	return models.BatchUpdate(
		db,
		models.ResourcesDramaSeries{}.TableName(),
		"id", []string{"sort"}, orderedList, func(item models.ResourcesDramaSeries) map[string]interface{} {
			return map[string]interface{}{
				"id":   item.ID,
				"sort": item.Sort,
			}
		},
	)
}

func (ResourcesDramaSeries) Create(tx *gorm.DB, resourceID, src string, sort int) (*models.ResourcesDramaSeries, error) {
	rules, err := models.MetadataRules(tx)
	if err != nil {
		return nil, err
	}
	dramaSeries := models.ResourcesDramaSeries{
		ID: core.GenerateUniqueID(), ResourcesID: resourceID, Src: src, Sort: sort,
		VideoMetadataExcluded:       models.MetadataSourceRule(rules, src) != "",
		VideoMetadataClassification: "system",
	}
	err = models.ResourcesDramaSeries{}.Creates(tx, &[]models.ResourcesDramaSeries{dramaSeries})
	return &dramaSeries, err
}

func (ResourcesDramaSeries) DeleteByResourcesID(tx *gorm.DB, resourceID string) error {
	return models.ResourcesDramaSeries{}.DeleteByResourcesID(tx, resourceID)
}
