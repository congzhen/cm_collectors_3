package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type CronJobs struct {
	ID             string                  `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesId   string                  `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);"`
	FilesBases     FilesBases              `json:"filesBases" gorm:"foreignKey:FilesBasesId;references:ID;"`
	JobsType       datatype.E_cronJobsType `json:"jobs_type" gorm:"type:varchar(100);"`
	CronExpression string                  `json:"cron_expression" gorm:"type:varchar(100);"`
	CreatedAt      *datatype.CustomTime    `json:"created_at" gorm:"column:created_at;type:datetime"`
	LastExecAt     *datatype.CustomTime    `json:"last_exec_at" gorm:"column:last_exec_at;type:datetime"`
	LastExecError  string                  `json:"last_exec_error" gorm:"type:varchar(255);"`
	LastExecStatus bool                    `json:"last_exec_status" gorm:"type:tinyint(1);"`
	Status         bool                    `json:"status" gorm:"type:tinyint(1);"`
}

func (CronJobs) TableName() string {
	return "cronJobs"
}

func (CronJobs) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("FilesBases")
}

func (t CronJobs) DataList(db *gorm.DB) (*[]CronJobs, error) {
	var cronJobs []CronJobs
	err := t.Preload(db).Model(&CronJobs{}).Order("id desc").Find(&cronJobs).Error
	return &cronJobs, err
}

func (t CronJobs) Info(db *gorm.DB, id string) (*CronJobs, error) {
	var info CronJobs
	err := db.First(&info, "id = ?", id).Error
	return &info, err
}

func (CronJobs) Update(db *gorm.DB, cronJobs *CronJobs, fields []string) error {
	result := db.Model(&cronJobs).Select(fields).Updates(cronJobs)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (CronJobs) Create(db *gorm.DB, cronJobs *CronJobs) error {
	return db.Create(&cronJobs).Error
}
func (CronJobs) DeleteById(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ? ", id).Delete(&CronJobs{}).Error
}
func (CronJobs) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Unscoped().Where("filesBases_id = ? ", filesBasesID).Delete(&CronJobs{}).Error
}
