package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type ResourcesDramaSeries struct {
	ID                string               `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID       string               `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_ResourcesDramaSeries_resourcesID"`
	Type              string               `json:"type" gorm:"column:type;type:varchar(50);"`
	Src               string               `json:"src" gorm:"column:src;type:text;"`
	Sort              int                  `json:"sort" gorm:"type:int;default:0"`
	M3u8BuilderTime   *datatype.CustomTime `json:"m3u8BuilderTime" gorm:"column:m3u8BuilderTime;type:datetime;"`
	M3u8BuilderStatus bool                 `json:"m3u8BuilderStatus" gorm:"column:m3u8BuilderStatus;type:tinyint(1);default:0"`
}

func (ResourcesDramaSeries) TableName() string {
	return "resourcesDramaSeries"
}

func (ResourcesDramaSeries) ListByResourceID(db *gorm.DB, resourceID string) (*[]ResourcesDramaSeries, error) {
	var dataList []ResourcesDramaSeries
	err := db.Where("resources_id = ?", resourceID).Order("sort").Find(&dataList).Error
	return &dataList, err
}

func (ResourcesDramaSeries) Creates(db *gorm.DB, resourcesDramaSeriesSlc *[]ResourcesDramaSeries) error {
	return db.Create(resourcesDramaSeriesSlc).Error
}
func (ResourcesDramaSeries) DeleteIDS(db *gorm.DB, ids []string) error {
	return db.Unscoped().Where("id in (?) ", ids).Delete(&ResourcesDramaSeries{}).Error
}
func (ResourcesDramaSeries) DeleteByResourcesID(db *gorm.DB, resourcesID string) error {
	return db.Unscoped().Where("resources_id = ? ", resourcesID).Delete(&ResourcesDramaSeries{}).Error
}
