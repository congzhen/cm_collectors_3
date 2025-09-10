package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

	"gorm.io/gorm"
)

type ResourcesDramaSeries struct{}

func (ResourcesDramaSeries) SearchPath(filesBasesIds []string, searchPath string) (*[]models.DramaSeriesWithResource, error) {
	return models.ResourcesDramaSeries{}.SearchPath(core.DBS(), filesBasesIds, searchPath)
}
func (ResourcesDramaSeries) ReplacePath(filesBasesIds []string, searchPath, replacePath string) (*[]models.DramaSeriesWithResource, error) {
	return models.ResourcesDramaSeries{}.ReplacePath(core.DBS(), filesBasesIds, searchPath, replacePath)
}

func (ResourcesDramaSeries) Info(id string) (*models.ResourcesDramaSeries, error) {
	return models.ResourcesDramaSeries{}.Info(core.DBS(), id)
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
	if len(dramaSeriesSlc) == 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		err := models.ResourcesDramaSeries{}.DeleteByResourcesID(tx, resourceID)
		if err != nil {
			return err
		}
		resourcesDramaSeriesModelsSlc := []models.ResourcesDramaSeries{}
		for i, v := range dramaSeriesSlc {
			resourcesDramaSeriesModelsSlc = append(resourcesDramaSeriesModelsSlc, models.ResourcesDramaSeries{
				ID:          core.GenerateUniqueID(),
				ResourcesID: resourceID,
				Src:         v.Src,
				Sort:        i,
			})
		}
		return models.ResourcesDramaSeries{}.Creates(tx, &resourcesDramaSeriesModelsSlc)
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
