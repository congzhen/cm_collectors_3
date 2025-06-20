package models

import (
	"cm_collectors_server/datatype"
)

type Performer struct {
	ID               string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerBasesID string               `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);;index:idx_performer_performerBasesID;"`
	Name             string               `json:"name" gorm:"type:varchar(200);"`
	AliasName        string               `json:"aliasName" gorm:"column:aliasName;type:varchar(500);"`
	Birthday         *datatype.CustomDate `json:"birthday" gorm:"type:date;"`
	Nationality      string               `json:"nationality" gorm:"type:varchar(200);"`
	CareerPerformer  bool                 `json:"careerPerformer" gorm:"column:careerPerformer;type:tinyint(1);default:1"`
	CareerDirector   bool                 `json:"careerDirector" gorm:"column:careerDirector;type:tinyint(1);default:0"`
	Photo            string               `json:"photo" gorm:"type:varchar(100);"`
	Introduction     string               `json:"introduction" gorm:"type:text;"`
	Cup              string               `json:"cup" gorm:"type:varchar(10);index:idx_performer_cup;"`
	Bust             int                  `json:"bust" gorm:"type:int;"`
	Waist            int                  `json:"waist" gorm:"type:int;"`
	Hip              int                  `json:"hip" gorm:"type:int;"`
	Stars            int                  `json:"stars" gorm:"type:int;"`
	RetreatStatus    bool                 `json:"retreatStatus" gorm:"column:retreatStatus;type:tinyint(1);default:0"`
	CreatedAt        datatype.CustomTime  `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status           bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (Performer) TableName() string {
	return "performer"
}
