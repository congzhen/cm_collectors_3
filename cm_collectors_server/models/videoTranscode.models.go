package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

const (
	VideoTranscodeStatusDraft              = "draft"
	VideoTranscodeStatusQueued             = "queued"
	VideoTranscodeStatusProbing            = "probing"
	VideoTranscodeStatusTranscoding        = "transcoding"
	VideoTranscodeStatusVerifying          = "verifying"
	VideoTranscodeStatusReplacing          = "replacing"
	VideoTranscodeStatusRefreshingMetadata = "refreshing_metadata"
	VideoTranscodeStatusSuccess            = "success"
	VideoTranscodeStatusFailed             = "failed"
	VideoTranscodeStatusCancelled          = "cancelled"
	VideoTranscodeStatusInterrupted        = "interrupted"
	VideoTranscodeStatusRollbackFailed     = "rollback_failed"
)

// VideoTranscodeTask 是可恢复的视频转码队列记录。
// ConfigJsonData 保存创建任务时的完整参数快照，避免全局设置变化影响已创建任务。
type VideoTranscodeTask struct {
	ID                 string               `json:"id" gorm:"primaryKey;type:char(20)"`
	DramaSeriesID      string               `json:"dramaSeriesId" gorm:"column:drama_series_id;type:char(20);index:idx_video_transcode_drama_status,priority:1"`
	ResourceID         string               `json:"resourceId" gorm:"column:resource_id;type:char(20);index"`
	ResourceTitle      string               `json:"resourceTitle" gorm:"column:resource_title;type:text"`
	SourcePath         string               `json:"sourcePath" gorm:"column:source_path;type:text"`
	SourceSize         int64                `json:"sourceSize" gorm:"column:source_size;type:bigint"`
	SourceModifiedTime int64                `json:"sourceModifiedTime" gorm:"column:source_modified_time;type:bigint"`
	SourceDuration     float64              `json:"sourceDuration" gorm:"column:source_duration;type:real;default:0"`
	SourceWidth        int                  `json:"sourceWidth" gorm:"column:source_width;type:int;default:0"`
	SourceHeight       int                  `json:"sourceHeight" gorm:"column:source_height;type:int;default:0"`
	SourceFrameRate    float64              `json:"sourceFrameRate" gorm:"column:source_frame_rate;type:real;default:0"`
	SourceVideoCodec   string               `json:"sourceVideoCodec" gorm:"column:source_video_codec;type:varchar(50)"`
	SourceAudioCodec   string               `json:"sourceAudioCodec" gorm:"column:source_audio_codec;type:varchar(50)"`
	SourceVideoBitRate int64                `json:"sourceVideoBitRate" gorm:"column:source_video_bit_rate;type:bigint;default:0"`
	ConfigJsonData     string               `json:"configJsonData" gorm:"column:config_json_data;type:text"`
	Status             string               `json:"status" gorm:"type:varchar(30);index:idx_video_transcode_status_created,priority:1;index:idx_video_transcode_drama_status,priority:2"`
	Progress           float64              `json:"progress" gorm:"type:real;default:0"`
	ProcessedSeconds   float64              `json:"processedSeconds" gorm:"column:processed_seconds;type:real;default:0"`
	Speed              string               `json:"speed" gorm:"type:varchar(30)"`
	TemporaryPath      string               `json:"temporaryPath" gorm:"column:temporary_path;type:text"`
	OutputPath         string               `json:"outputPath" gorm:"column:output_path;type:text"`
	BackupPath         string               `json:"backupPath" gorm:"column:backup_path;type:text"`
	OutputSize         int64                `json:"outputSize" gorm:"column:output_size;type:bigint;default:0"`
	OutputDuration     float64              `json:"outputDuration" gorm:"column:output_duration;type:real;default:0"`
	OutputWidth        int                  `json:"outputWidth" gorm:"column:output_width;type:int;default:0"`
	OutputHeight       int                  `json:"outputHeight" gorm:"column:output_height;type:int;default:0"`
	OutputFrameRate    float64              `json:"outputFrameRate" gorm:"column:output_frame_rate;type:real;default:0"`
	OutputVideoCodec   string               `json:"outputVideoCodec" gorm:"column:output_video_codec;type:varchar(50)"`
	OutputAudioCodec   string               `json:"outputAudioCodec" gorm:"column:output_audio_codec;type:varchar(50)"`
	OutputVideoBitRate int64                `json:"outputVideoBitRate" gorm:"column:output_video_bit_rate;type:bigint;default:0"`
	ErrorMessage       string               `json:"errorMessage" gorm:"column:error_message;type:text"`
	WarningMessage     string               `json:"warningMessage" gorm:"column:warning_message;type:text"`
	CreatedAt          *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime;index:idx_video_transcode_status_created,priority:2"`
	StartedAt          *datatype.CustomTime `json:"startedAt" gorm:"column:started_at;type:datetime"`
	FinishedAt         *datatype.CustomTime `json:"finishedAt" gorm:"column:finished_at;type:datetime"`
	UpdatedAt          *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (VideoTranscodeTask) TableName() string { return "video_transcode_tasks" }

func (VideoTranscodeTask) List(db *gorm.DB) ([]VideoTranscodeTask, error) {
	var tasks []VideoTranscodeTask
	err := db.Order("created_at DESC, id DESC").Find(&tasks).Error
	return tasks, err
}

func (VideoTranscodeTask) Info(db *gorm.DB, id string) (*VideoTranscodeTask, error) {
	var task VideoTranscodeTask
	err := db.Where("id = ?", id).First(&task).Error
	return &task, err
}

func (VideoTranscodeTask) FirstQueued(db *gorm.DB) (*VideoTranscodeTask, error) {
	var task VideoTranscodeTask
	result := db.Where("status = ?", VideoTranscodeStatusQueued).
		Order("created_at DESC, id DESC").
		Limit(1).
		Find(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		// 队列空闲是工作线程轮询时的正常状态，不让 GORM 把它打印成查询错误。
		return nil, gorm.ErrRecordNotFound
	}
	return &task, nil
}

func (VideoTranscodeTask) HasActiveByDramaSeriesID(db *gorm.DB, dramaSeriesID string) (bool, error) {
	var count int64
	err := db.Model(&VideoTranscodeTask{}).
		Where("drama_series_id = ? AND status IN ?", dramaSeriesID, []string{
			VideoTranscodeStatusDraft,
			VideoTranscodeStatusQueued,
			VideoTranscodeStatusProbing,
			VideoTranscodeStatusTranscoding,
			VideoTranscodeStatusVerifying,
			VideoTranscodeStatusReplacing,
			VideoTranscodeStatusRefreshingMetadata,
		}).Count(&count).Error
	return count > 0, err
}
