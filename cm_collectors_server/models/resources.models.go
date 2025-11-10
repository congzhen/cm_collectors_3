package models

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Resources struct {
	ID                    string                  `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID          string                  `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_resources_filesBasesID"`
	Title                 string                  `json:"title" gorm:"type:varchar(200);"`
	KeyWords              string                  `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	IssueNumber           string                  `json:"issueNumber" gorm:"column:issueNumber;type:varchar(200);"`
	Mode                  datatype.E_resourceMode `json:"mode" gorm:"type:varchar(20);"`
	CoverPoster           string                  `json:"coverPoster" gorm:"column:coverPoster;type:varchar(100);"`
	CoverPosterMode       int                     `json:"coverPosterMode" gorm:"column:coverPosterMode;type:int;default:0"`
	CoverPosterWidth      int                     `json:"coverPosterWidth" gorm:"column:coverPosterWidth;type:int;"`
	CoverPosterHeight     int                     `json:"coverPosterHeight" gorm:"column:coverPosterHeight;type:int;"`
	IssuingDate           *datatype.CustomDate    `json:"issuingDate" gorm:"column:issuingDate;type:date;"`
	Country               string                  `json:"country" gorm:"type:varchar(50);"`
	Definition            string                  `json:"definition" gorm:"type:varchar(50);"`
	Stars                 int                     `json:"stars" gorm:"type:int;"`
	Score                 float64                 `json:"score" gorm:"type:float64;"`
	Hot                   int                     `json:"hot" gorm:"type:int;"`
	LastPlayTime          *datatype.CustomTime    `json:"lastPlayTime" gorm:"column:lastPlayTime;type:datetime;"`
	LastPlayFile          string                  `json:"lastPlayFile" gorm:"column:lastPlayFile;type:varchar(500);"`
	Abstract              string                  `json:"abstract" gorm:"type:text;"`
	LastScraperUpdateTime *datatype.CustomDate    `json:"lastScraperUpdateTime" gorm:"column:lastScraperUpdateTime;type:date;default:NULL"`
	CreatedAt             *datatype.CustomTime    `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status                bool                    `json:"status" gorm:"type:tinyint(1);default:1"`
	Tags                  []Tag                   `json:"tags" gorm:"many2many:resourcesTags;joinForeignKey:ResourcesID;joinReferences:TagID"`
	Performers            []Performer             `json:"performers" gorm:"many2many:resourcesPerformers;joinForeignKey:ResourcesID;joinReferences:PerformerID"`
	Directors             []Performer             `json:"directors" gorm:"many2many:resourcesDirectors;joinForeignKey:ResourcesID;joinReferences:DirectorID"`
	ResourcesDramaSeries  []ResourcesDramaSeries  `json:"dramaSeries" gorm:"foreignKey:ResourcesID;references:ID"`
}

func (Resources) TableName() string {
	return "resources"
}

func (Resources) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("Performers").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Joins("LEFT JOIN tagClass ON tag.tagClass_id = tagClass.id").
				Order("tagClass.sort asc, tag.sort asc")
		}).
		Preload("ResourcesDramaSeries", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		})
}

func (t Resources) Info(db *gorm.DB, id string) (*Resources, error) {
	var info Resources
	err := t.Preload(db).First(&info, "id = ?", id).Error
	if info.Tags == nil {
		info.Tags = []Tag{}
	}
	if info.Performers == nil {
		info.Performers = []Performer{}
	}
	if info.Directors == nil {
		info.Directors = []Performer{}
	}
	return &info, err
}

func (t Resources) DataListAll(db *gorm.DB, page, limit int) (*[]Resources, error) {
	offset := (page - 1) * limit
	var dataList []Resources
	err := t.Preload(db).Model(&Resources{}).Order("addTime desc").Limit(limit).Offset(offset).Find(&dataList)
	return &dataList, err.Error
}
func (t Resources) DataListByIds(db *gorm.DB, ids []string) (*[]Resources, error) {
	var dataList []Resources
	err := t.Preload(db).Model(&Resources{}).Where("id in (?)", ids).Order("addTime desc").Find(&dataList).Error
	return &dataList, err
}

func (t Resources) DataList(db *gorm.DB, par *datatype.ReqParam_ResourcesList) (*[]Resources, int64, error) {
	var dataList []Resources
	var total int64
	offset := (par.Page - 1) * par.Limit
	query := t.Preload(db).Model(Resources{}).Where("filesBases_id = ?", par.FilesBasesId)
	query = t.setDbSearchData(query, &par.SearchData)
	if par.FetchCount {
		err := query.Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
	}
	query = t.setDbSearchDataOrder(query, par.SearchData.Sort).Limit(par.Limit).Offset(offset)
	err := query.Find(&dataList).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range dataList {
		if dataList[i].Tags == nil {
			dataList[i].Tags = []Tag{}
		}
		if dataList[i].Performers == nil {
			dataList[i].Performers = []Performer{}
		}
		if dataList[i].Directors == nil {
			dataList[i].Directors = []Performer{}
		}
	}
	return &dataList, total, err
}

func (t Resources) setDbSearchData(db *gorm.DB, searchData *datatype.ReqParam_SearchData) *gorm.DB {
	if len(searchData.SearchTextSlc) > 0 {
		// 使用参数化查询避免SQL注入风险
		var orConditions []string
		var params []interface{}

		for _, text := range searchData.SearchTextSlc {
			searchLike := "%" + text + "%"
			orConditions = append(orConditions, "(title LIKE ? OR issueNumber LIKE ?)")
			params = append(params, searchLike, searchLike)
		}

		// 将所有OR条件组合成一个整体条件，但仍使用参数化查询
		if len(orConditions) > 0 {
			condition := strings.Join(orConditions, " OR ")
			db = db.Where(condition, params...)
		}
	}
	db = t.setDbSearchGroup(db, "country", &searchData.Country)
	db = t.setDbSearchGroup(db, "definition", &searchData.Definition)
	db = t.setDbSearchGroup(db, "stars", &searchData.Star)
	db = t.setDbSearchYear(db, &searchData.Year)
	db = t.setDbSearchPerformer(db, &searchData.Performer)
	db = t.setDbSearchCup(db, &searchData.Cup)
	db = t.setDbSearchTags(db, &searchData.Tag)
	return db
}
func (Resources) setDbSearchGroup(db *gorm.DB, field string, searchGroup *datatype.I_searchGroup) *gorm.DB {
	if len(searchGroup.Options) == 0 {
		return db
	}
	switch searchGroup.Logic {
	case datatype.E_searchLogic_single:
		// 单个等于查询
		db = db.Where(field+" = ?", searchGroup.Options[0])
	case datatype.E_searchLogic_multiOr:
		// 多个 OR 查询
		subQuery := core.DBS().Where(field+" = ?", searchGroup.Options[0])
		for _, option := range searchGroup.Options[1:] {
			subQuery.Or(field+" = ?", option)
		}
		db = db.Where(subQuery)

	case datatype.E_searchLogic_multiAnd:
		// 多个 AND 查询（例如 tag=1 AND tag=2）
		for _, option := range searchGroup.Options {
			db = db.Where(field+" = ?", option)
		}

	case datatype.E_searchLogic_not:
		// NOT IN 查询
		db = db.Where(field+" NOT IN (?)", searchGroup.Options)
	}

	return db
}

func (Resources) setDbSearchYear(db *gorm.DB, searchYearGroup *datatype.I_searchGroup) *gorm.DB {
	if len(searchYearGroup.Options) == 0 {
		return db
	}

	var conditions []string
	var params []interface{}

	for _, option := range searchYearGroup.Options {
		if option == datatype.V_Search_Before2000 {
			conditions = append(conditions, "issuingDate < ?")
			params = append(params, "2000-01-01")
		} else if year, err := strconv.Atoi(option); err == nil {
			startDate := fmt.Sprintf("%d-01-01", year)
			endDate := fmt.Sprintf("%d-12-31", year)
			conditions = append(conditions, "issuingDate BETWEEN ? AND ?")
			params = append(params, startDate, endDate)
		}
	}

	if len(conditions) > 0 {
		// 构建 OR 连接的条件
		query := "(" + strings.Join(conditions, " OR ") + ")"
		db = db.Where(query, params...)
	}

	return db
}

func (Resources) setDbSearchPerformer(db *gorm.DB, searchPerformerGroup *datatype.I_searchGroup) *gorm.DB {
	if len(searchPerformerGroup.Options) == 0 {
		return db
	}
	switch searchPerformerGroup.Logic {
	case datatype.E_searchLogic_single:
		// 单个等于查询
		if searchPerformerGroup.Options[0] == datatype.V_Search_Not {
			db = db.Where("(select count(*) from resourcesPerformers where resources_id = resources.id) = 0")
		} else {
			db = db.Where("id in (select resources_id from resourcesPerformers where performer_id = ?)", searchPerformerGroup.Options[0])
		}
	case datatype.E_searchLogic_multiOr:
		// 多个 OR 查询
		// 构建 IN 查询，实现 OR 逻辑
		db = db.Where("id IN (SELECT resources_id FROM resourcesPerformers WHERE performer_id IN (?))", searchPerformerGroup.Options)

	case datatype.E_searchLogic_multiAnd:
		// 多个 AND 查询（例如 tag=1 AND tag=2）
		// 构建多个 EXISTS 子句，每个对应一个 performer_id
		for _, option := range searchPerformerGroup.Options {
			db = db.Where("EXISTS (SELECT 1 FROM resourcesPerformers WHERE performer_id = ? AND resources_id = resources.id)", option)
		}
	case datatype.E_searchLogic_not:
		// NOT IN 查询
		// NOT IN 查询，排除具有特定 performer_id 的资源
		db = db.Where("NOT EXISTS (SELECT 1 FROM resourcesPerformers WHERE performer_id IN (?) AND resources_id = resources.id)", searchPerformerGroup.Options)
	}
	return db
}

func (Resources) setDbSearchCup(db *gorm.DB, searchCupGroup *datatype.I_searchGroup) *gorm.DB {
	if len(searchCupGroup.Options) == 0 {
		return db
	}
	switch searchCupGroup.Logic {
	case datatype.E_searchLogic_single:
		db = db.Where("id in (select resources_id from resourcesPerformers where performer_id in (select id from performer where cup = ?))", searchCupGroup.Options[0])
	case datatype.E_searchLogic_multiOr:
		db = db.Where("id in (select resources_id from resourcesPerformers where performer_id in (select id from performer where cup in (?)))", searchCupGroup.Options)
	case datatype.E_searchLogic_multiAnd:
		for _, option := range searchCupGroup.Options {
			db = db.Where("EXISTS (SELECT 1 FROM resourcesPerformers WHERE performer_id in (select id from performer where cup = ?) AND resources_id = resources.id)", option)
		}
	case datatype.E_searchLogic_not:
		db = db.Where("NOT EXISTS (SELECT 1 FROM resourcesPerformers WHERE performer_id in (select id from performer where cup in (?)) AND resources_id = resources.id)", searchCupGroup.Options)
	}
	return db
}

func (Resources) setDbSearchTags(db *gorm.DB, tagGroupMap *map[string]datatype.I_searchGroup) *gorm.DB {
	for _, tagGroup := range *tagGroupMap {
		if len(tagGroup.Options) == 0 {
			continue
		}
		switch tagGroup.Logic {
		case datatype.E_searchLogic_single:
			db = db.Where("id in (select resources_id from resourcesTags where tag_id = ?)", tagGroup.Options[0])
		case datatype.E_searchLogic_multiOr:
			db = db.Where("id in (select resources_id from resourcesTags where tag_id in (?))", tagGroup.Options)
		case datatype.E_searchLogic_multiAnd:
			for _, option := range tagGroup.Options {
				db = db.Where("EXISTS (SELECT 1 FROM resourcesTags WHERE tag_id = ? AND resources_id = resources.id)", option)
			}
		case datatype.E_searchLogic_not:
			db = db.Where("NOT EXISTS (SELECT 1 FROM resourcesTags WHERE tag_id in (?) AND resources_id = resources.id)", tagGroup.Options)
		}
	}
	return db
}

func (Resources) setDbSearchDataOrder(db *gorm.DB, searchSort datatype.E_searchSort) *gorm.DB {
	switch searchSort {
	case datatype.E_searchSort_addTimeAsc:
		db = db.Order("addTime ASC")
	case datatype.E_searchSort_addTimeDesc:
		db = db.Order("addTime DESC")
	case datatype.E_searchSort_issueNumberAsc:
		db = db.Order("issueNumber ASC,addTime DESC")
	case datatype.E_searchSort_issueNumberDesc:
		db = db.Order("issueNumber DESC,addTime DESC")
	case datatype.E_searchSort_scoreAsc:
		db = db.Order("score ASC,addTime DESC")
	case datatype.E_searchSort_scoreDesc:
		db = db.Order("score DESC,addTime DESC")
	case datatype.E_searchSort_starAsc:
		db = db.Order("stars ASC,addTime DESC")
	case datatype.E_searchSort_starDesc:
		db = db.Order("stars DESC,addTime DESC")
	case datatype.E_searchSort_issuingDateAsc:
		db = db.Order("issuingDate ASC,addTime DESC")
	case datatype.E_searchSort_issuingDateDesc:
		db = db.Order("issuingDate DESC,addTime DESC")
	case datatype.E_searchSort_titleAsc:
		db = db.Order("title ASC,addTime DESC")
	case datatype.E_searchSort_titleDesc:
		db = db.Order("title DESC,addTime DESC")
	case datatype.E_searchSort_history:
		db = db.Order("lastPlayTime DESC,addTime DESC")
	case datatype.E_searchSort_hot:
		db = db.Order("hot DESC,addTime DESC")
	default:
		db = db.Order("addTime DESC")
	}
	return db
}

func (Resources) Update(db *gorm.DB, resources *Resources, fields []string) error {
	result := db.Model(&resources).Select(fields).Updates(resources)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (Resources) Create(db *gorm.DB, resources *Resources) error {
	return db.Create(&resources).Error
}

func (Resources) DeleteById(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ? ", id).Delete(&Resources{}).Error
}
