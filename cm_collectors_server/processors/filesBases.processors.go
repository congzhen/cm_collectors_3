package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
)

type FilesBases struct {
	models.FilesBases
}

func (FilesBases) DataList() (*[]FilesBases, error) {
	dataList, err := models.FilesBases{}.DataList(core.DBS())
	if err != nil {
		return nil, err
	}
	list := make([]FilesBases, len(*dataList))
	for i, v := range *dataList {
		list[i].FilesBases = v
	}
	return &list, nil
}
