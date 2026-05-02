package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type TvboxRecommend struct {
	ID         string               `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourceID string               `json:"resourceId" gorm:"column:resourceId;type:char(20);uniqueIndex:idx_tvboxRecommend_resourceId;"`
	Sort       int                  `json:"sort" gorm:"type:int;default:0"`
	CreatedAt  *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Resource   Resources            `json:"resource" gorm:"foreignKey:ResourceID;references:ID"`
}

func (TvboxRecommend) List(db *gorm.DB) (*[]TvboxRecommend, error) {
	var dataList []TvboxRecommend
	err := db.
		Preload("Resource.Performers").
		Preload("Resource.Directors").
		Preload("Resource.Tags").
		Order("sort desc").
		Find(&dataList).Error
	return &dataList, err
}

func (TvboxRecommend) Create(db *gorm.DB, record *TvboxRecommend) error {
	return db.Create(record).Error
}

func (TvboxRecommend) DeleteByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&TvboxRecommend{}).Error
}

func (TvboxRecommend) ExistsByResourceID(db *gorm.DB, resourceId string) (bool, error) {
	var count int64
	err := db.Model(&TvboxRecommend{}).Where("resourceId = ?", resourceId).Count(&count).Error
	return count > 0, err
}

func (TvboxRecommend) MaxSort(db *gorm.DB) (int, error) {
	var maxSort int
	err := db.Model(&TvboxRecommend{}).Select("COALESCE(MAX(sort), -1)").Scan(&maxSort).Error
	return maxSort, err
}

func (TvboxRecommend) UpdateSort(db *gorm.DB, id string, sort int) error {
	return db.Model(&TvboxRecommend{}).Where("id = ?", id).Update("sort", sort).Error
}
