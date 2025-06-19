package models

type ResourcesTags struct {
	ID          string `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID string `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_ResourcesTags_ResourcesID;index:idx_ResourcesTags_tagID,priority:2;"`
	TagID       string `json:"tag_id" gorm:"column:tag_id;type:char(20);index:idx_ResourcesTags_tagID,priority:1"`
	Sort        int    `json:"sort" gorm:"type:int(4);default:0"`
}

func (ResourcesTags) TableName() string {
	return "resourcesTags"
}
