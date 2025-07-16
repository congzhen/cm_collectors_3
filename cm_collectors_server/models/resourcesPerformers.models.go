package models

import "gorm.io/gorm"

type ResourcesPerformers struct {
	ID          string `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID string `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_ResourcesPerformers_resourcesID,priority:1;"`
	PerformerID string `json:"performer_id" gorm:"column:performer_id;type:char(20);index:idx_ResourcesPerformers_performerID;index:idx_ResourcesPerformers_resourcesID,priority:2"`
	Sort        int    `json:"sort" gorm:"type:int;default:0"`
}

func (ResourcesPerformers) TableName() string {
	return "resourcesPerformers"
}

func (ResourcesPerformers) ListByResourceID(db *gorm.DB, resourceID string) (*[]ResourcesPerformers, error) {
	var dataList []ResourcesPerformers
	err := db.Where("resources_id = ?", resourceID).Order("sort").Find(&dataList).Error
	return &dataList, err
}

func (ResourcesPerformers) Creates(db *gorm.DB, resourcesPerformerSlc *[]ResourcesPerformers) error {
	return db.Create(resourcesPerformerSlc).Error
}
func (ResourcesPerformers) DeleteIDS(db *gorm.DB, ids []string) error {
	return db.Unscoped().Where("id in (?) ", ids).Delete(&ResourcesPerformers{}).Error
}
