package models

import (
	"cm_collectors_server/datatype"
	"time"

	"gorm.io/gorm"
)

type SearchFavorite struct {
	ID                string                       `json:"id" gorm:"primaryKey;type:char(20)"`
	FilesBasesID      string                       `json:"filesBasesId" gorm:"column:files_bases_id;type:char(20);index:idx_search_favorite_library_sort"`
	SearchData        datatype.ReqParam_SearchData `json:"searchData" gorm:"column:search_data;serializer:json;type:text"`
	SchemaVersion     int                          `json:"schemaVersion" gorm:"column:schema_version;type:int;default:1"`
	Sort              int                          `json:"sort" gorm:"type:int;default:0;index:idx_search_favorite_library_sort"`
	InvalidConditions int                          `json:"invalidConditions" gorm:"-"`
	OptionLabels      map[string]string            `json:"optionLabels" gorm:"-"`
	CreatedAt         time.Time                    `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt         time.Time                    `json:"updatedAt" gorm:"column:updated_at"`
}

func (SearchFavorite) TableName() string {
	return "search_favorites"
}

func (SearchFavorite) List(db *gorm.DB, filesBasesID string) ([]SearchFavorite, error) {
	var list []SearchFavorite
	err := db.Where("files_bases_id = ?", filesBasesID).
		Order("sort asc, created_at asc, id asc").
		Find(&list).Error
	return list, err
}

func (SearchFavorite) CountByFilesBasesID(db *gorm.DB, filesBasesID string) (int64, error) {
	var count int64
	err := db.Model(&SearchFavorite{}).Where("files_bases_id = ?", filesBasesID).Count(&count).Error
	return count, err
}

func (SearchFavorite) MaxSort(db *gorm.DB, filesBasesID string) (int, error) {
	var maxSort int
	err := db.Model(&SearchFavorite{}).
		Where("files_bases_id = ?", filesBasesID).
		Select("COALESCE(MAX(sort), -1)").
		Scan(&maxSort).Error
	return maxSort, err
}

func (SearchFavorite) Create(db *gorm.DB, favorite *SearchFavorite) error {
	return db.Create(favorite).Error
}

func (SearchFavorite) Get(db *gorm.DB, id string) (*SearchFavorite, error) {
	var favorite SearchFavorite
	err := db.First(&favorite, "id = ?", id).Error
	return &favorite, err
}

func (SearchFavorite) Save(db *gorm.DB, favorite *SearchFavorite) error {
	return db.Model(&SearchFavorite{}).
		Where("id = ?", favorite.ID).
		Select("SearchData", "SchemaVersion", "UpdatedAt").
		Updates(favorite).Error
}

func (SearchFavorite) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&SearchFavorite{}).Error
}

func (SearchFavorite) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Where("files_bases_id = ?", filesBasesID).Delete(&SearchFavorite{}).Error
}
