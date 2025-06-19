package models

type ResourcesDirectors struct {
	ID          string `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID string `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_resources_directors_resourcesID,priority:1;"`
	PerformerID string `json:"performer_id" gorm:"column:performer_id;type:char(20);index:idx_resources_directors_performerID;index:idx_resources_directors_resourcesID,priority:2"`
	Sort        int    `json:"sort" gorm:"type:int(4);default:0"`
}

func (ResourcesDirectors) TableName() string {
	return "resourcesDirectors"
}
