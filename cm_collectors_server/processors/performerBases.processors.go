package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
)

type PerformerBases struct{}

func (PerformerBases) DataList() (*[]models.PerformerBases, error) {
	return models.PerformerBases{}.DataList(core.DBS())
}

func (PerformerBases) Update(par *datatype.ReqParam_UpdatePerformerBases) error {
	performerBasesModels := models.PerformerBases{
		ID:     par.ID,
		Name:   par.Name,
		Sort:   par.Sort,
		Status: par.Status,
	}
	return performerBasesModels.Update(core.DBS(), &performerBasesModels, []string{"name", "sort", "status"})
}
