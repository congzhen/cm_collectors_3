package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
)

type Performer struct{}

func (Performer) BasicList_Performer(performerBasesIds []string) (*[]models.PerformerBasic, error) {
	return models.Performer{}.BasicList_Performer(core.DBS(), performerBasesIds)
}

func (Performer) DataList(performerBasesId string, fetchCount bool, page, limit int, search, star, cup string) (*[]models.Performer, int64, error) {
	return models.Performer{}.DataList(core.DBS(), performerBasesId, fetchCount, page, limit, search, star, cup)
}

func (Performer) ListTopPreferredPerformers(preferredIds []string, mainPerformerBasesId string, shieldNoPerformerPhoto bool, limit int) (*[]models.Performer, error) {
	return models.Performer{}.ListTopPreferredPerformers(core.DBS(), preferredIds, mainPerformerBasesId, shieldNoPerformerPhoto, limit)
}
