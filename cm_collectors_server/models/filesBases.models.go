package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type FilesBases struct {
	ID                         string                       `json:"id" gorm:"primaryKey;type:char(20);"`
	Name                       string                       `json:"name" gorm:"type:varchar(200);"`
	Sort                       int                          `json:"sort" gorm:"type:int;"`
	CreatedAt                  *datatype.CustomTime         `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status                     bool                         `json:"status" gorm:"type:tinyint(1);"`
	FilesRelatedPerformerBases []FilesRelatedPerformerBases `json:"filesRelatedPerformerBases" gorm:"foreignKey:FilesBasesID;references:ID;"`
}

type FilesBasesDetails struct {
	FilesBases
	FilesBasesSetting FilesBasesSetting `json:"filesBasesSetting" gorm:"foreignKey:FilesBasesID;references:ID;"`
}

func (FilesBases) TableName() string {
	return "filesBases"
}

func (FilesBases) preloadTable(db *gorm.DB) *gorm.DB {
	return db.Preload("FilesBasesSetting").Preload("FilesRelatedPerformerBases")
}
func (FilesBases) preloadTableBase(db *gorm.DB) *gorm.DB {
	return db.Preload("FilesRelatedPerformerBases")
}

func (t FilesBases) DataList(db *gorm.DB) (*[]FilesBases, error) {
	var dataList []FilesBases
	err := t.preloadTableBase(db).Model(&FilesBases{}).Order("sort").Find(&dataList).Error
	return &dataList, err
}
func (t FilesBases) DataListByIds(db *gorm.DB, ids []string) (*[]FilesBases, error) {
	var dataList []FilesBases
	err := t.preloadTableBase(db).Model(&FilesBases{}).Where("id in (?)", ids).Order("sort").Find(&dataList).Error
	return &dataList, err
}
func (t FilesBases) Info(db *gorm.DB, id string) (*FilesBases, error) {
	var info FilesBases
	err := t.preloadTableBase(db).First(&info, "id = ?", id).Error
	return &info, err
}

func (t FilesBases) InfoDetails(db *gorm.DB, id string) (*FilesBasesDetails, error) {
	var data FilesBasesDetails
	err := t.preloadTable(db).Model(&FilesBases{}).Where("id = ?", id).First(&data).Error
	return &data, err
}

func (FilesBases) GetTotal(db *gorm.DB) (int64, error) {
	var total int64
	err := db.Model(&FilesBases{}).Count(&total).Error
	return total, err
}

func (FilesBases) Update(db *gorm.DB, filesBases *FilesBases, fields []string) error {
	result := db.Model(&filesBases).Select(fields).Updates(filesBases)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (FilesBases) Create(db *gorm.DB, filesBases *FilesBases) error {
	return db.Create(&filesBases).Error
}
