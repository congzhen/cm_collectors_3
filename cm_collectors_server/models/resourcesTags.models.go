package models

import "gorm.io/gorm"

type ResourcesTags struct {
	ID          string `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID string `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_ResourcesTags_ResourcesID;index:idx_ResourcesTags_tagID,priority:2;"`
	TagID       string `json:"tag_id" gorm:"column:tag_id;type:char(20);index:idx_ResourcesTags_tagID,priority:1"`
	Sort        int    `json:"sort" gorm:"type:int;default:0"`
}

func (ResourcesTags) TableName() string {
	return "resourcesTags"
}

func (ResourcesTags) ListByResourceID(db *gorm.DB, resourceID string) (*[]ResourcesTags, error) {
	var dataList []ResourcesTags
	err := db.Where("resources_id = ?", resourceID).Order("sort").Find(&dataList).Error
	return &dataList, err
}

func (ResourcesTags) Creates(db *gorm.DB, resourcesTagSlc *[]ResourcesTags) error {
	return db.Create(resourcesTagSlc).Error
}
func (ResourcesTags) DeleteIDS(db *gorm.DB, ids []string) error {
	return db.Unscoped().Where("id in (?) ", ids).Delete(&ResourcesTags{}).Error
}

func (ResourcesTags) DeleteByResourcesID(db *gorm.DB, resourcesID string) error {
	return db.Unscoped().Where("resources_id = ?", resourcesID).Delete(&ResourcesTags{}).Error
}
