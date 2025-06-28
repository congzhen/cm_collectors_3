package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type TagClass struct {
	ID           string               `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID string               `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_tagClass_filesBasesID;"`
	Name         string               `json:"name" gorm:"type:varchar(200);"`
	LeftShow     bool                 `json:"leftShow" gorm:"column:leftShow;type:tinyint(1);default:1"`
	Hot          int                  `json:"hot" gorm:"type:int;default:0"`
	Sort         int                  `json:"sort" gorm:"type:int;default:0"`
	CreatedAt    *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Status       bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (TagClass) TableName() string {
	return "tagClass"
}

func (t TagClass) DataListByFilesBasesId(db *gorm.DB, filesBasesID string) (*[]TagClass, error) {
	var dataList []TagClass
	err := db.Where("filesBases_id = ?", filesBasesID).Order("sort").Find(&dataList).Error
	return &dataList, err
}

func (TagClass) GetTotalByFilesBasesId(db *gorm.DB, filesBasesID string) (int64, error) {
	var total int64
	err := db.Model(&TagClass{}).Where("filesBases_id = ?", filesBasesID).Count(&total).Error
	return total, err
}

func (TagClass) Update(db *gorm.DB, tagClass *TagClass, fields []string) error {
	result := db.Model(&tagClass).Select(fields).Updates(tagClass)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (TagClass) Create(db *gorm.DB, tagClass *TagClass) error {
	return db.Create(&tagClass).Error
}
