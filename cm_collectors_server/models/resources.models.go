package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type Resources struct {
	ID                   string                 `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID         string                 `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_resources_filesBasesID"`
	Title                string                 `json:"title" gorm:"type:varchar(200);"`
	KeyWords             string                 `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	IssueNumber          string                 `json:"issueNumber" gorm:"column:issueNumber;type:varchar(200);"`
	Mode                 string                 `json:"mode" gorm:"type:varchar(20);"`
	CoverPoster          string                 `json:"coverPoster" gorm:"column:coverPoster;type:varchar(100);"`
	CoverPosterMode      string                 `json:"coverPosterMode" gorm:"column:coverPosterMode;type:varchar(50);"`
	CoverPosterWidth     int                    `json:"coverPosterWidth" gorm:"column:coverPosterWidth;type:int;"`
	CoverPosterHeight    int                    `json:"coverPosterHeight" gorm:"column:coverPosterHeight;type:int;"`
	IssuingDate          *datatype.CustomDate   `json:"issuingDate" gorm:"column:issuingDate;type:date;"`
	Country              string                 `json:"country" gorm:"type:varchar(50);"`
	Definition           string                 `json:"definition" gorm:"type:varchar(50);"`
	Stars                int                    `json:"stars" gorm:"type:int;"`
	Hot                  int                    `json:"hot" gorm:"type:int;"`
	LastPlayTime         *datatype.CustomTime   `json:"lastPlayTime" gorm:"column:lastPlayTime;type:datetime;"`
	LastPlayFile         string                 `json:"lastPlayFile" gorm:"column:lastPlayFile;type:varchar(500);"`
	Abstract             string                 `json:"abstract" gorm:"type:text;"`
	CreatedAt            datatype.CustomTime    `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status               bool                   `json:"status" gorm:"type:tinyint(1);default:1"`
	Tags                 []Tag                  `json:"tags" gorm:"many2many:resourcesTags;joinForeignKey:ResourcesID;joinReferences:TagID"`
	Performers           []Performer            `json:"performers" gorm:"many2many:resourcesPerformers;joinForeignKey:ResourcesID;joinReferences:PerformerID"`
	Directors            []Performer            `json:"directors" gorm:"many2many:resourcesDirectors;joinForeignKey:ResourcesID;joinReferences:DirectorID"`
	ResourcesDramaSeries []ResourcesDramaSeries `json:"dramaSeries" gorm:"foreignKey:ResourcesID;references:ID"`
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

func (t Resources) DataList(db *gorm.DB, par *datatype.ReqParam_ResourcesList) (*[]Resources, int64, error) {
	var dataList []Resources
	var total int64
	offset := (par.Page - 1) * par.Limit
	query := t.Preload(db).Model(Resources{}).Where("filesBases_id = ?", par.FilesBasesId)
	if par.FetchCount {
		err := query.Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
	}
	query = query.Order("addTime desc").Limit(par.Limit).Offset(offset)
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
