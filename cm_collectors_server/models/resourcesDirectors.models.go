package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ResourcesDirectors struct {
	ID          string `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID string `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_resources_directors_resourcesID,priority:1;"`
	PerformerID string `json:"performer_id" gorm:"column:performer_id;type:char(20);index:idx_resources_directors_performerID;index:idx_resources_directors_resourcesID,priority:2"`
	Sort        int    `json:"sort" gorm:"type:int;default:0"`
}

func (ResourcesDirectors) TableName() string {
	return "resourcesDirectors"
}

func (ResourcesDirectors) ListByResourceID(db *gorm.DB, resourceID string) (*[]ResourcesDirectors, error) {
	var dataList []ResourcesDirectors
	err := db.Where("resources_id = ?", resourceID).Order("sort").Find(&dataList).Error
	return &dataList, err
}

func (ResourcesDirectors) Creates(db *gorm.DB, resourcesDirectorSlc *[]ResourcesDirectors) error {
	return db.Create(resourcesDirectorSlc).Error
}
func (ResourcesDirectors) DeleteIDS(db *gorm.DB, ids []string) error {
	return db.Unscoped().Where("id in (?) ", ids).Delete(&ResourcesDirectors{}).Error
}

func (ResourcesDirectors) DeleteByResourcesID(db *gorm.DB, resourcesID string) error {
	return db.Unscoped().Where("resources_id = ?", resourcesID).Delete(&ResourcesDirectors{}).Error
}
func (ResourcesDirectors) DeleteByFilesBasesID(db *gorm.DB, filesBases_id string) error {
	sqlWhere := fmt.Sprintf("resources_id in (select id from %s where filesBases_id = ?)", Resources{}.TableName())
	return db.Unscoped().Where(sqlWhere, filesBases_id).Delete(&ResourcesDirectors{}).Error
}
