package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"fmt"

	"gorm.io/gorm"
)

type TvboxRecommend struct{}

func (TvboxRecommend) List() (*[]models.TvboxRecommend, error) {
	return models.TvboxRecommend{}.List(core.DBS())
}

func (TvboxRecommend) Add(resourceId string) error {
	db := core.DBS()
	exists, err := models.TvboxRecommend{}.ExistsByResourceID(db, resourceId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("该资源已在TVBox推荐列表中")
	}
	maxSort, err := models.TvboxRecommend{}.MaxSort(db)
	if err != nil {
		return err
	}
	record := &models.TvboxRecommend{
		ID:         core.GenerateUniqueID(),
		ResourceID: resourceId,
		Sort:       maxSort + 1,
	}
	return models.TvboxRecommend{}.Create(db, record)
}

func (TvboxRecommend) Delete(id string) error {
	return models.TvboxRecommend{}.DeleteByID(core.DBS(), id)
}

func (TvboxRecommend) UpdateSort(items []datatype.ReqParam_TvboxRecommendSort) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		m := models.TvboxRecommend{}
		for _, item := range items {
			if err := m.UpdateSort(tx, item.ID, item.Sort); err != nil {
				return err
			}
		}
		return nil
	})
}
