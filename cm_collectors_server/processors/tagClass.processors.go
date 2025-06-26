package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
)

type TagClass struct{}

func (TagClass) DataListByFilesBasesId(filesBasesId string) (*[]models.TagClass, error) {
	return models.TagClass{}.DataListByFilesBasesId(core.DBS(), filesBasesId)
}
