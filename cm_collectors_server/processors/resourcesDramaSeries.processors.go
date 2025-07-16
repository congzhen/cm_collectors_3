package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"

	"gorm.io/gorm"
)

type ResourcesDramaSeries struct{}

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
