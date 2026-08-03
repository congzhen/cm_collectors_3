package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type PerformerTagClass struct {
	ID               string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerBasesID string               `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);index:idx_performer_tag_class_base"`
	Name             string               `json:"name" gorm:"type:varchar(100);"`
	Sort             int                  `json:"sort" gorm:"type:int;default:0"`
	Status           bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
	CreatedAt        *datatype.CustomTime `json:"createdAt" gorm:"column:addTime;type:datetime"`
}

func (PerformerTagClass) TableName() string { return "performerTagClass" }

type PerformerTag struct {
	ID                  string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerTagClassID string               `json:"performerTagClass_id" gorm:"column:performerTagClass_id;type:char(20);index:idx_performer_tag_class"`
	Name                string               `json:"name" gorm:"type:varchar(100);"`
	Sort                int                  `json:"sort" gorm:"type:int;default:0"`
	Status              bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
	CreatedAt           *datatype.CustomTime `json:"createdAt" gorm:"column:addTime;type:datetime"`
	PerformerCount      int64                `json:"performerCount" gorm:"-"`
}

func (PerformerTag) TableName() string { return "performerTag" }

type PerformersTags struct {
	PerformerID    string `json:"performer_id" gorm:"column:performer_id;type:char(20);primaryKey;index:idx_performers_tags_performer"`
	PerformerTagID string `json:"performerTag_id" gorm:"column:performer_tag_id;type:char(20);primaryKey;index:idx_performers_tags_tag"`
}

func (PerformersTags) TableName() string { return "performersTags" }

type PerformerTagData struct {
	TagClasses []PerformerTagClass `json:"tagClasses"`
	Tags       []PerformerTag      `json:"tags"`
}

type performerTagCount struct {
	PerformerTagID string `gorm:"column:performer_tag_id"`
	Count          int64  `gorm:"column:performerCount"`
}

func (PerformerTag) Data(db *gorm.DB, performerBasesID string, includeDisabled bool) (*PerformerTagData, error) {
	classQuery := db.Where("performerBases_id = ?", performerBasesID)
	if !includeDisabled {
		classQuery = classQuery.Where("status = ?", true)
	}
	var classes []PerformerTagClass
	if err := classQuery.Order("sort asc, addTime asc").Find(&classes).Error; err != nil {
		return nil, err
	}

	classIDs := make([]string, 0, len(classes))
	for _, class := range classes {
		classIDs = append(classIDs, class.ID)
	}
	data := &PerformerTagData{TagClasses: classes, Tags: []PerformerTag{}}
	if len(classIDs) == 0 {
		return data, nil
	}

	tagQuery := db.Where("performerTagClass_id IN ?", classIDs)
	if !includeDisabled {
		tagQuery = tagQuery.Where("status = ?", true)
	}
	if err := tagQuery.Order("sort asc, addTime asc").Find(&data.Tags).Error; err != nil {
		return nil, err
	}

	tagIDs := make([]string, 0, len(data.Tags))
	for _, tag := range data.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	if len(tagIDs) == 0 {
		return data, nil
	}
	var counts []performerTagCount
	if err := db.Table("performersTags").
		Select("performersTags.performer_tag_id, COUNT(DISTINCT performersTags.performer_id) AS performerCount").
		Joins("INNER JOIN performer ON performer.id = performersTags.performer_id").
		Where("performersTags.performer_tag_id IN ? AND performer.status = ?", tagIDs, true).
		Group("performersTags.performer_tag_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	countMap := make(map[string]int64, len(counts))
	for _, item := range counts {
		countMap[item.PerformerTagID] = item.Count
	}
	for i := range data.Tags {
		data.Tags[i].PerformerCount = countMap[data.Tags[i].ID]
	}
	return data, nil
}

func (PerformersTags) Set(db *gorm.DB, performerID string, tagIDs []string) error {
	if err := db.Where("performer_id = ?", performerID).Delete(&PerformersTags{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	relations := make([]PerformersTags, 0, len(tagIDs))
	seen := make(map[string]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID == "" {
			continue
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		relations = append(relations, PerformersTags{PerformerID: performerID, PerformerTagID: tagID})
	}
	if len(relations) == 0 {
		return nil
	}
	return db.Create(&relations).Error
}
