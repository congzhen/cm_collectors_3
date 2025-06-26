package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type Tag struct {
	ID         string              `json:"id" gorm:"primaryKey;type:char(20);"`
	TagClassID string              `json:"tagClass_id" gorm:"column:tagClass_id;type:char(20);index:idx_tag_tagClassID;"`
	Name       string              `json:"name" gorm:"type:varchar(200);"`
	Hot        int                 `json:"hot" gorm:"type:int;default:0"`
	Sort       int                 `json:"sort" gorm:"type:int;default:0"`
	CreatedAt  datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Status     bool                `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (Tag) TableName() string {
	return "tag"
}

func (t Tag) DataListByTagClassIds(db *gorm.DB, tagClassIds []string) (*[]Tag, error) {
	var dataList []Tag
	err := db.Where("tagClass_id in (?)", tagClassIds).Order("sort").Find(&dataList).Error
	return &dataList, err
}
