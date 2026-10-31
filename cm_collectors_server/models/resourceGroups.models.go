package models

import (
	"cm_collectors_server/core"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

const maxResourceGroupMembers = 1000

type ResourceGroups struct{}

// ResourceGroupPreview keeps summaries small; full playback data is loaded by Info.
type ResourceGroupPreview struct {
	ID                   string `json:"id"`
	FilesBasesID         string `json:"filesBases_id" gorm:"column:filesBases_id"`
	Title                string `json:"title"`
	Subtitle             string `json:"subtitle"`
	CoverPoster          string `json:"coverPoster" gorm:"column:coverPoster"`
	CoverPosterWidth     int    `json:"coverPosterWidth" gorm:"column:coverPosterWidth"`
	CoverPosterHeight    int    `json:"coverPosterHeight" gorm:"column:coverPosterHeight"`
	AssociationGroupID   string `json:"associationGroupId" gorm:"column:association_group_id"`
	AssociationGroupSort int    `json:"associationGroupSort" gorm:"column:association_group_sort"`
}
type ResourceGroupSummary struct {
	ID           string                 `json:"id"`
	FilesBasesID string                 `json:"filesBasesId"`
	Count        int64                  `json:"count"`
	Preview      []ResourceGroupPreview `json:"preview"`
}

func groupResourceIDs(ids []string) ([]string, error) {
	if len(ids) == 0 || len(ids) > maxResourceGroupMembers {
		return nil, errors.New("请选择 1 至 1000 个资源")
	}
	unique := make(map[string]bool, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return nil, errors.New("资源 ID 不能为空")
		}
		if !unique[id] {
			unique[id] = true
			result = append(result, id)
		}
	}
	return result, nil
}

func groupMembers(db *gorm.DB, library, groupID string) ([]Resources, error) {
	members := []Resources{}
	if library == "" || groupID == "" {
		return members, errors.New("文件库或分组不能为空")
	}
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("filesBases_id = ? AND association_group_id = ?", library, groupID).
		Order("association_group_sort ASC, id ASC").Find(&members).Error
	if err == nil && len(members) < 2 {
		err = errors.New("分组不存在或已解散，请刷新后重试")
	}
	return members, err
}

func (ResourceGroups) Create(db *gorm.DB, library string, ids []string) (string, error) {
	ids, err := groupResourceIDs(ids)
	if err != nil {
		return "", err
	}
	if len(ids) < 2 || library == "" {
		return "", errors.New("请在同一文件库中选择至少两个未分组资源")
	}
	groupID := core.GenerateUniqueID()
	err = db.Transaction(func(tx *gorm.DB) error {
		var resources []Resources
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ? AND filesBases_id = ?", ids, library).Order("id ASC").Find(&resources).Error; err != nil {
			return err
		}
		if len(resources) != len(ids) {
			return errors.New("资源不存在或不属于当前文件库")
		}
		for _, resource := range resources {
			if resource.AssociationGroupID != "" {
				return errors.New("所选资源已加入其他分组，请刷新后重试")
			}
		}
		for index, id := range ids {
			result := tx.Model(&Resources{}).Where("id = ? AND (association_group_id = '' OR association_group_id IS NULL)", id).
				Updates(map[string]interface{}{"association_group_id": groupID, "association_group_sort": index})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("资源分组已变化，请刷新后重试")
			}
		}
		return nil
	})
	return groupID, err
}

func (ResourceGroups) Add(db *gorm.DB, library, groupID string, ids []string) error {
	ids, err := groupResourceIDs(ids)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		members, err := groupMembers(tx, library, groupID)
		if err != nil {
			return err
		}
		var resources []Resources
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ? AND filesBases_id = ?", ids, library).Order("id ASC").Find(&resources).Error; err != nil {
			return err
		}
		if len(resources) != len(ids) {
			return errors.New("资源不存在或不属于当前文件库")
		}
		byID := make(map[string]Resources, len(resources))
		for _, resource := range resources {
			if resource.AssociationGroupID != "" && resource.AssociationGroupID != groupID {
				return errors.New("不能将其他组的成员直接加入，请先移出原组")
			}
			byID[resource.ID] = resource
		}
		addedCount := 0
		for _, resource := range resources {
			if resource.AssociationGroupID == "" {
				addedCount++
			}
		}
		if len(members)+addedCount > maxResourceGroupMembers {
			return errors.New("每个关联组最多包含 1000 个资源")
		}
		next := members[len(members)-1].AssociationGroupSort + 1
		for _, id := range ids {
			if byID[id].AssociationGroupID == groupID {
				continue
			}
			result := tx.Model(&Resources{}).Where("id = ? AND (association_group_id = '' OR association_group_id IS NULL)", id).
				Updates(map[string]interface{}{"association_group_id": groupID, "association_group_sort": next})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("资源分组已变化，请刷新后重试")
			}
			next++
		}
		return nil
	})
}

func (ResourceGroups) Remove(db *gorm.DB, library, groupID string, ids []string) error {
	ids, err := groupResourceIDs(ids)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		members, err := groupMembers(tx, library, groupID)
		if err != nil {
			return err
		}
		known := map[string]bool{}
		for _, member := range members {
			known[member.ID] = true
		}
		for _, id := range ids {
			if !known[id] {
				return errors.New("资源已不属于此组，请刷新后重试")
			}
		}
		if err := tx.Model(&Resources{}).Where("id IN ? AND filesBases_id = ? AND association_group_id = ?", ids, library, groupID).
			Updates(map[string]interface{}{"association_group_id": "", "association_group_sort": 0}).Error; err != nil {
			return err
		}
		if len(members)-len(ids) < 2 {
			return (ResourceGroups{}).Dissolve(tx, library, groupID)
		}
		return nil
	})
}

func (ResourceGroups) Dissolve(db *gorm.DB, library, groupID string) error {
	if library == "" || groupID == "" {
		return errors.New("文件库或分组不能为空")
	}
	return db.Model(&Resources{}).Where("filesBases_id = ? AND association_group_id = ?", library, groupID).
		Updates(map[string]interface{}{"association_group_id": "", "association_group_sort": 0}).Error
}

func (ResourceGroups) Reorder(db *gorm.DB, library, groupID string, ids []string) error {
	ids, err := groupResourceIDs(ids)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		members, err := groupMembers(tx, library, groupID)
		if err != nil {
			return err
		}
		if len(ids) != len(members) {
			return errors.New("分组成员已变化，请刷新后重试")
		}
		known := map[string]bool{}
		for _, member := range members {
			known[member.ID] = true
		}
		for index, id := range ids {
			if !known[id] {
				return errors.New("分组成员已变化，请刷新后重试")
			}
			if err := tx.Model(&Resources{}).Where("id = ? AND association_group_id = ?", id, groupID).Update("association_group_sort", index).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (ResourceGroups) Info(db *gorm.DB, library, groupID string) ([]Resources, error) {
	resources := []Resources{}
	if library == "" || groupID == "" {
		return resources, errors.New("文件库或分组不能为空")
	}
	err := (Resources{}).Preload(db).Where("filesBases_id = ? AND association_group_id = ?", library, groupID).
		Order("association_group_sort ASC, resources.id ASC").Find(&resources).Error
	if err == nil && len(resources) < 2 {
		err = errors.New("分组不存在或已解散，请刷新后重试")
	}
	for i := range resources {
		if resources[i].Tags == nil {
			resources[i].Tags = []Tag{}
		}
		if resources[i].Performers == nil {
			resources[i].Performers = []Performer{}
		}
		if resources[i].Directors == nil {
			resources[i].Directors = []Performer{}
		}
		if resources[i].ResourcesDramaSeries == nil {
			resources[i].ResourcesDramaSeries = []ResourcesDramaSeries{}
		}
	}
	if err == nil {
		err = (Performer{}).fillResourceCountsForResources(db, resources)
	}
	return resources, err
}

// Summaries 只加载封面摘要，不预载分集、演员和标签，供叠牌和选组面板使用。
func (ResourceGroups) Summaries(db *gorm.DB, library string, groupIDs []string, search string, page, limit int) ([]ResourceGroupSummary, int64, error) {
	result := []ResourceGroupSummary{}
	if library == "" {
		return result, 0, errors.New("文件库不能为空")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if len(groupIDs) > 1000 {
		return result, 0, errors.New("分组数量过多")
	}
	query := db.Model(&Resources{}).Where("filesBases_id = ? AND association_group_id <> ''", library)
	if len(groupIDs) > 0 {
		query = query.Where("association_group_id IN ?", groupIDs)
	}
	if search != "" {
		matching := db.Model(&Resources{}).Select("association_group_id").Where("filesBases_id = ? AND (title LIKE ? OR subtitle LIKE ? OR issueNumber LIKE ?)", library, "%"+search+"%", "%"+search+"%", "%"+search+"%")
		query = query.Where("association_group_id IN (?)", matching)
	}
	type countedGroup struct {
		ID    string `gorm:"column:association_group_id"`
		Count int64
	}
	grouped := query.Select("association_group_id, COUNT(*) AS count").Group("association_group_id").Having("COUNT(*) >= 2")
	var total int64
	if err := db.Table("(?) AS grouped_resources", grouped).Count(&total).Error; err != nil {
		return result, 0, err
	}
	var groups []countedGroup
	// 指定分组摘要批量获取；选组面板才按组分页。
	if len(groupIDs) == 0 {
		grouped = grouped.Limit(limit).Offset((page - 1) * limit)
	}
	if err := grouped.Order("association_group_id ASC").Scan(&groups).Error; err != nil {
		return result, 0, err
	}
	for _, group := range groups {
		preview := []ResourceGroupPreview{}
		if err := db.Model(&Resources{}).Select("id", "filesBases_id", "title", "subtitle", "coverPoster", "coverPosterWidth", "coverPosterHeight", "association_group_id", "association_group_sort").
			Where("filesBases_id = ? AND association_group_id = ?", library, group.ID).Order("association_group_sort ASC, id ASC").Limit(7).Find(&preview).Error; err != nil {
			return result, 0, err
		}
		result = append(result, ResourceGroupSummary{ID: group.ID, FilesBasesID: library, Count: group.Count, Preview: preview})
	}
	return result, total, nil
}

// SetResourceGroupDisplay 只修改本库独立字段，不修改公共配置或跟随关系。
func SetResourceGroupDisplay(db *gorm.DB, library string, enabled bool) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var settings FilesBasesSetting
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("filesBases_id = ?", library).First(&settings).Error; err != nil {
			return err
		}
		config, err := parseConfig(settings.ConfigJsonData)
		if err != nil {
			return err
		}
		config["mergeResourceGroups"] = enabled
		raw, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return tx.Model(&settings).Update("config_json_data", string(raw)).Error
	})
}
