package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"

	"gorm.io/gorm"
)

type Tag struct{}

type TagData struct {
	TagClass *[]models.TagClass `json:"tagClass"`
	Tag      *[]models.Tag      `json:"tag"`
}

func (t Tag) TagData(filesBasesID string) (*TagData, error) {
	tagClass, err := TagClass{}.DataListByFilesBasesId(filesBasesID)
	if err != nil {
		return nil, err
	}
	tagClassIds := make([]string, len(*tagClass))
	for i, v := range *tagClass {
		tagClassIds[i] = v.ID
	}
	tags, err := t.TagListByTagClassIds(tagClassIds)
	if err != nil {
		return nil, err
	}
	return &TagData{
		TagClass: tagClass,
		Tag:      tags,
	}, nil
}

func (t Tag) TagDataUpdateSort(par *datatype.ReqParam_UpdateTagDataSort) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		err := t.TagUpdateSort(tx, &par.TagSort)
		if err != nil {
			return err
		}
		return TagClass{}.TagClassUpdateSort(tx, &par.TagClassSort)
	})
}

func (t Tag) TagUpdateSort(db *gorm.DB, tagSort *[]datatype.TagSort) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, v := range *tagSort {
			tagModels := models.Tag{
				ID:   v.ID,
				Sort: v.Sort,
			}
			err := tagModels.Update(db, &tagModels, []string{"sort"})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (Tag) TagListByTagClassIds(tagClassIds []string) (*[]models.Tag, error) {
	return models.Tag{}.DataListByTagClassIds(core.DBS(), tagClassIds)
}

func (Tag) GetTotalByTagClassID(tagClassID string) (int64, error) {
	return models.Tag{}.GetTotalByTagClassID(core.DBS(), tagClassID)
}

func (t Tag) Create(par *datatype.ReqParam_Tag) error {
	tagTotal, err := t.GetTotalByTagClassID(par.TagClassID)
	if err != nil {
		return err
	}
	timeNow := datatype.CustomTime(core.TimeNow())
	tagModels := models.Tag{
		ID:         core.GenerateUniqueID(),
		TagClassID: par.TagClassID,
		Name:       par.Name,
		Sort:       int(tagTotal) + 1,
		CreatedAt:  &timeNow,
		Status:     true,
	}
	return tagModels.Create(core.DBS(), &tagModels)
}

func (Tag) Update(tag *datatype.ReqParam_Tag) error {
	db := core.DBS()
	return models.Tag{}.Update(db, &models.Tag{
		ID:         tag.ID,
		Name:       tag.Name,
		TagClassID: tag.TagClassID,
		Sort:       tag.Sort,
		Status:     tag.Status,
	}, []string{"name", "tagClass_id", "sort", "status"})
}
