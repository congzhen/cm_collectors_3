package models

type FilesBasesSetting struct {
	FilesBasesID   string `json:"filesBases_id" gorm:"column:filesBases_id;primaryKey;type:char(20);"`
	ConfigJsonData string `json:"config_json_data" gorm:"type:text;"`
	NfoJsonData    string `json:"nfo_json_data" gorm:"type:text;"`
	SimpleJsonData string `json:"simple_json_data" gorm:"type:text;"`
}

func (FilesBasesSetting) TableName() string {
	return "filesBasesSetting"
}
