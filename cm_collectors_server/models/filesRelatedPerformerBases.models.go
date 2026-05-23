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

// CountByPerformerBasesID 统计有多少文件库仍关联着指定演员库。
// 即使演员库本身没有演员，只要还被文件库引用，就不能直接删除，否则文件库的演员库选择会失效。
func (FilesRelatedPerformerBases) CountByPerformerBasesID(db *gorm.DB, performerBasesID string) (int64, error) {
	var total int64
	err := db.Model(&FilesRelatedPerformerBases{}).Where("performerBases_id = ?", performerBasesID).Count(&total).Error
	return total, err
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

// DeleteByFilesBasesID 删除文件库与演员库的关联关系。
// 文件库真实删除成功时，它的关联关系也不再有意义，需要在同一个事务中一起清理。
func (FilesRelatedPerformerBases) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Unscoped().Where("filesBases_id = ?", filesBasesID).Delete(&FilesRelatedPerformerBases{}).Error
}
