package models

import "cm_collectors_server/datatype"

type TagClass struct {
	ID           string              `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID string              `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);index:idx_tagClass_filesBasesID;"`
	Name         string              `json:"name" gorm:"type:varchar(200);"`
	LeftShow     bool                `json:"leftShow" gorm:"column:leftShow;type:tinyint(1);default:1"`
	Hot          int                 `json:"hot" gorm:"type:int(4);default:0"`
	Sort         int                 `json:"sort" gorm:"type:int(4);default:0"`
	CreatedAt    datatype.CustomTime `json:"addTime" gorm:"column:addTime;type:datetime"`
	Status       bool                `json:"status" gorm:"type:tinyint(1);default:1"`
}

func (TagClass) TableName() string {
	return "tagClass"
}
