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
		vfM := models.VideoFingerprint{}
		dsM := models.ResourcesDramaSeries{}

		oldDramaSeries, err := dsM.ListByResourceID(tx, resourceID)
		if err != nil {
			return err
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
				current = models.ResourcesDramaSeries{
					ID:          core.GenerateUniqueID(),
					ResourcesID: resourceID,
					Src:         submitted.Src,
					Sort:        sort,
				}
				newDramaSeries = append(newDramaSeries, current)
				matched[current.ID] = struct{}{}
				continue
			}

			matched[current.ID] = struct{}{}
			pathChanged := current.Src != submitted.Src
			current.Src = submitted.Src
			current.Sort = sort
			existingUpdates = append(existingUpdates, current)
			if !pathChanged {
				continue
			}
			resetFingerprintIDs = append(resetFingerprintIDs, current.ID)
		}

		// 资源可能包含大量分集，按块批量更新路径和排序，避免逐条执行 SQL。
		const updateBatchSize = 200
		for start := 0; start < len(existingUpdates); start += updateBatchSize {
			end := min(start+updateBatchSize, len(existingUpdates))
			if err := models.BatchUpdate(
				tx,
				models.ResourcesDramaSeries{}.TableName(),
				"id",
				[]string{"src", "sort"},
				existingUpdates[start:end],
				func(item models.ResourcesDramaSeries) map[string]interface{} {
					return map[string]interface{}{"id": item.ID, "src": item.Src, "sort": item.Sort}
				},
			); err != nil {
				return err
			}
		}
		if len(resetFingerprintIDs) > 0 {
			// 路径变化时保留上一次采集值，仅将可重新生成的视频信息标记为失效。
			if err := tx.Model(&models.ResourcesVideoMetadata{}).
				Where("drama_series_id IN ?", resetFingerprintIDs).
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

func (ResourcesDramaSeries) SortBySrc(resourceID string) error {
	db := core.DBS()
	list, err := models.ResourcesDramaSeries{}.ListByResourceID(db, resourceID)
	if err != nil {
		return err
	}
	// 根据src字段进行正序排序
	sortedList := *list
	for i := 0; i < len(sortedList)-1; i++ {
		for j := i + 1; j < len(sortedList); j++ {
			if sortedList[i].Src > sortedList[j].Src {
				sortedList[i], sortedList[j] = sortedList[j], sortedList[i]
			}
		}
	}
	for i := range sortedList {
		sortedList[i].Sort = i
	}
	return models.BatchUpdate(
		db,
		models.ResourcesDramaSeries{}.TableName(),
		"id", []string{"sort"}, sortedList, func(item models.ResourcesDramaSeries) map[string]interface{} {
			return map[string]interface{}{
				"id":   item.ID,
				"sort": item.Sort,
			}
		},
	)
}

func (ResourcesDramaSeries) Create(tx *gorm.DB, resourceID, src string, sort int) error {
	return models.ResourcesDramaSeries{}.Creates(tx, &[]models.ResourcesDramaSeries{
		{ID: core.GenerateUniqueID(), ResourcesID: resourceID, Src: src, Sort: sort},
	})
}

func (ResourcesDramaSeries) DeleteByResourcesID(tx *gorm.DB, resourceID string) error {
	return models.ResourcesDramaSeries{}.DeleteByResourcesID(tx, resourceID)
}
