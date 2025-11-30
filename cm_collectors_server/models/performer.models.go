package models

import (
	"cm_collectors_server/datatype"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Performer struct {
	ID                    string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerBasesID      string               `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);;index:idx_performer_performerBasesID;"`
	Name                  string               `json:"name" gorm:"type:varchar(200);"`
	AliasName             string               `json:"aliasName" gorm:"column:aliasName;type:varchar(500);"`
	KeyWords              string               `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	Birthday              string               `json:"birthday" gorm:"column:birthday;type:varchar(10);"`
	Nationality           string               `json:"nationality" gorm:"type:varchar(200);"`
	CareerPerformer       bool                 `json:"careerPerformer" gorm:"column:careerPerformer;type:tinyint(1);default:1"`
	CareerDirector        bool                 `json:"careerDirector" gorm:"column:careerDirector;type:tinyint(1);default:0"`
	Photo                 string               `json:"photo" gorm:"type:varchar(100);"`
	Introduction          string               `json:"introduction" gorm:"type:text;"`
	Cup                   string               `json:"cup" gorm:"type:varchar(10);index:idx_performer_cup;"`
	Bust                  string               `json:"bust" gorm:"type:varchar(10);"`
	Waist                 string               `json:"waist" gorm:"type:varchar(10);"`
	Hip                   string               `json:"hip" gorm:"type:varchar(10);"`
	Stars                 int                  `json:"stars" gorm:"type:int;"`
	RetreatStatus         bool                 `json:"retreatStatus" gorm:"column:retreatStatus;type:tinyint(1);default:0"`
	CreatedAt             *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	LastScraperUpdateTime *datatype.CustomDate `json:"lastScraperUpdateTime" gorm:"column:lastScraperUpdateTime;type:date;default:NULL"`
	Status                bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
}

type PerformerBasic struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AliasName string `json:"aliasName"  gorm:"column:aliasName;"`
	KeyWords  string `json:"keyWords" gorm:"column:keyWords;"`
}

func (Performer) TableName() string {
	return "performer"
}
func (PerformerBasic) TableName() string {
	return "performer"
}

func (Performer) BasicList(db *gorm.DB, performerBasesIds []string, careerPerformer, careerDirector bool) (*[]PerformerBasic, error) {
	var list []PerformerBasic
	db = db.Model(&Performer{})
	if len(performerBasesIds) > 0 {
		db = db.Where("performerBases_id in (?) and status = 1", performerBasesIds)
	}
	if careerPerformer {
		db = db.Where("careerPerformer = 1")
	}
	if careerDirector {
		db = db.Where("careerDirector = 1")
	}
	err := db.Order("addTime desc").Find(&list).Error
	return &list, err
}
func (Performer) PhotosByPerformerBasesId_DB(db *gorm.DB, performerBasesId string) ([]string, error) {
	var photos []string
	err := db.Model(&Performer{}).Where("performerBases_id = ?", performerBasesId).Pluck("photo", &photos).Error
	return photos, err
}

func (Performer) SearchLastScraperUpdateTime(db *gorm.DB, performerBasesId, lastScraperUpdateTime string) (*[]PerformerBasic, error) {
	var list []PerformerBasic
	db = db.Model(&Performer{}).Where("performerBases_id = ?", performerBasesId)
	if lastScraperUpdateTime != "" {
		//转换成日期
		lastData := datatype.CustomDate{}
		lastData.SetValue(lastScraperUpdateTime)
		lastDataValue, err := lastData.Value()
		if err == nil {
			db = db.Where("lastScraperUpdateTime < ?", lastDataValue)
		}
	} else {
		db = db.Where("lastScraperUpdateTime is null")
	}
	err := db.Order("addTime desc").Find(&list).Error
	return &list, err
}

func (Performer) InfoByName(db *gorm.DB, performerBasesID, name string, searchAliasName bool) (*Performer, error) {
	var performer Performer
	if searchAliasName {
		db = db.Where("performerBases_id = ? and (name = ? or aliasName like ?)", performerBasesID, name, "%"+name+"%")
	} else {
		db = db.Where("performerBases_id = ? and name = ?", performerBasesID, name)
	}

	err := db.First(&performer).Error
	return &performer, err
}
func (Performer) InfoByID(db *gorm.DB, id string) (*Performer, error) {
	var performer Performer
	err := db.Where("id = ?", id).First(&performer).Error
	return &performer, err
}

func (Performer) DataList(db *gorm.DB, performerBasesId string, fetchCount bool, page, limit int, search, star, cup, charIndex string) (*[]Performer, int64, error) {
	var dataList []Performer
	var total int64
	offset := (page - 1) * limit
	query := db.Model(Performer{}).Where("performerBases_id = ? and status = 1", performerBasesId)
	if charIndex != "" {
		charIndex = strings.ToLower(charIndex)
		if charIndex != "all" {
			query = query.Where("name like ? or keyWords like ?", charIndex+"%", charIndex+"%")
		}
	}
	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("keyWords like ? or name like ? or aliasName like ?", searchQuery, searchQuery, searchQuery)
	}
	if star != "" {
		query = query.Where("stars = ?", star)
	}
	if cup != "" && cup != "ALL" {
		if cup == "noCup" {
			query = query.Where("cup = ''")
		} else {
			query = query.Where("cup = ?", cup)
		}
	}
	if fetchCount {
		err := query.Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
	}
	query = query.Order("addTime desc").Limit(limit).Offset(offset)
	err := query.Find(&dataList).Error
	if err != nil {
		return nil, 0, err
	}
	return &dataList, total, err
}
func (Performer) DataListByIds(db *gorm.DB, ids []string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Model(Performer{}).Where("id in (?)", ids).Find(&dataList).Error
	return &dataList, err
}

func (t Performer) ListTopPreferredPerformers(db *gorm.DB, preferredIds []string, mainPerformerBasesId string, shieldNoPerformerPhoto bool, limit int) (*[]Performer, error) {
	var dataList []Performer

	dataListByIds, err := t.ListByIds(db, preferredIds)
	if err != nil {
		return nil, err
	}
	surplus := limit - len(*dataListByIds)

	if surplus > 0 {
		var surplusDataList []Performer
		query := db.Model(Performer{}).Where("performerBases_id = ? and status = 1", mainPerformerBasesId)
		if shieldNoPerformerPhoto {
			query = query.Where("photo != ''")
		}
		err := query.Order("addTime desc").Limit(surplus).Find(&surplusDataList).Error
		if err != nil {
			return nil, err
		}
		//组合数据
		dataList = append(*dataListByIds, surplusDataList...)
	} else {
		dataList = *dataListByIds
	}

	return &dataList, err
}
func (t Performer) RecycleBin(db *gorm.DB, performerBasesId string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Where("performerBases_id = ?", performerBasesId).Where("status = ?", false).Find(&dataList).Error
	return &dataList, err
}

func (Performer) ListByIds(db *gorm.DB, ids []string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Where("id in (?)", ids).Find(&dataList).Error
	// 构建 id 到索引的映射
	idIndexMap := make(map[string]int)
	for i, id := range ids {
		idIndexMap[id] = i
	}
	// 按 ids 的顺序排序
	sort.Slice(dataList, func(i, j int) bool {
		return idIndexMap[dataList[i].ID] < idIndexMap[dataList[j].ID]
	})
	return &dataList, err
}

func (Performer) Update(db *gorm.DB, performer *Performer, fields []string) error {
	result := db.Model(&performer).Select(fields).Updates(performer)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (Performer) Create(db *gorm.DB, performer *Performer) error {
	return db.Create(&performer).Error
}

func (Performer) DeleteByPerformerBasesIds(db *gorm.DB, performerBasesIds []string) error {
	return db.Unscoped().Where("performerBases_id in (?)", performerBasesIds).Delete(&Performer{}).Error
}
