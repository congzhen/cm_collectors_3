package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/internal/maintenance"
	"cm_collectors_server/models"
	ff "cm_collectors_server/processorsFFmpeg"
	"context"
	"crypto/sha256"
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

	"github.com/skratchdot/open-golang/open"
	"gorm.io/gorm"
)

type VideoMoments struct{}

const previewRoot = "./runtime/cache/videoPreviews"

// Injectable database accessor keeps feature tests isolated from the application database.
var videoMomentsDB = core.DBS
var videoMomentDuration = func(src string) (float64, error) {
	info, err := (ff.VideoInfo{}).GetVideoFormatInfo(src)
	if err != nil {
		return 0, err
	}
	return (ff.VideoInfo{}).GetVideoDuration(info), nil
}

func momentSource(id string) (models.ResourcesDramaSeries, models.Resources, string, error) {
	var d models.ResourcesDramaSeries
	var r models.Resources
	if err := videoMomentsDB().First(&d, "id = ?", id).Error; err != nil {
		return d, r, "", errors.New("分集不存在")
	}
	if err := videoMomentsDB().First(&r, "id = ? AND status = ?", d.ResourcesID, true).Error; err != nil {
		return d, r, "", errors.New("资源已删除")
	}
	if r.Mode != datatype.E_resourceMode_Movies {
		return d, r, "", errors.New("仅支持本地视频")
	}
	stat, err := os.Stat(d.Src)
	if err != nil || !stat.Mode().IsRegular() {
		return d, r, "", errors.New("源视频不存在或不可读取")
	}
	// Used only to name a new cache directory. Existing previews are never invalidated by this key.
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d", d.ID, stat.Size(), stat.ModTime().UnixNano()))))
	return d, r, key, nil
}

func validateMoment(v *models.VideoMoment, duration float64) error {
	v.Title = strings.TrimSpace(v.Title)
	if v.Title == "" {
		v.Title = "精彩时刻"
	}
	if len([]rune(v.Title)) > 200 || len([]rune(v.Note)) > 4000 {
		return errors.New("名称最多200字，备注最多4000字")
	}
	if math.IsNaN(v.Start) || math.IsInf(v.Start, 0) || math.IsNaN(v.End) || math.IsInf(v.End, 0) || v.Start < 0 || v.End < 0 || (v.End > 0 && v.End <= v.Start) {
		return errors.New("结束时间必须大于起点，时间点必须为非负数")
	}
	if duration > 0 && (v.Start >= duration || v.End > duration) {
		return errors.New("收藏时间超出视频时长")
	}
	return nil
}

func (VideoMoments) Save(id string, input models.VideoMoment) (*models.VideoMoment, error) {
	var old models.VideoMoment
	if id != "" {
		if err := videoMomentsDB().First(&old, "id = ?", id).Error; err != nil {
			return nil, err
		}
		input.DramaSeriesID = old.DramaSeriesID
		input.CreatedAt = old.CreatedAt
	}
	d, r, key, err := momentSource(input.DramaSeriesID)
	// Text-only edits preserve the original times, source snapshot and cover,
	// even after the source is replaced with a shorter or unreadable video.
	if id != "" && input.Start == old.Start && input.End == old.End {
		if e := validateMoment(&input, 0); e != nil {
			return nil, e
		}
		old.Title, old.Note, old.UpdatedAt = input.Title, input.Note, time.Now()
		old.Available, old.ResourceTitle = err == nil, r.Title
		e := videoMomentsDB().Model(&models.VideoMoment{}).Where("id = ?", id).Select("Title", "Note", "UpdatedAt").Updates(&old).Error
		return &old, e
	}
	if err != nil {
		return nil, err
	}
	// DurationSeconds is rounded metadata and may belong to a replaced source.
	// Only reuse a precise duration tied to the current source snapshot.
	duration := float64(0)
	var preview models.VideoPreview
	if videoMomentsDB().First(&preview, "drama_series_id = ?", d.ID).Error == nil && preview.SourceKey == key && preview.Duration > 0 {
		duration = preview.Duration
	}
	if duration <= 0 {
		duration, err = videoMomentDuration(d.Src)
		if err != nil {
			return nil, errors.New("无法读取视频时长")
		}
	}
	if duration <= 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return nil, errors.New("无法读取视频时长")
	}
	if err = validateMoment(&input, duration); err != nil {
		return nil, err
	}
	input.ID = id
	if id == "" {
		input.ID = core.GenerateUniqueID()
		input.CreatedAt = time.Now()
	}
	input.ResourceID = r.ID
	input.SourceKey = key
	input.UpdatedAt = time.Now()
	if id == "" {
		err = videoMomentsDB().Create(&input).Error
	} else {
		err = videoMomentsDB().Model(&models.VideoMoment{}).Where("id = ?", id).Select("Title", "Note", "Start", "End", "ResourceID", "SourceKey", "UpdatedAt").Updates(&input).Error
	}
	if err != nil {
		return nil, err
	}
	input.Available = true
	input.ResourceTitle = r.Title
	// A cover failure must not lose the bookmark; the cover request can retry later.
	_, _ = persistMomentCover(input)
	return &input, nil
}

func (VideoMoments) List(resourceID string, dramaSeriesID string, page int) ([]models.VideoMoment, error) {
	if page < 1 {
		page = 1
	}
	db := videoMomentsDB().Model(&models.VideoMoment{})
	if resourceID != "" {
		db = db.Where("resource_id = ?", resourceID)
	}
	if dramaSeriesID != "" {
		db = db.Where("drama_series_id = ?", dramaSeriesID)
	}
	list := []models.VideoMoment{}
	if err := db.Order("created_at desc, id desc").Offset((page - 1) * 50).Limit(50).Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		_, r, _, err := momentSource(list[i].DramaSeriesID)
		list[i].Available = err == nil
		list[i].ResourceTitle = r.Title
	}
	return list, nil
}
func (VideoMoments) ListLibrary(filesBasesID, keyword string, page int) ([]models.VideoMoment, int64, error) {
	if filesBasesID == "" {
		return nil, 0, errors.New("请选择文件库")
	}
	if page < 1 {
		page = 1
	}
	resources := videoMomentsDB().Model(&models.Resources{}).Select("id").Where("filesBases_id = ?", filesBasesID)
	db := videoMomentsDB().Model(&models.VideoMoment{}).Where("resource_id IN (?)", resources)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		titles := videoMomentsDB().Model(&models.Resources{}).Select("id").Where("filesBases_id = ? AND title LIKE ?", filesBasesID, pattern)
		db = db.Where("title LIKE ? OR note LIKE ? OR resource_id IN (?)", pattern, pattern, titles)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := []models.VideoMoment{}
	if err := db.Order("created_at desc, id desc").Offset((page - 1) * 50).Limit(50).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	for i := range list {
		_, r, _, err := momentSource(list[i].DramaSeriesID)
		list[i].Available = err == nil
		list[i].ResourceTitle = r.Title
	}
	return list, total, nil
}
func (VideoMoments) Delete(id string) error {
	momentCoverMutex.Lock()
	defer momentCoverMutex.Unlock()
	if err := os.RemoveAll(momentCoverDir(id)); err != nil {
		return err
	}
	return videoMomentsDB().Where("id = ?", id).Delete(&models.VideoMoment{}).Error
}

var previewQueue = struct {
	sync.Mutex
	initialized bool
	lastError   string
	paused      bool
	running     bool
}{}

func initPreviewQueue() error {
	previewQueue.Lock()
	defer previewQueue.Unlock()
	if previewQueue.initialized {
		return nil
	}
	if err := videoMomentsDB().Model(&models.VideoPreview{}).Where("status = ?", "processing").Update("status", "queued").Error; err != nil {
		return err
	}
	previewQueue.initialized = true
	return nil
}

type libraryVideoPreviewConfig struct {
	Generation string `json:"videoPreviewGeneration"`
	Interval   int    `json:"videoPreviewInterval"`
}

func videoPreviewLibraryConfig(libraryID string) (libraryVideoPreviewConfig, error) {
	config := libraryVideoPreviewConfig{Generation: "playback", Interval: 30}
	var setting models.FilesBasesSetting
	err := videoMomentsDB().First(&setting, "filesBases_id = ?", libraryID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	effective, err := models.EffectiveLibraryConfig(videoMomentsDB(), libraryID, "display", setting.ConfigJsonData)
	if err != nil {
		return config, err
	}
	if effective != "" {
		if err := json.Unmarshal([]byte(effective), &config); err != nil {
			return config, err
		}
	}
	if config.Interval == 0 {
		config.Interval = 30
	}
	if config.Interval != 10 && config.Interval != 30 && config.Interval != 60 && config.Interval != 120 {
		return config, errors.New("默认抽图间隔必须为10、30、60或120秒")
	}
	return config, nil
}

func (VideoMoments) Preview(id string) (models.VideoPreview, error) {
	var v models.VideoPreview
	err := videoMomentsDB().First(&v, "drama_series_id = ?", id).Error
	if err == nil {
		// Only check physical cache loss, never source-file changes.
		if v.Status == "ready" || (v.Status == "failed" && v.Completed > 0) {
			missing := v.Status == "ready" && (v.Completed <= 0 || v.Completed != v.Count)
			for i := 0; !missing && i < v.Completed; i++ {
				missing = !previewFileExists(filepath.Join(previewDir(v), fmt.Sprintf("%06d.jpg", i)))
			}
			if missing {
				v.Status = "missing"
				v.Completed = 0
				v.Error = "时间轴缓存已删除或不完整，可重新生成"
			}
		}
		return v, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return v, err
	}
	_, r, key, err := momentSource(id)
	if err != nil {
		return v, err
	}
	config, err := videoPreviewLibraryConfig(r.FilesBasesID)
	return models.VideoPreview{DramaSeriesID: id, Status: "none", Interval: config.Interval, SourceKey: key}, err
}

// Source replacement never removes previews; explicit deletion and configured capacity eviction do.
func (VideoMoments) DeletePreview(id string) error {
	if err := initPreviewQueue(); err != nil {
		return err
	}
	previewQueue.Lock()
	defer previewQueue.Unlock()
	var v models.VideoPreview
	if err := videoMomentsDB().First(&v, "drama_series_id = ?", id).Error; err != nil {
		return err
	}
	if v.Status == "processing" || (v.Status == "queued" && (!previewQueue.paused || previewQueue.running)) {
		return errors.New("请先暂停队列，等待当前抽图完成后再删除")
	}
	if err := os.RemoveAll(filepath.Join(previewRoot, v.SourceKey)); err != nil {
		return err
	}
	return videoMomentsDB().Where("drama_series_id = ?", id).Delete(&models.VideoPreview{}).Error
}
func (VideoMoments) Queue(id string, interval int) (models.VideoPreview, error) {
	if err := initPreviewQueue(); err != nil {
		return models.VideoPreview{}, err
	}
	if interval != 10 && interval != 30 && interval != 60 && interval != 120 {
		return models.VideoPreview{}, errors.New("抽图间隔必须为10、30、60或120秒")
	}
	previewQueue.Lock()
	defer previewQueue.Unlock()
	v, err := (VideoMoments{}).Preview(id)
	if err != nil {
		return v, err
	}
	if v.Status == "queued" || v.Status == "processing" || v.Status == "ready" {
		startPreviewWorkerLocked()
		return v, nil
	}
	_, _, key, err := momentSource(id)
	if err != nil {
		return v, err
	}
	setting, err := (VideoMoments{}).Setting()
	if err != nil {
		return v, err
	}
	if v.Status == "none" || v.Status == "missing" {
		if v.Status == "missing" {
			if err := os.RemoveAll(filepath.Join(previewRoot, v.SourceKey)); err != nil {
				return v, err
			}
		}
		v = models.VideoPreview{DramaSeriesID: id, SourceKey: key, Status: "queued", Interval: interval, Width: setting.Width}
	} else {
		v.Status = "queued"
		v.Error = ""
	}
	if err = videoMomentsDB().Save(&v).Error; err != nil {
		return v, err
	}
	startPreviewWorkerLocked()
	return v, nil
}

// QueueOnPlayback respects the actual resource library, independent of the client's selected library.
// Existing cache is retained; physical cache loss permits rebuilding without source-change invalidation.
func (VideoMoments) QueueOnPlayback(id string) (bool, error) {
	var d models.ResourcesDramaSeries
	if err := videoMomentsDB().First(&d, "id = ?", id).Error; err != nil {
		return false, err
	}
	var r models.Resources
	if err := videoMomentsDB().First(&r, "id = ? AND status = ?", d.ResourcesID, true).Error; err != nil {
		return false, err
	}
	if r.Mode != datatype.E_resourceMode_Movies {
		return false, nil
	}
	config, err := videoPreviewLibraryConfig(r.FilesBasesID)
	if err != nil {
		return false, err
	}
	if config.Generation != "" && config.Generation != "playback" {
		return false, nil
	}
	current, err := (VideoMoments{}).Preview(id)
	if err != nil {
		return false, err
	}
	if current.Status != "none" && current.Status != "missing" {
		return false, nil
	}
	_, err = (VideoMoments{}).Queue(id, config.Interval)
	return err == nil, err
}

func startPreviewWorkerLocked() {
	if !previewQueue.running && !previewQueue.paused {
		previewQueue.running = true
		maintenance.Go(runPreviewQueue)
	}
}
func runPreviewQueue() {
	for {
		previewQueue.Lock()
		if previewQueue.paused {
			previewQueue.running = false
			previewQueue.Unlock()
			return
		}
		var v models.VideoPreview
		err := videoMomentsDB().Where("status = ?", "queued").Order("updated_at asc").First(&v).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				previewQueue.lastError = "读取预览队列失败：" + err.Error()
			}
			previewQueue.running = false
			previewQueue.Unlock()
			return
		}
		v.Status = "processing"
		err = videoMomentsDB().Save(&v).Error
		previewQueue.Unlock()
		if err != nil {
			previewQueue.Lock()
			previewQueue.lastError = "保存预览任务失败：" + err.Error()
			previewQueue.running = false
			previewQueue.Unlock()
			return
		}
		err = generatePreview(&v)
		if err != nil {
			v.Status = "failed"
			v.Error = err.Error()
		} else if v.Status != "queued" {
			v.Status = "ready"
		}
		previewQueue.Lock()
		if err := videoMomentsDB().Save(&v).Error; err != nil {
			previewQueue.lastError = "无法保存预览任务状态：" + err.Error()
			previewQueue.running = false
			previewQueue.Unlock()
			return
		}
		if err := evictPreviews(); err != nil {
			previewQueue.lastError = "清理预览缓存失败：" + err.Error()
		}
		previewQueue.Unlock()
	}
}
func previewDir(v models.VideoPreview) string {
	return filepath.Join(previewRoot, v.SourceKey, fmt.Sprintf("%d-%d", v.Interval, v.Width))
}
func generatePreview(v *models.VideoPreview) error {
	d, _, _, err := momentSource(v.DramaSeriesID)
	if err != nil {
		return err
	}
	tool, err := (ff.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return err
	}
	// Keep an existing sampling grid when resuming; source replacement is not an invalidation signal.
	if v.Count == 0 {
		info, err := (ff.VideoInfo{}).GetVideoFormatInfo(d.Src)
		if err != nil {
			return errors.New("无法读取视频信息")
		}
		v.Duration = (ff.VideoInfo{}).GetVideoDuration(info)
		if v.Duration <= 0 || math.IsInf(v.Duration, 0) || math.IsNaN(v.Duration) {
			return errors.New("视频时长无效")
		}
		// Bound disk and decoding costs for unusually long media.
		v.Interval = max(v.Interval, int(math.Ceil(v.Duration/1200)))
		v.Count = int(math.Ceil(v.Duration / float64(v.Interval)))
	}
	if err = os.MkdirAll(previewDir(*v), 0755); err != nil {
		return err
	}
	// A paused task may lose files that were previously completed. Resume at
	// the first missing frame, while retaining every intact cached image.
	for i := 0; i < v.Completed; i++ {
		if !previewFileExists(filepath.Join(previewDir(*v), fmt.Sprintf("%06d.jpg", i))) {
			v.Completed = i
			break
		}
	}
	for i := v.Completed; i < v.Count; i++ {
		previewQueue.Lock()
		paused := previewQueue.paused
		previewQueue.Unlock()
		if paused {
			v.Status = "queued"
			return nil
		}
		file := filepath.Join(previewDir(*v), fmt.Sprintf("%06d.jpg", i))
		if !previewFileExists(file) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			cmd := ff.CreateCommandContext(ctx, tool, "-nostdin", "-v", "error", "-ss", strconv.Itoa(i*v.Interval), "-i", d.Src, "-map", "0:v:0", "-frames:v", "1", "-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", v.Width, v.Width*9/16/2*2, v.Width, v.Width*9/16/2*2), "-threads", "1", "-q:v", "5", "-f", "image2", "-y", file+".tmp")
			err = cmd.Run()
			cancel()
			if err != nil {
				os.Remove(file + ".tmp")
				return errors.New("预览图生成失败或超时，请重试")
			}
			if err = os.Rename(file+".tmp", file); err != nil {
				return err
			}
		}
		v.Completed = i + 1
		if err = videoMomentsDB().Save(v).Error; err != nil {
			return err
		}
	}
	return nil
}
func (VideoMoments) Image(id string, index int) (string, error) {
	// Image requests validate only their own frame. Calling Preview here would
	// scan every cached frame for every thumbnail (quadratic filesystem work).
	var v models.VideoPreview
	if err := videoMomentsDB().First(&v, "drama_series_id = ?", id).Error; err != nil {
		return "", err
	}
	if index < 0 || index >= v.Completed {
		return "", errors.New("预览图尚未生成")
	}
	path := filepath.Join(previewDir(v), fmt.Sprintf("%06d.jpg", index))
	if !previewFileExists(path) {
		return "", errors.New("预览图已删除或不完整")
	}
	return path, nil
}
func (VideoMoments) QueueControl(action string) error {
	if err := initPreviewQueue(); err != nil {
		return err
	}
	previewQueue.Lock()
	defer previewQueue.Unlock()
	switch action {
	case "pause":
		previewQueue.paused = true
	case "resume":
		if !previewQueue.running {
			if err := videoMomentsDB().Model(&models.VideoPreview{}).Where("status = ?", "processing").Update("status", "queued").Error; err != nil {
				return err
			}
		}
		previewQueue.lastError = ""
		previewQueue.paused = false
		startPreviewWorkerLocked()
	case "clear":
		if previewQueue.running {
			return errors.New("请先暂停队列并等待当前抽图结束")
		}
		// Fixed application-owned cache root, never a source path.
		if err := os.RemoveAll(previewRoot); err != nil {
			return err
		}
		return videoMomentsDB().Where("1 = 1").Delete(&models.VideoPreview{}).Error
	default:
		return errors.New("未知操作")
	}
	return nil
}
func (VideoMoments) QueueStatus() (map[string]interface{}, error) {
	if err := initPreviewQueue(); err != nil {
		return nil, err
	}
	previewQueue.Lock()
	paused, running, queueError := previewQueue.paused, previewQueue.running, previewQueue.lastError
	previewQueue.Unlock()
	var size int64
	filepath.Walk(previewRoot, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	var queued int64
	if err := videoMomentsDB().Model(&models.VideoPreview{}).Where("status IN ?", []string{"queued", "processing"}).Count(&queued).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"paused": paused, "running": running, "bytes": size, "queued": queued, "error": queueError}, nil
}

func (VideoMoments) Setting() (models.VideoPreviewSetting, error) {
	v := models.VideoPreviewSetting{ID: 1, MaxMB: 1024, Width: 240}
	err := videoMomentsDB().First(&v, "id = ?", 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.VideoPreviewSetting{ID: 1, MaxMB: 1024, Width: 240}, nil
	}
	return v, err
}
func (VideoMoments) SaveSetting(v models.VideoPreviewSetting) error {
	if v.MaxMB != 0 && (v.MaxMB < 64 || v.MaxMB > 1048576) {
		return errors.New("缓存上限应为64至1048576 MB，0表示不限")
	}
	if v.Width != 160 && v.Width != 240 && v.Width != 320 && v.Width != 480 {
		return errors.New("请选择有效的缩略图宽度")
	}
	v.ID = 1
	previewQueue.Lock()
	defer previewQueue.Unlock()
	if err := videoMomentsDB().Save(&v).Error; err != nil {
		return err
	}
	return evictPreviews()
}
func cacheBytes() int64 {
	var size int64
	filepath.Walk(previewRoot, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// Capacity eviction is independent of source-file changes. Oldest completed previews go first.
func evictPreviews() error {
	setting, err := (VideoMoments{}).Setting()
	if err != nil || setting.MaxMB == 0 {
		return err
	}
	size := cacheBytes()
	limit := setting.MaxMB * 1024 * 1024
	if size <= limit {
		return nil
	}
	var list []models.VideoPreview
	if err = videoMomentsDB().Where("status NOT IN ?", []string{"queued", "processing"}).Order("updated_at asc, drama_series_id asc").Find(&list).Error; err != nil {
		return err
	}
	for _, v := range list {
		if size <= limit {
			break
		}
		if err = os.RemoveAll(filepath.Join(previewRoot, v.SourceKey)); err != nil {
			return err
		}
		if err = videoMomentsDB().Where("drama_series_id = ?", v.DramaSeriesID).Delete(&models.VideoPreview{}).Error; err != nil {
			return err
		}
		size = cacheBytes()
	}
	return nil
}

var openVideoPreviewFolder = open.Run

func (VideoMoments) OpenCacheFolder() error {
	if err := (Play{}).AllowServerOpenFile(); err != nil {
		return err
	}
	path, err := filepath.Abs(previewRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return openVideoPreviewFolder(path)
}
