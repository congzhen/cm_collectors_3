package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

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
func (t Tag) TagListByTagClassId(tagClassId string) (*[]models.Tag, error) {
	return t.TagListByTagClassIds([]string{tagClassId})
}

func (t Tag) InfoByID(id string) (*models.Tag, error) {
	info, err := models.Tag{}.InfoByID(core.DBS(), id)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_Tag_Not_Found
	}
	return info, err
}

func (t Tag) InfoByName(filesBasesID, name string) (*models.Tag, error) {
	info, err := models.Tag{}.InfoByName(core.DBS(), filesBasesID, name)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_Tag_Not_Found
	}
	return info, err
}

// TagInfoByNameNotFoundCreate 根据名称查找标签，如果未找到则创建新标签
// filesBasesID: 文件基础ID，用于在特定范围内查找标签
// name: 标签名称
// 返回值: 标签信息指针和错误信息
// 如果根据名称未找到标签，则创建一个新标签并返回新创建的标签信息
func (t Tag) TagInfoByNameNotFoundCreate(filesBasesID, name string) (*models.Tag, error) {
	info, err := t.InfoByName(filesBasesID, name)
	if err != nil && err == errorMessage.Err_Tag_Not_Found {
		tagClass, err := TagClass{}.GetFirstTagClassByFilesBasesIDNotFoundCreate(filesBasesID)
		if err != nil {
			return nil, err
		}
		par := datatype.ReqParam_Tag{
			TagClassID: tagClass.ID,
			Name:       name,
		}
		tagId, err := t.Create(&par)
		if err != nil {
			return nil, err
		}
		return t.InfoByID(tagId)
	}
	return info, err
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

func (t Tag) Create(par *datatype.ReqParam_Tag) (string, error) {
	tagTotal, err := t.GetTotalByTagClassID(par.TagClassID)
	if err != nil {
		return "", err
	}
	timeNow := datatype.CustomTime(core.TimeNow())
	id := core.GenerateUniqueID()
	tagModels := models.Tag{
		ID:         id,
		TagClassID: par.TagClassID,
		Name:       par.Name,
		KeyWords:   utils.PinyinInitials(par.Name),
		Sort:       int(tagTotal) + 1,
		CreatedAt:  &timeNow,
		Status:     true,
	}
	return id, tagModels.Create(core.DBS(), &tagModels)
}

func (Tag) Update(tag *datatype.ReqParam_Tag) error {
	db := core.DBS()
	return models.Tag{}.Update(db, &models.Tag{
		ID:         tag.ID,
		Name:       tag.Name,
		TagClassID: tag.TagClassID,
		KeyWords:   utils.PinyinInitials(tag.Name),
		Sort:       tag.Sort,
		Status:     tag.Status,
	}, []string{"name", "keyWords", "tagClass_id", "sort", "status"})
}

func (Tag) DeleteTag(tagID string) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		//删除tag
		err := models.Tag{}.DeleteById(tx, tagID)
		if err != nil {
			return err
		}
		//删除tag关联
		err = ResourcesTags{}.DeleteByTagID(tx, tagID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (Tag) UpdateHot(db *gorm.DB, ids []string) error {
	return models.Tag{}.UpdateHot(db, ids)
}
