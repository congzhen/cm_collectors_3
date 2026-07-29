package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorscache "cm_collectors_server/processorsCache"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

const CurrentVideoMetadataVersion = 1

const (
	videoMetadataPriorityLow = iota + 1
	videoMetadataPriorityMedium
	videoMetadataPriorityHigh
)

const (
	VideoMetadataRunMissing      = "missing"
	VideoMetadataRunMissingStale = "missing_stale"
	VideoMetadataRunFailed       = "failed"
	VideoMetadataRunAll          = "all"
)

type videoMetadataQueueItem struct {
	dramaSeries models.ResourcesDramaSeries
	priority    int
	waiters     []chan error
	batchTaskID string
}

type videoMetadataCoordinator struct {
	sync.Mutex
	queued  map[string]*videoMetadataQueueItem
	running map[string]*videoMetadataQueueItem
	items   []*videoMetadataQueueItem
	worker  bool
}

var globalVideoMetadataCoordinator = videoMetadataCoordinator{
	queued:  map[string]*videoMetadataQueueItem{},
	running: map[string]*videoMetadataQueueItem{},
}
var videoMetadataBatchMu sync.Mutex
var lastInteractiveVideoMetadataAt atomic.Int64
var lastIdleVideoMetadataCheckID string

type VideoMetadata struct{}

type VideoMetadataSettingData struct {
	Setting       models.VideoMetadataSetting `json:"setting"`
	FilesBasesIDs []string                    `json:"filesBasesIds"`
}

type VideoMetadataRunRequest struct {
	ScopeMode      string   `json:"scopeMode"`
	FilesBasesIDs  []string `json:"filesBasesIds"`
	RunMode        string   `json:"runMode"`
	MaxItemsPerRun int      `json:"maxItemsPerRun"`
}

type VideoMetadataStats struct {
	FilesBasesID string `json:"filesBasesId"`
	Name         string `json:"name"`
	Total        int64  `json:"total"`
	Completed    int64  `json:"completed"`
	Pending      int64  `json:"pending"`
	Failed       int64  `json:"failed"`
	Processing   int64  `json:"processing"`
	Stale        int64  `json:"stale"`
}

type videoMetadataCandidate struct {
	models.ResourcesDramaSeries
	FilesBasesID string `gorm:"column:files_bases_id"`
}

type collectedVideoMetadataFile struct {
	DramaSeriesID    string `gorm:"column:drama_series_id"`
	Src              string `gorm:"column:src"`
	FileSize         int64  `gorm:"column:file_size"`
	FileModifiedTime int64  `gorm:"column:file_modified_time"`
}

var initVideoMetadataOnce sync.Once

func InitVideoMetadata() {
	initVideoMetadataOnce.Do(func() {
		lastInteractiveVideoMetadataAt.Store(core.TimeNow().UnixMilli())
		// 异常退出遗留的 processing 记录恢复为 stale，避免永久阻塞。
		core.DBS().Model(&models.ResourcesVideoMetadata{}).
			Where("probe_status = ?", models.VideoMetadataStatusProcessing).
			Update("probe_status", models.VideoMetadataStatusStale)
		core.DBS().Model(&models.VideoMetadataBatchTask{}).
			Where("status = ?", models.VideoMetadataBatchStatusRunning).
			Updates(map[string]interface{}{
				"status":      models.VideoMetadataBatchStatusPaused,
				"last_error":  "服务曾中断，任务已暂停，可手动继续",
				"current_src": "",
			})
		go (VideoMetadata{}).idleLoop()
	})
}

func (VideoMetadata) idleLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		setting, err := (models.VideoMetadataSetting{}).Ensure(core.DBS())
		if err != nil || !setting.IdleBackfillEnabled || setting.Paused {
			continue
		}
		lastInteractiveAt := time.UnixMilli(lastInteractiveVideoMetadataAt.Load())
		if core.TimeNow().Sub(lastInteractiveAt) < time.Duration(setting.IdleWaitMinutes)*time.Minute {
			continue
		}
		if videoMetadataHasActiveBatch() {
			continue
		}
		ids, err := (models.VideoMetadataSettingFilesBases{}).ListIDs(core.DBS())
		if err != nil || (setting.IdleScopeMode == models.VideoMetadataScopeSelected && len(ids) == 0) {
			continue
		}
		if err := markNextChangedVideoMetadataScope(
			setting.IdleScopeMode, ids, setting.IdleBatchSize,
		); err != nil {
			core.LogErr(err)
			continue
		}
		candidates, err := findVideoMetadataCandidates(
			ids, setting.IdleScopeMode, VideoMetadataRunMissingStale, setting.IdleBatchSize, nil,
		)
		if err != nil {
			core.LogErr(err)
			continue
		}
		for _, candidate := range candidates {
			(VideoMetadata{}).enqueueIfNeeded(
				candidate.ResourcesDramaSeries, videoMetadataPriorityLow, "", false, false, false,
			)
		}
	}
}

func videoMetadataHasActiveBatch() bool {
	var count int64
	err := core.DBS().Model(&models.VideoMetadataBatchTask{}).
		Where("status IN ?", []string{models.VideoMetadataBatchStatusRunning, models.VideoMetadataBatchStatusPaused}).
		Count(&count).Error
	return err == nil && count > 0
}

func (VideoMetadata) Setting() (*VideoMetadataSettingData, error) {
	db := core.DBS()
	setting, err := (models.VideoMetadataSetting{}).Ensure(db)
	if err != nil {
		return nil, err
	}
	ids, err := (models.VideoMetadataSettingFilesBases{}).ListIDs(db)
	if err != nil {
		return nil, err
	}
	return &VideoMetadataSettingData{Setting: *setting, FilesBasesIDs: ids}, nil
}

func (VideoMetadata) SaveSetting(data *VideoMetadataSettingData) (*VideoMetadataSettingData, error) {
	if data == nil {
		return nil, errors.New("视频信息采集设置不能为空")
	}
	normalizeVideoMetadataSetting(&data.Setting)
	if data.Setting.IdleBackfillEnabled {
		if err := validateVideoMetadataScope(data.Setting.IdleScopeMode, data.FilesBasesIDs); err != nil {
			return nil, err
		}
	}
	err := core.DBS().Transaction(func(tx *gorm.DB) error {
		if _, err := (models.VideoMetadataSetting{}).Ensure(tx); err != nil {
			return err
		}
		if err := (models.VideoMetadataSetting{}).Save(tx, &data.Setting); err != nil {
			return err
		}
		return (models.VideoMetadataSettingFilesBases{}).Replace(tx, uniqueVideoMetadataStrings(data.FilesBasesIDs), core.GenerateUniqueID)
	})
	if err != nil {
		return nil, err
	}
	return (VideoMetadata{}).Setting()
}

func normalizeVideoMetadataSetting(setting *models.VideoMetadataSetting) {
	setting.ID = "default"
	if setting.IdleScopeMode != models.VideoMetadataScopeAll {
		setting.IdleScopeMode = models.VideoMetadataScopeSelected
	}
	if setting.IdleWaitMinutes < 1 {
		setting.IdleWaitMinutes = 5
	}
	if setting.ProbeIntervalMilliseconds < 0 {
		setting.ProbeIntervalMilliseconds = 0
	}
	if setting.ProbeIntervalMilliseconds > 60000 {
		setting.ProbeIntervalMilliseconds = 60000
	}
	if setting.IdleBatchSize < 1 {
		setting.IdleBatchSize = 20
	}
	if setting.IdleBatchSize > 1000 {
		setting.IdleBatchSize = 1000
	}
}

func validateVideoMetadataScope(scopeMode string, ids []string) error {
	if scopeMode != models.VideoMetadataScopeSelected && scopeMode != models.VideoMetadataScopeAll {
		return errors.New("无效的视频信息采集范围")
	}
	if scopeMode == models.VideoMetadataScopeSelected && len(uniqueVideoMetadataStrings(ids)) == 0 {
		return errors.New("请选择至少一个文件库")
	}
	return nil
}

func uniqueVideoMetadataStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// TriggerForResources 只提交当前列表页，是否启用由全局列表触发设置控制。
func (VideoMetadata) TriggerForResources(resources *[]models.Resources) {
	(VideoMetadata{}).MarkInteractiveActivity()
	if resources == nil || !videoMetadataTriggerEnabled(func(s *models.VideoMetadataSetting) bool { return s.CollectOnList }) {
		return
	}
	for _, resource := range *resources {
		if !isVideoResourceMode(resource.Mode) {
			continue
		}
		for _, ds := range resource.ResourcesDramaSeries {
			(VideoMetadata{}).enqueueIfNeeded(ds, videoMetadataPriorityMedium, "", false, false, false)
		}
	}
}

// TriggerForResource 用于新增或修改后的资源。
func (VideoMetadata) TriggerForResource(resource *models.Resources) {
	(VideoMetadata{}).MarkInteractiveActivity()
	if resource == nil || !isVideoResourceMode(resource.Mode) ||
		!videoMetadataTriggerEnabled(func(s *models.VideoMetadataSetting) bool { return s.CollectOnNewOrChanged }) {
		return
	}
	for _, ds := range resource.ResourcesDramaSeries {
		(VideoMetadata{}).enqueueIfNeeded(ds, videoMetadataPriorityHigh, "", false, false, false)
	}
}

// TriggerForDetail 用于详情页按需补齐。
func (VideoMetadata) TriggerForDetail(resource *models.Resources) {
	(VideoMetadata{}).MarkInteractiveActivity()
	if resource == nil || !isVideoResourceMode(resource.Mode) ||
		!videoMetadataTriggerEnabled(func(s *models.VideoMetadataSetting) bool { return s.CollectOnDetailOrPlay }) {
		return
	}
	for _, ds := range resource.ResourcesDramaSeries {
		(VideoMetadata{}).enqueueIfNeeded(ds, videoMetadataPriorityHigh, "", false, false, false)
	}
}

func (VideoMetadata) MarkInteractiveActivity() {
	lastInteractiveVideoMetadataAt.Store(core.TimeNow().UnixMilli())
}

func videoMetadataTriggerEnabled(selectValue func(*models.VideoMetadataSetting) bool) bool {
	setting, err := (models.VideoMetadataSetting{}).Ensure(core.DBS())
	if err != nil {
		core.LogErr(err)
		return false
	}
	return !setting.Paused && selectValue(setting)
}

func isVideoResourceMode(mode datatype.E_resourceMode) bool {
	return mode == datatype.E_resourceMode_Movies || mode == datatype.E_resourceMode_VideoLink
}

func (VideoMetadata) enqueueIfNeeded(
	ds models.ResourcesDramaSeries,
	priority int,
	batchTaskID string,
	force bool,
	wantCompletion bool,
	waitExisting bool,
) (bool, <-chan error) {
	if ds.ID == "" || ds.Src == "" {
		return false, nil
	}
	if !force {
		needed, err := (VideoMetadata{}).needsProbe(ds.ID, false)
		if err != nil || !needed {
			return false, nil
		}
	}
	return enqueueVideoMetadata(videoMetadataQueueItem{
		dramaSeries: ds,
		priority:    priority,
		batchTaskID: batchTaskID,
	}, wantCompletion, waitExisting)
}

func enqueueVideoMetadata(item videoMetadataQueueItem, wantCompletion, waitExisting bool) (bool, <-chan error) {
	if item.priority > videoMetadataPriorityLow {
		lastInteractiveVideoMetadataAt.Store(core.TimeNow().UnixMilli())
	}
	q := &globalVideoMetadataCoordinator
	q.Lock()
	defer q.Unlock()

	id := item.dramaSeries.ID
	if existing, ok := q.running[id]; ok {
		if wantCompletion && waitExisting {
			done := make(chan error, 1)
			existing.waiters = append(existing.waiters, done)
			return false, done
		}
		return false, nil
	}
	if existing, ok := q.queued[id]; ok {
		if item.priority > existing.priority {
			existing.priority = item.priority
		}
		if wantCompletion && waitExisting {
			done := make(chan error, 1)
			existing.waiters = append(existing.waiters, done)
			return false, done
		}
		return false, nil
	}
	copyItem := item
	var done chan error
	if wantCompletion {
		done = make(chan error, 1)
		copyItem.waiters = append(copyItem.waiters, done)
	}
	q.queued[id] = &copyItem
	q.items = append(q.items, &copyItem)
	if !q.worker {
		q.worker = true
		go (VideoMetadata{}).runQueue()
	}
	return true, done
}

func (VideoMetadata) runQueue() {
	for {
		item, ok := popVideoMetadata()
		if !ok {
			return
		}
		err := (VideoMetadata{}).probe(item.dramaSeries)
		finishVideoMetadata(item, err)

		// 文件处理间隔只用于低优先级的空闲补齐，避免拖慢详情、播放和手动任务。
		setting, settingErr := (models.VideoMetadataSetting{}).Ensure(core.DBS())
		if item.priority == videoMetadataPriorityLow &&
			settingErr == nil && setting.ProbeIntervalMilliseconds > 0 {
			time.Sleep(time.Duration(setting.ProbeIntervalMilliseconds) * time.Millisecond)
		}
	}
}

func popVideoMetadata() (*videoMetadataQueueItem, bool) {
	q := &globalVideoMetadataCoordinator
	q.Lock()
	defer q.Unlock()
	if len(q.items) == 0 {
		q.worker = false
		return nil, false
	}
	best := 0
	for i := 1; i < len(q.items); i++ {
		if q.items[i].priority > q.items[best].priority {
			best = i
		}
	}
	item := q.items[best]
	q.items = append(q.items[:best], q.items[best+1:]...)
	delete(q.queued, item.dramaSeries.ID)
	q.running[item.dramaSeries.ID] = item
	return item, true
}

func finishVideoMetadata(item *videoMetadataQueueItem, err error) {
	q := &globalVideoMetadataCoordinator
	q.Lock()
	delete(q.running, item.dramaSeries.ID)
	q.Unlock()
	for _, waiter := range item.waiters {
		waiter <- err
		close(waiter)
	}
}

func (VideoMetadata) needsProbe(dramaSeriesID string, includeFailed bool) (bool, error) {
	item, err := (models.ResourcesVideoMetadata{}).Get(core.DBS(), dramaSeriesID)
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if item.ProbeStatus == models.VideoMetadataStatusFailed {
		if !includeFailed {
			return false, nil
		}
		return item.NextRetryTime == nil || item.NextRetryTime.IsZero() ||
			!core.TimeNow().Before(time.Time(*item.NextRetryTime)), nil
	}
	if item.ProbeStatus == models.VideoMetadataStatusProcessing {
		return false, nil
	}
	if item.MetadataVersion < CurrentVideoMetadataVersion || item.ProbeStatus == models.VideoMetadataStatusStale {
		return true, nil
	}
	if item.ProbeStatus == models.VideoMetadataStatusSuccess {
		ds, dsErr := (models.ResourcesDramaSeries{}).Info(core.DBS(), dramaSeriesID)
		if dsErr != nil {
			return false, dsErr
		}
		stat, statErr := os.Stat(ds.Src)
		if statErr != nil {
			return true, nil
		}
		return stat.Size() != item.FileSize || stat.ModTime().UnixMilli() != item.FileModifiedTime, nil
	}
	return true, nil
}

func (VideoMetadata) probe(ds models.ResourcesDramaSeries) error {
	// 队列中的路径可能在等待期间被用户修改，执行前始终读取最新分集。
	current, err := (models.ResourcesDramaSeries{}).Info(core.DBS(), ds.ID)
	if err != nil {
		return err
	}
	ds = *current
	now := core.TimeNow()
	probeTime := datatype.CustomTime(now)
	if err := (VideoMetadata{}).markProcessing(ds.ID, probeTime); err != nil {
		return err
	}
	stat, err := os.Stat(ds.Src)
	if err != nil {
		return (VideoMetadata{}).saveFailure(ds, videoMetadataErrorCode(err), err, probeTime)
	}

	formatInfo, err := (processorscache.CacheVideoInfoLastUse{}).GetVideoInfoHandle(ds.Src)
	if err != nil {
		return (VideoMetadata{}).saveFailure(ds, videoMetadataErrorCode(err), err, probeTime)
	}
	basic := (processorsffmpeg.VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(formatInfo)
	duration, err := strconv.ParseFloat(basic.Duration, 64)
	if err != nil || duration <= 0 || basic.Width <= 0 || basic.Height <= 0 {
		if err == nil {
			err = errors.New("FFprobe 未返回有效的视频时长或分辨率")
		}
		return (VideoMetadata{}).saveFailure(ds, "invalid_metadata", err, probeTime)
	}
	videoBitRate, _ := strconv.ParseInt(basic.BitRate, 10, 64)
	fileSize, _ := strconv.ParseInt(basic.Size, 10, 64)
	if fileSize <= 0 {
		fileSize = stat.Size()
	}

	// FFprobe 运行期间路径也可能被编辑；旧文件结果不能覆盖新路径的元数据。
	latest, err := (models.ResourcesDramaSeries{}).Info(core.DBS(), ds.ID)
	if err != nil {
		return err
	}
	if latest.Src != ds.Src {
		core.DBS().Model(&models.ResourcesVideoMetadata{}).
			Where("drama_series_id = ?", ds.ID).
			Updates(map[string]interface{}{
				"probe_status":     models.VideoMetadataStatusStale,
				"metadata_version": 0,
			})
		return errors.New("分集路径在采集过程中发生变化，已等待重新采集")
	}

	metadata := models.ResourcesVideoMetadata{
		DramaSeriesID:    ds.ID,
		MetadataVersion:  CurrentVideoMetadataVersion,
		ProbeStatus:      models.VideoMetadataStatusSuccess,
		ProbeTime:        &probeTime,
		FileSize:         fileSize,
		FileModifiedTime: stat.ModTime().UnixMilli(),
		Width:            basic.Width,
		Height:           basic.Height,
		FrameRate:        basic.FrameRate,
		FrameRateRaw:     basic.FrameRateRaw,
		VideoCodec:       basic.VideoCodec,
		VideoProfile:     basic.VideoProfile,
		PixelFormat:      basic.PixelFormat,
		BitDepth:         basic.BitDepth,
		VideoBitRate:     videoBitRate,
		ContainerFormat:  basic.ContainerFormat,
		AudioCodec:       basic.AudioCodec,
		AudioChannels:    basic.AudioChannels,
		AudioSampleRate:  basic.AudioSampleRate,
	}
	return core.DBS().Transaction(func(tx *gorm.DB) error {
		if err := (models.ResourcesDramaSeries{}).UpdateDuration(
			tx, ds.ID, int(duration), models.DurationProbeStatusSuccess, &probeTime,
		); err != nil {
			return err
		}
		return (models.ResourcesVideoMetadata{}).Upsert(tx, &metadata)
	})
}

func (VideoMetadata) markProcessing(dramaSeriesID string, probeTime datatype.CustomTime) error {
	existing, err := (models.ResourcesVideoMetadata{}).Get(core.DBS(), dramaSeriesID)
	if err == gorm.ErrRecordNotFound {
		return core.DBS().Create(&models.ResourcesVideoMetadata{
			DramaSeriesID: dramaSeriesID,
			ProbeStatus:   models.VideoMetadataStatusProcessing,
			ProbeTime:     &probeTime,
		}).Error
	}
	if err != nil {
		return err
	}
	return core.DBS().Model(existing).Updates(map[string]interface{}{
		"probe_status": models.VideoMetadataStatusProcessing,
		"probe_time":   &probeTime,
	}).Error
}

func (VideoMetadata) saveFailure(
	ds models.ResourcesDramaSeries,
	code string,
	probeErr error,
	probeTime datatype.CustomTime,
) error {
	existing, err := (models.ResourcesVideoMetadata{}).Get(core.DBS(), ds.ID)
	retryCount := 1
	if err == nil {
		retryCount = existing.RetryCount + 1
	}
	nextRetry := datatype.CustomTime(videoMetadataNextRetry(retryCount, time.Time(probeTime)))
	dbErr := core.DBS().Transaction(func(tx *gorm.DB) error {
		duration := ds.DurationSeconds
		if err := (models.ResourcesDramaSeries{}).UpdateDuration(
			tx, ds.ID, duration, models.DurationProbeStatusFailed, &probeTime,
		); err != nil {
			return err
		}
		return tx.Model(&models.ResourcesVideoMetadata{}).Where("drama_series_id = ?", ds.ID).
			Updates(map[string]interface{}{
				"probe_status":    models.VideoMetadataStatusFailed,
				"probe_time":      &probeTime,
				"next_retry_time": &nextRetry,
				"retry_count":     retryCount,
				"error_code":      code,
				"error_message":   probeErr.Error(),
			}).Error
	})
	if dbErr != nil {
		return dbErr
	}
	core.LogErr(fmt.Errorf("获取视频元数据失败：%s，%w", ds.Src, probeErr))
	return probeErr
}

func videoMetadataNextRetry(retryCount int, now time.Time) time.Time {
	switch retryCount {
	case 1:
		return now.Add(time.Hour)
	case 2:
		return now.Add(6 * time.Hour)
	case 3:
		return now.Add(24 * time.Hour)
	default:
		return now.Add(7 * 24 * time.Hour)
	}
}

func videoMetadataErrorCode(err error) string {
	if os.IsNotExist(err) {
		return "file_not_found"
	}
	if os.IsPermission(err) {
		return "permission_denied"
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "ffprobe不可用"):
		return "ffprobe_unavailable"
	case strings.Contains(text, "timeout"):
		return "probe_timeout"
	case strings.Contains(text, "unsupported"):
		return "unsupported_format"
	default:
		return "io_error"
	}
}

// EnsureForPlay 在播放需要 FFprobe 信息时复用统一队列，避免缺失元数据的视频同时执行两次探测。
func (VideoMetadata) EnsureForPlay(ds models.ResourcesDramaSeries) error {
	(VideoMetadata{}).MarkInteractiveActivity()
	if !videoMetadataTriggerEnabled(func(s *models.VideoMetadataSetting) bool { return s.CollectOnDetailOrPlay }) {
		return nil
	}
	needed, err := (VideoMetadata{}).needsProbe(ds.ID, false)
	if err != nil || !needed {
		return err
	}
	_, done := (VideoMetadata{}).enqueueIfNeeded(ds, videoMetadataPriorityHigh, "", false, true, true)
	if done == nil {
		return nil
	}
	return <-done
}

func (VideoMetadata) MetadataInfo(dramaSeriesID string) (*models.ResourcesVideoMetadata, error) {
	return (models.ResourcesVideoMetadata{}).Get(core.DBS(), dramaSeriesID)
}

func (VideoMetadata) Stats() ([]VideoMetadataStats, error) {
	var result []VideoMetadataStats
	err := core.DBS().Raw(`
		SELECT fb.id AS files_bases_id, fb.name,
			COUNT(ds.id) AS total,
			SUM(CASE WHEN vm.probe_status = 'success' AND vm.metadata_version >= ? THEN 1 ELSE 0 END) AS completed,
			SUM(CASE WHEN ds.id IS NOT NULL AND vm.drama_series_id IS NULL THEN 1 ELSE 0 END) AS pending,
			SUM(CASE WHEN vm.probe_status = 'failed' THEN 1 ELSE 0 END) AS failed,
			SUM(CASE WHEN vm.probe_status = 'processing' THEN 1 ELSE 0 END) AS processing,
			SUM(CASE WHEN vm.probe_status = 'stale'
				OR (vm.drama_series_id IS NOT NULL AND vm.probe_status NOT IN ('failed', 'processing') AND vm.metadata_version < ?)
				THEN 1 ELSE 0 END) AS stale
		FROM filesBases fb
		LEFT JOIN resources r ON r.filesBases_id = fb.id AND r.mode IN ('movies', 'videoLink')
		LEFT JOIN resourcesDramaSeries ds ON ds.resources_id = r.id
		LEFT JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id
		GROUP BY fb.id, fb.name
		ORDER BY fb.sort, fb.id`, CurrentVideoMetadataVersion, CurrentVideoMetadataVersion).Scan(&result).Error
	return result, err
}

func (VideoMetadata) StartBatch(request VideoMetadataRunRequest) (*models.VideoMetadataBatchTask, error) {
	videoMetadataBatchMu.Lock()
	defer videoMetadataBatchMu.Unlock()

	request.FilesBasesIDs = uniqueVideoMetadataStrings(request.FilesBasesIDs)
	if request.ScopeMode == "" {
		request.ScopeMode = models.VideoMetadataScopeSelected
	}
	if err := validateVideoMetadataScope(request.ScopeMode, request.FilesBasesIDs); err != nil {
		return nil, err
	}
	if request.RunMode == "" {
		request.RunMode = VideoMetadataRunMissingStale
	}
	if !validVideoMetadataRunMode(request.RunMode) {
		return nil, errors.New("无效的视频信息补齐方式")
	}
	var running int64
	if err := core.DBS().Model(&models.VideoMetadataBatchTask{}).
		Where("status IN ?", []string{models.VideoMetadataBatchStatusRunning, models.VideoMetadataBatchStatusPaused}).
		Count(&running).Error; err != nil {
		return nil, err
	}
	if running > 0 {
		return nil, errors.New("已有视频信息补齐任务正在运行或暂停")
	}
	now := datatype.CustomTime(core.TimeNow())
	task := models.VideoMetadataBatchTask{
		ID:        core.GenerateUniqueID(),
		ScopeMode: request.ScopeMode,
		RunMode:   request.RunMode,
		Status:    models.VideoMetadataBatchStatusRunning,
		StartedAt: &now,
	}
	err := core.DBS().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		for _, id := range request.FilesBasesIDs {
			scope := models.VideoMetadataBatchTaskFilesBases{
				ID: core.GenerateUniqueID(), TaskID: task.ID, FilesBasesID: id,
			}
			if err := tx.Create(&scope).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if request.RunMode == VideoMetadataRunAll {
		if err := markVideoMetadataScopeStale(request.ScopeMode, request.FilesBasesIDs); err != nil {
			updateVideoMetadataBatchError(task.ID, err)
			return nil, err
		}
	}
	total, err := countVideoMetadataCandidates(
		request.FilesBasesIDs, request.ScopeMode, request.RunMode, nil,
	)
	if err != nil {
		updateVideoMetadataBatchError(task.ID, err)
		return nil, err
	}
	if request.MaxItemsPerRun > 0 && total > int64(request.MaxItemsPerRun) {
		total = int64(request.MaxItemsPerRun)
	}
	if err := core.DBS().Model(&models.VideoMetadataBatchTask{}).
		Where("id = ?", task.ID).Update("total", total).Error; err != nil {
		updateVideoMetadataBatchError(task.ID, err)
		return nil, err
	}
	task.Total = total
	go (VideoMetadata{}).runBatch(task.ID, request.MaxItemsPerRun)
	return &task, nil
}

// RunForCron 按计划任务范围同步处理一个有上限的批次。
// 所有单条任务仍进入统一调度器，因此与页面、播放和手动补齐保持分集级去重。
func (VideoMetadata) RunForCron(request VideoMetadataRunRequest) error {
	request.FilesBasesIDs = uniqueVideoMetadataStrings(request.FilesBasesIDs)
	if request.ScopeMode == "" {
		request.ScopeMode = models.VideoMetadataScopeSelected
	}
	if err := validateVideoMetadataScope(request.ScopeMode, request.FilesBasesIDs); err != nil {
		return err
	}
	if request.MaxItemsPerRun <= 0 {
		request.MaxItemsPerRun = 100
	}
	if request.MaxItemsPerRun > 10000 {
		request.MaxItemsPerRun = 10000
	}
	if request.RunMode == "" {
		request.RunMode = VideoMetadataRunMissingStale
	}
	if !validVideoMetadataRunMode(request.RunMode) || request.RunMode == VideoMetadataRunAll {
		return errors.New("无效的视频信息计划任务处理方式")
	}
	if request.RunMode == VideoMetadataRunMissingStale {
		if err := markChangedVideoMetadataScope(request.ScopeMode, request.FilesBasesIDs); err != nil {
			return err
		}
	}
	candidates, err := findVideoMetadataCandidates(
		request.FilesBasesIDs, request.ScopeMode, request.RunMode, request.MaxItemsPerRun, nil,
	)
	if err != nil {
		return err
	}
	failed := 0
	for _, candidate := range candidates {
		force := request.RunMode == VideoMetadataRunFailed
		queued, done := (VideoMetadata{}).enqueueIfNeeded(
			candidate.ResourcesDramaSeries, videoMetadataPriorityMedium, "", force, true, false,
		)
		if !queued || done == nil {
			continue
		}
		if probeErr := <-done; probeErr != nil {
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("本轮有 %d 个视频信息采集失败", failed)
	}
	return nil
}

func validVideoMetadataRunMode(mode string) bool {
	return mode == VideoMetadataRunMissing ||
		mode == VideoMetadataRunMissingStale ||
		mode == VideoMetadataRunFailed ||
		mode == VideoMetadataRunAll
}

func (VideoMetadata) runBatch(taskID string, maxItems int) {
	var task models.VideoMetadataBatchTask
	if err := core.DBS().Where("id = ?", taskID).First(&task).Error; err != nil {
		return
	}
	ids, err := videoMetadataBatchScopeIDs(task)
	if err != nil {
		updateVideoMetadataBatchError(taskID, err)
		return
	}
	if task.RunMode == VideoMetadataRunMissingStale {
		if err := markChangedVideoMetadataScope(task.ScopeMode, ids); err != nil {
			updateVideoMetadataBatchError(taskID, err)
			return
		}
		total, err := countVideoMetadataCandidates(ids, task.ScopeMode, task.RunMode, nil)
		if err != nil {
			updateVideoMetadataBatchError(taskID, err)
			return
		}
		if maxItems > 0 && total > int64(maxItems) {
			total = int64(maxItems)
		}
		if err := core.DBS().Model(&models.VideoMetadataBatchTask{}).
			Where("id = ?", taskID).Update("total", total).Error; err != nil {
			updateVideoMetadataBatchError(taskID, err)
			return
		}
	}
	processed := 0
	seen := map[string]struct{}{}
	for {
		if maxItems > 0 && processed >= maxItems {
			break
		}
		if !videoMetadataBatchCanContinue(taskID) {
			return
		}
		limit := 20
		if maxItems > 0 && maxItems-processed < limit {
			limit = maxItems - processed
		}
		excludeIDs := make([]string, 0, len(seen))
		for id := range seen {
			excludeIDs = append(excludeIDs, id)
		}
		candidates, err := findVideoMetadataCandidates(ids, task.ScopeMode, task.RunMode, limit, excludeIDs)
		if err != nil {
			updateVideoMetadataBatchError(taskID, err)
			return
		}
		if len(candidates) == 0 {
			break
		}
		for _, candidate := range candidates {
			seen[candidate.ID] = struct{}{}
			if !videoMetadataBatchCanContinue(taskID) {
				return
			}
			updateVideoMetadataBatchCurrent(taskID, candidate.Src)
			force := task.RunMode == VideoMetadataRunAll || task.RunMode == VideoMetadataRunFailed
			queued, done := (VideoMetadata{}).enqueueIfNeeded(
				candidate.ResourcesDramaSeries, videoMetadataPriorityHigh, taskID, force, true, false,
			)
			if !queued || done == nil {
				incrementVideoMetadataBatch(taskID, "skipped")
				processed++
				continue
			}
			if err := <-done; err != nil {
				incrementVideoMetadataBatch(taskID, "failed")
			} else {
				incrementVideoMetadataBatch(taskID, "success")
			}
			processed++
		}
	}
	now := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.VideoMetadataBatchTask{}).Where("id = ? AND status = ?", taskID, models.VideoMetadataBatchStatusRunning).
		Updates(map[string]interface{}{"status": models.VideoMetadataBatchStatusCompleted, "finished_at": &now, "current_src": ""})
}

func videoMetadataBatchScopeIDs(task models.VideoMetadataBatchTask) ([]string, error) {
	if task.ScopeMode == models.VideoMetadataScopeAll {
		var ids []string
		err := core.DBS().Model(&models.FilesBases{}).Order("sort, id").Pluck("id", &ids).Error
		return ids, err
	}
	var ids []string
	err := core.DBS().Model(&models.VideoMetadataBatchTaskFilesBases{}).
		Where("task_id = ?", task.ID).Pluck("files_bases_id", &ids).Error
	return ids, err
}

func findVideoMetadataCandidates(ids []string, scopeMode, runMode string, limit int, excludeIDs []string) ([]videoMetadataCandidate, error) {
	var list []videoMetadataCandidate
	q := videoMetadataCandidateQuery(ids, scopeMode, runMode, excludeIDs).
		Select("ds.*, r.filesBases_id")
	err := q.Order("ds.id").Limit(limit).Scan(&list).Error
	return list, err
}

func countVideoMetadataCandidates(ids []string, scopeMode, runMode string, excludeIDs []string) (int64, error) {
	var count int64
	err := videoMetadataCandidateQuery(ids, scopeMode, runMode, excludeIDs).
		Distinct("ds.id").Count(&count).Error
	return count, err
}

// markChangedVideoMetadataScope 在用户明确执行“缺失并更新失效项”时检查已采集文件。
// 普通列表查询不访问磁盘；这里可以承担 NAS I/O，并只把确实变化或无法访问的记录标记为 stale。
func markChangedVideoMetadataScope(scopeMode string, ids []string) error {
	var files []collectedVideoMetadataFile
	if err := collectedVideoMetadataScopeQuery(scopeMode, ids).
		Order("ds.id").Scan(&files).Error; err != nil {
		return err
	}
	return markChangedVideoMetadataFiles(files)
}

// markNextChangedVideoMetadataScope 让空闲补齐每轮只检查有限数量的成功记录，
// 并以分集 ID 游标逐步遍历整个范围，避免每分钟从头扫描 NAS。
func markNextChangedVideoMetadataScope(scopeMode string, ids []string, limit int) error {
	if limit < 1 {
		return nil
	}
	var files []collectedVideoMetadataFile
	q := collectedVideoMetadataScopeQuery(scopeMode, ids)
	if lastIdleVideoMetadataCheckID != "" {
		q = q.Where("ds.id > ?", lastIdleVideoMetadataCheckID)
	}
	if err := q.Order("ds.id").Limit(limit).Scan(&files).Error; err != nil {
		return err
	}
	if len(files) == 0 {
		lastIdleVideoMetadataCheckID = ""
		return nil
	}
	lastIdleVideoMetadataCheckID = files[len(files)-1].DramaSeriesID
	return markChangedVideoMetadataFiles(files)
}

func collectedVideoMetadataScopeQuery(scopeMode string, ids []string) *gorm.DB {
	q := core.DBS().Table("resourcesDramaSeries ds").
		Select(`ds.id AS drama_series_id, ds.src,
			vm.file_size, vm.file_modified_time`).
		Joins("JOIN resources r ON r.id = ds.resources_id").
		Joins("JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id").
		Where("r.mode IN ?", []datatype.E_resourceMode{
			datatype.E_resourceMode_Movies,
			datatype.E_resourceMode_VideoLink,
		}).
		Where("vm.probe_status = ? AND vm.metadata_version >= ?",
			models.VideoMetadataStatusSuccess, CurrentVideoMetadataVersion)
	if scopeMode == models.VideoMetadataScopeSelected {
		q = q.Where("r.filesBases_id IN ?", ids)
	}
	return q
}

func markChangedVideoMetadataFiles(files []collectedVideoMetadataFile) error {
	staleIDs := make([]string, 0)
	for _, file := range files {
		stat, err := os.Stat(file.Src)
		if err != nil ||
			stat.Size() != file.FileSize ||
			stat.ModTime().UnixMilli() != file.FileModifiedTime {
			staleIDs = append(staleIDs, file.DramaSeriesID)
		}
	}
	if len(staleIDs) == 0 {
		return nil
	}
	return core.DBS().Model(&models.ResourcesVideoMetadata{}).
		Where("drama_series_id IN ?", staleIDs).
		Updates(map[string]interface{}{
			"probe_status":     models.VideoMetadataStatusStale,
			"metadata_version": 0,
			"next_retry_time":  nil,
			"retry_count":      0,
			"error_code":       "",
			"error_message":    "",
		}).Error
}

func videoMetadataCandidateQuery(ids []string, scopeMode, runMode string, excludeIDs []string) *gorm.DB {
	q := core.DBS().Table("resourcesDramaSeries ds").
		Joins("JOIN resources r ON r.id = ds.resources_id").
		Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id = ds.id").
		Where("r.mode IN ?", []datatype.E_resourceMode{datatype.E_resourceMode_Movies, datatype.E_resourceMode_VideoLink}).
		Where("ds.src <> ''")
	if scopeMode == models.VideoMetadataScopeSelected {
		q = q.Where("r.filesBases_id IN ?", ids)
	}
	if len(excludeIDs) > 0 {
		q = q.Where("ds.id NOT IN ?", excludeIDs)
	}
	now := core.TimeNow()
	switch runMode {
	case VideoMetadataRunFailed:
		q = q.Where("vm.probe_status = ? AND (vm.next_retry_time IS NULL OR vm.next_retry_time <= ?)", models.VideoMetadataStatusFailed, now)
	case VideoMetadataRunAll:
		q = q.Where(`vm.drama_series_id IS NULL
			OR vm.probe_status = ?
			OR (vm.probe_status NOT IN (?, ?) AND vm.metadata_version < ?)`,
			models.VideoMetadataStatusStale,
			models.VideoMetadataStatusFailed,
			models.VideoMetadataStatusProcessing,
			CurrentVideoMetadataVersion)
	case VideoMetadataRunMissing:
		q = q.Where("vm.drama_series_id IS NULL")
	default:
		q = q.Where(`vm.drama_series_id IS NULL
			OR vm.probe_status = ?
			OR (vm.probe_status NOT IN (?, ?) AND vm.metadata_version < ?)`,
			models.VideoMetadataStatusStale,
			models.VideoMetadataStatusFailed,
			models.VideoMetadataStatusProcessing,
			CurrentVideoMetadataVersion)
	}
	return q
}

func markVideoMetadataScopeStale(scopeMode string, ids []string) error {
	q := core.DBS().Model(&models.ResourcesVideoMetadata{}).
		Where(`drama_series_id IN (
			SELECT ds.id FROM resourcesDramaSeries ds
			JOIN resources r ON r.id = ds.resources_id
			WHERE r.mode IN ('movies', 'videoLink')
		)`)
	if scopeMode == models.VideoMetadataScopeSelected {
		q = q.Where(`drama_series_id IN (
			SELECT ds.id FROM resourcesDramaSeries ds
			JOIN resources r ON r.id = ds.resources_id
			WHERE r.filesBases_id IN ?
		)`, ids)
	}
	return q.Updates(map[string]interface{}{
		"probe_status":     models.VideoMetadataStatusStale,
		"metadata_version": 0,
	}).Error
}

func (VideoMetadata) BatchStatus() (*models.VideoMetadataBatchTask, error) {
	var task models.VideoMetadataBatchTask
	err := core.DBS().Order("created_at DESC").First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return &task, nil
	}
	return &task, err
}

func (VideoMetadata) PauseBatch() error {
	return core.DBS().Model(&models.VideoMetadataBatchTask{}).
		Where("status = ?", models.VideoMetadataBatchStatusRunning).
		Update("status", models.VideoMetadataBatchStatusPaused).Error
}

func (VideoMetadata) ResumeBatch() error {
	videoMetadataBatchMu.Lock()
	defer videoMetadataBatchMu.Unlock()

	var task models.VideoMetadataBatchTask
	err := core.DBS().Where("status = ?", models.VideoMetadataBatchStatusPaused).Order("created_at DESC").First(&task).Error
	if err != nil {
		return err
	}
	if err := core.DBS().Model(&task).Update("status", models.VideoMetadataBatchStatusRunning).Error; err != nil {
		return err
	}
	go (VideoMetadata{}).runBatch(task.ID, 0)
	return nil
}

func (VideoMetadata) StopBatch() error {
	now := datatype.CustomTime(core.TimeNow())
	return core.DBS().Model(&models.VideoMetadataBatchTask{}).
		Where("status IN ?", []string{models.VideoMetadataBatchStatusRunning, models.VideoMetadataBatchStatusPaused}).
		Updates(map[string]interface{}{"status": models.VideoMetadataBatchStatusStopped, "finished_at": &now, "current_src": ""}).Error
}

func videoMetadataBatchCanContinue(taskID string) bool {
	var status string
	err := core.DBS().Model(&models.VideoMetadataBatchTask{}).Where("id = ?", taskID).Pluck("status", &status).Error
	return err == nil && status == models.VideoMetadataBatchStatusRunning
}

func incrementVideoMetadataBatch(taskID, field string) {
	if field != "success" && field != "failed" && field != "skipped" {
		return
	}
	core.DBS().Model(&models.VideoMetadataBatchTask{}).Where("id = ?", taskID).
		UpdateColumn(field, gorm.Expr(field+" + ?", 1))
}

func updateVideoMetadataBatchCurrent(taskID, src string) {
	core.DBS().Model(&models.VideoMetadataBatchTask{}).Where("id = ?", taskID).
		Update("current_src", src)
}

func updateVideoMetadataBatchError(taskID string, err error) {
	now := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.VideoMetadataBatchTask{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status": models.VideoMetadataBatchStatusStopped, "last_error": err.Error(),
			"finished_at": &now, "current_src": "",
		})
}
