package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	VideoMetadataStatusProcessing = "processing"
	VideoMetadataStatusSuccess    = "success"
	VideoMetadataStatusFailed     = "failed"
	VideoMetadataStatusStale      = "stale"

	VideoMetadataScopeSelected = "selected"
	VideoMetadataScopeAll      = "all"

	VideoMetadataBatchStatusRunning   = "running"
	VideoMetadataBatchStatusPaused    = "paused"
	VideoMetadataBatchStatusStopped   = "stopped"
	VideoMetadataBatchStatusCompleted = "completed"
)

// ResourcesVideoMetadata 保存一个视频分集的可重新生成元数据。
// 时长继续保存在 resourcesDramaSeries.durationSeconds，避免破坏已有接口和旧数据库。
type ResourcesVideoMetadata struct {
	DramaSeriesID    string               `json:"dramaSeriesId" gorm:"column:drama_series_id;primaryKey;type:char(20);"`
	MetadataVersion  int                  `json:"metadataVersion" gorm:"column:metadata_version;type:int;default:0;index:idx_video_metadata_version"`
	ProbeStatus      string               `json:"probeStatus" gorm:"column:probe_status;type:varchar(20);index:idx_video_metadata_status"`
	ProbeTime        *datatype.CustomTime `json:"probeTime" gorm:"column:probe_time;type:datetime;"`
	NextRetryTime    *datatype.CustomTime `json:"nextRetryTime" gorm:"column:next_retry_time;type:datetime;"`
	RetryCount       int                  `json:"retryCount" gorm:"column:retry_count;type:int;default:0"`
	ErrorCode        string               `json:"errorCode" gorm:"column:error_code;type:varchar(50);"`
	ErrorMessage     string               `json:"errorMessage" gorm:"column:error_message;type:text;"`
	FileSize         int64                `json:"fileSize" gorm:"column:file_size;type:bigint;default:0"`
	FileModifiedTime int64                `json:"fileModifiedTime" gorm:"column:file_modified_time;type:bigint;default:0"`
	Width            int                  `json:"width" gorm:"type:int;default:0;index:idx_video_metadata_height_codec,priority:2"`
	Height           int                  `json:"height" gorm:"type:int;default:0;index:idx_video_metadata_height_codec,priority:1"`
	FrameRate        float64              `json:"frameRate" gorm:"column:frame_rate;type:real;default:0"`
	FrameRateRaw     string               `json:"frameRateRaw" gorm:"column:frame_rate_raw;type:varchar(30);"`
	VideoCodec       string               `json:"videoCodec" gorm:"column:video_codec;type:varchar(50);index:idx_video_metadata_height_codec,priority:3"`
	VideoProfile     string               `json:"videoProfile" gorm:"column:video_profile;type:varchar(100);"`
	PixelFormat      string               `json:"pixelFormat" gorm:"column:pixel_format;type:varchar(50);"`
	BitDepth         int                  `json:"bitDepth" gorm:"column:bit_depth;type:int;default:0"`
	VideoBitRate     int64                `json:"videoBitRate" gorm:"column:video_bit_rate;type:bigint;default:0"`
	ContainerFormat  string               `json:"containerFormat" gorm:"column:container_format;type:varchar(100);"`
	AudioCodec       string               `json:"audioCodec" gorm:"column:audio_codec;type:varchar(50);"`
	AudioChannels    int                  `json:"audioChannels" gorm:"column:audio_channels;type:int;default:0"`
	AudioSampleRate  int                  `json:"audioSampleRate" gorm:"column:audio_sample_rate;type:int;default:0"`
	CreatedAt        *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt        *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (ResourcesVideoMetadata) TableName() string { return "resources_video_metadata" }

func (ResourcesVideoMetadata) Get(db *gorm.DB, dramaSeriesID string) (*ResourcesVideoMetadata, error) {
	var item ResourcesVideoMetadata
	err := db.Where("drama_series_id = ?", dramaSeriesID).First(&item).Error
	return &item, err
}

func (ResourcesVideoMetadata) Upsert(db *gorm.DB, item *ResourcesVideoMetadata) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "drama_series_id"}},
		UpdateAll: true,
	}).Create(item).Error
}

func (ResourcesVideoMetadata) DeleteByDramaSeriesIDs(db *gorm.DB, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return db.Unscoped().Where("drama_series_id IN ?", ids).Delete(&ResourcesVideoMetadata{}).Error
}

func (ResourcesVideoMetadata) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Exec(`DELETE FROM resources_video_metadata
		WHERE drama_series_id IN (
			SELECT ds.id FROM resourcesDramaSeries ds
			JOIN resources r ON r.id = ds.resources_id
			WHERE r.filesBases_id = ?
		)`, filesBasesID).Error
}

// VideoMetadataSetting 保存全局采集触发和空闲补齐策略。
type VideoMetadataSetting struct {
	ID                        string               `json:"id" gorm:"primaryKey;type:char(20);"`
	CollectOnNewOrChanged     bool                 `json:"collectOnNewOrChanged" gorm:"column:collect_on_new_or_changed;type:tinyint(1);default:1"`
	CollectOnDetailOrPlay     bool                 `json:"collectOnDetailOrPlay" gorm:"column:collect_on_detail_or_play;type:tinyint(1);default:1"`
	CollectOnList             bool                 `json:"collectOnList" gorm:"column:collect_on_list;type:tinyint(1);default:0"`
	IdleBackfillEnabled       bool                 `json:"idleBackfillEnabled" gorm:"column:idle_backfill_enabled;type:tinyint(1);default:0"`
	IdleScopeMode             string               `json:"idleScopeMode" gorm:"column:idle_scope_mode;type:varchar(20);default:selected"`
	IdleWaitMinutes           int                  `json:"idleWaitMinutes" gorm:"column:idle_wait_minutes;type:int;default:5"`
	ProbeIntervalMilliseconds int                  `json:"probeIntervalMilliseconds" gorm:"column:probe_interval_milliseconds;type:int;default:1500"`
	IdleBatchSize             int                  `json:"idleBatchSize" gorm:"column:idle_batch_size;type:int;default:20"`
	Paused                    bool                 `json:"paused" gorm:"type:tinyint(1);default:0"`
	CreatedAt                 *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt                 *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (VideoMetadataSetting) TableName() string { return "video_metadata_settings" }

func (VideoMetadataSetting) Default() VideoMetadataSetting {
	return VideoMetadataSetting{
		ID:                        "default",
		CollectOnNewOrChanged:     true,
		CollectOnDetailOrPlay:     true,
		IdleScopeMode:             VideoMetadataScopeSelected,
		IdleWaitMinutes:           5,
		ProbeIntervalMilliseconds: 1500,
		IdleBatchSize:             20,
	}
}

func (m VideoMetadataSetting) Ensure(db *gorm.DB) (*VideoMetadataSetting, error) {
	var item VideoMetadataSetting
	err := db.Where("id = ?", "default").First(&item).Error
	if err == nil {
		return &item, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	item = m.Default()
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (VideoMetadataSetting) Save(db *gorm.DB, item *VideoMetadataSetting) error {
	item.ID = "default"
	return db.Model(&VideoMetadataSetting{}).Where("id = ?", item.ID).Select("*").Updates(item).Error
}

type VideoMetadataSettingFilesBases struct {
	ID           string               `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID string               `json:"filesBasesId" gorm:"column:files_bases_id;type:char(20);uniqueIndex;"`
	CreatedAt    *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
}

func (VideoMetadataSettingFilesBases) TableName() string {
	return "video_metadata_setting_files_bases"
}

func (VideoMetadataSettingFilesBases) ListIDs(db *gorm.DB) ([]string, error) {
	var ids []string
	err := db.Model(&VideoMetadataSettingFilesBases{}).Order("files_bases_id").Pluck("files_bases_id", &ids).Error
	return ids, err
}

func (VideoMetadataSettingFilesBases) Replace(db *gorm.DB, ids []string, newID func() string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&VideoMetadataSettingFilesBases{}).Error; err != nil {
			return err
		}
		for _, id := range ids {
			if id == "" {
				continue
			}
			item := VideoMetadataSettingFilesBases{ID: newID(), FilesBasesID: id}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (VideoMetadataSettingFilesBases) DeleteByFilesBasesID(db *gorm.DB, id string) error {
	return db.Unscoped().Where("files_bases_id = ?", id).Delete(&VideoMetadataSettingFilesBases{}).Error
}

// VideoMetadataBatchTask 保存手动批量补齐的范围和可恢复状态。
type VideoMetadataBatchTask struct {
	ID         string               `json:"id" gorm:"primaryKey;type:char(20);"`
	ScopeMode  string               `json:"scopeMode" gorm:"column:scope_mode;type:varchar(20);"`
	RunMode    string               `json:"runMode" gorm:"column:run_mode;type:varchar(30);"`
	Status     string               `json:"status" gorm:"type:varchar(20);index:idx_video_metadata_batch_status"`
	Total      int64                `json:"total" gorm:"type:bigint;default:0"`
	Success    int64                `json:"success" gorm:"type:bigint;default:0"`
	Failed     int64                `json:"failed" gorm:"type:bigint;default:0"`
	Skipped    int64                `json:"skipped" gorm:"type:bigint;default:0"`
	CurrentSrc string               `json:"currentSrc" gorm:"column:current_src;type:text;"`
	LastError  string               `json:"lastError" gorm:"column:last_error;type:text;"`
	CreatedAt  *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	StartedAt  *datatype.CustomTime `json:"startedAt" gorm:"column:started_at;type:datetime;"`
	FinishedAt *datatype.CustomTime `json:"finishedAt" gorm:"column:finished_at;type:datetime;"`
	UpdatedAt  *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (VideoMetadataBatchTask) TableName() string { return "video_metadata_batch_tasks" }

type VideoMetadataBatchTaskFilesBases struct {
	ID           string `json:"id" gorm:"primaryKey;type:char(20);"`
	TaskID       string `json:"taskId" gorm:"column:task_id;type:char(20);index:idx_video_metadata_batch_scope"`
	FilesBasesID string `json:"filesBasesId" gorm:"column:files_bases_id;type:char(20);index:idx_video_metadata_batch_scope"`
}

func (VideoMetadataBatchTaskFilesBases) TableName() string {
	return "video_metadata_batch_task_files_bases"
}
