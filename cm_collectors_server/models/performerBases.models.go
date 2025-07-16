package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type PerformerBases struct {
	ID        string               `json:"id" gorm:"primaryKey;type:char(20);"`
	Name      string               `json:"name" gorm:"type:varchar(200);"`
	Sort      int                  `json:"sort" gorm:"type:int;"`
	CreatedAt *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Status    bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (PerformerBases) TableName() string {
	return "performerBases"
}

func (PerformerBases) DataList(db *gorm.DB) (*[]PerformerBases, error) {
	var dataList []PerformerBases
	err := db.Model(&PerformerBases{}).Order("sort").Find(&dataList).Error
	return &dataList, err
}
func (PerformerBases) Update(db *gorm.DB, performerBases *PerformerBases, fields []string) error {
	result := db.Model(&performerBases).Select(fields).Updates(performerBases)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
