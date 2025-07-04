package models

import (
	"cm_collectors_server/datatype"
	"sort"

	"gorm.io/gorm"
)

type Performer struct {
	ID               string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerBasesID string               `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);;index:idx_performer_performerBasesID;"`
	Name             string               `json:"name" gorm:"type:varchar(200);"`
	AliasName        string               `json:"aliasName" gorm:"column:aliasName;type:varchar(500);"`
	KeyWords         string               `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	Birthday         string               `json:"birthday" gorm:"column:birthday;type:varchar(10);"`
	Nationality      string               `json:"nationality" gorm:"type:varchar(200);"`
	CareerPerformer  bool                 `json:"careerPerformer" gorm:"column:careerPerformer;type:tinyint(1);default:1"`
	CareerDirector   bool                 `json:"careerDirector" gorm:"column:careerDirector;type:tinyint(1);default:0"`
	Photo            string               `json:"photo" gorm:"type:varchar(100);"`
	Introduction     string               `json:"introduction" gorm:"type:text;"`
	Cup              string               `json:"cup" gorm:"type:varchar(10);index:idx_performer_cup;"`
	Bust             string               `json:"bust" gorm:"type:varchar(10);"`
	Waist            string               `json:"waist" gorm:"type:varchar(10);"`
	Hip              string               `json:"hip" gorm:"type:varchar(10);"`
	Stars            int                  `json:"stars" gorm:"type:int;"`
	RetreatStatus    bool                 `json:"retreatStatus" gorm:"column:retreatStatus;type:tinyint(1);default:0"`
	CreatedAt        *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	Status           bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
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

func (Performer) BasicList_Performer(db *gorm.DB, performerBasesIds []string) (*[]PerformerBasic, error) {
	var list []PerformerBasic
	if len(performerBasesIds) > 0 {
		db = db.Where("performerBases_id in (?) and status = 1", performerBasesIds)
	}
	err := db.Model(&Performer{}).Where("careerPerformer = ?", true).Order("addTime desc").Find(&list).Error
	return &list, err
}

func (Performer) DataList(db *gorm.DB, performerBasesId string, fetchCount bool, page, limit int, search, star, cup string) (*[]Performer, int64, error) {
	var dataList []Performer
	var total int64
	offset := (page - 1) * limit
	query := db.Model(Performer{}).Where("performerBases_id = ? and status = 1", performerBasesId)
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
