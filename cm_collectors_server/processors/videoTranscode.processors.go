package processors

import (
	"bufio"
	"bytes"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/internal/maintenance"
	"cm_collectors_server/models"
	processorscache "cm_collectors_server/processorsCache"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type VideoTranscodeConfig struct {
	Container        string  `json:"container"`
	OutputMode       string  `json:"outputMode"`
	OutputFileName   string  `json:"outputFileName"`
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

type VideoEditSegment struct {
	ID         string           `json:"id"`
	Start      float64          `json:"start"`
	End        float64          `json:"end"`
	Transition *VideoTransition `json:"transition,omitempty"`
}

type VideoTransition struct {
	Type      string  `json:"type"`
	Duration  float64 `json:"duration"`
	AudioFade bool    `json:"audioFade,omitempty"`
}

type VideoEditPlan struct {
	Version  int                `json:"version"`
	Segments []VideoEditSegment `json:"segments"`
}

type VideoTranscodeEditPlanRequest struct {
	Plan           VideoEditPlan `json:"plan"`
	OutputMode     string        `json:"outputMode"`
	OutputFileName string        `json:"outputFileName"`
}

type VideoTranscodeTransitionPreviewRequest struct {
	Left  VideoEditSegment `json:"left"`
	Right VideoEditSegment `json:"right"`
}

type VideoTranscodeThumbnailBatchRequest struct {
	Times []float64 `json:"times"`
}

type VideoTranscodeTimelineThumbnail struct {
	Time float64 `json:"time"`
	URL  string  `json:"url"`
}

func DefaultVideoTranscodeConfig() VideoTranscodeConfig {
	return VideoTranscodeConfig{
		Container:        "mp4",
		OutputMode:       "new_file",
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
	Config              VideoTranscodeConfig `json:"config"`
	EditPlan            VideoEditPlan        `json:"editPlan"`
	FilesBasesID        string               `json:"filesBasesId"`
	CoverPoster         string               `json:"coverPoster"`
	CanRetryReplacement bool                 `json:"canRetryReplacement"`
	CanSaveAsNewFile    bool                 `json:"canSaveAsNewFile"`
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

type VideoTranscodeStartRequest struct {
	IDs       []string `json:"ids"`
	EnableGPU bool     `json:"enableGpu"`
}

type VideoTranscodeStartResult struct {
	Queued     int `json:"queued"`
	EnabledGPU int `json:"enabledGpu"`
}

type VideoTranscodeUpdateConfigRequest struct {
	IDs    []string             `json:"ids"`
	Config VideoTranscodeConfig `json:"config"`
}

type VideoTranscodeResetResult struct {
	Reset   int `json:"reset"`
	Skipped int `json:"skipped"`
}

type VideoTranscodeEditPlanResult struct {
	Plan           VideoEditPlan        `json:"plan"`
	EditedDuration float64              `json:"editedDuration"`
	HasEdits       bool                 `json:"hasEdits"`
	Config         VideoTranscodeConfig `json:"config"`
	ConfigAdjusted bool                 `json:"configAdjusted"`
}

const videoTranscodeVerificationTimeout = 2 * time.Minute
const videoTranscodeReplacementLockTimeout = 5 * time.Second
const videoTranscodeReplacementLockAttempts = 3
const videoTransitionMinDuration = 0.1
const videoTransitionMaxDuration = 3.0
const videoTransitionSegmentMargin = 0.05
const videoTransitionPreviewCacheMaxAge = 7 * 24 * time.Hour
const videoTransitionPreviewCacheMaxBytes int64 = 512 << 20

var videoTransitionPreviewMu sync.Mutex
var videoTransitionLegacyCacheCleanupOnce sync.Once

type videoTransitionPreviewCacheFile struct {
	path    string
	size    int64
	modTime time.Time
}

func applicationCachePath(parts ...string) (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取软件根目录失败: %w", err)
	}
	cacheRoot := filepath.Join(filepath.Dir(executablePath), "runtime", "cache")
	return filepath.Join(append([]string{cacheRoot}, parts...)...), nil
}

func cleanupLegacyVideoTransitionPreviewCache(currentCacheRoot string) {
	videoTransitionLegacyCacheCleanupOnce.Do(func() {
		userCacheRoot, err := os.UserCacheDir()
		if err != nil {
			return
		}
		legacyRoot := filepath.Join(userCacheRoot, "cm_collectors_3", "transition-previews")
		if filepath.Clean(legacyRoot) == filepath.Clean(currentCacheRoot) {
			return
		}
		entries, err := os.ReadDir(legacyRoot)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".webm") {
				continue
			}
			_ = os.Remove(filepath.Join(legacyRoot, entry.Name()))
		}
		_ = os.Remove(legacyRoot)
	})
}

type videoTranscodeVerifyResult struct {
	Info processorsffmpeg.VideoFormatInfo
	Stat os.FileInfo
	Err  error
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

type mediaReadRegistry struct {
	mu      sync.Mutex
	nextID  uint64
	readers map[string]map[uint64]context.CancelFunc
	blocked map[string]int
}

var activeMediaReads = &mediaReadRegistry{
	readers: make(map[string]map[uint64]context.CancelFunc),
	blocked: make(map[string]int),
}

func mediaPathKey(path string) string {
	cleaned := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(cleaned)
	}
	return cleaned
}

func (r *mediaPathLockRegistry) get(path string) *sync.RWMutex {
	cleaned := mediaPathKey(path)
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

// registerMediaRead 记录仍在读取媒体文件的 HTTP 请求。替换源文件前会取消这些请求，
// 让 Windows 上的播放句柄及时关闭，避免替换流程永久等待读锁。
func registerMediaRead(path string, parent context.Context) (context.Context, func()) {
	cleaned := mediaPathKey(path)
	ctx, cancel := context.WithCancel(parent)
	activeMediaReads.mu.Lock()
	if activeMediaReads.blocked[cleaned] > 0 {
		activeMediaReads.mu.Unlock()
		cancel()
		return ctx, cancel
	}
	activeMediaReads.nextID++
	id := activeMediaReads.nextID
	readers := activeMediaReads.readers[cleaned]
	if readers == nil {
		readers = make(map[uint64]context.CancelFunc)
		activeMediaReads.readers[cleaned] = readers
	}
	readers[id] = cancel
	activeMediaReads.mu.Unlock()

	var once sync.Once
	release := func() {
		once.Do(func() {
			cancel()
			activeMediaReads.mu.Lock()
			delete(activeMediaReads.readers[cleaned], id)
			if len(activeMediaReads.readers[cleaned]) == 0 {
				delete(activeMediaReads.readers, cleaned)
			}
			activeMediaReads.mu.Unlock()
		})
	}
	return ctx, release
}

func cancelMediaReads(path string) {
	cleaned := mediaPathKey(path)
	activeMediaReads.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(activeMediaReads.readers[cleaned]))
	for _, cancel := range activeMediaReads.readers[cleaned] {
		cancels = append(cancels, cancel)
	}
	activeMediaReads.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

// blockMediaReads 在替换期间阻止新的播放请求再次取得读锁，并取消已经存在的请求。
func blockMediaReads(path string) func() {
	cleaned := mediaPathKey(path)
	activeMediaReads.mu.Lock()
	activeMediaReads.blocked[cleaned]++
	cancels := make([]context.CancelFunc, 0, len(activeMediaReads.readers[cleaned]))
	for _, cancel := range activeMediaReads.readers[cleaned] {
		cancels = append(cancels, cancel)
	}
	activeMediaReads.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			activeMediaReads.mu.Lock()
			activeMediaReads.blocked[cleaned]--
			if activeMediaReads.blocked[cleaned] <= 0 {
				delete(activeMediaReads.blocked, cleaned)
			}
			activeMediaReads.mu.Unlock()
		})
	}
}

func lockMediaForReplacement(ctx context.Context, path string, timeout time.Duration) (func(), error) {
	lock := mediaPathLocks.get(path)
	deadline := time.Now().Add(timeout)
	for {
		if lock.TryLock() {
			return lock.Unlock, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("等待源视频播放请求释放超过 %s", timeout)
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func lockMediaForReplacementWithRetry(ctx context.Context, path string) (func(), error) {
	var lastErr error
	for attempt := 1; attempt <= videoTranscodeReplacementLockAttempts; attempt++ {
		unlock, err := lockMediaForReplacement(ctx, path, videoTranscodeReplacementLockTimeout)
		if err == nil {
			return unlock, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		cancelMediaReads(path)
		if attempt < videoTranscodeReplacementLockAttempts {
			timer := time.NewTimer(500 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return nil, fmt.Errorf("自动重试 %d 次后仍无法释放源视频读取句柄: %w",
		videoTranscodeReplacementLockAttempts, lastErr)
}

type VideoTranscode struct{}

func InitVideoTranscode() {
	db := core.DBS()
	interruptedAt := datatype.CustomTime(core.TimeNow())
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
			"finished_at":   &interruptedAt,
		})
	recoverCriticalVideoTranscodeTasks()
	go transcodeManager.loop()
	// 服务重启后只主动检查一次，恢复数据库中已经排队的任务。
	transcodeManager.notify()
}

func recoverCriticalVideoTranscodeTasks() {
	var tasks []models.VideoTranscodeTask
	if err := core.DBS().Where("status IN ?", []string{
		models.VideoTranscodeStatusSavingOutput,
		models.VideoTranscodeStatusRetryingReplace,
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
	if task.Status == models.VideoTranscodeStatusSavingOutput {
		recoverVideoTranscodeNewOutput(task)
		return
	}
	if task.Status == models.VideoTranscodeStatusRetryingReplace &&
		(task.BackupPath == "" || !fileExistsForTranscode(task.BackupPath)) {
		setVideoTranscodeFailure(task.ID, errors.New("服务在重试替换前退出，已保留校验成功的临时成片，可再次手动重试替换"))
		return
	}
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

func recoverVideoTranscodeNewOutput(task models.VideoTranscodeTask) {
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil || config.OutputMode != "new_file" {
		markVideoTranscodeInterrupted(task.ID, "另存新文件任务的参数无法恢复，请重新执行")
		return
	}
	lock := mediaPathLocks.get(task.OutputPath)
	lock.Lock()
	defer lock.Unlock()
	if !fileExistsForTranscode(task.OutputPath) {
		if !isVideoTranscodeTemporaryPathForTask(task) || !fileExistsForTranscode(task.TemporaryPath) {
			markVideoTranscodeInterrupted(task.ID, "服务在保存新文件时退出，未找到可恢复的输出文件")
			return
		}
		if err := publishVideoTranscodeOutput(task.TemporaryPath, task.OutputPath); err != nil {
			markVideoTranscodeInterrupted(task.ID, "恢复另存的新文件失败: "+err.Error())
			return
		}
	}
	stat, err := os.Stat(task.OutputPath)
	if err != nil || (task.OutputSize > 0 && stat.Size() != task.OutputSize) {
		markVideoTranscodeInterrupted(task.ID, "另存的新文件校验失败，请人工检查输出文件")
		return
	}
	if err := completeVideoTranscodeNewOutput(task.ID, stat.Size(), "服务重启后已恢复另存的新文件"); err != nil {
		markVideoTranscodeInterrupted(task.ID, "恢复另存的新文件状态失败: "+err.Error())
		return
	}
	if isVideoTranscodeTemporaryPathForTask(task) {
		_ = os.Remove(task.TemporaryPath)
	}
	processorscache.CacheVideoInfoLastUse{}.Invalidate(task.OutputPath)
}

func fileExistsForTranscode(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func markVideoTranscodeInterrupted(id, message string) {
	finished := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.VideoTranscodeStatusInterrupted,
		"error_message": message,
		"finished_at":   &finished,
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
	if !maintenance.Begin() {
		m.scheduleRetry()
		return
	}
	defer maintenance.End()
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

	maintenance.Go(func() {
		defer func() {
			cancel()
			m.mu.Lock()
			m.currentID = ""
			m.cancel = nil
			m.mu.Unlock()
			m.notify()
		}()
		(VideoTranscode{}).execute(ctx, task)
	})
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
		editPlan := VideoEditPlan{Version: 1}
		if task.EditPlanJsonData != "" {
			_ = json.Unmarshal([]byte(task.EditPlanJsonData), &editPlan)
		}
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
		if task.OutputWidth <= 0 && config.OutputMode == "replace" &&
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
			VideoTranscodeTask:  task,
			Config:              config,
			EditPlan:            editPlan,
			FilesBasesID:        image.FilesBasesID,
			CoverPoster:         image.CoverPoster,
			CanRetryReplacement: canRetryVideoTranscodeReplacement(task, config),
			CanSaveAsNewFile:    canReuseVerifiedVideoTranscodeOutput(task),
		})
	}
	return result, nil
}

func canRetryVideoTranscodeReplacement(task models.VideoTranscodeTask, config VideoTranscodeConfig) bool {
	if config.OutputMode != "replace" || !canReuseVerifiedVideoTranscodeOutput(task) {
		return false
	}
	if !strings.Contains(task.ErrorMessage, "释放源视频读取句柄失败") &&
		!strings.Contains(task.ErrorMessage, "创建临时回滚文件失败") &&
		!strings.Contains(task.ErrorMessage, "可再次手动重试替换") {
		return false
	}
	if task.BackupPath != "" {
		if _, err := os.Stat(task.BackupPath); err == nil {
			return false
		}
	}
	return true
}

func canReuseVerifiedVideoTranscodeOutput(task models.VideoTranscodeTask) bool {
	if task.Status != models.VideoTranscodeStatusFailed ||
		!isVideoTranscodeTemporaryPathForTask(task) || task.OutputSize <= 0 {
		return false
	}
	tempStat, err := os.Stat(task.TemporaryPath)
	if err != nil || tempStat.IsDir() || tempStat.Size() != task.OutputSize {
		return false
	}
	_, err = os.Stat(task.SourcePath)
	return err == nil
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
	if request.Config.VideoCodec == "copy" || request.Config.AudioCodec == "copy" {
		query := core.DBS().Model(&models.VideoTranscodeTask{}).
			Where("status = ? AND edit_plan_json_data <> ''", models.VideoTranscodeStatusDraft)
		if len(request.IDs) > 0 {
			query = query.Where("id IN ?", request.IDs)
		}
		var editedTaskCount int64
		if err := query.Count(&editedTaskCount).Error; err != nil {
			return err
		}
		if editedTaskCount > 0 {
			return errors.New("剪辑任务不能复制音视频流，请选择 H.264/H.265 和 AAC；也可以只选择未剪辑任务应用复制参数")
		}
	}
	query := core.DBS().Where("status = ?", models.VideoTranscodeStatusDraft)
	if len(request.IDs) > 0 {
		query = query.Where("id IN ?", request.IDs)
	}
	var tasks []models.VideoTranscodeTask
	if err := query.Find(&tasks).Error; err != nil {
		return err
	}
	return core.DBS().Transaction(func(tx *gorm.DB) error {
		for _, task := range tasks {
			existing := DefaultVideoTranscodeConfig()
			if err := json.Unmarshal([]byte(task.ConfigJsonData), &existing); err != nil {
				return fmt.Errorf("解析任务 %s 的转码参数失败: %w", task.ID, err)
			}
			updated := request.Config
			updated.OutputMode = existing.OutputMode
			updated.OutputFileName = existing.OutputFileName
			if updated.OutputMode != "new_file" {
				updated.OutputMode = "replace"
				updated.OutputFileName = ""
			}
			if updated.OutputMode == "new_file" {
				updated.KeepBackup = false
			}
			data, err := json.Marshal(updated)
			if err != nil {
				return err
			}
			result := tx.Model(&models.VideoTranscodeTask{}).
				Where("id = ? AND status = ?", task.ID, models.VideoTranscodeStatusDraft).
				Update("config_json_data", string(data))
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

func (VideoTranscode) SaveEditPlan(
	id string,
	request VideoTranscodeEditPlanRequest,
) (*VideoTranscodeEditPlanResult, error) {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return nil, err
	}
	if task.Status != models.VideoTranscodeStatusDraft {
		return nil, errors.New("只能编辑待转码任务")
	}
	sourceInfo, duration, err := videoTranscodeTaskPreciseSourceInfo(task)
	if err != nil {
		return nil, err
	}
	plan, editedDuration, hasEdits, err := normalizeVideoEditPlan(request.Plan, duration)
	if err != nil {
		return nil, err
	}
	if hasEdits && countVideoTranscodeStreams(sourceInfo, "subtitle") > 0 {
		return nil, errors.New("第一版视频剪辑暂不支持包含字幕轨的源文件")
	}
	planJSON := ""
	if hasEdits {
		data, marshalErr := json.Marshal(plan)
		if marshalErr != nil {
			return nil, marshalErr
		}
		planJSON = string(data)
	}
	storedEditedDuration := 0.0
	if hasEdits {
		storedEditedDuration = editedDuration
	}
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
		return nil, fmt.Errorf("解析转码参数失败: %w", err)
	}
	configAdjusted := false
	if hasEdits {
		config, configAdjusted = videoTranscodeConfigForEdit(config)
	}
	if request.OutputMode != "" {
		config.OutputMode = request.OutputMode
	}
	if config.OutputMode == "new_file" {
		config.OutputFileName = strings.TrimSpace(request.OutputFileName)
		config.KeepBackup = false
	} else {
		config.OutputFileName = ""
	}
	if err := validateVideoTranscodeConfig(config); err != nil {
		return nil, fmt.Errorf("转码参数无效: %w", err)
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	result := core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id = ? AND status = ?", id, models.VideoTranscodeStatusDraft).
		Updates(map[string]interface{}{
			"edit_plan_json_data": planJSON,
			"edited_duration":     storedEditedDuration,
			"source_duration":     duration,
			"config_json_data":    string(configJSON),
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, errors.New("任务状态已经变化，剪辑方案未保存")
	}
	return &VideoTranscodeEditPlanResult{
		Plan:           plan,
		EditedDuration: editedDuration,
		HasEdits:       hasEdits,
		Config:         config,
		ConfigAdjusted: configAdjusted,
	}, nil
}

func videoTranscodeConfigForEdit(config VideoTranscodeConfig) (VideoTranscodeConfig, bool) {
	adjusted := false
	if config.VideoCodec != "h264" && config.VideoCodec != "h265" {
		config.VideoCodec = "h264"
		config.GPUEncoder = ""
		adjusted = true
	}
	if config.AudioCodec != "aac" {
		config.AudioCodec = "aac"
		adjusted = true
	}
	return config, adjusted
}

func (VideoTranscode) Thumbnail(ctx context.Context, id string, at float64) ([]byte, error) {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return nil, err
	}
	if task.Status != models.VideoTranscodeStatusDraft {
		return nil, errors.New("只能为待转码任务生成剪辑缩略图")
	}
	duration, err := videoTranscodeTaskSourceDuration(task)
	if err != nil {
		return nil, err
	}
	if math.IsNaN(at) || math.IsInf(at, 0) {
		return nil, errors.New("缩略图时间点无效")
	}
	at = math.Max(0, math.Min(at, math.Max(0, duration-0.01)))
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return nil, err
	}
	readCtx, releaseRead := registerMediaRead(task.SourcePath, ctx)
	defer releaseRead()
	if readCtx.Err() != nil {
		return nil, errors.New("源视频正在替换，请稍后重试")
	}
	unlockMedia := lockMediaForRead(task.SourcePath)
	defer unlockMedia()
	return generateVideoTranscodeThumbnail(readCtx, ffmpegPath, task.SourcePath, at)
}

func generateVideoTranscodeThumbnail(ctx context.Context, ffmpegPath, sourcePath string, at float64) ([]byte, error) {
	cmd := processorsffmpeg.CreateCommandContext(ctx, ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-ss", strconv.FormatFloat(at, 'f', 3, 64),
		"-i", sourcePath,
		"-frames:v", "1", "-an", "-sn",
		"-vf", "scale=240:-2", "-q:v", "4",
		"-f", "image2pipe", "-vcodec", "mjpeg", "pipe:1",
	)
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("生成剪辑缩略图失败: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("生成的剪辑缩略图为空")
	}
	return data, nil
}

func (VideoTranscode) ThumbnailBatch(
	ctx context.Context,
	id string,
	request VideoTranscodeThumbnailBatchRequest,
) ([]VideoTranscodeTimelineThumbnail, error) {
	if len(request.Times) == 0 || len(request.Times) > 160 {
		return nil, errors.New("缩略图采样数量必须在 1～160 之间")
	}
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return nil, err
	}
	if task.Status != models.VideoTranscodeStatusDraft {
		return nil, errors.New("只能为待转码任务生成剪辑缩略图")
	}
	duration, err := videoTranscodeTaskSourceDuration(task)
	if err != nil {
		return nil, err
	}
	times := make([]float64, len(request.Times))
	for index, at := range request.Times {
		if math.IsNaN(at) || math.IsInf(at, 0) {
			return nil, errors.New("缩略图时间点无效")
		}
		times[index] = math.Max(0, math.Min(at, math.Max(0, duration-0.01)))
	}
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return nil, err
	}
	readCtx, releaseRead := registerMediaRead(task.SourcePath, ctx)
	defer releaseRead()
	if readCtx.Err() != nil {
		return nil, errors.New("源视频正在替换，请稍后重试")
	}
	unlockMedia := lockMediaForRead(task.SourcePath)
	defer unlockMedia()
	result := make([]VideoTranscodeTimelineThumbnail, len(times))
	jobs := make(chan int)
	workerCount := min(4, len(times))
	var workers sync.WaitGroup
	var firstErr error
	successCount := 0
	var errMu sync.Mutex
	workers.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer workers.Done()
			for index := range jobs {
				if readCtx.Err() != nil {
					continue
				}
				data, thumbnailErr := generateVideoTranscodeThumbnail(
					readCtx, ffmpegPath, task.SourcePath, times[index],
				)
				if thumbnailErr != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = thumbnailErr
					}
					errMu.Unlock()
					continue
				}
				result[index] = VideoTranscodeTimelineThumbnail{
					Time: times[index],
					URL:  "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data),
				}
				errMu.Lock()
				successCount++
				errMu.Unlock()
			}
		}()
	}
	for index := range times {
		if readCtx.Err() != nil {
			break
		}
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	if readCtx.Err() != nil {
		return nil, readCtx.Err()
	}
	if successCount == 0 && firstErr != nil {
		return nil, fmt.Errorf("批量生成时间轴缩略图失败: %w", firstErr)
	}
	return result, nil
}

func cleanupVideoTransitionPreviewCache(
	cacheRoot, preservePath string,
	now time.Time,
	maxAge time.Duration,
	maxBytes int64,
) {
	entries, err := os.ReadDir(cacheRoot)
	if err != nil {
		return
	}
	files := make([]videoTransitionPreviewCacheFile, 0, len(entries))
	totalBytes := int64(0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(cacheRoot, entry.Name())
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".tmp.webm") {
			if path != preservePath && now.Sub(info.ModTime()) > time.Hour {
				_ = os.Remove(path)
			}
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".webm") {
			continue
		}
		if path != preservePath && now.Sub(info.ModTime()) > maxAge {
			if os.Remove(path) == nil {
				continue
			}
		}
		files = append(files, videoTransitionPreviewCacheFile{
			path: path, size: info.Size(), modTime: info.ModTime(),
		})
		totalBytes += info.Size()
	}
	if totalBytes <= maxBytes {
		return
	}
	sort.Slice(files, func(left, right int) bool {
		return files[left].modTime.Before(files[right].modTime)
	})
	for _, file := range files {
		if totalBytes <= maxBytes {
			break
		}
		if file.path == preservePath {
			continue
		}
		if os.Remove(file.path) == nil {
			totalBytes -= file.size
		}
	}
}

// TransitionPreview 生成只覆盖切点附近的浏览器兼容 WebM 预览。
func (VideoTranscode) TransitionPreview(
	ctx context.Context,
	id string,
	request VideoTranscodeTransitionPreviewRequest,
) (string, error) {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return "", err
	}
	if task.Status != models.VideoTranscodeStatusDraft {
		return "", errors.New("只能预览待转码任务的转场")
	}
	duration, err := videoTranscodeTaskSourceDuration(task)
	if err != nil {
		return "", err
	}
	plan, _, _, err := normalizeVideoEditPlan(VideoEditPlan{
		Version:  1,
		Segments: []VideoEditSegment{request.Left, request.Right},
	}, duration)
	if err != nil {
		return "", err
	}
	left, right := plan.Segments[0], plan.Segments[1]
	if left.Transition == nil {
		return "", errors.New("请先选择转场效果")
	}
	stat, err := os.Stat(task.SourcePath)
	if err != nil {
		return "", err
	}
	identity := fmt.Sprintf(
		"%s|%d|%d|transition-v5-colored-fade|%.6f|%.6f|%.6f|%.6f|%s|%.6f|%t",
		mediaPathKey(task.SourcePath), stat.Size(), stat.ModTime().UnixNano(),
		left.Start, left.End, right.Start, right.End,
		left.Transition.Type, left.Transition.Duration, left.Transition.AudioFade,
	)
	cacheRoot, err := applicationCachePath("video-transcode", "transition-previews")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return "", fmt.Errorf("创建转场预览缓存目录失败: %w", err)
	}
	cleanupLegacyVideoTransitionPreviewCache(cacheRoot)
	previewPath := filepath.Join(cacheRoot, fmt.Sprintf("%x.webm", sha256.Sum256([]byte(identity))))
	videoTransitionPreviewMu.Lock()
	defer videoTransitionPreviewMu.Unlock()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	cleanupVideoTransitionPreviewCache(
		cacheRoot, previewPath, time.Now(),
		videoTransitionPreviewCacheMaxAge, videoTransitionPreviewCacheMaxBytes,
	)
	if cached, statErr := os.Stat(previewPath); statErr == nil && cached.Size() > 0 {
		now := time.Now()
		_ = os.Chtimes(previewPath, now, now)
		return previewPath, nil
	}
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return "", err
	}
	sourceInfo, _, err := videoTranscodeTaskPreciseSourceInfo(task)
	if err != nil {
		return "", err
	}
	primaryVideo := (processorsffmpeg.VideoInfo{}).PrimaryVideoStream(sourceInfo)
	if primaryVideo == nil {
		return "", errors.New("源文件不包含有效视频流")
	}
	var audioIndex = -1
	for _, stream := range sourceInfo.Streams {
		if stream.CodecType == "audio" {
			audioIndex = stream.Index
			break
		}
	}
	halfDuration := left.Transition.Duration / 2
	lead := math.Min(0.75, math.Max(0, left.End-left.Start-halfDuration))
	tail := math.Min(0.75, math.Max(0, right.End-right.Start-halfDuration))
	leftStart := left.End - halfDuration - lead
	rightEnd := right.Start + halfDuration + tail
	videoNormalize := "scale=640:-2:force_original_aspect_ratio=decrease,setsar=1,fps=24,format=yuv420p,settb=AVTB"
	graph := make([]string, 0, 14)
	leftParts := make([]string, 0, 2)
	if lead > 0.001 {
		graph = append(graph, fmt.Sprintf(
			"[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,%s[lpre]",
			primaryVideo.Index, leftStart, left.End-halfDuration, videoNormalize))
		leftParts = append(leftParts, "[lpre]")
	}
	graph = append(graph,
		fmt.Sprintf("[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,%s,split=2[lreal][lbackdropsrc]",
			primaryVideo.Index, left.End-halfDuration, left.End, videoNormalize),
		fmt.Sprintf("[lbackdropsrc]%s[lbackdrop]", overlayTransitionBackdropFilter(left.Transition.Type)),
		fmt.Sprintf("[lreal][lbackdrop]xfade=transition=%s:duration=%.6f:offset=0[ltransition]",
			overlayTransitionFilterType(left.Transition.Type), halfDuration),
	)
	leftParts = append(leftParts, "[ltransition]")
	leftOutput := leftParts[0]
	if len(leftParts) > 1 {
		graph = append(graph, strings.Join(leftParts, "")+"concat=n=2:v=1:a=0[lout]")
		leftOutput = "[lout]"
	}
	graph = append(graph,
		fmt.Sprintf("[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,%s,split=2[rreal][rbackdropsrc]",
			primaryVideo.Index, right.Start, right.Start+halfDuration, videoNormalize),
		fmt.Sprintf("[rbackdropsrc]%s[rbackdrop]", overlayTransitionBackdropFilter(left.Transition.Type)),
		fmt.Sprintf("[rbackdrop][rreal]xfade=transition=%s:duration=%.6f:offset=0[rtransition]",
			overlayTransitionFilterType(left.Transition.Type), halfDuration),
	)
	rightParts := []string{"[rtransition]"}
	if tail > 0.001 {
		graph = append(graph, fmt.Sprintf(
			"[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,%s[rtail]",
			primaryVideo.Index, right.Start+halfDuration, rightEnd, videoNormalize))
		rightParts = append(rightParts, "[rtail]")
	}
	rightOutput := rightParts[0]
	if len(rightParts) > 1 {
		graph = append(graph, strings.Join(rightParts, "")+"concat=n=2:v=1:a=0[rout]")
		rightOutput = "[rout]"
	}
	graph = append(graph, leftOutput+rightOutput+"concat=n=2:v=1:a=0[vout]")
	if audioIndex >= 0 {
		leftAudio := fmt.Sprintf("[0:%d]atrim=start=%.6f:end=%.6f,asetpts=PTS-STARTPTS",
			audioIndex, leftStart, left.End)
		rightAudio := fmt.Sprintf("[0:%d]atrim=start=%.6f:end=%.6f,asetpts=PTS-STARTPTS",
			audioIndex, right.Start, rightEnd)
		if left.Transition.AudioFade {
			leftAudio += fmt.Sprintf(",afade=t=out:st=%.6f:d=%.6f", lead, halfDuration)
			rightAudio += fmt.Sprintf(",afade=t=in:st=0:d=%.6f", halfDuration)
		}
		graph = append(graph, leftAudio+"[la]", rightAudio+"[ra]", "[la][ra]concat=n=2:v=0:a=1[aout]")
	}
	temporaryPath := previewPath + ".tmp.webm"
	_ = os.Remove(temporaryPath)
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", task.SourcePath,
		"-filter_complex", strings.Join(graph, ";"), "-map", "[vout]"}
	if audioIndex >= 0 {
		args = append(args, "-map", "[aout]")
	}
	args = append(args, "-c:v", "libvpx", "-deadline", "realtime", "-cpu-used", "8",
		"-b:v", "900k", "-crf", "34", "-pix_fmt", "yuv420p")
	if audioIndex >= 0 {
		args = append(args, "-c:a", "libopus", "-b:a", "64k", "-ac", "2")
	} else {
		args = append(args, "-an")
	}
	args = append(args, "-f", "webm", temporaryPath)
	readCtx, releaseRead := registerMediaRead(task.SourcePath, ctx)
	defer releaseRead()
	if readCtx.Err() != nil {
		return "", errors.New("源视频正在替换，请稍后重试")
	}
	unlockMedia := lockMediaForRead(task.SourcePath)
	defer unlockMedia()
	cmd := processorsffmpeg.CreateCommandContext(readCtx, ffmpegPath, args...)
	if output, runErr := cmd.CombinedOutput(); runErr != nil {
		_ = os.Remove(temporaryPath)
		message := strings.TrimSpace(string(output))
		if len(message) > 2000 {
			message = message[len(message)-2000:]
		}
		return "", fmt.Errorf("生成转场预览失败: %v: %s", runErr, message)
	}
	if err := os.Rename(temporaryPath, previewPath); err != nil {
		_ = os.Remove(temporaryPath)
		return "", fmt.Errorf("保存转场预览失败: %w", err)
	}
	cleanupVideoTransitionPreviewCache(
		cacheRoot, previewPath, time.Now(),
		videoTransitionPreviewCacheMaxAge, videoTransitionPreviewCacheMaxBytes,
	)
	return previewPath, nil
}

func videoTranscodeTaskSourceDuration(task *models.VideoTranscodeTask) (float64, error) {
	if err := validateVideoTranscodeTaskSourceSnapshot(task); err != nil {
		return 0, err
	}
	if task.SourceDuration > 0 {
		return task.SourceDuration, nil
	}
	duration, err := videoTranscodeTaskPreciseSourceDuration(task)
	if err == nil {
		task.SourceDuration = duration
		_ = core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
			Update("source_duration", duration).Error
	}
	return duration, err
}

func videoTranscodeTaskPreciseSourceDuration(task *models.VideoTranscodeTask) (float64, error) {
	_, duration, err := videoTranscodeTaskPreciseSourceInfo(task)
	return duration, err
}

func videoTranscodeTaskPreciseSourceInfo(
	task *models.VideoTranscodeTask,
) (processorsffmpeg.VideoFormatInfo, float64, error) {
	if err := validateVideoTranscodeTaskSourceSnapshot(task); err != nil {
		return processorsffmpeg.VideoFormatInfo{}, 0, err
	}
	info, err := (processorsffmpeg.VideoInfo{}).GetVideoFormatInfo(task.SourcePath)
	if err != nil {
		return processorsffmpeg.VideoFormatInfo{}, 0, err
	}
	duration := (processorsffmpeg.VideoInfo{}).GetVideoDuration(info)
	if duration <= 0 {
		return processorsffmpeg.VideoFormatInfo{}, 0, errors.New("无法获得有效的视频时长")
	}
	return info, duration, nil
}

func validateVideoTranscodeTaskSourceSnapshot(task *models.VideoTranscodeTask) error {
	stat, err := os.Stat(task.SourcePath)
	if err != nil {
		return err
	}
	if stat.IsDir() || stat.Size() != task.SourceSize || stat.ModTime().UnixMilli() != task.SourceModifiedTime {
		return errors.New("源文件已经变化，请重新加入转码列表")
	}
	return nil
}

func normalizeVideoEditPlan(
	plan VideoEditPlan,
	sourceDuration float64,
) (VideoEditPlan, float64, bool, error) {
	if sourceDuration <= 0 {
		return VideoEditPlan{}, 0, false, errors.New("源视频时长无效")
	}
	if len(plan.Segments) == 0 {
		return VideoEditPlan{}, 0, false, errors.New("剪辑方案至少需要保留一个片段")
	}
	if len(plan.Segments) > 200 {
		return VideoEditPlan{}, 0, false, errors.New("剪辑片段不能超过 200 个")
	}
	normalized := VideoEditPlan{Version: 1, Segments: make([]VideoEditSegment, len(plan.Segments))}
	seenIDs := make(map[string]struct{}, len(plan.Segments))
	for index, segment := range plan.Segments {
		if math.IsNaN(segment.Start) || math.IsInf(segment.Start, 0) ||
			math.IsNaN(segment.End) || math.IsInf(segment.End, 0) {
			return VideoEditPlan{}, 0, false, errors.New("剪辑片段时间无效")
		}
		segment.Start = math.Round(segment.Start*1000) / 1000
		segment.End = math.Round(segment.End*1000) / 1000
		if segment.Start < 0 || segment.End > sourceDuration+0.05 || segment.End-segment.Start < 0.1 {
			return VideoEditPlan{}, 0, false, fmt.Errorf("第 %d 个剪辑片段超出源视频范围或短于 0.1 秒", index+1)
		}
		segment.End = math.Min(segment.End, sourceDuration)
		if segment.ID == "" {
			segment.ID = core.GenerateUniqueID()
		}
		if _, exists := seenIDs[segment.ID]; exists {
			return VideoEditPlan{}, 0, false, errors.New("剪辑片段 ID 重复")
		}
		seenIDs[segment.ID] = struct{}{}
		if index == len(plan.Segments)-1 {
			segment.Transition = nil
		} else if segment.Transition != nil {
			transition := *segment.Transition
			transition.Duration = math.Round(transition.Duration*1000) / 1000
			if !supportedVideoTransition(transition.Type) {
				return VideoEditPlan{}, 0, false, fmt.Errorf("第 %d 个剪辑片段的转场类型不受支持", index+1)
			}
			if transition.Duration < videoTransitionMinDuration ||
				transition.Duration > videoTransitionMaxDuration+0.001 {
				return VideoEditPlan{}, 0, false, fmt.Errorf(
					"第 %d 个剪辑片段的转场时长必须在 %.1f～%.1f 秒之间",
					index+1, videoTransitionMinDuration, videoTransitionMaxDuration,
				)
			}
			segment.Transition = &transition
		}
		normalized.Segments[index] = segment
	}
	bySourceTime := append([]VideoEditSegment(nil), normalized.Segments...)
	sort.Slice(bySourceTime, func(i, j int) bool { return bySourceTime[i].Start < bySourceTime[j].Start })
	for index := 1; index < len(bySourceTime); index++ {
		if bySourceTime[index].Start < bySourceTime[index-1].End-0.01 {
			return VideoEditPlan{}, 0, false, errors.New("剪辑片段不能重叠或重复使用同一段画面")
		}
	}
	for index, segment := range normalized.Segments {
		incomingHalf := 0.0
		if index > 0 && normalized.Segments[index-1].Transition != nil {
			incomingHalf = normalized.Segments[index-1].Transition.Duration / 2
		}
		outgoingHalf := 0.0
		if segment.Transition != nil {
			outgoingHalf = segment.Transition.Duration / 2
		}
		if incomingHalf+outgoingHalf > segment.End-segment.Start-videoTransitionSegmentMargin+0.001 {
			return VideoEditPlan{}, 0, false, fmt.Errorf("第 %d 个剪辑片段太短，无法容纳前后转场", index+1)
		}
	}
	editedDuration := 0.0
	for _, segment := range normalized.Segments {
		editedDuration += segment.End - segment.Start
	}
	hasEdits := len(normalized.Segments) != 1 ||
		normalized.Segments[0].Start > 0.01 ||
		math.Abs(normalized.Segments[0].End-sourceDuration) > 0.05
	return normalized, editedDuration, hasEdits, nil
}

func supportedVideoTransition(transitionType string) bool {
	switch transitionType {
	case "fade", "fadeblack", "dissolve", "wipeleft", "wiperight",
		"slideleft", "slideright", "hlslice", "vuslice", "circleopen", "pixelize":
		return true
	default:
		return false
	}
}

func overlayTransitionFilterType(transitionType string) string {
	if transitionType == "fadeblack" {
		return "fade"
	}
	return transitionType
}

func overlayTransitionBackdropFilter(transitionType string) string {
	color := "black"
	if transitionType == "fade" {
		color = "white"
	}
	return fmt.Sprintf("drawbox=x=0:y=0:w=iw:h=ih:color=%s@0.98:t=fill", color)
}

func (VideoTranscode) Start(request VideoTranscodeStartRequest) (*VideoTranscodeStartResult, error) {
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
		return nil, err
	}
	if len(tasks) == 0 {
		return &VideoTranscodeStartResult{}, nil
	}
	startResult := &VideoTranscodeStartResult{}
	if request.EnableGPU {
		capabilities := (processorsffmpeg.FFmpeg{}).AvailableGPUEncoders()
		for index := range tasks {
			config := DefaultVideoTranscodeConfig()
			if err := json.Unmarshal([]byte(tasks[index].ConfigJsonData), &config); err != nil {
				return nil, fmt.Errorf("任务 %s 的转码参数无效: %w", tasks[index].ID, err)
			}
			encoder, changed := selectVideoTranscodeGPUEncoder(config, capabilities)
			if !changed {
				continue
			}
			config.GPUEncoder = encoder
			if err := validateVideoTranscodeConfig(config); err != nil {
				return nil, fmt.Errorf("任务 %s 的 GPU 转码参数无效: %w", tasks[index].ID, err)
			}
			data, err := json.Marshal(config)
			if err != nil {
				return nil, err
			}
			update := core.DBS().Model(&models.VideoTranscodeTask{}).
				Where("id = ? AND status IN ?", tasks[index].ID, statuses).
				Update("config_json_data", string(data))
			if update.Error != nil {
				return nil, update.Error
			}
			if update.RowsAffected == 1 {
				tasks[index].ConfigJsonData = string(data)
				startResult.EnabledGPU++
			}
		}
	}
	taskIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if task.EditPlanJsonData != "" {
			config := DefaultVideoTranscodeConfig()
			if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
				return nil, fmt.Errorf("任务 %s 的转码参数无效: %w", task.ID, err)
			}
			if config.VideoCodec == "copy" || config.AudioCodec == "copy" {
				return nil, fmt.Errorf("任务“%s”包含剪辑方案，请先将视频编码设为 H.264/H.265，并将音频编码设为 AAC", task.ResourceTitle)
			}
		}
		if isVideoTranscodeTemporaryPathForTask(task) {
			if err := os.Remove(task.TemporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("清理任务 %s 的旧转码临时文件失败: %w", task.ID, err)
			}
		}
		taskIDs = append(taskIDs, task.ID)
	}
	queueUpdate := core.DBS().Model(&models.VideoTranscodeTask{}).
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
		})
	if queueUpdate.Error != nil {
		return nil, queueUpdate.Error
	}
	startResult.Queued = int(queueUpdate.RowsAffected)
	transcodeManager.notify()
	return startResult, nil
}

func selectVideoTranscodeGPUEncoder(
	config VideoTranscodeConfig,
	capabilities []processorsffmpeg.GPUEncoderCapability,
) (string, bool) {
	if config.VideoCodec != "h264" && config.VideoCodec != "h265" {
		return config.GPUEncoder, false
	}
	selected := ""
	for _, capability := range capabilities {
		for _, codec := range capability.VideoCodecs {
			if codec != config.VideoCodec {
				continue
			}
			if capability.ID == config.GPUEncoder {
				return config.GPUEncoder, false
			}
			if selected == "" {
				selected = capability.ID
			}
		}
	}
	return selected, selected != "" && selected != config.GPUEncoder
}

// RetryReplacement 复用已经校验成功的临时成片，只重新执行文件替换和元数据刷新。
func (VideoTranscode) RetryReplacement(id string) error {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return err
	}
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
		return fmt.Errorf("解析转码参数失败: %w", err)
	}
	if !canRetryVideoTranscodeReplacement(*task, config) {
		return errors.New("当前任务没有可安全复用的临时成片，请重置后重新转码")
	}

	transcodeManager.mu.Lock()
	if transcodeManager.currentID != "" {
		transcodeManager.mu.Unlock()
		return errors.New("当前有其它转码任务正在执行，请稍后重试替换")
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := datatype.CustomTime(core.TimeNow())
	claim := core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id = ? AND status = ?", id, models.VideoTranscodeStatusFailed).
		Updates(map[string]interface{}{
			"status":          models.VideoTranscodeStatusRetryingReplace,
			"error_message":   "",
			"warning_message": "",
			"started_at":      &now,
			"finished_at":     nil,
		})
	if claim.Error != nil || claim.RowsAffected != 1 {
		transcodeManager.mu.Unlock()
		cancel()
		return fmt.Errorf("领取重试替换任务失败: %w", videoTranscodeUpdateError(claim))
	}
	transcodeManager.currentID = id
	transcodeManager.cancel = cancel
	transcodeManager.mu.Unlock()

	maintenance.Go(func() {
		defer func() {
			cancel()
			transcodeManager.mu.Lock()
			transcodeManager.currentID = ""
			transcodeManager.cancel = nil
			transcodeManager.mu.Unlock()
			transcodeManager.notify()
		}()
		(VideoTranscode{}).retryReplacement(ctx, task, config)
	})
	return nil
}

func (VideoTranscode) retryReplacement(
	ctx context.Context,
	task *models.VideoTranscodeTask,
	config VideoTranscodeConfig,
) {
	current, outputInfo, outputStat, err := verifyReusableVideoTranscodeOutput(ctx, task, config)
	if err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	if err := saveVideoTranscodeOutputInfo(task.ID, outputInfo, outputStat); err != nil {
		setVideoTranscodeFailure(task.ID, fmt.Errorf("保存临时成片信息失败: %w", err))
		return
	}
	if err := replaceVideoTranscodeSource(
		ctx, task, current, task.TemporaryPath, task.OutputPath, task.BackupPath,
		outputInfo, outputStat, config,
	); err != nil {
		setVideoTranscodeFailure(task.ID, err)
	}
}

func verifyReusableVideoTranscodeOutput(
	ctx context.Context,
	task *models.VideoTranscodeTask,
	config VideoTranscodeConfig,
) (*models.ResourcesDramaSeries, processorsffmpeg.VideoFormatInfo, os.FileInfo, error) {
	if err := validateVideoTranscodeTaskSourceSnapshot(task); err != nil {
		return nil, processorsffmpeg.VideoFormatInfo{}, nil,
			fmt.Errorf("复用临时成片前检查源文件失败: %w", err)
	}
	current, err := (models.ResourcesDramaSeries{}).Info(core.DBS(), task.DramaSeriesID)
	if err != nil {
		return nil, processorsffmpeg.VideoFormatInfo{}, nil, err
	}
	if filepath.Clean(current.Src) != filepath.Clean(task.SourcePath) {
		return nil, processorsffmpeg.VideoFormatInfo{}, nil,
			errors.New("资源路径已经变化，已拒绝复用临时成片")
	}
	sourceInfo, err := (processorsffmpeg.VideoInfo{}).GetVideoFormatInfo(current.Src)
	if err != nil {
		return nil, processorsffmpeg.VideoFormatInfo{}, nil, err
	}
	expectedDuration := task.SourceDuration
	hasEdits := task.EditPlanJsonData != ""
	if hasEdits && task.EditedDuration > 0 {
		expectedDuration = task.EditedDuration
	}
	if expectedDuration <= 0 {
		expectedDuration = (processorsffmpeg.VideoInfo{}).GetVideoDuration(sourceInfo)
	}
	outputInfo, outputStat, err := verifyVideoTranscodeOutputWithTimeout(
		ctx, videoTranscodeVerificationTimeout, task.TemporaryPath,
		expectedDuration, sourceInfo, config, hasEdits,
	)
	if err != nil {
		return nil, outputInfo, outputStat, fmt.Errorf("重新校验临时成片失败: %w", err)
	}
	if task.OutputSize > 0 && outputStat.Size() != task.OutputSize {
		return nil, outputInfo, outputStat, errors.New("临时成片大小已经变化，已拒绝复用")
	}
	return current, outputInfo, outputStat, nil
}

// SaveVerifiedOutputAsNewFile 在替换失败后复用临时成片，保留源文件并另存新文件。
func (VideoTranscode) SaveVerifiedOutputAsNewFile(id string) error {
	task, err := (models.VideoTranscodeTask{}).Info(core.DBS(), id)
	if err != nil {
		return err
	}
	if !canReuseVerifiedVideoTranscodeOutput(*task) {
		return errors.New("当前任务没有可安全复用的临时成片")
	}
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
		return fmt.Errorf("解析转码参数失败: %w", err)
	}
	config.OutputMode = "new_file"
	config.OutputFileName = ""
	_, outputPath, _, err := videoTranscodePaths(task, config)
	if err != nil {
		return err
	}
	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	transcodeManager.mu.Lock()
	if transcodeManager.currentID != "" {
		transcodeManager.mu.Unlock()
		return errors.New("当前有其它转码任务正在执行，请稍后保存新文件")
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := datatype.CustomTime(core.TimeNow())
	claim := core.DBS().Model(&models.VideoTranscodeTask{}).
		Where("id = ? AND status = ?", id, models.VideoTranscodeStatusFailed).
		Updates(map[string]interface{}{
			"status":           models.VideoTranscodeStatusSavingOutput,
			"config_json_data": string(configData),
			"output_path":      outputPath,
			"backup_path":      "",
			"error_message":    "",
			"warning_message":  "",
			"started_at":       &now,
			"finished_at":      nil,
		})
	if claim.Error != nil || claim.RowsAffected != 1 {
		transcodeManager.mu.Unlock()
		cancel()
		return fmt.Errorf("领取保存新文件任务失败: %w", videoTranscodeUpdateError(claim))
	}
	task.OutputPath = outputPath
	task.BackupPath = ""
	transcodeManager.currentID = id
	transcodeManager.cancel = cancel
	transcodeManager.mu.Unlock()

	maintenance.Go(func() {
		defer func() {
			cancel()
			transcodeManager.mu.Lock()
			transcodeManager.currentID = ""
			transcodeManager.cancel = nil
			transcodeManager.mu.Unlock()
			transcodeManager.notify()
		}()
		_, outputInfo, outputStat, verifyErr := verifyReusableVideoTranscodeOutput(ctx, task, config)
		if verifyErr != nil {
			setVideoTranscodeFailure(task.ID, verifyErr)
			return
		}
		if err := saveVideoTranscodeOutputInfo(task.ID, outputInfo, outputStat); err != nil {
			setVideoTranscodeFailure(task.ID, fmt.Errorf("保存临时成片信息失败: %w", err))
			return
		}
		if err := saveVideoTranscodeNewOutput(ctx, task, task.TemporaryPath, outputPath, outputStat); err != nil {
			setVideoTranscodeFailure(task.ID, err)
		}
	})
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
				"edit_plan_json_data":   "",
				"edited_duration":       0,
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

// cleanupVideoTranscodeTemporaryFile 只清理由当前任务创建、且仍位于源文件目录中的临时成片。
// 数据库中的路径即使被意外污染，也不能借由“移除任务”删除任意文件。
func cleanupVideoTranscodeTemporaryFile(task models.VideoTranscodeTask) error {
	if task.TemporaryPath == "" || !isVideoTranscodeTemporaryPathForTask(task) {
		return nil
	}
	if err := os.Remove(task.TemporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("清理任务 %s 的转码临时文件失败: %w", task.ID, err)
	}
	return nil
}

func deleteVideoTranscodeTasksWithCleanup(
	db *gorm.DB,
	query *gorm.DB,
	allowedStatuses []string,
) (int64, error) {
	var tasks []models.VideoTranscodeTask
	if err := query.Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	var deleted int64
	for _, task := range tasks {
		// 先通过带状态条件的删除领取该任务：若它已进入重试或执行态，RowsAffected 为 0，
		// 随后也不会触碰其正在使用的临时成片。
		result := db.Unscoped().Where("id = ? AND status IN ?", task.ID, allowedStatuses).
			Delete(&models.VideoTranscodeTask{})
		if result.Error != nil {
			return deleted, result.Error
		}
		if result.RowsAffected == 0 {
			continue
		}
		if err := cleanupVideoTranscodeTemporaryFile(task); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}

// deleteDiscardableVideoTranscodeTasksByResourceIDs 在资源删除时只清理不再有历史价值、
// 也不可能仍在执行的草稿和失败任务。成功记录和活动任务保持不变。
func deleteDiscardableVideoTranscodeTasksByResourceIDs(db *gorm.DB, resourceIDs []string) error {
	if len(resourceIDs) == 0 {
		return nil
	}
	statuses := []string{
		models.VideoTranscodeStatusDraft,
		models.VideoTranscodeStatusFailed,
	}
	_, err := deleteVideoTranscodeTasksWithCleanup(
		db,
		db.Where("resource_id IN ? AND status IN ?", resourceIDs, statuses),
		statuses,
	)
	return err
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
	if task.Status == models.VideoTranscodeStatusSavingOutput ||
		task.Status == models.VideoTranscodeStatusReplacing ||
		task.Status == models.VideoTranscodeStatusRefreshingMetadata {
		return errors.New("文件正在保存、替换或刷新元数据，当前阶段不可取消")
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
	db := core.DBS()
	statuses := []string{
		models.VideoTranscodeStatusDraft,
		models.VideoTranscodeStatusSuccess,
		models.VideoTranscodeStatusFailed,
		models.VideoTranscodeStatusCancelled,
		models.VideoTranscodeStatusInterrupted,
		models.VideoTranscodeStatusRollbackFailed,
	}
	var count int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var deleteErr error
		count, deleteErr = deleteVideoTranscodeTasksWithCleanup(
			tx,
			tx.Where("id = ? AND status IN ?", id, statuses),
			statuses,
		)
		return deleteErr
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("只能删除草稿或已经结束的任务")
	}
	return nil
}

func (VideoTranscode) DeleteBatch(request VideoTranscodeIDsRequest) (int64, error) {
	if len(request.IDs) == 0 {
		return 0, errors.New("请选择需要移除的任务")
	}
	db := core.DBS()
	statuses := []string{
		models.VideoTranscodeStatusDraft,
		models.VideoTranscodeStatusSuccess,
		models.VideoTranscodeStatusFailed,
		models.VideoTranscodeStatusCancelled,
		models.VideoTranscodeStatusInterrupted,
		models.VideoTranscodeStatusRollbackFailed,
	}
	var count int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var deleteErr error
		count, deleteErr = deleteVideoTranscodeTasksWithCleanup(
			tx,
			tx.Where("id IN ? AND status IN ?", request.IDs, statuses),
			statuses,
		)
		return deleteErr
	})
	return count, err
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
	if config.OutputMode != "replace" && config.OutputMode != "new_file" {
		return errors.New("输出方式只支持替换源文件或保存为新文件")
	}
	if config.OutputMode == "replace" && config.OutputFileName != "" {
		return errors.New("替换源文件模式不能设置新文件名")
	}
	if config.OutputMode == "new_file" {
		if err := validateVideoTranscodeOutputFileName(config.OutputFileName); err != nil {
			return err
		}
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

func validateVideoTranscodeOutputFileName(name string) error {
	if name == "" {
		return nil
	}
	if len([]byte(name)) > 180 || name == "." || name == ".." ||
		strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") ||
		strings.ContainsAny(name, `<>:"/\|?*`) {
		return errors.New("新文件名无效，请勿包含路径或系统保留字符")
	}
	for _, char := range name {
		if char < 32 {
			return errors.New("新文件名不能包含控制字符")
		}
	}
	reservedBase := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	reserved := map[string]struct{}{
		"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
		"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {}, "COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
		"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {}, "LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
	}
	if _, exists := reserved[reservedBase]; exists {
		return errors.New("新文件名是系统保留名称，请更换名称")
	}
	return nil
}

func (VideoTranscode) execute(ctx context.Context, task *models.VideoTranscodeTask) {
	config := DefaultVideoTranscodeConfig()
	if err := json.Unmarshal([]byte(task.ConfigJsonData), &config); err != nil {
		setVideoTranscodeFailure(task.ID, fmt.Errorf("解析转码参数失败: %w", err))
		return
	}
	if err := validateVideoTranscodeConfig(config); err != nil {
		setVideoTranscodeFailure(task.ID, fmt.Errorf("转码参数无效: %w", err))
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
	editPlan := VideoEditPlan{Version: 1}
	expectedDuration := duration
	hasEdits := task.EditPlanJsonData != ""
	if hasEdits {
		if err := json.Unmarshal([]byte(task.EditPlanJsonData), &editPlan); err != nil {
			setVideoTranscodeFailure(task.ID, fmt.Errorf("解析剪辑方案失败: %w", err))
			return
		}
		var normalizeErr error
		editPlan, expectedDuration, hasEdits, normalizeErr = normalizeVideoEditPlan(editPlan, duration)
		if normalizeErr != nil {
			setVideoTranscodeFailure(task.ID, fmt.Errorf("剪辑方案无效: %w", normalizeErr))
			return
		}
		if hasEdits && (config.VideoCodec == "copy" || config.AudioCodec == "copy") {
			setVideoTranscodeFailure(task.ID, errors.New("剪辑任务不能复制音视频流，请选择 H.264/H.265 和 AAC"))
			return
		}
		if hasEdits && countVideoTranscodeStreams(sourceInfo, "subtitle") > 0 {
			setVideoTranscodeFailure(task.ID, errors.New("第一版视频剪辑暂不支持包含字幕轨的源文件"))
			return
		}
	}
	if err := validateVideoTranscodeSourceCompatibility(sourceInfo, config); err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	sourceBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(sourceInfo)
	sourceVideoBitRate, _ := strconv.ParseInt(sourceBasic.BitRate, 10, 64)
	storedEditedDuration := 0.0
	if hasEdits {
		storedEditedDuration = expectedDuration
	}

	tempPath, outputPath, backupPath, err := videoTranscodePaths(task, config)
	if err != nil {
		setVideoTranscodeFailure(task.ID, err)
		return
	}
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
		"edited_duration":       storedEditedDuration,
	})
	if runStateResult.Error != nil || runStateResult.RowsAffected != 1 {
		setVideoTranscodeFailure(task.ID, fmt.Errorf(
			"保存转码运行状态失败: %w",
			videoTranscodeUpdateError(runStateResult),
		))
		return
	}

	_ = os.Remove(tempPath)
	if err := runVideoTranscodeFFmpeg(
		ctx, task.ID, current.Src, tempPath, expectedDuration,
		primaryVideo.Index, sourceInfo, config, editPlan, hasEdits,
	); err != nil {
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
	outputInfo, outputStat, err := verifyVideoTranscodeOutputWithTimeout(
		ctx, videoTranscodeVerificationTimeout,
		tempPath, expectedDuration, sourceInfo, config, hasEdits,
	)
	if err != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeFailure(task.ID, err)
		return
	}
	if err := saveVideoTranscodeOutputInfo(task.ID, outputInfo, outputStat); err != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeFailure(task.ID, fmt.Errorf("保存转码输出信息失败: %w", err))
		return
	}
	if ctx.Err() != nil {
		_ = os.Remove(tempPath)
		setVideoTranscodeCancelled(task.ID)
		return
	}
	if config.OutputMode == "new_file" {
		err = saveVideoTranscodeNewOutput(ctx, task, tempPath, outputPath, outputStat)
	} else {
		err = replaceVideoTranscodeSource(ctx, task, current, tempPath, outputPath, backupPath, outputInfo, outputStat, config)
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			_ = os.Remove(tempPath)
			setVideoTranscodeCancelled(task.ID)
		} else {
			setVideoTranscodeFailure(task.ID, err)
		}
		return
	}
}

func saveVideoTranscodeOutputInfo(
	taskID string,
	outputInfo processorsffmpeg.VideoFormatInfo,
	outputStat os.FileInfo,
) error {
	outputBasic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(outputInfo)
	outputVideoBitRate, _ := strconv.ParseInt(outputBasic.BitRate, 10, 64)
	return core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"output_size":           outputStat.Size(),
			"output_duration":       (processorsffmpeg.VideoInfo{}).GetVideoDuration(outputInfo),
			"output_width":          outputBasic.Width,
			"output_height":         outputBasic.Height,
			"output_frame_rate":     outputBasic.FrameRate,
			"output_video_codec":    outputBasic.VideoCodec,
			"output_audio_codec":    outputBasic.AudioCodec,
			"output_video_bit_rate": outputVideoBitRate,
		}).Error
}

func videoTranscodePaths(task *models.VideoTranscodeTask, config VideoTranscodeConfig) (string, string, string, error) {
	sourceExt := filepath.Ext(task.SourcePath)
	base := strings.TrimSuffix(filepath.Base(task.SourcePath), sourceExt)
	ext := "." + config.Container
	if config.Container == "mkv" {
		ext = ".mkv"
	}
	dir := filepath.Dir(task.SourcePath)
	runID := core.GenerateUniqueID()
	output := filepath.Join(dir, base+ext)
	if config.OutputMode == "new_file" {
		outputBase := config.OutputFileName
		if outputBase == "" {
			outputBase = base + "_转码"
			if task.EditPlanJsonData != "" {
				outputBase = base + "_剪辑"
			}
		}
		var err error
		output, err = nextVideoTranscodeOutputPath(dir, outputBase, ext)
		if err != nil {
			return "", "", "", err
		}
	}
	temp := filepath.Join(dir, "."+base+".cmtranscode-"+task.ID+"-"+runID+ext+".part")
	backup := ""
	if config.OutputMode == "replace" {
		backup = filepath.Join(dir, "."+filepath.Base(task.SourcePath)+".cmbackup-"+task.ID+"-"+runID)
	}
	return temp, output, backup, nil
}

func nextVideoTranscodeOutputPath(dir, outputBase, ext string) (string, error) {
	for index := 1; index <= 10000; index++ {
		suffix := ""
		if index > 1 {
			suffix = "_" + strconv.Itoa(index)
		}
		candidate := filepath.Join(dir, outputBase+suffix+ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("检查新文件输出路径失败: %w", err)
		}
	}
	return "", errors.New("无法生成可用的新文件名称")
}

func runVideoTranscodeFFmpeg(
	ctx context.Context,
	taskID, source, output string,
	duration float64,
	primaryVideoIndex int,
	sourceInfo processorsffmpeg.VideoFormatInfo,
	config VideoTranscodeConfig,
	editPlan VideoEditPlan,
	hasEdits bool,
) error {
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return err
	}
	args := []string{"-hide_banner", "-loglevel", "warning", "-y", "-i", source}
	if hasEdits {
		filterGraph, outputLabels := buildVideoEditFilterGraph(
			primaryVideoIndex, sourceInfo, editPlan, config,
		)
		args = append(args, "-filter_complex", filterGraph)
		for _, label := range outputLabels {
			args = append(args, "-map", label)
		}
		args = append(args, "-map_metadata", "0", "-map_chapters", "-1")
	} else {
		args = append(args,
			"-map", fmt.Sprintf("0:%d", primaryVideoIndex), "-map", "0:a?", "-map", "0:s?",
			"-map_metadata", "0", "-map_chapters", "0",
		)
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
		if config.ResolutionHeight > 0 && !hasEdits {
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

func buildVideoEditFilterGraph(
	primaryVideoIndex int,
	sourceInfo processorsffmpeg.VideoFormatInfo,
	plan VideoEditPlan,
	config VideoTranscodeConfig,
) (string, []string) {
	filters := make([]string, 0, len(plan.Segments)*8)
	videoLabels := make([]string, 0, len(plan.Segments))
	for index, segment := range plan.Segments {
		var incoming *VideoTransition
		if index > 0 {
			incoming = plan.Segments[index-1].Transition
		}
		incomingHalf := 0.0
		if incoming != nil {
			incomingHalf = incoming.Duration / 2
		}
		outgoingHalf := 0.0
		if segment.Transition != nil {
			outgoingHalf = segment.Transition.Duration / 2
		}
		videoParts := make([]string, 0, 3)
		if incoming != nil {
			realLabel := fmt.Sprintf("v%d_in_real", index)
			backdropSourceLabel := fmt.Sprintf("v%d_in_backdrop_src", index)
			backdropLabel := fmt.Sprintf("v%d_in_backdrop", index)
			partLabel := fmt.Sprintf("v%d_in", index)
			filters = append(filters,
				fmt.Sprintf("[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,settb=AVTB,split=2[%s][%s]",
					primaryVideoIndex, segment.Start, segment.Start+incomingHalf, realLabel, backdropSourceLabel),
				fmt.Sprintf("[%s]%s[%s]", backdropSourceLabel, overlayTransitionBackdropFilter(incoming.Type), backdropLabel),
				fmt.Sprintf("[%s][%s]xfade=transition=%s:duration=%.6f:offset=0[%s]",
					backdropLabel, realLabel, overlayTransitionFilterType(incoming.Type), incomingHalf, partLabel),
			)
			videoParts = append(videoParts, "["+partLabel+"]")
		}
		middleStart := segment.Start + incomingHalf
		middleEnd := segment.End - outgoingHalf
		if middleEnd-middleStart > 0.001 {
			partLabel := fmt.Sprintf("v%d_mid", index)
			filters = append(filters, fmt.Sprintf(
				"[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,settb=AVTB[%s]",
				primaryVideoIndex, middleStart, middleEnd, partLabel,
			))
			videoParts = append(videoParts, "["+partLabel+"]")
		}
		if segment.Transition != nil {
			realLabel := fmt.Sprintf("v%d_out_real", index)
			backdropSourceLabel := fmt.Sprintf("v%d_out_backdrop_src", index)
			backdropLabel := fmt.Sprintf("v%d_out_backdrop", index)
			partLabel := fmt.Sprintf("v%d_out", index)
			filters = append(filters,
				fmt.Sprintf("[0:%d]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS,settb=AVTB,split=2[%s][%s]",
					primaryVideoIndex, segment.End-outgoingHalf, segment.End, realLabel, backdropSourceLabel),
				fmt.Sprintf("[%s]%s[%s]", backdropSourceLabel, overlayTransitionBackdropFilter(segment.Transition.Type), backdropLabel),
				fmt.Sprintf("[%s][%s]xfade=transition=%s:duration=%.6f:offset=0[%s]",
					realLabel, backdropLabel, overlayTransitionFilterType(segment.Transition.Type), outgoingHalf, partLabel),
			)
			videoParts = append(videoParts, "["+partLabel+"]")
		}
		videoLabel := fmt.Sprintf("v%d", index)
		if len(videoParts) == 1 {
			videoLabels = append(videoLabels, videoParts[0])
		} else {
			filters = append(filters, strings.Join(videoParts, "")+
				fmt.Sprintf("concat=n=%d:v=1:a=0[%s]", len(videoParts), videoLabel))
			videoLabels = append(videoLabels, "["+videoLabel+"]")
		}
	}
	currentVideo := videoLabels[0]
	for index := 1; index < len(videoLabels); index++ {
		outputLabel := fmt.Sprintf("vjoin%d", index)
		filters = append(filters, currentVideo+videoLabels[index]+"concat=n=2:v=1:a=0["+outputLabel+"]")
		currentVideo = "[" + outputLabel + "]"
	}
	videoOutputLabel := "vout"
	if config.ResolutionHeight > 0 {
		filters = append(filters, fmt.Sprintf(
			"%sscale=-2:min(ih\\,%d)[%s]",
			currentVideo, config.ResolutionHeight, videoOutputLabel,
		))
	} else {
		filters = append(filters, currentVideo+"null["+videoOutputLabel+"]")
	}
	outputLabels := []string{"[" + videoOutputLabel + "]"}
	audioOutputIndex := 0
	for _, stream := range sourceInfo.Streams {
		if stream.CodecType != "audio" {
			continue
		}
		audioInputs := strings.Builder{}
		for segmentIndex, segment := range plan.Segments {
			label := fmt.Sprintf("a%d_%d", audioOutputIndex, segmentIndex)
			audioFilter := fmt.Sprintf(
				"[0:%d]atrim=start=%.6f:end=%.6f,asetpts=PTS-STARTPTS",
				stream.Index, segment.Start, segment.End,
			)
			if segmentIndex > 0 {
				incoming := plan.Segments[segmentIndex-1].Transition
				if incoming != nil && incoming.AudioFade {
					audioFilter += fmt.Sprintf(",afade=t=in:st=0:d=%.6f", incoming.Duration/2)
				}
			}
			if segment.Transition != nil && segment.Transition.AudioFade {
				outgoingHalf := segment.Transition.Duration / 2
				audioFilter += fmt.Sprintf(",afade=t=out:st=%.6f:d=%.6f",
					segment.End-segment.Start-outgoingHalf, outgoingHalf)
			}
			filters = append(filters, audioFilter+"["+label+"]")
			audioInputs.WriteString("[")
			audioInputs.WriteString(label)
			audioInputs.WriteString("]")
		}
		outputLabel := fmt.Sprintf("aout%d", audioOutputIndex)
		filters = append(filters, fmt.Sprintf(
			"%sconcat=n=%d:v=0:a=1[%s]",
			audioInputs.String(), len(plan.Segments), outputLabel,
		))
		outputLabels = append(outputLabels, "["+outputLabel+"]")
		audioOutputIndex++
	}
	return strings.Join(filters, ";"), outputLabels
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
	hasEdits bool,
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
	if err := validateVideoTranscodeOutputInfo(sourceInfo, info, sourceDuration, config, hasEdits); err != nil {
		return info, nil, err
	}
	return info, stat, nil
}

func verifyVideoTranscodeOutputWithTimeout(
	ctx context.Context,
	timeout time.Duration,
	path string,
	sourceDuration float64,
	sourceInfo processorsffmpeg.VideoFormatInfo,
	config VideoTranscodeConfig,
	hasEdits bool,
) (processorsffmpeg.VideoFormatInfo, os.FileInfo, error) {
	result := waitVideoTranscodeVerification(ctx, timeout, func() videoTranscodeVerifyResult {
		info, stat, err := verifyVideoTranscodeOutput(path, sourceDuration, sourceInfo, config, hasEdits)
		return videoTranscodeVerifyResult{Info: info, Stat: stat, Err: err}
	})
	return result.Info, result.Stat, result.Err
}

func waitVideoTranscodeVerification(
	ctx context.Context,
	timeout time.Duration,
	verify func() videoTranscodeVerifyResult,
) videoTranscodeVerifyResult {
	resultChannel := make(chan videoTranscodeVerifyResult, 1)
	maintenance.Go(func() { resultChannel <- verify() })
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-resultChannel:
		return result
	case <-ctx.Done():
		return videoTranscodeVerifyResult{Err: ctx.Err()}
	case <-timer.C:
		return videoTranscodeVerifyResult{Err: fmt.Errorf(
			"校验转码输出超过 %s，任务已停止，源文件未替换", timeout,
		)}
	}
}

func validateVideoTranscodeOutputInfo(
	sourceInfo, outputInfo processorsffmpeg.VideoFormatInfo,
	sourceDuration float64,
	config VideoTranscodeConfig,
	hasEdits bool,
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
	if hasEdits {
		if countVideoTranscodeStreams(outputInfo, "subtitle") != 0 {
			return errors.New("剪辑输出包含了非预期的字幕轨")
		}
	} else if countVideoTranscodeStreams(sourceInfo, "subtitle") != countVideoTranscodeStreams(outputInfo, "subtitle") {
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

func saveVideoTranscodeNewOutput(
	ctx context.Context,
	task *models.VideoTranscodeTask,
	tempPath, outputPath string,
	outputStat os.FileInfo,
) error {
	lock := mediaPathLocks.get(outputPath)
	lock.Lock()
	defer lock.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	stateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Update("status", models.VideoTranscodeStatusSavingOutput)
	if stateResult.Error != nil || stateResult.RowsAffected != 1 {
		return fmt.Errorf("保存新文件状态失败: %w", videoTranscodeUpdateError(stateResult))
	}
	if err := publishVideoTranscodeOutput(tempPath, outputPath); err != nil {
		return fmt.Errorf("保存转码结果为新文件失败: %w", err)
	}
	finalStat, err := os.Stat(outputPath)
	if err != nil {
		_ = os.Remove(outputPath)
		return err
	}
	if outputStat != nil && finalStat.Size() != outputStat.Size() {
		_ = os.Remove(outputPath)
		return errors.New("另存的新文件大小与校验结果不一致")
	}
	if err := completeVideoTranscodeNewOutput(task.ID, finalStat.Size(), ""); err != nil {
		_ = os.Remove(outputPath)
		return err
	}
	if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		warning := "新文件已保存，但转码临时文件清理失败: " + err.Error()
		_ = core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
			Update("warning_message", warning).Error
	}
	processorscache.CacheVideoInfoLastUse{}.Invalidate(outputPath)
	return nil
}

func publishVideoTranscodeOutput(tempPath, outputPath string) error {
	if err := os.Link(tempPath, outputPath); err == nil {
		return nil
	}
	source, err := os.Open(tempPath)
	if err != nil {
		return err
	}
	defer source.Close()
	target, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		_ = target.Close()
		if !succeeded {
			_ = os.Remove(outputPath)
		}
	}()
	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	if err := target.Sync(); err != nil {
		return err
	}
	if err := target.Close(); err != nil {
		return err
	}
	succeeded = true
	return nil
}

func completeVideoTranscodeNewOutput(taskID string, outputSize int64, warning string) error {
	finished := datatype.CustomTime(core.TimeNow())
	result := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":          models.VideoTranscodeStatusSuccess,
		"progress":        100,
		"output_size":     outputSize,
		"backup_path":     "",
		"warning_message": warning,
		"error_message":   "",
		"finished_at":     &finished,
	})
	if result.Error != nil || result.RowsAffected != 1 {
		return videoTranscodeUpdateError(result)
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
	replaceStateResult := core.DBS().Model(&models.VideoTranscodeTask{}).Where("id = ?", task.ID).
		Update("status", models.VideoTranscodeStatusReplacing)
	if replaceStateResult.Error != nil || replaceStateResult.RowsAffected != 1 {
		return fmt.Errorf("保存文件替换状态失败: %w", videoTranscodeUpdateError(replaceStateResult))
	}
	// 编辑器或其它页面可能仍保持源视频的 Range/HLS 请求。先让这些请求结束，
	// 它们会关闭文件句柄并释放读锁，然后替换流程才能安全取得写锁。
	unblockMediaReads := blockMediaReads(current.Src)
	defer unblockMediaReads()
	processorscache.CacheVideoInfoLastUse{}.Invalidate(current.Src)
	_ = (processorscache.CacheFileLastUse{}).Invalidate(current.Src)
	unlockMedia, err := lockMediaForReplacementWithRetry(ctx, current.Src)
	if err != nil {
		return fmt.Errorf("释放源视频读取句柄失败: %w", err)
	}
	defer unlockMedia()
	if ctx.Err() != nil {
		return ctx.Err()
	}
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
