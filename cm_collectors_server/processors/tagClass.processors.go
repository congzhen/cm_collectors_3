package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"

	"gorm.io/gorm"
)

type TagClass struct{}

func (TagClass) DataListByFilesBasesId(filesBasesId string) (*[]models.TagClass, error) {
	return models.TagClass{}.DataListByFilesBasesId(core.DBS(), filesBasesId)
}

func (TagClass) TagClassUpdateSort(db *gorm.DB, tagClassSort *[]datatype.TagSort) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, v := range *tagClassSort {
			tagClassModels := models.TagClass{
				ID:   v.ID,
				Sort: v.Sort,
			}
			err := tagClassModels.Update(db, &tagClassModels, []string{"sort"})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (TagClass) GetTotalByFilesBasesId(filesBasesId string) (int64, error) {
	return models.TagClass{}.GetTotalByFilesBasesId(core.DBS(), filesBasesId)
}

func (t TagClass) Create(par *datatype.ReqParam_TagClass) error {
	tagClassTotal, err := t.GetTotalByFilesBasesId(par.FilesBasesID)
	if err != nil {
		return err
	}
	timeNow := datatype.CustomTime(core.TimeNow())
	tagClassModels := models.TagClass{
		ID:           core.GenerateUniqueID(),
		FilesBasesID: par.FilesBasesID,
		Name:         par.Name,
		LeftShow:     true,
		Sort:         int(tagClassTotal) + 1,
		CreatedAt:    &timeNow,
		Status:       true,
	}
	return tagClassModels.Create(core.DBS(), &tagClassModels)
}

func (TagClass) Update(tagClass *datatype.ReqParam_TagClass) error {
	db := core.DBS()
	return models.TagClass{}.Update(db, &models.TagClass{
		ID:       tagClass.ID,
		Name:     tagClass.Name,
		LeftShow: tagClass.LeftShow,
		Sort:     tagClass.Sort,
		Status:   tagClass.Status,
	}, []string{"name", "leftShow", "sort", "status"})
}
