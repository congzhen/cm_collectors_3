package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type FilesBases struct {
	ID                         string                       `json:"id" gorm:"primaryKey;type:char(20);"`
	Name                       string                       `json:"name" gorm:"type:varchar(200);"`
	Sort                       int                          `json:"sort" gorm:"type:int;"`
	CreatedAt                  datatype.CustomTime          `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status                     bool                         `json:"status" gorm:"type:tinyint(1);"`
	FilesRelatedPerformerBases []FilesRelatedPerformerBases `json:"filesRelatedPerformerBases" gorm:"foreignKey:FilesBasesID;references:ID;"`
	FilesBasesSetting          FilesBasesSetting            `json:"filesBasesSetting" gorm:"foreignKey:FilesBasesID;references:ID;"`
}

func (FilesBases) TableName() string {
	return "filesBases"
}

func (FilesBases) preloadTable(db *gorm.DB) *gorm.DB {
	return db.Preload("FilesBasesSetting").Preload("FilesRelatedPerformerBases")
}

func (t FilesBases) DataList(db *gorm.DB) (*[]FilesBases, error) {
	var dataList []FilesBases
	err := t.preloadTable(db).Model(&FilesBases{}).Order("sort").Find(&dataList).Error
	return &dataList, err
}
