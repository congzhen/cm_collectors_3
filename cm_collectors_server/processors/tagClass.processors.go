package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"

	"gorm.io/gorm"
)

type TagClass struct{}

func (TagClass) DataListByFilesBasesId(filesBasesId string) (*[]models.TagClass, error) {
	return models.TagClass{}.DataListByFilesBasesId(core.DBS(), filesBasesId)
}

func (TagClass) InfoByID(id string) (*models.TagClass, error) {
	info, err := models.TagClass{}.InfoByID(core.DBS(), id)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_TagClaSS_Not_Found
	}
	return info, err
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

func (t TagClass) GetFirstTagClassByFilesBasesIDNotFoundCreate(filesBasesID string) (*models.TagClass, error) {
	tagClassList, err := t.DataListByFilesBasesId(filesBasesID)
	if err != nil {
		return nil, err
	}
	if len(*tagClassList) > 0 {
		return &(*tagClassList)[0], nil
	}
	par := datatype.ReqParam_TagClass{
		FilesBasesID: filesBasesID,
		Name:         "Default",
	}
	id, err := t.Create(&par)
	if err != nil {
		return nil, err
	}
	return t.InfoByID(id)
}

func (t TagClass) Create(par *datatype.ReqParam_TagClass) (string, error) {
	return t.Create_DB(core.DBS(), par)
}
func (t TagClass) Create_DB(db *gorm.DB, par *datatype.ReqParam_TagClass) (string, error) {
	tagClassTotal, err := t.GetTotalByFilesBasesId(par.FilesBasesID)
	if err != nil {
		return "", err
	}
	timeNow := datatype.CustomTime(core.TimeNow())
	id := core.GenerateUniqueID()
	tagClassModels := models.TagClass{
		ID:           id,
		FilesBasesID: par.FilesBasesID,
		Name:         par.Name,
		LeftShow:     true,
		Sort:         int(tagClassTotal) + 1,
		CreatedAt:    &timeNow,
		Status:       true,
	}
	return id, tagClassModels.Create(db, &tagClassModels)
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
func (TagClass) UpdateNameByID_DB(db *gorm.DB, id, name string) error {
	return models.TagClass{}.Update(db, &models.TagClass{
		ID:   id,
		Name: name,
	}, []string{"name"})
}

func (TagClass) DeleteByFilesBasesID_DB(db *gorm.DB, filesBasesID string) error {
	return models.TagClass{}.DeleteByFilesBasesID(db, filesBasesID)
}
