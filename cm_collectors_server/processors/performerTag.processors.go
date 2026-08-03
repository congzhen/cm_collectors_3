package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PerformerTag struct{}

func (PerformerTag) Data(performerBasesID string, includeDisabled bool) (*models.PerformerTagData, error) {
	return models.PerformerTag{}.Data(core.DBS(), performerBasesID, includeDisabled)
}

func (PerformerTag) CreateClass(par *datatype.ReqParam_PerformerTagClass) (*models.PerformerTagClass, error) {
	name := strings.TrimSpace(par.Name)
	if name == "" {
		return nil, fmt.Errorf("标签分类名称不能为空")
	}
	var baseTotal int64
	if err := core.DBS().Model(&models.PerformerBases{}).Where("id = ?", par.PerformerBasesID).Count(&baseTotal).Error; err != nil {
		return nil, err
	}
	if baseTotal == 0 {
		return nil, fmt.Errorf("演员库不存在")
	}
	createdAt := datatype.CustomTime(core.TimeNow())
	item := &models.PerformerTagClass{
		ID: core.GenerateUniqueID(), PerformerBasesID: par.PerformerBasesID,
		Name: name, Sort: par.Sort, Status: par.Status, CreatedAt: &createdAt,
	}
	if err := core.DBS().Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (PerformerTag) UpdateClass(par *datatype.ReqParam_PerformerTagClass) error {
	name := strings.TrimSpace(par.Name)
	if name == "" {
		return fmt.Errorf("标签分类名称不能为空")
	}
	return core.DBS().Model(&models.PerformerTagClass{}).Where("id = ?", par.ID).Updates(map[string]any{
		"name": name, "sort": par.Sort, "status": par.Status,
	}).Error
}

func (PerformerTag) DeleteClass(id string) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		var tagIDs []string
		if err := tx.Model(&models.PerformerTag{}).Where("performerTagClass_id = ?", id).Pluck("id", &tagIDs).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			if err := tx.Where("performer_tag_id IN ?", tagIDs).Delete(&models.PerformersTags{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", tagIDs).Delete(&models.PerformerTag{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", id).Delete(&models.PerformerTagClass{}).Error
	})
}

func (PerformerTag) CreateTag(par *datatype.ReqParam_PerformerTag) (*models.PerformerTag, error) {
	name := strings.TrimSpace(par.Name)
	if name == "" {
		return nil, fmt.Errorf("标签名称不能为空")
	}
	var classTotal int64
	if err := core.DBS().Model(&models.PerformerTagClass{}).Where("id = ?", par.PerformerTagClassID).Count(&classTotal).Error; err != nil {
		return nil, err
	}
	if classTotal == 0 {
		return nil, fmt.Errorf("演员标签分类不存在")
	}
	createdAt := datatype.CustomTime(core.TimeNow())
	item := &models.PerformerTag{
		ID: core.GenerateUniqueID(), PerformerTagClassID: par.PerformerTagClassID,
		Name: name, Sort: par.Sort, Status: par.Status, CreatedAt: &createdAt,
	}
	if err := core.DBS().Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (PerformerTag) UpdateTag(par *datatype.ReqParam_PerformerTag) error {
	name := strings.TrimSpace(par.Name)
	if name == "" {
		return fmt.Errorf("标签名称不能为空")
	}
	return core.DBS().Model(&models.PerformerTag{}).Where("id = ?", par.ID).Updates(map[string]any{
		"name": name, "sort": par.Sort, "status": par.Status,
	}).Error
}

func (PerformerTag) DeleteTag(id string) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("performer_tag_id = ?", id).Delete(&models.PerformersTags{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.PerformerTag{}).Error
	})
}

func (PerformerTag) UpdateSort(par *datatype.ReqParam_PerformerTagSort) error {
	return core.DBS().Transaction(func(tx *gorm.DB) error {
		for _, item := range par.TagClassSort {
			if err := tx.Model(&models.PerformerTagClass{}).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		for _, item := range par.TagSort {
			if err := tx.Model(&models.PerformerTag{}).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func validatePerformerTagIDs(tx *gorm.DB, performerBasesID string, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}
	var count int64
	err := tx.Model(&models.PerformerTag{}).
		Joins("INNER JOIN performerTagClass ON performerTagClass.id = performerTag.performerTagClass_id").
		Where("performerTag.id IN ? AND performerTagClass.performerBases_id = ?", tagIDs, performerBasesID).
		Count(&count).Error
	if err != nil {
		return err
	}
	unique := make(map[string]struct{}, len(tagIDs))
	for _, id := range tagIDs {
		if id != "" {
			unique[id] = struct{}{}
		}
	}
	if count != int64(len(unique)) {
		return fmt.Errorf("演员标签不属于当前演员库")
	}
	return nil
}
