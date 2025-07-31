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

func (PerformerBases) InfoById(id string) (*models.PerformerBases, error) {
	return models.PerformerBases{}.InfoById(core.DBS(), id)
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

func (PerformerBases) GetTotal() (int64, error) {
	return models.PerformerBases{}.GetTotal(core.DBS())
}

func (t PerformerBases) Create(name string) (string, error) {
	id := core.GenerateUniqueID()
	db := core.DBS()
	createdAt := datatype.CustomTime(core.TimeNow())
	tagTotal, err := t.GetTotal()
	if err != nil {
		return id, err
	}
	performerBasesModels := models.PerformerBases{
		ID:        id,
		Name:      name,
		Sort:      int(tagTotal) + 1,
		CreatedAt: &createdAt,
		Status:    true,
	}
	return id, performerBasesModels.Create(db, &performerBasesModels)
}
