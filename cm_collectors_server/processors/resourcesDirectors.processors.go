package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

	"gorm.io/gorm"
)

type ResourcesDirectors struct{}

func (t ResourcesDirectors) SetResourcesDirectors(tx *gorm.DB, resourceID string, directorSlc []string) error {
	list, err := models.ResourcesDirectors{}.ListByResourceID(tx, resourceID)
	if err != nil {
		return err
	}
	existingMap := make(map[string]models.ResourcesDirectors)
	existingIDS := make([]string, len(*list))
	for i, v := range *list {
		existingIDS[i] = v.PerformerID
		existingMap[v.PerformerID] = v
	}
	_, toDelete, toAdd := utils.ArrayIntersectDiff(existingIDS, directorSlc)
	if err := t.handleDeletes(tx, existingMap, toDelete); err != nil {
		return err
	}
	if err := t.handleAdds(tx, toAdd, resourceID); err != nil {
		return err
	}

	return nil
}

func (ResourcesDirectors) handleDeletes(tx *gorm.DB, existingMap map[string]models.ResourcesDirectors, toDelete []string) error {
	if len(toDelete) > 0 {
		ids := make([]string, len(toDelete))
		for i, id := range toDelete {
			ids[i] = existingMap[id].ID
		}
		return models.ResourcesDirectors{}.DeleteIDS(tx, ids)
	}
	return nil
}

func (ResourcesDirectors) handleAdds(tx *gorm.DB, toAdd []string, resourceID string) error {
	if len(toAdd) == 0 {
		return nil
	}
	var newRecords []models.ResourcesDirectors
	for _, id := range toAdd {
		newRecords = append(newRecords, models.ResourcesDirectors{
			ID:          core.GenerateUniqueID(),
			ResourcesID: resourceID,
			PerformerID: id,
		})
	}

	return models.ResourcesDirectors{}.Creates(tx, &newRecords)
}

func (ResourcesDirectors) DeleteByResourcesID(tx *gorm.DB, resourceID string) error {
	return models.ResourcesDirectors{}.DeleteByResourcesID(tx, resourceID)
}
