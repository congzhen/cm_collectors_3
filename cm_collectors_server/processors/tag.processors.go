package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
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

func (Tag) TagListByTagClassIds(tagClassIds []string) (*[]models.Tag, error) {
	return models.Tag{}.DataListByTagClassIds(core.DBS(), tagClassIds)
}
