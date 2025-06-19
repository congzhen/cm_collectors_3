package models

import (
	"cm_collectors_server/datatype"
)

type PerformerBases struct {
	ID        string              `json:"id" gorm:"primaryKey;type:char(20);"`
	Name      string              `json:"name" gorm:"type:varchar(200);"`
	Sort      int                 `json:"sort" gorm:"type:int(4);"`
	CreatedAt datatype.CustomTime `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status    bool                `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (PerformerBases) TableName() string {
	return "performerBases"
}
