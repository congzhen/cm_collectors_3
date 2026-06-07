package models

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

const AutoBackupStateID = "default"

type AutoBackupState struct {
	ID                         string               `json:"id" gorm:"primaryKey;type:varchar(32);"`
	LastSuccessBackupAt        *datatype.CustomTime `json:"last_success_backup_at" gorm:"column:last_success_backup_at;type:datetime"`
	PendingResourceChangeCount int                  `json:"pending_resource_change_count" gorm:"column:pending_resource_change_count;type:int;default:0"`
	LastResourceChangeAt       *datatype.CustomTime `json:"last_resource_change_at" gorm:"column:last_resource_change_at;type:datetime"`
	LastTimeCheckAt            *datatype.CustomTime `json:"last_time_check_at" gorm:"column:last_time_check_at;type:datetime"`
	Running                    bool                 `json:"running" gorm:"column:running;type:tinyint(1);default:0"`
	RunningStartedAt           *datatype.CustomTime `json:"running_started_at" gorm:"column:running_started_at;type:datetime"`
	LastBackupPath             string               `json:"last_backup_path" gorm:"column:last_backup_path;type:varchar(1000);"`
	LastBackupReason           string               `json:"last_backup_reason" gorm:"column:last_backup_reason;type:varchar(100);"`
	LastError                  string               `json:"last_error" gorm:"column:last_error;type:text;"`
	UpdatedAt                  *datatype.CustomTime `json:"updated_at" gorm:"column:updated_at;type:datetime"`
}

func (AutoBackupState) TableName() string {
	return "auto_backup_state"
}

func (AutoBackupState) Ensure(db *gorm.DB) (*AutoBackupState, error) {
	var state AutoBackupState
	err := db.First(&state, "id = ?", AutoBackupStateID).Error
	if err == nil {
		return &state, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	now := datatype.CustomTime{}
	state = AutoBackupState{
		ID:        AutoBackupStateID,
		UpdatedAt: &now,
	}
	now.SetValue(core.TimeNow())
	if err := db.Create(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}
