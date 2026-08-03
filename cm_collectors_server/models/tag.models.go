package models

import (
	"cm_collectors_server/datatype"
	"fmt"

	"gorm.io/gorm"
)

type Tag struct {
	ID            string               `json:"id" gorm:"primaryKey;type:char(20);"`
	TagClassID    string               `json:"tagClass_id" gorm:"column:tagClass_id;type:char(20);index:idx_tag_tagClassID;"`
	Name          string               `json:"name" gorm:"type:varchar(200);"`
	KeyWords      string               `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	AIDescription string               `json:"aiDescription" gorm:"column:aiDescription;type:text;"`
	AIEnabled     bool                 `json:"aiEnabled" gorm:"column:aiEnabled;type:tinyint(1);default:1"`
	Hot           int                  `json:"hot" gorm:"type:int;default:0"`
	Sort          int                  `json:"sort" gorm:"type:int;default:0"`
	CreatedAt     *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Status        bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
	ResourceCount int64                `json:"resourceCount" gorm:"-"`
}

type tagResourceCount struct {
	TagID         string `gorm:"column:tag_id"`
	ResourceCount int64  `gorm:"column:resourceCount"`
}

func (Tag) TableName() string {
	return "tag"
}

func (t Tag) DataListByTagClassIds(db *gorm.DB, tagClassIds []string, includeResourceCount ...bool) (*[]Tag, error) {
	var dataList []Tag
	if err := db.Where("tagClass_id in (?)", tagClassIds).Order("sort").Find(&dataList).Error; err != nil {
		return &dataList, err
	}
	includeCount := len(includeResourceCount) == 0 || includeResourceCount[0]
	if includeCount {
		if err := t.fillResourceCounts(db, dataList); err != nil {
			return &dataList, err
		}
	}
	return &dataList, nil
}

// fillResourceCounts 批量回填每个自定义标签关联的有效资源数量，避免逐标签查询。
func (Tag) fillResourceCounts(db *gorm.DB, dataList []Tag) error {
	if len(dataList) == 0 {
		return nil
	}

	tagIDs := make([]string, 0, len(dataList))
	for _, tag := range dataList {
		tagIDs = append(tagIDs, tag.ID)
	}

	var counts []tagResourceCount
	if err := db.Raw(`
		SELECT resourcesTags.tag_id, COUNT(DISTINCT resourcesTags.resources_id) AS resourceCount
		FROM resourcesTags
		INNER JOIN resources ON resources.id = resourcesTags.resources_id
		WHERE resources.status = 1
			AND resourcesTags.tag_id IN ?
		GROUP BY resourcesTags.tag_id
	`, tagIDs).Scan(&counts).Error; err != nil {
		return err
	}

	countMap := make(map[string]int64, len(counts))
	for _, item := range counts {
		countMap[item.TagID] = item.ResourceCount
	}
	for i := range dataList {
		dataList[i].ResourceCount = countMap[dataList[i].ID]
	}
	return nil
}

func (t Tag) InfoByID(db *gorm.DB, id string) (*Tag, error) {
	var tag Tag
	err := db.Where("id = ? ", id).First(&tag).Error
	return &tag, err
}

func (t Tag) InfoByName(db *gorm.DB, filesBasesID, name string) (*Tag, error) {
	var tag Tag
	err := db.Table(fmt.Sprintf("%s as t", t.TableName())).
		Joins(fmt.Sprintf("LEFT JOIN %s as tc ON t.tagClass_id = tc.id", TagClass{}.TableName())).
		Where("tc.filesBases_id = ? and t.name = ?", filesBasesID, name).First(&tag).Error
	return &tag, err
}

func (Tag) GetTotalByTagClassID(db *gorm.DB, tagClassID string) (int64, error) {
	var total int64
	err := db.Model(&Tag{}).Where("tagClass_id = ?", tagClassID).Count(&total).Error
	return total, err
}

func (Tag) Update(db *gorm.DB, tag *Tag, fields []string) error {
	result := db.Model(&tag).Select(fields).Updates(tag)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}

// 根据给定的ids数组，将匹配的记录hot值加1
func (Tag) UpdateHot(db *gorm.DB, ids []string) error {
	result := db.Model(&Tag{}).Where("id in (?)", ids).Update("hot", gorm.Expr("hot + ?", 1))
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}

func (Tag) Create(db *gorm.DB, tag *Tag) error {
	return db.Create(&tag).Error
}

func (Tag) DeleteById(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ? ", id).Delete(&Tag{}).Error
}
func (Tag) DeleteTagByTagClassSlc(db *gorm.DB, tagClassSlc []string) error {
	return db.Unscoped().Where("tagClass_id in (?) ", tagClassSlc).Delete(&Tag{}).Error
}
