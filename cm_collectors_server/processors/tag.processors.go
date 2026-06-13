package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"encoding/json"
	"strings"

	"gorm.io/gorm"
)

type Tag struct{}

type TagData struct {
	TagClass *[]models.TagClass `json:"tagClass"`
	Tag      *[]models.Tag      `json:"tag"`
}

type importTagItem struct {
	Name             string  `json:"name"`
	AIDescription    *string `json:"aiDescription,omitempty"`
	AIEnabled        *bool   `json:"aiEnabled,omitempty"`
	HasAIDescription bool    `json:"-"`
}

func (i *importTagItem) UnmarshalJSON(data []byte) error {
	type alias importTagItem
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*i = importTagItem(raw)
	var fieldMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &fieldMap); err == nil {
		_, i.HasAIDescription = fieldMap["aiDescription"]
	}
	return nil
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

func (t Tag) ImportTag(filesBasesID string, importDataRaw json.RawMessage) error {
	importData, err := t.normalizeImportTagData(importDataRaw)
	if err != nil {
		return err
	}
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		tagData, err := t.TagData(filesBasesID)
		if err != nil {
			return err
		}

		tagClassMap := make(map[string]*models.TagClass)
		tagMap := make(map[string]*models.Tag)

		for _, tagClass := range *tagData.TagClass {
			tagClassMap[tagClass.Name] = &tagClass
		}

		for _, tag := range *tagData.Tag {
			tagMap[tag.Name] = &tag
		}

		for tagClassName, tagSlc := range importData {
			oldNameTagClass, newNameTagClass := t.ParseTagName(tagClassName)
			var tagClassID string
			if existingTagClass, exists := tagClassMap[oldNameTagClass]; exists {
				tagClassID = existingTagClass.ID
				if oldNameTagClass != newNameTagClass {
					if err := (TagClass{}).UpdateNameByID_DB(tx, existingTagClass.ID, newNameTagClass); err != nil {
						return err
					}
				}
			} else {
				tagClassID, err = TagClass{}.Create_DB(tx, &datatype.ReqParam_TagClass{
					FilesBasesID: filesBasesID,
					Name:         newNameTagClass,
				})
				if err != nil {
					return err
				}
			}

			for _, tagItem := range tagSlc {
				oldNameTag, newNameTag := t.ParseTagName(tagItem.Name)
				if existingTag, exists := tagMap[oldNameTag]; exists {
					fields := []string{}
					tagUpdate := models.Tag{ID: existingTag.ID}
					if oldNameTag != newNameTag || existingTag.TagClassID != tagClassID {
						tagUpdate.Name = newNameTag
						tagUpdate.TagClassID = tagClassID
						tagUpdate.KeyWords = utils.PinyinInitials(newNameTag)
						fields = append(fields, "name", "keyWords", "tagClass_id")
					}
					if tagItem.HasAIDescription {
						aiDescription := ""
						if tagItem.AIDescription != nil {
							aiDescription = *tagItem.AIDescription
						}
						tagUpdate.AIDescription = strings.TrimSpace(aiDescription)
						fields = append(fields, "aiDescription")
					}
					if tagItem.AIEnabled != nil {
						tagUpdate.AIEnabled = *tagItem.AIEnabled
						fields = append(fields, "aiEnabled")
					}
					if len(fields) > 0 {
						if err := (models.Tag{}).Update(tx, &tagUpdate, fields); err != nil {
							return err
						}
					}
				} else {
					par := datatype.ReqParam_Tag{
						TagClassID: tagClassID,
						Name:       newNameTag,
						AIEnabled:  tagItem.AIEnabled,
					}
					if tagItem.HasAIDescription && tagItem.AIDescription != nil {
						par.AIDescription = *tagItem.AIDescription
					}
					if _, err = t.Create_DB(tx, &par); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func (t Tag) normalizeImportTagData(importDataRaw json.RawMessage) (map[string][]importTagItem, error) {
	result := map[string][]importTagItem{}
	if len(importDataRaw) == 0 || string(importDataRaw) == "null" {
		return result, nil
	}

	var newData map[string][]importTagItem
	if err := json.Unmarshal(importDataRaw, &newData); err == nil {
		for tagClassName, tags := range newData {
			tagClassName = strings.TrimSpace(tagClassName)
			if tagClassName == "" {
				continue
			}
			for _, tag := range tags {
				tag.Name = strings.TrimSpace(tag.Name)
				if tag.Name == "" {
					continue
				}
				result[tagClassName] = append(result[tagClassName], tag)
			}
		}
		return result, nil
	}

	var oldData map[string][]string
	if err := json.Unmarshal(importDataRaw, &oldData); err != nil {
		return nil, err
	}
	for tagClassName, tags := range oldData {
		tagClassName = strings.TrimSpace(tagClassName)
		if tagClassName == "" {
			continue
		}
		for _, tagName := range tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			result[tagClassName] = append(result[tagClassName], importTagItem{Name: tagName})
		}
	}
	return result, nil
}

// ParseTagName 解析标签名称，如果含有 => 则分解为旧名称和新名称，否则两个名称相同
func (t Tag) ParseTagName(tagName string) (oldName, newName string) {
	index := strings.Index(tagName, "=>")
	if index != -1 {
		// 如果包含 =>，则分割为旧名称和新名称
		oldName = strings.TrimSpace(tagName[:index])
		newName = strings.TrimSpace(tagName[index+2:]) // +2 是为了跳过 "=>" 两个字符
	} else {
		// 如果不包含 =>，则新旧名称相同
		oldName = strings.TrimSpace(tagName)
		newName = oldName
	}
	return oldName, newName
}

func (t Tag) Create(par *datatype.ReqParam_Tag) (string, error) {
	return t.Create_DB(core.DBS(), par)
}
func (t Tag) Create_DB(db *gorm.DB, par *datatype.ReqParam_Tag) (string, error) {
	tagTotal, err := t.GetTotalByTagClassID(par.TagClassID)
	if err != nil {
		return "", err
	}
	timeNow := datatype.CustomTime(core.TimeNow())
	id := core.GenerateUniqueID()
	name := strings.TrimSpace(par.Name)
	aiEnabled := true
	if par.AIEnabled != nil {
		aiEnabled = *par.AIEnabled
	}
	tagModels := models.Tag{
		ID:            id,
		TagClassID:    par.TagClassID,
		Name:          name,
		KeyWords:      utils.PinyinInitials(name),
		AIDescription: strings.TrimSpace(par.AIDescription),
		AIEnabled:     aiEnabled,
		Sort:          int(tagTotal) + 1,
		CreatedAt:     &timeNow,
		Status:        true,
	}
	return id, tagModels.Create(db, &tagModels)
}

func (Tag) Update(tag *datatype.ReqParam_Tag) error {
	db := core.DBS()
	name := strings.TrimSpace(tag.Name)
	aiEnabled := true
	if tag.AIEnabled != nil {
		aiEnabled = *tag.AIEnabled
	}
	return models.Tag{}.Update(db, &models.Tag{
		ID:            tag.ID,
		Name:          name,
		TagClassID:    tag.TagClassID,
		KeyWords:      utils.PinyinInitials(name),
		AIDescription: strings.TrimSpace(tag.AIDescription),
		AIEnabled:     aiEnabled,
		Sort:          tag.Sort,
		Status:        tag.Status,
	}, []string{"name", "keyWords", "tagClass_id", "aiDescription", "aiEnabled", "sort", "status"})
}
func (Tag) UpdateNameAndTagClassIDByID_DB(db *gorm.DB, id, name, tagClassID string) error {
	name = strings.TrimSpace(name)
	return models.Tag{}.Update(db, &models.Tag{
		ID:         id,
		Name:       name,
		TagClassID: tagClassID,
		KeyWords:   utils.PinyinInitials(name),
	}, []string{"name", "keyWords", "tagClass_id"})
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

func (Tag) DeleteTagByTagClassSlc(db *gorm.DB, tagClassSlc []string) error {
	return models.Tag{}.DeleteTagByTagClassSlc(db, tagClassSlc)
}

func (Tag) UpdateHot(db *gorm.DB, ids []string) error {
	return models.Tag{}.UpdateHot(db, ids)
}
