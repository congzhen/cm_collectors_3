package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

	"gorm.io/gorm"
)

type ResourcesPerformers struct{}

func (t ResourcesPerformers) SetResourcesPerformers(tx *gorm.DB, resourceID string, performerSlc []string) error {
	list, err := models.ResourcesPerformers{}.ListByResourceID(tx, resourceID)
	if err != nil {
		return err
	}
	existingMap := make(map[string]models.ResourcesPerformers)
	existingIDS := make([]string, len(*list))
	for i, v := range *list {
		existingIDS[i] = v.PerformerID
		existingMap[v.PerformerID] = v
	}
	_, toDelete, toAdd := utils.ArrayIntersectDiff(existingIDS, performerSlc)
	if err := t.handleDeletes(tx, existingMap, toDelete); err != nil {
		return err
	}
	if err := t.handleAdds(tx, toAdd, resourceID); err != nil {
		return err
	}

	return nil
}

func (ResourcesPerformers) handleDeletes(tx *gorm.DB, existingMap map[string]models.ResourcesPerformers, toDelete []string) error {
	if len(toDelete) > 0 {
		ids := make([]string, len(toDelete))
		for i, id := range toDelete {
			ids[i] = existingMap[id].ID
		}
		return models.ResourcesPerformers{}.DeleteIDS(tx, ids)
	}
	return nil
}

func (ResourcesPerformers) handleAdds(tx *gorm.DB, toAdd []string, resourceID string) error {
	if len(toAdd) == 0 {
		return nil
	}
	var newRecords []models.ResourcesPerformers
	for _, id := range toAdd {
		newRecords = append(newRecords, models.ResourcesPerformers{
			ID:          core.GenerateUniqueID(),
			ResourcesID: resourceID,
			PerformerID: id,
		})
	}

	return models.ResourcesPerformers{}.Creates(tx, &newRecords)
}
