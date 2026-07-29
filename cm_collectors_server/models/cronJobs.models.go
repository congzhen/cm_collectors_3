package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

type CronJobs struct {
	ID             string                  `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesId   string                  `json:"filesBases_id" gorm:"column:filesBases_id;type:char(20);"`
	FilesBases     FilesBases              `json:"filesBases" gorm:"foreignKey:FilesBasesId;references:ID;"`
	FilesBasesList []FilesBases            `json:"filesBasesList" gorm:"many2many:cron_jobs_files_bases;joinForeignKey:CronJobsID;joinReferences:FilesBasesID"`
	ScopeMode      string                  `json:"scopeMode" gorm:"column:scope_mode;type:varchar(20);default:selected"`
	ConfigJsonData string                  `json:"configJsonData" gorm:"column:config_json_data;type:text;"`
	JobsType       datatype.E_cronJobsType `json:"jobs_type" gorm:"type:varchar(100);"`
	CronExpression string                  `json:"cron_expression" gorm:"type:varchar(100);"`
	CreatedAt      *datatype.CustomTime    `json:"created_at" gorm:"column:created_at;type:datetime"`
	LastExecAt     *datatype.CustomTime    `json:"last_exec_at" gorm:"column:last_exec_at;type:datetime"`
	LastExecError  string                  `json:"last_exec_error" gorm:"type:varchar(255);"`
	LastExecStatus bool                    `json:"last_exec_status" gorm:"type:tinyint(1);"`
	Status         bool                    `json:"status" gorm:"type:tinyint(1);"`
	Running        bool                    `json:"running" gorm:"-"`
}

func (CronJobs) TableName() string {
	return "cronJobs"
}

func (CronJobs) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("FilesBases").Preload("FilesBasesList")
}

type CronJobsFilesBases struct {
	CronJobsID   string `json:"cronJobsId" gorm:"column:cron_jobs_id;primaryKey;type:char(20);"`
	FilesBasesID string `json:"filesBasesId" gorm:"column:files_bases_id;primaryKey;type:char(20);"`
}

func (CronJobsFilesBases) TableName() string { return "cron_jobs_files_bases" }

func (CronJobsFilesBases) Replace(db *gorm.DB, cronJobsID string, filesBasesIDs []string) error {
	if err := db.Unscoped().Where("cron_jobs_id = ?", cronJobsID).Delete(&CronJobsFilesBases{}).Error; err != nil {
		return err
	}
	for _, id := range filesBasesIDs {
		if id == "" {
			continue
		}
		if err := db.Create(&CronJobsFilesBases{CronJobsID: cronJobsID, FilesBasesID: id}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (t CronJobs) DataList(db *gorm.DB) (*[]CronJobs, error) {
	var cronJobs []CronJobs
	err := t.Preload(db).Model(&CronJobs{}).Order("id desc").Find(&cronJobs).Error
	return &cronJobs, err
}

func (t CronJobs) Info(db *gorm.DB, id string) (*CronJobs, error) {
	var info CronJobs
	err := t.Preload(db).First(&info, "id = ?", id).Error
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
	if err := db.Unscoped().Where("cron_jobs_id = ?", id).Delete(&CronJobsFilesBases{}).Error; err != nil {
		return err
	}
	return db.Unscoped().Where("id = ? ", id).Delete(&CronJobs{}).Error
}
func (CronJobs) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	var jobIDs []string
	if err := db.Model(&CronJobsFilesBases{}).
		Where("files_bases_id = ?", filesBasesID).
		Pluck("cron_jobs_id", &jobIDs).Error; err != nil {
		return err
	}
	if err := db.Unscoped().Where("files_bases_id = ?", filesBasesID).Delete(&CronJobsFilesBases{}).Error; err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		var remaining []string
		if err := db.Model(&CronJobsFilesBases{}).Where("cron_jobs_id = ?", jobID).
			Order("files_bases_id").Pluck("files_bases_id", &remaining).Error; err != nil {
			return err
		}
		if len(remaining) == 0 {
			if err := db.Unscoped().Where("id = ? AND scope_mode = ?", jobID, VideoMetadataScopeSelected).
				Delete(&CronJobs{}).Error; err != nil {
				return err
			}
			continue
		}
		if err := db.Model(&CronJobs{}).Where("id = ?", jobID).Update("filesBases_id", remaining[0]).Error; err != nil {
			return err
		}
	}
	return nil
}
