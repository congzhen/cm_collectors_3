package processors

import (
	"bufio"
	"bytes"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorscache "cm_collectors_server/processorsCache"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type VideoTranscodeConfig struct {
	Container        string  `json:"container"`
	VideoCodec       string  `json:"videoCodec"`
	QualityMode      string  `json:"qualityMode"`
	CRF              int     `json:"crf"`
	VideoBitrateKbps int     `json:"videoBitrateKbps"`
	Preset           string  `json:"preset"`
	ResolutionHeight int     `json:"resolutionHeight"`
	FrameRate        float64 `json:"frameRate"`
	AudioCodec       string  `json:"audioCodec"`
	AudioBitrateKbps int     `json:"audioBitrateKbps"`
	Threads          int     `json:"threads"`
	KeepBackup       bool    `json:"keepBackup"`
	GPUEncoder       string  `json:"gpuEncoder"`
}

func DefaultVideoTranscodeConfig() VideoTranscodeConfig {
	return VideoTranscodeConfig{
		Container:        "mp4",
		VideoCodec:       "copy",
		QualityMode:      "crf",
		CRF:              23,
		VideoBitrateKbps: 4000,
		Preset:           "medium",
		AudioCodec:       "copy",
		AudioBitrateKbps: 192,
	}
}

type VideoTranscodeTaskView struct {
	models.VideoTranscodeTask
	Config       VideoTranscodeConfig `json:"config"`
	FilesBasesID string               `json:"filesBasesId"`
	CoverPoster  string               `json:"coverPoster"`
}

type VideoTranscodeAddRequest struct {
	ResourceIDs    []string              `json:"resourceIds"`
	DramaSeriesIDs []string              `json:"dramaSeriesIds"`
	Config         *VideoTranscodeConfig `json:"config"`
}

type VideoTranscodeAddResult struct {
	Added            int `json:"added"`
	SkippedDuplicate int `json:"skippedDuplicate"`
	SkippedMissing   int `json:"skippedMissing"`
}

type VideoTranscodeIDsRequest struct {
	IDs []string `json:"ids"`
}

type VideoTranscodeUpdateConfigRequest struct {
	IDs    []string             `json:"ids"`
	Config VideoTranscodeConfig `json:"config"`
}

type VideoTranscodeResetResult struct {
	Reset   int `json:"reset"`
	Skipped int `json:"skipped"`
}

type videoTranscodeCandidate struct {
	DramaSeriesID            string
	ResourceID               string
	ResourceTitle            string
	SourcePath               string
	SourceDuration           float64
	SourceWidth              int
	SourceHeight             int
	SourceFrameRate          float64
	SourceVideoCodec         string
	SourceAudioCodec         string
	SourceVideoBitRate       int64
	MetadataStatus           string
	MetadataFileSize         int64
	MetadataFileModifiedTime int64
}

type videoTranscodeManager struct {
	mu             sync.Mutex
	paused         bool
	currentID      string
	cancel         context.CancelFunc
	wake           chan struct{}
	retryScheduled bool
}

var transcodeManager = &videoTranscodeManager{wake: make(chan struct{}, 1)}

type mediaPathLockRegistry struct {
	mu    sync.Mutex
	locks map[string]*sync.RWMutex
}

var mediaPathLocks = &mediaPathLockRegistry{locks: make(map[string]*sync.RWMutex)}

func (r *mediaPathLockRegistry) get(path string) *sync.RWMutex {
	cleaned := filepath.Clean(path)
	r.mu.Lock()
	defer r.mu.Unlock()
	lock := r.locks[cleaned]
	if lock == nil {
		lock = &sync.RWMutex{}
		r.locks[cleaned] = lock
	}
	return lock
}

// lockMediaForRead 让播放请求与转码替换阶段共享同一把路径锁。
func lockMediaForRead(path string) func() {
	lock := mediaPathLocks.get(path)
	lock.RLock()
	return lock.RUnlock
}

type VideoTranscode struct{}

func InitVideoTranscode() {
	db := core.DBS()
	var ordinaryInterrupted []models.VideoTranscodeTask
	_ = db.Where("status IN ?", []string{
		models.VideoTranscodeStatusProbing,
		models.VideoTranscodeStatusTranscoding,
		models.VideoTranscodeStatusVerifying,
	}).Find(&ordinaryInterrupted).Error
	for _, task := range ordinaryInterrupted {
		if isVideoTranscodeTemporaryPathForTask(task) {
			_ = os.Remove(task.TemporaryPath)
		}
	}
	db.Model(&models.VideoTranscodeTask{}).
		Where("status IN ?", []string{
			models.VideoTranscodeStatusProbing,
			models.VideoTranscodeStatusTranscoding,
			models.VideoTranscodeStatusVerifying,
		}).
		Updates(map[string]interface{}{
			"status":        models.VideoTranscodeStatusInterrupted,
			"error_message": "服务在任务执行期间退出，请确认源文件后重试",
		})
	recoverCriticalVideoTranscodeTasks()
	go transcodeManager.loop()
	// 服务重启后只主动检查一次，恢复数据库中已经排队的任务。
	transcodeManager.notify()
}

func recoverCriticalVideoTranscodeTasks() {
	var tasks []models.VideoTranscodeTask
	if err := core.DBS().Where("status IN ?", []string{
		models.VideoTranscodeStatusReplacing,
		models.VideoTranscodeStatusRefreshingMetadata,
	}).Find(&tasks).Error; err != nil {
		return
	}
	for _, task := range tasks {
		recoverCriticalVideoTranscodeTask(task)
	}
}

func recoverCriticalVideoTranscodeTask(task models.VideoTranscodeTask) {
	if task.BackupPath == "" {
		markVideoTranscodeInterrupted(task.ID, "服务在文件替换阶段退出，但任务没有记录备份路径，请人工检查")
		return
	}
	if _, err := os.Stat(task.BackupPath); err != nil {
		markVideoTranscodeInterrupted(task.ID, "服务在文件替换阶段退出，未发现旧文件备份，请人工检查")
		return
	}

	lock := mediaPathLocks.get(task.SourcePath)
	lock.Lock()
	defer lock.Unlock()

	recoveredOutput := ""
	sourceExists := fileExistsForTranscode(task.SourcePath)
	outputExists := task.OutputPath != "" && fileExistsForTranscode(task.OutputPath)
	samePath := filepath.Clean(task.SourcePath) == filepath.Clean(task.OutputPath)
	if sourceExists && !samePath {
		core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":        models.VideoTranscodeStatusRollbackFailed,
			"error_message": "恢复时发现源路径和备份文件同时存在，已停止自动处理以避免覆盖文件",
		})
		return
	}
	if (samePath && sourceExists) || (!samePath && outputExists) {
		recoveredOutput = task.OutputPath + ".interrupted-" + task.ID
		if fileExistsForTranscode(recoveredOutput) {
			core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
				"status":        models.VideoTranscodeStatusRollbackFailed,
				"error_message": "恢复输出文件已存在，已停止自动处理: " + recoveredOutput,
			})
			return
		}
		if err := os.Rename(task.OutputPath, recoveredOutput); err != nil {
			core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
				"status":        models.VideoTranscodeStatusRollbackFailed,
				"error_message": "移动中断输出文件失败: " + err.Error(),
			})
			return
		}
	}
	if err := os.Rename(task.BackupPath, task.SourcePath); err != nil {
		core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":        models.VideoTranscodeStatusRollbackFailed,
			"error_message": "自动恢复旧文件失败: " + err.Error(),
		})
		return
	}
	err := core.DBS().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ResourcesDramaSeries{}).Where("id = ?", task.DramaSeriesID).
			Update("src", task.SourcePath).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ResourcesVideoMetadata{}).
			Where("drama_series_id = ?", task.DramaSeriesID).
			Updates(map[string]interface{}{
				"probe_status":     models.VideoMetadataStatusStale,
				"metadata_version": 0,
			}).Error; err != nil {
			return err
		}
		return (models.VideoFingerprint{}).RebuildByDramaSeriesIDs(
			tx, []string{task.DramaSeriesID}, core.GenerateUniqueID,
		)
	})
	if err != nil {
		core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":        models.VideoTranscodeStatusRollbackFailed,
			"error_message": "旧文件已恢复，但数据库路径恢复失败: " + err.Error(),
		})
		return
	}
	message := "服务异常退出后已自动恢复旧文件"
	if recoveredOutput != "" {
		message += "；中断时的新输出保留在 " + recoveredOutput
	}
	markVideoTranscodeInterrupted(task.ID, message)
}

func fileExistsForTranscode(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func markVideoTranscodeInterrupted(id, message string) {
	core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.VideoTranscodeStatusInterrupted,
		"error_message": message,
	})
}

func (m *videoTranscodeManager) notify() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *videoTranscodeManager) loop() {
	for range m.wake {
		m.runNext()
	}
}

func (m *videoTranscodeManager) runNext() {
	m.mu.Lock()
	if m.paused || m.currentID != "" {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	task, err := (models.VideoTranscodeTask{}).FirstQueued(core.DBS())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			core.LogErr(fmt.Errorf("读取视频转码队列失败: %w", err))
			m.scheduleRetry()
		}
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	if m.paused || m.currentID != "" {
		m.mu.Unlock()
		cancel()
		return
	}
	now := datatype.CustomTime(core.TimeNow())
	claim := core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id = ? AND status = ?", task.ID, models.VideoTranscodeStatusQueued).
		Updates(map[string]interface{}{
			"status":     models.VideoTranscodeStatusProbing,
			"started_at": &now,
		})
	if claim.Error != nil {
		m.mu.Unlock()
		cancel()
		core.LogErr(fmt.Errorf("领取视频转码任务失败: %w", claim.Error))
		m.scheduleRetry()
		return
	}
	if claim.RowsAffected == 0 {
		m.mu.Unlock()
		cancel()
		m.notify()
		return
	}
	m.currentID = task.ID
	m.cancel = cancel
	m.mu.Unlock()

	go func() {
		defer func() {
			cancel()
			m.mu.Lock()
			m.currentID = ""
			m.cancel = nil
			m.mu.Unlock()
			m.notify()
		}()
		(VideoTranscode{}).execute(ctx, task)
	}()
}

func (m *videoTranscodeManager) scheduleRetry() {
	m.mu.Lock()
	if m.retryScheduled {
		m.mu.Unlock()
		return
	}
	m.retryScheduled = true
	m.mu.Unlock()
	time.AfterFunc(2*time.Second, func() {
		m.mu.Lock()
		m.retryScheduled = false
		m.mu.Unlock()
		m.notify()
	})
}

func (VideoTranscode) List() ([]VideoTranscodeTaskView, error) {
	tasks, err := (models.VideoTranscodeTask{}).List(core.DBS())
	if err != nil {
		return nil, err
	}
	type resourceImage struct {
		ID           string
		FilesBasesID string
		CoverPoster  string
	}
	resourceIDs := make([]string, 0, len(tasks))
	dramaSeriesIDs := make([]string, 0, len(tasks))
	resourceIDSet := make(map[string]struct{}, len(tasks))
	dramaSeriesIDSet := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if _, exists := resourceIDSet[task.ResourceID]; !exists {
			resourceIDSet[task.ResourceID] = struct{}{}
			resourceIDs = append(resourceIDs, task.ResourceID)
		}
		if task.SourceWidth <= 0 || (task.Status == models.VideoTranscodeStatusSuccess && task.OutputWidth <= 0) {
			if _, exists := dramaSeriesIDSet[task.DramaSeriesID]; !exists {
				dramaSeriesIDSet[task.DramaSeriesID] = struct{}{}
				dramaSeriesIDs = append(dramaSeriesIDs, task.DramaSeriesID)
			}
		}
	}
	var resourceImages []resourceImage
	if len(resourceIDs) > 0 {
		if err := core.DBS().Model(&models.Resources{}).
			Select("id, filesBases_id AS files_bases_id, coverPoster AS cover_poster").
			Where("id IN ?", resourceIDs).
			Scan(&resourceImages).Error; err != nil {
			return nil, err
		}
	}
	imagesByResourceID := make(map[string]resourceImage, len(resourceImages))
	for _, image := range resourceImages {
		imagesByResourceID[image.ID] = image
	}
	var sourceMetadata []videoTranscodeCandidate
	if len(dramaSeriesIDs) > 0 {
		if err := core.DBS().Table(models.ResourcesDramaSeries{}.TableName()+" AS ds").
			Select(`ds.id AS drama_series_id,
				ds.durationSeconds AS source_duration,
				vm.width AS source_width,
				vm.height AS source_height,
				vm.frame_rate AS source_frame_rate,
				vm.video_codec AS source_video_codec,
				vm.audio_codec AS source_audio_codec,
				vm.video_bit_rate AS source_video_bit_rate,
				vm.probe_status AS metadata_status,
				vm.file_size AS metadata_file_size,
				vm.file_modified_time AS metadata_file_modified_time`).
			Joins("LEFT JOIN "+models.ResourcesVideoMetadata{}.TableName()+" AS vm ON vm.drama_series_id = ds.id").
			Where("ds.id IN ?", dramaSeriesIDs).
			Scan(&sourceMetadata).Error; err != nil {
			return nil, err
		}
	}
	metadataByDramaSeriesID := make(map[string]videoTranscodeCandidate, len(sourceMetadata))
	for _, metadata := range sourceMetadata {
		metadataByDramaSeriesID[metadata.DramaSeriesID] = metadata
	}
	result := make([]VideoTranscodeTaskView, 0, len(tasks))
	for _, task := range tasks {
		config := DefaultVideoTranscodeConfig()
		_ = json.Unmarshal([]byte(task.ConfigJsonData), &config)
		metadata := metadataByDramaSeriesID[task.DramaSeriesID]
		if task.SourceWidth <= 0 &&
			metadata.MetadataStatus == models.VideoMetadataStatusSuccess &&
			metadata.MetadataFileModifiedTime == task.SourceModifiedTime {
			task.SourceDuration = metadata.SourceDuration
			task.SourceWidth = metadata.SourceWidth
			task.SourceHeight = metadata.SourceHeight
			task.SourceFrameRate = metadata.SourceFrameRate
			task.SourceVideoCodec = metadata.SourceVideoCodec
			task.SourceAudioCodec = metadata.SourceAudioCodec
			task.SourceVideoBitRate = metadata.SourceVideoBitRate
		}
		if task.OutputWidth <= 0 &&
			task.Status == models.VideoTranscodeStatusSuccess &&
			metadata.MetadataStatus == models.VideoMetadataStatusSuccess &&
			(task.OutputSize <= 0 || metadata.MetadataFileSize == task.OutputSize) {
			task.OutputDuration = metadata.SourceDuration
			task.OutputWidth = metadata.SourceWidth
			task.OutputHeight = metadata.SourceHeight
			task.OutputFrameRate = metadata.SourceFrameRate
			task.OutputVideoCodec = metadata.SourceVideoCodec
			task.OutputAudioCodec = metadata.SourceAudioCodec
			task.OutputVideoBitRate = metadata.SourceVideoBitRate
		}
		image := imagesByResourceID[task.ResourceID]
		result = append(result, VideoTranscodeTaskView{
			VideoTranscodeTask: task,
			Config:             config,
			FilesBasesID:       image.FilesBasesID,
			CoverPoster:        image.CoverPoster,
		})
	}
	return result, nil
}

func (VideoTranscode) Capabilities() map[string]interface{} {
	return map[string]interface{}{
		"gpuEncoders": (processorsffmpeg.FFmpeg{}).AvailableGPUEncoders(),
	}
}

func (VideoTranscode) Add(request VideoTranscodeAddRequest) (*VideoTranscodeAddResult, error) {
	config := DefaultVideoTranscodeConfig()
	if request.Config != nil {
		config = *request.Config
	}
	if err := validateVideoTranscodeConfig(config); err != nil {
		return nil, err
	}
	candidates, err := videoTranscodeCandidates(request)
	if err != nil {
		return nil, err
	}
	configJSON, _ := json.Marshal(config)
	result := &VideoTranscodeAddResult{}
	err = core.DBS().Transaction(func(tx *gorm.DB) error {
		for _, candidate := range candidates {
			active, err := (models.VideoTranscodeTask{}).HasActiveByDramaSeriesID(tx, candidate.DramaSeriesID)
			if err != nil {
				return err
			}
			if active {
				result.SkippedDuplicate++
				continue
			}
			stat, err := os.Stat(candidate.SourcePath)
			if err != nil || stat.IsDir() {
				result.SkippedMissing++
				continue
			}
			now := datatype.CustomTime(core.TimeNow())
			task := models.VideoTranscodeTask{
				ID:                 core.GenerateUniqueID(),
				DramaSeriesID:      candidate.DramaSeriesID,
				ResourceID:         candidate.ResourceID,
				ResourceTitle:      candidate.ResourceTitle,
				SourcePath:         candidate.SourcePath,
				SourceSize:         stat.Size(),
				SourceModifiedTime: stat.ModTime().UnixMilli(),
				ConfigJsonData:     string(configJSON),
				Status:             models.VideoTranscodeStatusDraft,
				CreatedAt:          &now,
			}
			if candidate.MetadataStatus == models.VideoMetadataStatusSuccess &&
				candidate.MetadataFileModifiedTime == stat.ModTime().UnixMilli() {
				task.SourceDuration = candidate.SourceDuration
				task.SourceWidth = candidate.SourceWidth
				task.SourceHeight = candidate.SourceHeight
				task.SourceFrameRate = candidate.SourceFrameRate
				task.SourceVideoCodec = candidate.SourceVideoCodec
				task.SourceAudioCodec = candidate.SourceAudioCodec
				task.SourceVideoBitRate = candidate.SourceVideoBitRate
			}
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
			result.Added++
		}
		return nil
	})
	return result, err
}

func videoTranscodeCandidates(request VideoTranscodeAddRequest) ([]videoTranscodeCandidate, error) {
	if len(request.ResourceIDs) == 0 && len(request.DramaSeriesIDs) == 0 {
		return nil, errors.New("请选择至少一个资源或视频分集")
	}
	var candidates []videoTranscodeCandidate
	query := core.DBS().Table(models.ResourcesDramaSeries{}.TableName() + " AS ds").
		Select(`ds.id AS drama_series_id,
			ds.resources_id AS resource_id,
			r.title AS resource_title,
			ds.src AS source_path,
			ds.durationSeconds AS source_duration,
			vm.width AS source_width,
			vm.height AS source_height,
			vm.frame_rate AS source_frame_rate,
			vm.video_codec AS source_video_codec,
			vm.audio_codec AS source_audio_codec,
			vm.video_bit_rate AS source_video_bit_rate,
			vm.probe_status AS metadata_status,
			vm.file_size AS metadata_file_size,
			vm.file_modified_time AS metadata_file_modified_time`).
		Joins("JOIN " + models.Resources{}.TableName() + " AS r ON r.id = ds.resources_id").
		Joins("LEFT JOIN " + models.ResourcesVideoMetadata{}.TableName() + " AS vm ON vm.drama_series_id = ds.id")
	if len(request.ResourceIDs) > 0 && len(request.DramaSeriesIDs) > 0 {
		query = query.Where("ds.resources_id IN ? OR ds.id IN ?", request.ResourceIDs, request.DramaSeriesIDs)
	} else if len(request.ResourceIDs) > 0 {
		query = query.Where("ds.resources_id IN ?", request.ResourceIDs)
	} else {
		query = query.Where("ds.id IN ?", request.DramaSeriesIDs)
	}
	err := query.Where("ds.src <> ''").Order("r.id, ds.sort").Scan(&candidates).Error
	return candidates, err
}

func (VideoTranscode) UpdateConfig(request VideoTranscodeUpdateConfigRequest) error {
	if err := validateVideoTranscodeConfig(request.Config); err != nil {
		return err
	}
	data, _ := json.Marshal(request.Config)
	query := core.DBS().Model(&models.VideoTranscodeTask{}).Where("status = ?", models.VideoTranscodeStatusDraft)
	if len(request.IDs) > 0 {
		query = query.Where("id IN ?", request.IDs)
	}
	return query.Update("config_json_data", string(data)).Error
}

func (VideoTranscode) Start(request VideoTranscodeIDsRequest) error {
	statuses := []string{
		models.VideoTranscodeStatusDraft,
		models.VideoTranscodeStatusFailed,
		models.VideoTranscodeStatusCancelled,
		models.VideoTranscodeStatusInterrupted,
	}
	query := core.DBS().Where("status IN ?", statuses)
	if len(request.IDs) > 0 {
		query = query.Where("id IN ?", request.IDs)
	}
	var tasks []models.VideoTranscodeTask
	if err := query.Find(&tasks).Error; err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	taskIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if isVideoTranscodeTemporaryPathForTask(task) {
			if err := os.Remove(task.TemporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("清理任务 %s 的旧转码临时文件失败: %w", task.ID, err)
			}
		}
		taskIDs = append(taskIDs, task.ID)
	}
	if err := core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id IN ? AND status IN ?", taskIDs, statuses).
		Updates(map[string]interface{}{
			"status":                models.VideoTranscodeStatusQueued,
			"progress":              0,
			"processed_seconds":     0,
			"speed":                 "",
			"temporary_path":        "",
			"output_path":           "",
			"backup_path":           "",
			"output_size":           0,
			"output_duration":       0,
			"output_width":          0,
			"output_height":         0,
			"output_frame_rate":     0,
			"output_video_codec":    "",
			"output_audio_codec":    "",
			"output_video_bit_rate": 0,
			"error_message":         "",
			"warning_message":       "",
			"started_at":            nil,
			"finished_at":           nil,
		}).Error; err != nil {
		return err
	}
	transcodeManager.notify()
	return nil
}

func (VideoTranscode) ResetBatch(request VideoTranscodeIDsRequest) (*VideoTranscodeResetResult, error) {
	if len(request.IDs) == 0 {
		return nil, errors.New("请选择需要重置的任务")
	}
	var tasks []models.VideoTranscodeTask
	if err := core.DBS().Where("id IN ? AND status IN ?", request.IDs, []string{
		models.VideoTranscodeStatusSuccess,
		models.VideoTranscodeStatusFailed,
		models.VideoTranscodeStatusCancelled,
		models.VideoTranscodeStatusInterrupted,
	}).Find(&tasks).Error; err != nil {
		return nil, err
	}
	result := &VideoTranscodeResetResult{Skipped: len(request.IDs) - len(tasks)}
	for _, task := range tasks {
		current, err := (models.ResourcesDramaSeries{}).Info(core.DBS(), task.DramaSeriesID)
		if err != nil {
			result.Skipped++
			continue
		}
		stat, err := os.Stat(current.Src)
		if err != nil || stat.IsDir() {
			result.Skipped++
			continue
		}
		if isVideoTranscodeTemporaryPathForTask(task) {
			_ = os.Remove(task.TemporaryPath)
		}
		if err := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"source_path":           current.Src,
				"source_size":           stat.Size(),
				"source_modified_time":  stat.ModTime().UnixMilli(),
				"source_duration":       0,
				"source_width":          0,
				"source_height":         0,
				"source_frame_rate":     0,
				"source_video_codec":    "",
				"source_audio_codec":    "",
				"source_video_bit_rate": 0,
				"status":                models.VideoTranscodeStatusDraft,
				"progress":              0,
				"processed_seconds":     0,
				"speed":                 "",
				"temporary_path":        "",
				"output_path":           "",
				"backup_path":           "",
				"output_size":           0,
				"output_duration":       0,
				"output_width":          0,
				"output_height":         0,
				"output_frame_rate":     0,
				"output_video_codec":    "",
				"output_audio_codec":    "",
				"output_video_bit_rate": 0,
				"error_message":         "",
				"warning_message":       "",
				"started_at":            nil,
				"finished_at":           nil,
			}).Error; err != nil {
			result.Skipped++
			continue
		}
		result.Reset++
	}
	return result, nil
}

func isVideoTranscodeTemporaryPath(path, taskID string) bool {
	name := filepath.Base(path)
	return path != "" &&
		strings.Contains(name, ".cmtranscode-"+taskID) &&
		strings.HasSuffix(name, ".part")
}

func isVideoTranscodeTemporaryPathForTask(task models.VideoTranscodeTask) bool {
	return isVideoTranscodeTemporaryPath(task.TemporaryPath, task.ID) &&
		filepath.Clean(filepath.Dir(task.TemporaryPath)) == filepath.Clean(filepath.Dir(task.SourcePath))
}

func (VideoTranscode) Pause() {
	transcodeManager.mu.Lock()
	transcodeManager.paused = true
	transcodeManager.mu.Unlock()
}

func (VideoTranscode) Resume() {
	transcodeManager.mu.Lock()
	transcodeManager.paused = false
	transcodeManager.mu.Unlock()
	transcodeManager.notify()
}

func (VideoTranscode) QueueStatus() map[string]interface{} {
	transcodeManager.mu.Lock()
	defer transcodeManager.mu.Unlock()
	return map[string]interface{}{"paused": transcodeManager.paused, "currentId": transcodeManager.currentID}
}

func (VideoTranscode) Cancel(id string) error {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return err
	}
	if task.Status == models.VideoTranscodeStatusReplacing ||
		task.Status == models.VideoTranscodeStatusRefreshingMetadata {
		return errors.New("文件正在替换或刷新元数据，当前阶段不可取消")
	}
	transcodeManager.mu.Lock()
	if transcodeManager.currentID == id && transcodeManager.cancel != nil {
		transcodeManager.cancel()
		transcodeManager.mu.Unlock()
		return nil
	}
	transcodeManager.mu.Unlock()
	return core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", id).
		Where("status IN ?", []string{models.VideoTranscodeStatusDraft, models.VideoTranscodeStatusQueued}).
		Update("status", models.VideoTranscodeStatusCancelled).Error
}

func (VideoTranscode) Delete(id string) error {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return err
	}
	if !isVideoTranscodeTerminal(task.Status) && task.Status != models.VideoTranscodeStatusDraft {
		return errors.New("只能删除草稿或已经结束的任务")
	}
	return core.DBS().Unscoped().Where("id = ?", id).Delete(&models.VideoTranscodeTask{}).Error
}

func (VideoTranscode) DeleteBatch(request VideoTranscodeIDsRequest) (int64, error) {
	if len(request.IDs) == 0 {
		return 0, errors.New("请选择需要移除的任务")
	}
	result := core.DBS().Unscoped().
		Where("id IN ? AND status IN ?", request.IDs, []string{
			models.VideoTranscodeStatusDraft,
			models.VideoTranscodeStatusSuccess,
			models.VideoTranscodeStatusFailed,
			models.VideoTranscodeStatusCancelled,
			models.VideoTranscodeStatusInterrupted,
			models.VideoTranscodeStatusRollbackFailed,
		}).
		Delete(&models.VideoTranscodeTask{})
	return result.RowsAffected, result.Error
}

func isVideoTranscodeTerminal(status string) bool {
	switch status {
	case models.VideoTranscodeStatusSuccess, models.VideoTranscodeStatusFailed,
		models.VideoTranscodeStatusCancelled, models.VideoTranscodeStatusInterrupted,
		models.VideoTranscodeStatusRollbackFailed:
		return true
	default:
		return false
	}
}

func validateVideoTranscodeConfig(config VideoTranscodeConfig) error {
	if config.Container != "mp4" && config.Container != "mkv" {
		return errors.New("输出容器只支持 mp4 或 mkv")
	}
	if config.VideoCodec != "copy" && config.VideoCodec != "h264" && config.VideoCodec != "h265" {
		return errors.New("视频编码只支持 copy、h264 或 h265")
	}
	if config.AudioCodec != "copy" && config.AudioCodec != "aac" {
		return errors.New("音频编码只支持 copy 或 aac")
	}
	if config.VideoCodec == "copy" && (config.ResolutionHeight > 0 || config.FrameRate > 0) {
		return errors.New("复制视频流时不能修改分辨率或帧率")
	}
	if config.FrameRate < 0 || config.FrameRate > 240 {
		return errors.New("视频帧率必须在 0 到 240 之间")
	}
	if config.VideoCodec == "copy" && config.GPUEncoder != "" {
		return errors.New("复制视频流时不能启用 GPU 编码")
	}
	switch config.GPUEncoder {
	case "", "nvenc", "qsv", "amf":
	default:
		return errors.New("GPU 编码器无效")
	}
	if config.QualityMode != "crf" && config.QualityMode != "bitrate" {
		return errors.New("视频质量模式无效")
	}
	if config.CRF < 0 || config.CRF > 51 {
		return errors.New("CRF 必须在 0 到 51 之间")
	}
	if config.VideoBitrateKbps < 1 || config.AudioBitrateKbps < 1 {
		return errors.New("码率必须大于 0")
	}
	switch config.Preset {
	case "ultrafast", "fast", "medium", "slow":
	default:
		return errors.New("编码预设无效")
	}
	return nil
}

func (VideoTranscode) execute(ctx context.Context, task *models.VideoTranscodeTask) {
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
		setVideoTranscodeFailure(task.ID, fmt.Errorf("解析转码参数失败: %w", err))
		return
	}

	current, err := (models.ResourcesDramaSeries{}).Info(core.DBS(), task.DramaSeriesID)
	if err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	stat, err := os.Stat(current.Src)
	if err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	if filepath.Clean(current.Src) != filepath.Clean(task.SourcePath) ||
		stat.Size() != task.SourceSize || stat.ModTime().UnixMilli() != task.SourceModifiedTime {
		setVideoTranscodeFailure(task.ID, errors.New("源文件路径、大小或修改时间已经变化，请重新加入队列"))
		return
	}
	sourceInfo, err := (processorsffmpeg.VideoInfo{}).GetVideoFormatInfo(current.Src)
	if err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	duration := (processorsffmpeg.VideoInfo{}).GetVideoDuration(sourceInfo)
	if duration <= 0 {
		setVideoTranscodeFailure(task.ID, errors.New("无法获得有效的视频时长"))
		return
	}
	primaryVideo := (processorsffmpeg.VideoInfo{}).PrimaryVideoStream(sourceInfo)
	if primaryVideo == nil {
		setVideoTranscodeFailure(task.ID, errors.New("源文件不包含有效视频流"))
		return
	}
	if err := validateVideoTranscodeSourceCompatibility(sourceInfo, config); err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	sourceBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(sourceInfo)
	sourceVideoBitRate, _ := strconv.ParseInt(sourceBasic.BitRate, 10, 64)

	tempPath, outputPath, backupPath := videoTranscodePaths(task, config)
	if filepath.Clean(outputPath) != filepath.Clean(current.Src) {
		if _, err := os.Stat(outputPath); err == nil {
			setVideoTranscodeFailure(task.ID, fmt.Errorf("输出文件已经存在: %s", outputPath))
			return
		}
	}
	runStateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"temporary_path":        tempPath,
		"output_path":           outputPath,
		"backup_path":           backupPath,
		"status":                models.VideoTranscodeStatusTranscoding,
		"source_duration":       duration,
		"source_width":          sourceBasic.Width,
		"source_height":         sourceBasic.Height,
		"source_frame_rate":     sourceBasic.FrameRate,
		"source_video_codec":    sourceBasic.VideoCodec,
		"source_audio_codec":    sourceBasic.AudioCodec,
		"source_video_bit_rate": sourceVideoBitRate,
	})
	if runStateResult.Error != nil || runStateResult.RowsAffected != 1 {
		setVideoTranscodeFailure(task.ID, fmt.Errorf(
			"保存转码运行状态失败: %w",
			videoTranscodeUpdateError(runStateResult),
		))
		return
	}

	_ = os.Remove(tempPath)
	if err := runVideoTranscodeFFmpeg(ctx, task.ID, current.Src, tempPath, duration, primaryVideo.Index, config); err != nil {
		_ = os.Remove(tempPath)
		if errors.Is(err, context.Canceled) {
			setVideoTranscodeCancelled(task.ID)
		} else {
			setVideoTranscodeFailure(task.ID, err)
		}
		return
	}
	verifyStateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Update("status", models.VideoTranscodeStatusVerifying)
	if verifyStateResult.Error != nil || verifyStateResult.RowsAffected != 1 {
		_ = os.Remove(tempPath)
		setVideoTranscodeFailure(task.ID, fmt.Errorf(
			"保存输出校验状态失败: %w",
			videoTranscodeUpdateError(verifyStateResult),
		))
		return
	}
	outputInfo, outputStat, err := verifyVideoTranscodeOutput(tempPath, duration, sourceInfo, config)
	if err != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	outputBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(outputInfo)
	outputVideoBitRate, _ := strconv.ParseInt(outputBasic.BitRate, 10, 64)
	if err := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"output_size":           outputStat.Size(),
			"output_duration":       (processorsffmpeg.VideoInfo{}).GetVideoDuration(outputInfo),
			"output_width":          outputBasic.Width,
			"output_height":         outputBasic.Height,
			"output_frame_rate":     outputBasic.FrameRate,
			"output_video_codec":    outputBasic.VideoCodec,
			"output_audio_codec":    outputBasic.AudioCodec,
			"output_video_bit_rate": outputVideoBitRate,
		}).Error; err != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeFailure(task.ID, fmt.Errorf("保存转码输出信息失败: %w", err))
		return
	}
	if ctx.Err() != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeCancelled(task.ID)
		return
	}
	if err := replaceVideoTranscodeSource(ctx, task, current, tempPath, outputPath, backupPath, outputInfo, outputStat, config); err != nil {
		if errors.Is(err, context.Canceled) {
			_ = os.Remove(tempPath)
			setVideoTranscodeCancelled(task.ID)
		} else {
			setVideoTranscodeFailure(task.ID, err)
		}
		return
	}
}

func videoTranscodePaths(task *models.VideoTranscodeTask, config VideoTranscodeConfig) (string, string, string) {
	sourceExt := filepath.Ext(task.SourcePath)
	base := strings.TrimSuffix(filepath.Base(task.SourcePath), sourceExt)
	ext := "." + config.Container
	if config.Container == "mkv" {
		ext = ".mkv"
	}
	dir := filepath.Dir(task.SourcePath)
	runID := core.GenerateUniqueID()
	output := filepath.Join(dir, base+ext)
	temp := filepath.Join(dir, "."+base+".cmtranscode-"+task.ID+"-"+runID+ext+".part")
	backup := filepath.Join(dir, "."+filepath.Base(task.SourcePath)+".cmbackup-"+task.ID+"-"+runID)
	return temp, output, backup
}

func runVideoTranscodeFFmpeg(
	ctx context.Context,
	taskID, source, output string,
	duration float64,
	primaryVideoIndex int,
	config VideoTranscodeConfig,
) error {
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return err
	}
	args := []string{
		"-hide_banner", "-loglevel", "warning", "-y", "-i", source,
		"-map", fmt.Sprintf("0:%d", primaryVideoIndex), "-map", "0:a?", "-map", "0:s?",
		"-map_metadata", "0", "-map_chapters", "0",
	}
	videoEncoder := ""
	switch config.VideoCodec {
	case "copy":
		args = append(args, "-c:v", "copy")
	case "h264":
		videoEncoder = "libx264"
	case "h265":
		videoEncoder = "libx265"
	}
	if config.VideoCodec != "copy" {
		videoEncoder = hardwareVideoEncoder(config.VideoCodec, config.GPUEncoder, videoEncoder)
		args = append(args, "-c:v", videoEncoder, "-pix_fmt", "yuv420p")
		args = appendVideoEncoderQualityArgs(args, config)
	}
	if config.VideoCodec != "copy" {
		if config.ResolutionHeight > 0 {
			args = append(args, "-vf", fmt.Sprintf("scale=-2:min(ih\\,%d)", config.ResolutionHeight))
		}
		if config.FrameRate > 0 {
			args = append(args, "-r", strconv.FormatFloat(config.FrameRate, 'f', -1, 64))
		}
	}
	if config.AudioCodec == "copy" {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "aac", "-b:a", fmt.Sprintf("%dk", config.AudioBitrateKbps))
	}
	if config.Threads > 0 {
		args = append(args, "-threads", strconv.Itoa(config.Threads))
	}
	if config.Container == "mp4" {
		args = append(args, "-c:s", "mov_text", "-movflags", "+faststart")
	} else {
		args = append(args, "-c:s", "copy")
	}
	muxer := "mp4"
	if config.Container == "mkv" {
		muxer = "matroska"
	}
	args = append(args, "-progress", "pipe:1", "-nostats", "-f", muxer, output)

	cmd := processorsffmpeg.CreateCommandContext(ctx, ffmpegPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	processed := 0.0
	speed := ""
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		switch key {
		case "out_time_us", "out_time_ms":
			raw, _ := strconv.ParseFloat(value, 64)
			processed = raw / 1_000_000
		case "speed":
			speed = value
		case "progress":
			progress := math.Min(99.9, math.Max(0, processed/duration*100))
			core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
				"progress":          progress,
				"processed_seconds": processed,
				"speed":             speed,
			})
		}
	}
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		message := strings.TrimSpace(stderr.String())
		if len(message) > 4000 {
			message = message[len(message)-4000:]
		}
		return fmt.Errorf("FFmpeg 转码失败: %v\n%s", err, message)
	}
	return nil
}

func hardwareVideoEncoder(codec, gpuEncoder, fallback string) string {
	encoders := map[string]map[string]string{
		"nvenc": {"h264": "h264_nvenc", "h265": "hevc_nvenc"},
		"qsv":   {"h264": "h264_qsv", "h265": "hevc_qsv"},
		"amf":   {"h264": "h264_amf", "h265": "hevc_amf"},
	}
	if family, ok := encoders[gpuEncoder]; ok && family[codec] != "" {
		return family[codec]
	}
	return fallback
}

func appendVideoEncoderQualityArgs(args []string, config VideoTranscodeConfig) []string {
	if config.GPUEncoder == "" {
		args = append(args, "-preset", config.Preset)
		if config.QualityMode == "crf" {
			return append(args, "-crf", strconv.Itoa(config.CRF))
		}
		return append(args, "-b:v", fmt.Sprintf("%dk", config.VideoBitrateKbps))
	}
	if config.QualityMode == "bitrate" {
		return append(args, "-b:v", fmt.Sprintf("%dk", config.VideoBitrateKbps))
	}
	switch config.GPUEncoder {
	case "nvenc":
		presets := map[string]string{"ultrafast": "p1", "fast": "p3", "medium": "p4", "slow": "p6"}
		return append(args, "-preset", presets[config.Preset], "-rc", "vbr", "-cq", strconv.Itoa(config.CRF), "-b:v", "0")
	case "qsv":
		presets := map[string]string{"ultrafast": "veryfast", "fast": "fast", "medium": "medium", "slow": "slow"}
		return append(args, "-preset", presets[config.Preset], "-global_quality", strconv.Itoa(config.CRF))
	case "amf":
		qualities := map[string]string{"ultrafast": "speed", "fast": "speed", "medium": "balanced", "slow": "quality"}
		return append(args, "-quality", qualities[config.Preset], "-rc", "cqp", "-qp_i", strconv.Itoa(config.CRF), "-qp_p", strconv.Itoa(config.CRF))
	default:
		return args
	}
}

func countVideoTranscodeStreams(info processorsffmpeg.VideoFormatInfo, codecType string) int {
	count := 0
	for _, stream := range info.Streams {
		if stream.CodecType == codecType {
			count++
		}
	}
	return count
}

func validateVideoTranscodeSourceCompatibility(
	info processorsffmpeg.VideoFormatInfo,
	config VideoTranscodeConfig,
) error {
	if config.Container != "mp4" {
		return nil
	}
	bitmapSubtitleCodecs := map[string]struct{}{
		"hdmv_pgs_subtitle": {},
		"dvd_subtitle":      {},
		"dvb_subtitle":      {},
		"xsub":              {},
	}
	for _, stream := range info.Streams {
		if stream.CodecType != "subtitle" {
			continue
		}
		if _, unsupported := bitmapSubtitleCodecs[stream.CodecName]; unsupported {
			return fmt.Errorf(
				"MP4 不支持将位图字幕 %s 转为文本字幕；请改用 MKV 容器后重试",
				stream.CodecName,
			)
		}
	}
	return nil
}

func videoTranscodeStreamCodecCounts(
	info processorsffmpeg.VideoFormatInfo,
	codecType string,
) map[string]int {
	counts := make(map[string]int)
	for _, stream := range info.Streams {
		if stream.CodecType == codecType {
			counts[stream.CodecName]++
		}
	}
	return counts
}

func equalVideoTranscodeCodecCounts(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for codec, count := range left {
		if right[codec] != count {
			return false
		}
	}
	return true
}

func verifyVideoTranscodeOutput(
	path string,
	sourceDuration float64,
	sourceInfo processorsffmpeg.VideoFormatInfo,
	config VideoTranscodeConfig,
) (processorsffmpeg.VideoFormatInfo, os.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return processorsffmpeg.VideoFormatInfo{}, nil, err
	}
	if stat.Size() <= 0 {
		return processorsffmpeg.VideoFormatInfo{}, nil, errors.New("转码输出为空文件")
	}
	info, err := (processorsffmpeg.VideoInfo{}).GetVideoFormatInfo(path)
	if err != nil {
		return info, nil, fmt.Errorf("转码输出无法通过 FFprobe 校验: %w", err)
	}
	if err := validateVideoTranscodeOutputInfo(sourceInfo, info, sourceDuration, config); err != nil {
		return info, nil, err
	}
	return info, stat, nil
}

func validateVideoTranscodeOutputInfo(
	sourceInfo, outputInfo processorsffmpeg.VideoFormatInfo,
	sourceDuration float64,
	config VideoTranscodeConfig,
) error {
	video := (processorsffmpeg.VideoInfo{}).PrimaryVideoStream(outputInfo)
	if video == nil {
		return errors.New("转码输出不包含有效视频流")
	}
	outputDuration := (processorsffmpeg.VideoInfo{}).GetVideoDuration(outputInfo)
	tolerance := math.Max(2, sourceDuration*0.01)
	if outputDuration <= 0 || math.Abs(outputDuration-sourceDuration) > tolerance {
		return fmt.Errorf("输出时长异常：源视频 %.3f 秒，输出 %.3f 秒", sourceDuration, outputDuration)
	}
	sourceVideo := (processorsffmpeg.VideoInfo{}).PrimaryVideoStream(sourceInfo)
	if sourceVideo == nil {
		return errors.New("源文件不包含有效视频流，无法校验输出")
	}
	if config.VideoCodec == "copy" && video.CodecName != sourceVideo.CodecName {
		return fmt.Errorf("复制视频流后的编码发生变化：源文件为 %s，输出为 %s", sourceVideo.CodecName, video.CodecName)
	}
	if config.VideoCodec == "h264" && video.CodecName != "h264" {
		return fmt.Errorf("输出视频编码不是预期的 H.264，而是 %s", video.CodecName)
	}
	if config.VideoCodec == "h265" && video.CodecName != "hevc" {
		return fmt.Errorf("输出视频编码不是预期的 H.265，而是 %s", video.CodecName)
	}
	sourceAudioCodecs := videoTranscodeStreamCodecCounts(sourceInfo, "audio")
	outputAudioCodecs := videoTranscodeStreamCodecCounts(outputInfo, "audio")
	if countVideoTranscodeStreams(sourceInfo, "audio") != countVideoTranscodeStreams(outputInfo, "audio") {
		return errors.New("输出音频轨数量与源文件不一致，已拒绝替换源文件")
	}
	if config.AudioCodec == "copy" && !equalVideoTranscodeCodecCounts(sourceAudioCodecs, outputAudioCodecs) {
		return errors.New("复制音频流后的编码或轨道组成发生变化，已拒绝替换源文件")
	}
	if config.AudioCodec == "aac" {
		for codec := range outputAudioCodecs {
			if codec != "aac" {
				return fmt.Errorf("输出音频编码不是预期的 AAC，而是 %s", codec)
			}
		}
	}
	if countVideoTranscodeStreams(sourceInfo, "subtitle") != countVideoTranscodeStreams(outputInfo, "subtitle") {
		return errors.New("输出字幕轨数量与源文件不一致，已拒绝替换源文件")
	}
	if config.Container == "mp4" {
		for codec := range videoTranscodeStreamCodecCounts(outputInfo, "subtitle") {
			if codec != "mov_text" {
				return fmt.Errorf("MP4 输出字幕编码不是预期的 mov_text，而是 %s", codec)
			}
		}
	}
	outputBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(outputInfo)
	sourceBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(sourceInfo)
	if config.ResolutionHeight > 0 {
		expectedHeight := config.ResolutionHeight
		if sourceBasic.Height < expectedHeight {
			expectedHeight = sourceBasic.Height
		}
		if math.Abs(float64(outputBasic.Height-expectedHeight)) > 2 {
			return fmt.Errorf("输出高度不是预期值：预期约 %d，实际 %d", expectedHeight, outputBasic.Height)
		}
	}
	if config.FrameRate > 0 && math.Abs(outputBasic.FrameRate-config.FrameRate) > 0.1 {
		return fmt.Errorf("输出帧率不是预期值：预期 %.3f，实际 %.3f", config.FrameRate, outputBasic.FrameRate)
	}
	return nil
}

func replaceVideoTranscodeSource(
	ctx context.Context,
	task *models.VideoTranscodeTask,
	current *models.ResourcesDramaSeries,
	tempPath, outputPath, backupPath string,
	outputInfo processorsffmpeg.VideoFormatInfo,
	outputStat os.FileInfo,
	config VideoTranscodeConfig,
) error {
	lock := mediaPathLocks.get(current.Src)
	lock.Lock()
	defer lock.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	replaceStateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Update("status", models.VideoTranscodeStatusReplacing)
	if replaceStateResult.Error != nil || replaceStateResult.RowsAffected != 1 {
		return fmt.Errorf("保存文件替换状态失败: %w", videoTranscodeUpdateError(replaceStateResult))
	}
	processorscache.CacheVideoInfoLastUse{}.Invalidate(current.Src)
	_ = (processorscache.CacheFileLastUse{}).Invalidate(current.Src)
	if _, err := os.Stat(backupPath); err == nil {
		return fmt.Errorf("备份路径已经存在，已拒绝覆盖: %s", backupPath)
	}
	if err := renameVideoTranscodeFileWithRetry(current.Src, backupPath); err != nil {
		return fmt.Errorf("替换源文件前创建临时回滚文件失败，源视频可能仍被播放器或其他程序占用: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		if rollbackErr := os.Rename(backupPath, current.Src); rollbackErr != nil {
			core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
				Update("status", models.VideoTranscodeStatusRollbackFailed)
			return fmt.Errorf("启用转码文件失败: %v；恢复源文件也失败: %v", err, rollbackErr)
		}
		return fmt.Errorf("启用转码文件失败: %w", err)
	}
	refreshStateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Update("status", models.VideoTranscodeStatusRefreshingMetadata)
	if refreshStateResult.Error != nil || refreshStateResult.RowsAffected != 1 {
		return rollbackVideoTranscodeFiles(task.ID, current.Src, outputPath, backupPath,
			fmt.Errorf("保存元数据刷新状态失败: %w", videoTranscodeUpdateError(refreshStateResult)))
	}
	finalStat, err := os.Stat(outputPath)
	if err != nil {
		return rollbackVideoTranscodeFiles(task.ID, current.Src, outputPath, backupPath, err)
	}
	if outputStat != nil && finalStat.Size() != outputStat.Size() {
		return rollbackVideoTranscodeFiles(task.ID, current.Src, outputPath, backupPath, errors.New("替换后的文件大小与校验结果不一致"))
	}
	if err := saveVideoTranscodeMetadata(task, current, outputPath, outputInfo, finalStat); err != nil {
		return rollbackVideoTranscodeFiles(task.ID, current.Src, outputPath, backupPath, err)
	}

	if !config.KeepBackup {
		if err := os.Remove(backupPath); err != nil {
			warning := "转码和替换已完成，但旧文件备份清理失败: " + err.Error()
			if updateErr := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
				Update("warning_message", warning).Error; updateErr != nil {
				core.LogErr(fmt.Errorf("%s；保存警告信息也失败: %w", warning, updateErr))
			}
		}
	}
	processorscache.CacheVideoInfoLastUse{}.Invalidate(current.Src)
	processorscache.CacheVideoInfoLastUse{}.Invalidate(outputPath)
	return nil
}

func renameVideoTranscodeFileWithRetry(sourcePath, targetPath string) error {
	deadline := time.Now().Add(3 * time.Second)
	var lastErr error
	for {
		if err := os.Rename(sourcePath, targetPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			return lastErr
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func rollbackVideoTranscodeFiles(taskID, sourcePath, outputPath, backupPath string, cause error) error {
	rollbackOutput := outputPath + ".rollback-output"
	_ = os.Remove(rollbackOutput)
	moveOutputErr := os.Rename(outputPath, rollbackOutput)
	restoreErr := os.Rename(backupPath, sourcePath)
	if restoreErr != nil {
		core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", taskID).
			Update("status", models.VideoTranscodeStatusRollbackFailed)
		return fmt.Errorf("刷新转码结果失败: %v；恢复源文件失败: %v", cause, restoreErr)
	}
	if moveOutputErr == nil {
		_ = os.Remove(rollbackOutput)
	}
	return fmt.Errorf("刷新转码结果失败，已恢复源文件: %w", cause)
}

func saveVideoTranscodeMetadata(
	task *models.VideoTranscodeTask,
	current *models.ResourcesDramaSeries,
	outputPath string,
	info processorsffmpeg.VideoFormatInfo,
	stat os.FileInfo,
) error {
	basic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(info)
	duration, err := strconv.ParseFloat(basic.Duration, 64)
	if err != nil || duration <= 0 {
		return errors.New("无法写入有效的视频时长")
	}
	bitRate, _ := strconv.ParseInt(basic.BitRate, 10, 64)
	now := datatype.CustomTime(core.TimeNow())
	metadata := models.ResourcesVideoMetadata{
		DramaSeriesID:    current.ID,
		MetadataVersion:  CurrentVideoMetadataVersion,
		ProbeStatus:      models.VideoMetadataStatusSuccess,
		ProbeTime:        &now,
		FileSize:         stat.Size(),
		FileModifiedTime: stat.ModTime().UnixMilli(),
		Width:            basic.Width,
		Height:           basic.Height,
		FrameRate:        basic.FrameRate,
		FrameRateRaw:     basic.FrameRateRaw,
		VideoCodec:       basic.VideoCodec,
		VideoProfile:     basic.VideoProfile,
		PixelFormat:      basic.PixelFormat,
		BitDepth:         basic.BitDepth,
		VideoBitRate:     bitRate,
		ContainerFormat:  basic.ContainerFormat,
		AudioCodec:       basic.AudioCodec,
		AudioChannels:    basic.AudioChannels,
		AudioSampleRate:  basic.AudioSampleRate,
	}
	return core.DBS().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ResourcesDramaSeries{}).Where("id = ?", current.ID).Updates(map[string]interface{}{
			"src":                 outputPath,
			"durationSeconds":     int(duration),
			"durationProbeStatus": models.DurationProbeStatusSuccess,
			"durationProbeTime":   &now,
			"m3u8BuilderStatus":   false,
			"m3u8BuilderTime":     nil,
		}).Error; err != nil {
			return err
		}
		if err := (models.ResourcesVideoMetadata{}).Upsert(tx, &metadata); err != nil {
			return err
		}
		if err := (models.VideoFingerprint{}).RebuildByDramaSeriesIDs(
			tx, []string{current.ID}, core.GenerateUniqueID,
		); err != nil {
			return err
		}
		finished := datatype.CustomTime(core.TimeNow())
		taskResult := tx.Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":          models.VideoTranscodeStatusSuccess,
			"progress":        100,
			"output_size":     stat.Size(),
			"warning_message": "",
			"error_message":   "",
			"finished_at":     &finished,
		})
		if taskResult.Error != nil || taskResult.RowsAffected != 1 {
			return videoTranscodeUpdateError(taskResult)
		}
		return nil
	})
}

func videoTranscodeUpdateError(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	return fmt.Errorf("任务记录不存在或状态未能持久化（影响行数: %d）", result.RowsAffected)
}

func setVideoTranscodeFailure(id string, err error) {
	finished := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id = ? AND status <> ?", id, models.VideoTranscodeStatusRollbackFailed).
		Updates(map[string]interface{}{
			"status":        models.VideoTranscodeStatusFailed,
			"error_message": err.Error(),
			"finished_at":   &finished,
		})
}

func setVideoTranscodeCancelled(id string) {
	finished := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      models.VideoTranscodeStatusCancelled,
		"finished_at": &finished,
	})
}
