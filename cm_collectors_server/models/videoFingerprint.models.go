package models

import (
	"cm_collectors_server/datatype"

	"gorm.io/gorm"
)

const (
	VideoFingerprintStatus_Pending = 0
	VideoFingerprintStatus_Done    = 1
	VideoFingerprintStatus_Failed  = 2
)

type VideoFingerprint struct {
	ID            string               `json:"id" gorm:"primaryKey;type:char(20);"`
	DramaSeriesID string               `json:"drama_series_id" gorm:"column:drama_series_id;type:char(20);uniqueIndex;"`
	ResourcesID   string               `json:"resources_id" gorm:"column:resources_id;type:char(20);index:idx_vf_resources_id;"`
	FilesBasesID  string               `json:"files_bases_id" gorm:"column:files_bases_id;type:char(20);index:idx_vf_files_bases_id;"`
	Src           string               `json:"src" gorm:"column:src;type:text;"`
	Duration      float64              `json:"duration" gorm:"column:duration;type:double;default:0;"`
	PHash05       string               `json:"p_hash_05" gorm:"column:p_hash_05;type:varchar(16);"`
	PHash15       string               `json:"p_hash_15" gorm:"column:p_hash_15;type:varchar(16);"`
	PHash25       string               `json:"p_hash_25" gorm:"column:p_hash_25;type:varchar(16);"`
	PHash35       string               `json:"p_hash_35" gorm:"column:p_hash_35;type:varchar(16);"`
	PHash45       string               `json:"p_hash_45" gorm:"column:p_hash_45;type:varchar(16);"`
	PHash55       string               `json:"p_hash_55" gorm:"column:p_hash_55;type:varchar(16);"`
	PHash65       string               `json:"p_hash_65" gorm:"column:p_hash_65;type:varchar(16);"`
	PHash75       string               `json:"p_hash_75" gorm:"column:p_hash_75;type:varchar(16);"`
	PHash85       string               `json:"p_hash_85" gorm:"column:p_hash_85;type:varchar(16);"`
	PHash95       string               `json:"p_hash_95" gorm:"column:p_hash_95;type:varchar(16);"`
	Status        int                  `json:"status" gorm:"column:status;type:tinyint;default:0;"`
	FailReason    string               `json:"fail_reason" gorm:"column:fail_reason;type:varchar(500);"`
	CreatedAt     *datatype.CustomTime `json:"created_at" gorm:"column:created_at;type:datetime;autoCreateTime;"`
	UpdatedAt     *datatype.CustomTime `json:"updated_at" gorm:"column:updated_at;type:datetime;autoUpdateTime;"`
}

type VideoFingerprintSeedItem struct {
	DramaSeriesID string `gorm:"column:drama_series_id"`
	ResourcesID   string `gorm:"column:resources_id"`
	FilesBasesID  string `gorm:"column:files_bases_id"`
	Src           string `gorm:"column:src"`
}

func (VideoFingerprint) TableName() string {
	return "video_fingerprints"
}

func (VideoFingerprint) Create(db *gorm.DB, vf *VideoFingerprint) error {
	return db.Create(vf).Error
}

func (VideoFingerprint) Creates(db *gorm.DB, vfs *[]VideoFingerprint) error {
	if len(*vfs) == 0 {
		return nil
	}
	return db.Create(vfs).Error
}

func (VideoFingerprint) GetByDramaSeriesID(db *gorm.DB, dramaSeriesID string) (*VideoFingerprint, error) {
	var vf VideoFingerprint
	err := db.Where("drama_series_id = ?", dramaSeriesID).First(&vf).Error
	return &vf, err
}

func (VideoFingerprint) UpdateComputed(db *gorm.DB, vf *VideoFingerprint) error {
	return db.Model(vf).Updates(map[string]interface{}{
		"duration":    vf.Duration,
		"p_hash_05":   vf.PHash05,
		"p_hash_15":   vf.PHash15,
		"p_hash_25":   vf.PHash25,
		"p_hash_35":   vf.PHash35,
		"p_hash_45":   vf.PHash45,
		"p_hash_55":   vf.PHash55,
		"p_hash_65":   vf.PHash65,
		"p_hash_75":   vf.PHash75,
		"p_hash_85":   vf.PHash85,
		"p_hash_95":   vf.PHash95,
		"status":      vf.Status,
		"fail_reason": vf.FailReason,
	}).Error
}

func (VideoFingerprint) DeleteByDramaSeriesID(db *gorm.DB, dramaSeriesID string) error {
	return db.Unscoped().Where("drama_series_id = ?", dramaSeriesID).Delete(&VideoFingerprint{}).Error
}

func (VideoFingerprint) DeleteByDramaSeriesIDs(db *gorm.DB, dramaSeriesIDs []string) error {
	if len(dramaSeriesIDs) == 0 {
		return nil
	}
	return db.Unscoped().Where("drama_series_id IN ?", dramaSeriesIDs).Delete(&VideoFingerprint{}).Error
}

func (VideoFingerprint) DeleteByResourcesID(db *gorm.DB, resourcesID string) error {
	return db.Unscoped().Where("resources_id = ?", resourcesID).Delete(&VideoFingerprint{}).Error
}

func (VideoFingerprint) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	return db.Unscoped().Where("files_bases_id = ?", filesBasesID).Delete(&VideoFingerprint{}).Error
}

func (VideoFingerprint) DeleteAll(db *gorm.DB) error {
	return db.Unscoped().Where("1 = 1").Delete(&VideoFingerprint{}).Error
}

func (VideoFingerprint) ListPending(db *gorm.DB, filesBasesID string, limit int) (*[]VideoFingerprint, error) {
	var list []VideoFingerprint
	q := db.Where("status = ?", VideoFingerprintStatus_Pending)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	err := q.Limit(limit).Find(&list).Error
	return &list, err
}

func (VideoFingerprint) ListDone(db *gorm.DB, filesBasesID string) (*[]VideoFingerprint, error) {
	var list []VideoFingerprint
	q := db.Where("status = ?", VideoFingerprintStatus_Done)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	err := q.Find(&list).Error
	return &list, err
}

func (VideoFingerprint) MissingVideoDramaSeries(db *gorm.DB, filesBasesID string) ([]VideoFingerprintSeedItem, error) {
	var list []VideoFingerprintSeedItem
	q := db.Table(ResourcesDramaSeries{}.TableName()+" AS ds").
		Select("ds.id AS drama_series_id, ds.resources_id AS resources_id, r.filesBases_id AS files_bases_id, ds.src AS src").
		Joins("JOIN "+Resources{}.TableName()+" AS r ON r.id = ds.resources_id").
		Joins("LEFT JOIN "+VideoFingerprint{}.TableName()+" AS vf ON vf.drama_series_id = ds.id").
		Where("vf.id IS NULL").
		Where("r.mode IN ?", []string{"movies", "videoLink"})
	if filesBasesID != "" {
		q = q.Where("r.filesBases_id = ?", filesBasesID)
	}
	err := q.Scan(&list).Error
	return list, err
}

func (VideoFingerprint) Stats(db *gorm.DB, filesBasesID string) (*VideoFingerprintStats, error) {
	var total, pending, done, failed int64
	q := db.Model(&VideoFingerprint{})
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	q.Count(&total)

	q = db.Model(&VideoFingerprint{}).Where("status = ?", VideoFingerprintStatus_Pending)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	q.Count(&pending)

	q = db.Model(&VideoFingerprint{}).Where("status = ?", VideoFingerprintStatus_Done)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	q.Count(&done)

	q = db.Model(&VideoFingerprint{}).Where("status = ?", VideoFingerprintStatus_Failed)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	q.Count(&failed)

	var dramaTotal int64
	dramaTotalQ := db.Model(&ResourcesDramaSeries{}).
		Joins("JOIN resources ON resources.id = resourcesDramaSeries.resources_id").
		Where("resources.mode IN ?", []string{"movies", "videoLink"})
	if filesBasesID != "" {
		dramaTotalQ = dramaTotalQ.Where("resources.filesBases_id = ?", filesBasesID)
	}
	dramaTotalQ.Count(&dramaTotal)

	return &VideoFingerprintStats{
		Total:      total,
		Pending:    pending,
		Done:       done,
		Failed:     failed,
		DramaTotal: dramaTotal,
	}, nil
}

func (VideoFingerprint) ExistsByDramaSeriesID(db *gorm.DB, dramaSeriesID string) bool {
	var count int64
	db.Model(&VideoFingerprint{}).Where("drama_series_id = ?", dramaSeriesID).Count(&count)
	return count > 0
}

func (VideoFingerprint) ResetFailed(db *gorm.DB, filesBasesID string) (int64, error) {
	q := db.Model(&VideoFingerprint{}).Where("status = ?", VideoFingerprintStatus_Failed)
	if filesBasesID != "" {
		q = q.Where("files_bases_id = ?", filesBasesID)
	}
	result := q.Updates(map[string]interface{}{
		"status":      VideoFingerprintStatus_Pending,
		"fail_reason": "",
	})
	return result.RowsAffected, result.Error
}

type VideoFingerprintStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Done       int64 `json:"done"`
	Failed     int64 `json:"failed"`
	DramaTotal int64 `json:"drama_total"`
}
