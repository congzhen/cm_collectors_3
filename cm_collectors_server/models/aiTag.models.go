package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

const (
	AiTagRecordStatusPending    = "pending"
	AiTagRecordStatusProcessing = "processing"
	AiTagRecordStatusSuccess    = "success"
	AiTagRecordStatusFailed     = "failed"
	AiTagRecordStatusSkipped    = "skipped"
)

const (
	AiTagWriteModeAppend    = "append"
	AiTagWriteModeReplace   = "replace"
	AiTagWriteModeOnlyEmpty = "only_empty"
)

const (
	AiTagFrameStrategyQuick        = "quick"
	AiTagFrameStrategyHighAccuracy = "high_accuracy_adaptive"
	AiTagFrameStrategyUltra        = "ultra_accuracy"
)

const (
	AiTagImageResizeOriginal     = "original"
	AiTagImageResizeAutoFallback = "auto_fallback"
	AiTagImageResizeFixedResize  = "fixed_resize"
)

// AiTagSetting 保存 AI 自动标签的全局配置。
// 这里包含 AI 服务、截图策略、写入策略和任务节流参数；文件库启用范围单独存在 AiTagEnabledFilesBases。
type AiTagSetting struct {
	ID                    string               `json:"id" gorm:"primaryKey;type:char(20);"`
	Enabled               bool                 `json:"enabled" gorm:"type:tinyint(1);default:0"`
	Paused                bool                 `json:"paused" gorm:"type:tinyint(1);default:0"`
	Provider              string               `json:"provider" gorm:"type:varchar(50);default:openai"`
	BaseURL               string               `json:"baseUrl" gorm:"column:base_url;type:varchar(500);"`
	APIKey                string               `json:"apiKey" gorm:"column:api_key;type:text;"`
	Model                 string               `json:"model" gorm:"type:varchar(100);"`
	RequestTimeoutSeconds int                  `json:"requestTimeoutSeconds" gorm:"column:request_timeout_seconds;type:int;default:1800"`
	MaxResourcesPerRun    int                  `json:"maxResourcesPerRun" gorm:"column:max_resources_per_run;type:int;default:20"`
	MaxFramesPerResource  int                  `json:"maxFramesPerResource" gorm:"column:max_frames_per_resource;type:int;default:80"`
	MaxFramesPerVideo     int                  `json:"maxFramesPerVideo" gorm:"column:max_frames_per_video;type:int;default:40"`
	MaxVideosPerResource  int                  `json:"maxVideosPerResource" gorm:"column:max_videos_per_resource;type:int;default:12"`
	MaxImagesPerAIRequest int                  `json:"maxImagesPerAiRequest" gorm:"column:max_images_per_ai_request;type:int;default:5"`
	FrameStrategy         string               `json:"frameStrategy" gorm:"column:frame_strategy;type:varchar(50);default:high_accuracy_adaptive"`
	ImageResizeMode       string               `json:"imageResizeMode" gorm:"column:image_resize_mode;type:varchar(50);default:original"`
	FallbackImageMaxWidth int                  `json:"fallbackImageMaxWidth" gorm:"column:fallback_image_max_width;type:int;default:1440"`
	ImageJPEGQuality      int                  `json:"imageJpegQuality" gorm:"column:image_jpeg_quality;type:int;default:100"`
	MinConfidence         float64              `json:"minConfidence" gorm:"column:min_confidence;type:double;default:0.7"`
	MaxTagsPerResource    int                  `json:"maxTagsPerResource" gorm:"column:max_tags_per_resource;type:int;default:8"`
	WriteMode             string               `json:"writeMode" gorm:"column:write_mode;type:varchar(30);default:append"`
	CreatedAt             *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt             *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (AiTagSetting) TableName() string { return "ai_tag_settings" }

// Default 是功能首次启用时的保守默认值。
// 默认不启用功能，但截图和分批策略偏高准确，避免用户一开启就因过度压缩影响判断。
func (AiTagSetting) Default() AiTagSetting {
	return AiTagSetting{
		Provider:              "openai",
		RequestTimeoutSeconds: 1800,
		MaxResourcesPerRun:    20,
		MaxFramesPerResource:  80,
		MaxFramesPerVideo:     40,
		MaxVideosPerResource:  12,
		MaxImagesPerAIRequest: 5,
		FrameStrategy:         AiTagFrameStrategyHighAccuracy,
		ImageResizeMode:       AiTagImageResizeOriginal,
		FallbackImageMaxWidth: 1440,
		ImageJPEGQuality:      100,
		MinConfidence:         0.7,
		MaxTagsPerResource:    8,
		WriteMode:             AiTagWriteModeAppend,
	}
}

func (m AiTagSetting) First(db *gorm.DB) (*AiTagSetting, error) {
	var info AiTagSetting
	err := db.First(&info).Error
	return &info, err
}

func (AiTagSetting) Create(db *gorm.DB, item *AiTagSetting) error {
	return db.Create(item).Error
}

func (AiTagSetting) Update(db *gorm.DB, item *AiTagSetting) error {
	return db.Model(item).Select("*").Updates(item).Error
}

// AiTagEnabledFilesBases 保存哪些文件库参与 AI 自动标签。
// include/exclude 是标签分类筛选，影响喂给 AI 的标签池，不会改变文件库内真实标签。
type AiTagEnabledFilesBases struct {
	ID                 string               `json:"id" gorm:"primaryKey;type:char(20);"`
	FilesBasesID       string               `json:"filesBasesId" gorm:"column:files_bases_id;type:char(20);uniqueIndex;"`
	Enabled            bool                 `json:"enabled" gorm:"type:tinyint(1);default:0"`
	IncludeTagClassIDs string               `json:"includeTagClassIds" gorm:"column:include_tag_class_ids;type:text;"`
	ExcludeTagClassIDs string               `json:"excludeTagClassIds" gorm:"column:exclude_tag_class_ids;type:text;"`
	CreatedAt          *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt          *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (AiTagEnabledFilesBases) TableName() string { return "ai_tag_enabled_files_bases" }

func (AiTagEnabledFilesBases) List(db *gorm.DB) (*[]AiTagEnabledFilesBases, error) {
	var list []AiTagEnabledFilesBases
	err := db.Find(&list).Error
	return &list, err
}

func (AiTagEnabledFilesBases) ListEnabled(db *gorm.DB) (*[]AiTagEnabledFilesBases, error) {
	var list []AiTagEnabledFilesBases
	err := db.Where("enabled = ?", true).Find(&list).Error
	return &list, err
}

// Upsert 按 files_bases_id 更新或创建配置。
// 一个文件库只允许一条启用配置，避免计划任务扫描时出现重复范围。
func (AiTagEnabledFilesBases) Upsert(db *gorm.DB, item *AiTagEnabledFilesBases) error {
	var existing AiTagEnabledFilesBases
	err := db.Where("files_bases_id = ?", item.FilesBasesID).First(&existing).Error
	if err == nil {
		item.ID = existing.ID
		return db.Model(&existing).Updates(map[string]interface{}{
			"enabled":               item.Enabled,
			"include_tag_class_ids": item.IncludeTagClassIDs,
			"exclude_tag_class_ids": item.ExcludeTagClassIDs,
		}).Error
	}
	return db.Create(item).Error
}

// AiTagAnalysisRecord 是资源级分析记录。
// src_hash 标记资源视频源是否变化，tag_version_hash 标记可参与 AI 的标签池是否变化；
// 两者都没变时，success 记录不会被重复分析。
type AiTagAnalysisRecord struct {
	ID                string               `json:"id" gorm:"primaryKey;type:char(20);"`
	ResourcesID       string               `json:"resourcesId" gorm:"column:resources_id;type:char(20);index:idx_ai_tag_resources_id;uniqueIndex"`
	FilesBasesID      string               `json:"filesBasesId" gorm:"column:files_bases_id;type:char(20);index:idx_ai_tag_files_bases_id;"`
	Status            string               `json:"status" gorm:"type:varchar(30);index:idx_ai_tag_status;"`
	SrcHash           string               `json:"srcHash" gorm:"column:src_hash;type:varchar(64);"`
	TagVersionHash    string               `json:"tagVersionHash" gorm:"column:tag_version_hash;type:varchar(64);"`
	PromptVersion     string               `json:"promptVersion" gorm:"column:prompt_version;type:varchar(30);"`
	Model             string               `json:"model" gorm:"type:varchar(100);"`
	RecommendedTagIDs string               `json:"recommendedTagIds" gorm:"column:recommended_tag_ids;type:text;"`
	WrittenTagIDs     string               `json:"writtenTagIds" gorm:"column:written_tag_ids;type:text;"`
	ResultJSON        string               `json:"resultJson" gorm:"column:result_json;type:text;"`
	EvidenceJSON      string               `json:"evidenceJson" gorm:"column:evidence_json;type:text;"`
	FailReason        string               `json:"failReason" gorm:"column:fail_reason;type:text;"`
	AnalyzedAt        *datatype.CustomTime `json:"analyzedAt" gorm:"column:analyzed_at;type:datetime;"`
	CreatedAt         *datatype.CustomTime `json:"createdAt" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt         *datatype.CustomTime `json:"updatedAt" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

func (AiTagAnalysisRecord) TableName() string { return "ai_tag_analysis_records" }

func (AiTagAnalysisRecord) GetByResourceID(db *gorm.DB, resourceID string) (*AiTagAnalysisRecord, error) {
	var item AiTagAnalysisRecord
	err := db.Where("resources_id = ?", resourceID).First(&item).Error
	return &item, err
}

func (AiTagAnalysisRecord) Create(db *gorm.DB, item *AiTagAnalysisRecord) error {
	return db.Create(item).Error
}

func (AiTagAnalysisRecord) Update(db *gorm.DB, item *AiTagAnalysisRecord) error {
	return db.Model(item).Select("*").Updates(item).Error
}

func (AiTagAnalysisRecord) List(db *gorm.DB, filesBasesID, status string, page, limit int) (*[]AiTagAnalysisRecord, int64, error) {
	var list []AiTagAnalysisRecord
	var total int64
	q := db.Model(&AiTagAnalysisRecord{})
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return &list, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	err := q.Order("updated_at desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error
	return &list, total, err
}

func (AiTagAnalysisRecord) Stats(db *gorm.DB, filesBasesID string) (map[string]int64, error) {
	statuses := []string{AiTagRecordStatusPending, AiTagRecordStatusProcessing, AiTagRecordStatusSuccess, AiTagRecordStatusFailed, AiTagRecordStatusSkipped}
	result := map[string]int64{}
	for _, status := range statuses {
		q := db.Model(&AiTagAnalysisRecord{}).Where("status = ?", status)
		if filesBasesID != "" {
			q = q.Where("files_bases_id = ?", filesBasesID)
		}
		var count int64
		if err := q.Count(&count).Error; err != nil {
			return result, err
		}
		result[status] = count
	}
	return result, nil
}

func (AiTagAnalysisRecord) ResetFailed(db *gorm.DB, filesBasesID string) (int64, error) {
	q := db.Model(&AiTagAnalysisRecord{}).Where("status = ?", AiTagRecordStatusFailed)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	res := q.Updates(map[string]interface{}{"status": AiTagRecordStatusPending, "fail_reason": ""})
	return res.RowsAffected, res.Error
}

func (AiTagAnalysisRecord) ResetProcessing(db *gorm.DB, filesBasesID string) (int64, error) {
	q := db.Model(&AiTagAnalysisRecord{}).Where("status = ?", AiTagRecordStatusProcessing)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	res := q.Updates(map[string]interface{}{"status": AiTagRecordStatusPending, "fail_reason": "任务中断，已恢复为待处理"})
	return res.RowsAffected, res.Error
}

func (AiTagAnalysisRecord) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Unscoped().Where("files_bases_id = ?", filesBasesID).Delete(&AiTagAnalysisRecord{}).Error
}
