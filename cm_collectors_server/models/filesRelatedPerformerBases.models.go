package models

import "gorm.io/gorm"

type FilesRelatedPerformerBases struct {
	ID               string `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID     string `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);"`
	PerformerBasesID string `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);"`
	Main             bool   `json:"main" gorm:"column:main;type:tinyint(1);default:0"`
}

func (FilesRelatedPerformerBases) TableName() string {
	return "filesRelatedPerformerBases"
}

func (FilesRelatedPerformerBases) ListByFilesBasesID(db *gorm.DB, filesBasesID string) (*[]FilesRelatedPerformerBases, error) {
	var dataList []FilesRelatedPerformerBases
	err := db.Model(&FilesRelatedPerformerBases{}).Where("filesBases_id = ?", filesBasesID).Find(&dataList).Error
	return &dataList, err
}

func (FilesRelatedPerformerBases) Creates(db *gorm.DB, filesRelatedPerformerBasesSlc *[]FilesRelatedPerformerBases) error {
	return db.Create(filesRelatedPerformerBasesSlc).Error
}
func (FilesRelatedPerformerBases) Update(db *gorm.DB, filesRelatedPerformerBases *FilesRelatedPerformerBases, fields []string) error {
	result := db.Model(&filesRelatedPerformerBases).Select(fields).Updates(filesRelatedPerformerBases)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (FilesRelatedPerformerBases) DeleteIDS(db *gorm.DB, ids []string) error {
	return db.Unscoped().Where("id in (?) ", ids).Delete(&FilesRelatedPerformerBases{}).Error
}
