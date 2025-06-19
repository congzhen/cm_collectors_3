package models

type FilesRelatedPerformerBases struct {
	ID               string `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID     string `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);"`
	PerformerBasesID string `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);"`
	Main             bool   `json:"main" gorm:"column:main;type:tinyint(1);default:0"`
}

func (FilesRelatedPerformerBases) TableName() string {
	return "filesRelatedPerformerBases"
}
