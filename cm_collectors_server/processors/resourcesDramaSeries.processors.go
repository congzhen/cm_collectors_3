package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"

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

func (ResourcesDramaSeries) DeleteByResourcesID(tx *gorm.DB, resourceID string) error {
	return models.ResourcesDramaSeries{}.DeleteByResourcesID(tx, resourceID)
}
