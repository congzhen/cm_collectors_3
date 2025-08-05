package models

import (
	"cm_collectors_server/datatype"
	"fmt"
	"strings"

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

type DramaSeriesWithResource struct {
	ID          string `json:"id"`
	ResourcesID string `json:"resources_id" `
	Src         string `json:"src"`
	Title       string `json:"title"`
}

func (t ResourcesDramaSeries) SearchPath(db *gorm.DB, filesBasesIds []string, searchPath string) (*[]DramaSeriesWithResource, error) {
	var dataList []DramaSeriesWithResource
	db = db.Table(fmt.Sprintf("%s AS t", t.TableName())).
		Joins(fmt.Sprintf("left join %s as  r on t.resources_id = r.id", Resources{}.TableName())).
		Select("t.id,t.resources_id,t.src, r.title")
	if len(filesBasesIds) > 0 {
		db = db.Where("r.filesBases_id in (?)", filesBasesIds)
	}
	searchPath = "%" + searchPath + "%"
	db = db.Where("t.src like ?", searchPath)
	err := db.Find(&dataList).Error
	return &dataList, err
}

func (t ResourcesDramaSeries) ReplacePath(db *gorm.DB, filesBasesIds []string, searchPath, replacePath string) (*[]DramaSeriesWithResource, error) {
	dataList, err := t.SearchPath(db, filesBasesIds, searchPath)
	if err != nil {
		return nil, err
	}
	if len(*dataList) == 0 {
		return nil, nil
	}
	// 准备批量更新的 cases
	var cases []string
	var ids []string
	var args []interface{}
	for i, v := range *dataList {
		newSrc := strings.Replace(v.Src, searchPath, replacePath, -1)
		(*dataList)[i].Src = newSrc // 更新返回值
		cases = append(cases, "WHEN ? THEN ?")
		ids = append(ids, v.ID)
		args = append(args, v.ID, newSrc)
	}
	if len(ids) > 0 {
		// 构建批量更新 SQL
		sql := fmt.Sprintf("UPDATE %s SET src = CASE id %s END WHERE id IN (?)",
			t.TableName(), strings.Join(cases, " "))
		args = append(args, ids)

		err = db.Exec(sql, args...).Error
		if err != nil {
			return nil, err
		}
	}

	return dataList, nil
}

func (ResourcesDramaSeries) Info(db *gorm.DB, id string) (*ResourcesDramaSeries, error) {
	var info ResourcesDramaSeries
	err := db.Where("id = ?", id).First(&info).Error
	return &info, err
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
