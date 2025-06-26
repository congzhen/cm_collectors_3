package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
)

type Resources struct{}

func (Resources) DataList(par *datatype.ReqParam_ResourcesList) (*[]models.Resources, int64, error) {
	return models.Resources{}.DataList(core.DBS(), par)
}
