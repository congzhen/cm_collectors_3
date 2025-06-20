package models

import "cm_collectors_server/datatype"

type Resources struct {
	ID                string               `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID      string               `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_resources_filesBasesID"`
	Title             string               `json:"title" gorm:"type:varchar(200);"`
	IssueNumber       string               `json:"issueNumber" gorm:"column:issueNumber;type:varchar(200);"`
	Mode              string               `json:"mode" gorm:"type:varchar(20);"`
	CoverPoster       string               `json:"coverPoster" gorm:"column:coverPoster;type:varchar(100);"`
	CoverPosterMode   int                  `json:"coverPosterMode" gorm:"column:coverPosterMode;type:int;"`
	CoverPosterWidth  int                  `json:"coverPosterWidth" gorm:"column:coverPosterWidth;type:int;"`
	CoverPosterHeight int                  `json:"coverPosterHeight" gorm:"column:coverPosterHeight;type:int;"`
	IssuingDate       *datatype.CustomDate `json:"issuingDate" gorm:"column:issuingDate;type:date;"`
	Country           string               `json:"country" gorm:"type:varchar(50);"`
	Definition        string               `json:"definition" gorm:"type:varchar(50);"`
	Stars             int                  `json:"stars" gorm:"type:int;"`
	Hot               int                  `json:"hot" gorm:"type:int;"`
	LastPlayTime      *datatype.CustomTime `json:"lastPlayTime" gorm:"column:lastPlayTime;type:datetime;"`
	LastPlayFile      string               `json:"lastPlayFile" gorm:"column:lastPlayFile;type:varchar(500);"`
	Abstract          string               `json:"abstract" gorm:"type:text;"`
	CreatedAt         datatype.CustomTime  `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status            bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (Resources) TableName() string {
	return "resources"
}
