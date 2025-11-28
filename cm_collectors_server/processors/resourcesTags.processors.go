package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

	"gorm.io/gorm"
)

type ResourcesTags struct{}

func (ResourcesTags) ListByResourceID(tx *gorm.DB, resourceID string) (*[]models.ResourcesTags, error) {
	return models.ResourcesTags{}.ListByResourceID(tx, resourceID)
}

func (t ResourcesTags) GetTagIdsByResourceID(tx *gorm.DB, resourceID string) ([]string, error) {
	list, err := t.ListByResourceID(tx, resourceID)
	if err != nil {
		return []string{}, err
	}
	tagIds := make([]string, len(*list))
	for i, v := range *list {
		tagIds[i] = v.TagID
	}
	return tagIds, nil
}

func (t ResourcesTags) SetResourcesTag(tx *gorm.DB, resourceID string, tagSlc []string) error {
	list, err := models.ResourcesTags{}.ListByResourceID(tx, resourceID)
	if err != nil {
		return err
	}
	existingMap := make(map[string]models.ResourcesTags)
	existingIDS := make([]string, len(*list))
	for i, v := range *list {
		existingIDS[i] = v.TagID
		existingMap[v.TagID] = v
	}
	_, toDelete, toAdd := utils.ArrayIntersectDiff(existingIDS, tagSlc)
	if err := t.handleDeletes(tx, existingMap, toDelete); err != nil {
		return err
	}
	if err := t.handleAdds(tx, toAdd, resourceID); err != nil {
		return err
	}

	return nil
}

func (ResourcesTags) handleDeletes(tx *gorm.DB, existingMap map[string]models.ResourcesTags, toDelete []string) error {
	if len(toDelete) > 0 {
		ids := make([]string, len(toDelete))
		for i, id := range toDelete {
			ids[i] = existingMap[id].ID
		}
		return models.ResourcesTags{}.DeleteIDS(tx, ids)
	}
	return nil
}

func (ResourcesTags) handleAdds(tx *gorm.DB, toAdd []string, resourceID string) error {
	if len(toAdd) == 0 {
		return nil
	}
	var newRecords []models.ResourcesTags
	for _, id := range toAdd {
		newRecords = append(newRecords, models.ResourcesTags{
			ID:          core.GenerateUniqueID(),
			ResourcesID: resourceID,
			TagID:       id,
		})
	}

	return models.ResourcesTags{}.Creates(tx, &newRecords)
}

func (ResourcesTags) DeleteByResourcesID(tx *gorm.DB, resourceID string) error {
	return models.ResourcesTags{}.DeleteByResourcesID(tx, resourceID)
}
func (ResourcesTags) DeleteByTagID(tx *gorm.DB, tagID string) error {
	return models.ResourcesTags{}.DeleteByTagID(tx, tagID)
}
func (ResourcesTags) DeleteByTagIDS(tx *gorm.DB, tagIDS []string) error {
	return models.ResourcesTags{}.DeleteByTagIDS(tx, tagIDS)
}
