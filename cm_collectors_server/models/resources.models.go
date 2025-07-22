package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type Resources struct {
	ID                   string                  `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID         string                  `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_resources_filesBasesID"`
	Title                string                  `json:"title" gorm:"type:varchar(200);"`
	KeyWords             string                  `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	IssueNumber          string                  `json:"issueNumber" gorm:"column:issueNumber;type:varchar(200);"`
	Mode                 datatype.E_resourceMode `json:"mode" gorm:"type:varchar(20);"`
	CoverPoster          string                  `json:"coverPoster" gorm:"column:coverPoster;type:varchar(100);"`
	CoverPosterMode      int                     `json:"coverPosterMode" gorm:"column:coverPosterMode;type:int;default:0"`
	CoverPosterWidth     int                     `json:"coverPosterWidth" gorm:"column:coverPosterWidth;type:int;"`
	CoverPosterHeight    int                     `json:"coverPosterHeight" gorm:"column:coverPosterHeight;type:int;"`
	IssuingDate          *datatype.CustomDate    `json:"issuingDate" gorm:"column:issuingDate;type:date;"`
	Country              string                  `json:"country" gorm:"type:varchar(50);"`
	Definition           string                  `json:"definition" gorm:"type:varchar(50);"`
	Stars                int                     `json:"stars" gorm:"type:int;"`
	Hot                  int                     `json:"hot" gorm:"type:int;"`
	LastPlayTime         *datatype.CustomTime    `json:"lastPlayTime" gorm:"column:lastPlayTime;type:datetime;"`
	LastPlayFile         string                  `json:"lastPlayFile" gorm:"column:lastPlayFile;type:varchar(500);"`
	Abstract             string                  `json:"abstract" gorm:"type:text;"`
	CreatedAt            *datatype.CustomTime    `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status               bool                    `json:"status" gorm:"type:tinyint(1);default:1"`
	Tags                 []Tag                   `json:"tags" gorm:"many2many:resourcesTags;joinForeignKey:ResourcesID;joinReferences:TagID"`
	Performers           []Performer             `json:"performers" gorm:"many2many:resourcesPerformers;joinForeignKey:ResourcesID;joinReferences:PerformerID"`
	Directors            []Performer             `json:"directors" gorm:"many2many:resourcesDirectors;joinForeignKey:ResourcesID;joinReferences:DirectorID"`
	ResourcesDramaSeries []ResourcesDramaSeries  `json:"dramaSeries" gorm:"foreignKey:ResourcesID;references:ID"`
}

func (Resources) TableName() string {
	return "resources"
}

func (Resources) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("Performers").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
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
		// 构建第一个 LIKE 条件
		db = db.Where("title LIKE ?", "%"+searchData.SearchTextSlc[0]+"%")

		// 剩下的条件使用 Or 追加
		for _, text := range searchData.SearchTextSlc[1:] {
			db = db.Or("title LIKE ?", "%"+text+"%")
		}
	}
	db = t.setDbSearchGroup(db, "country", &searchData.Country)
	db = t.setDbSearchGroup(db, "definition", &searchData.Definition)
	db = t.setDbSearchGroup(db, "year", &searchData.Year)
	db = t.setDbSearchGroup(db, "stars", &searchData.Star)

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
		db = db.Where(field+" = ?", searchGroup.Options[0])
		for _, option := range searchGroup.Options[1:] {
			db = db.Or(field+" = ?", option)
		}

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
func (Resources) setDbSearchDataOrder(db *gorm.DB, searchSort datatype.E_searchSort) *gorm.DB {
	switch searchSort {
	case datatype.E_searchSort_addTimeAsc:
		db = db.Order("addTime ASC")
	case datatype.E_searchSort_addTimeDesc:
		db = db.Order("addTime DESC")
	case datatype.E_searchSort_issueNumberAsc:
		db = db.Order("issueNumber ASC")
	case datatype.E_searchSort_issueNumberDesc:
		db = db.Order("issueNumber DESC")
	case datatype.E_searchSort_starAsc:
		db = db.Order("stars ASC")
	case datatype.E_searchSort_starDesc:
		db = db.Order("stars DESC")
	case datatype.E_searchSort_titleAsc:
		db = db.Order("title ASC")
	case datatype.E_searchSort_titleDesc:
		db = db.Order("title DESC")
	case datatype.E_searchSort_history:
		db = db.Order("lastPlayTime DESC")
	case datatype.E_searchSort_hot:
		db = db.Order("hot DESC")
	case datatype.E_searchSort_youLike:
		db = db.Order("youLike DESC")
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
