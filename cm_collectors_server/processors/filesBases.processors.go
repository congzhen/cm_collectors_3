package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
)

type FilesBases struct{}

func (FilesBases) DataList() (*[]models.FilesBases, error) {
	return models.FilesBases{}.DataList(core.DBS())
}

func (FilesBases) InfoDetailsById(id string) (*models.FilesBasesDetails, error) {
	return models.FilesBases{}.InfoDetails(core.DBS(), id)
}
