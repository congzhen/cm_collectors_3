package processors

import (
	"cm_collectors_server/config"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	ff "cm_collectors_server/processorsFFmpeg"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func momentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err = db.AutoMigrate(&models.VideoMoment{}, &models.VideoPreview{}, &models.VideoPreviewSetting{}, &models.FilesBasesSetting{}, &models.Resources{}, &models.ResourcesDramaSeries{}, &models.SharedLibraryConfig{}, &models.LibraryConfigFollow{}); err != nil {
		t.Fatal(err)
	}
	previous := videoMomentsDB
	previousProbe := videoMomentDuration
	previousCapture := videoMomentCapture
	videoMomentCapture = func(_ string, _ float64, _ int, path string) error {
		return os.WriteFile(path, []byte("bookmark cover"), 0600)
	}
	videoMomentDuration = func(string) (float64, error) { return 120.75, nil }
	videoMomentsDB = func() *gorm.DB { return db }
	previewQueue.paused = true
	previewQueue.running = false
	previewQueue.initialized = false
	previewQueue.lastError = ""
	t.Chdir(t.TempDir())
	t.Cleanup(func() {
		videoMomentsDB = previous
		videoMomentDuration = previousProbe
		videoMomentCapture = previousCapture
		previewQueue.paused = false
		previewQueue.initialized = false
		previewQueue.lastError = ""
		sqlDB.Close()
	})
	return db
}
func TestMomentTimeValidation(t *testing.T) {
	for _, v := range []models.VideoMoment{{Start: -1}, {Start: 2, End: 1}, {Start: 2, End: 2}, {Start: 100}, {Start: 1, End: 101}, {Start: math.NaN()}, {End: math.Inf(1)}} {
		if validateMoment(&v, 100) == nil {
			t.Fatalf("accepted invalid time: %+v", v)
		}
	}
	v := models.VideoMoment{Start: 0, End: 100}
	if err := validateMoment(&v, 100); err != nil {
		t.Fatal(err)
	}
	v = models.VideoMoment{Start: 4, End: 0}
	if err := validateMoment(&v, 100); err != nil {
		t.Fatal(err)
	}
}
func TestLibraryMomentsPaginationAndSearch(t *testing.T) {
	db := momentTestDB(t)
	for _, r := range []models.Resources{{ID: "a", FilesBasesID: "library-a", Title: "资源甲"}, {ID: "b", FilesBasesID: "library-b", Title: "资源乙"}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 65; i++ {
		item := models.VideoMoment{ID: fmt.Sprintf("a-%03d", i), ResourceID: "a", DramaSeriesID: "missing", Title: fmt.Sprintf("片段%d", i)}
		if i == 0 {
			item.Note = "特别备注"
		}
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.VideoMoment{ID: "other", ResourceID: "b", Title: "特别备注"}).Error; err != nil {
		t.Fatal(err)
	}
	service := VideoMoments{}
	first, total, err := service.ListLibrary("library-a", "", 1)
	if err != nil || total != 65 || len(first) != 50 {
		t.Fatalf("first page: %d/%d, %v", len(first), total, err)
	}
	second, total, err := service.ListLibrary("library-a", "", 2)
	if err != nil || total != 65 || len(second) != 15 {
		t.Fatalf("second page: %d/%d, %v", len(second), total, err)
	}
	seen := map[string]bool{}
	for _, item := range append(first, second...) {
		if seen[item.ID] || item.ResourceID != "a" {
			t.Fatalf("duplicate or cross-library result: %+v", item)
		}
		seen[item.ID] = true
	}
	for _, query := range []struct {
		text  string
		count int64
	}{{"特别备注", 1}, {"片段64", 1}, {"资源甲", 65}, {"资源乙", 0}} {
		_, total, err = service.ListLibrary("library-a", query.text, 1)
		if err != nil || total != query.count {
			t.Fatalf("search %q: %d, %v", query.text, total, err)
		}
	}
	if _, _, err = service.ListLibrary("", "", 1); err == nil {
		t.Fatal("missing library must not expose all bookmarks")
	}
}
func TestVideoPreviewUsesEffectiveSharedConfig(t *testing.T) {
	db := momentTestDB(t)
	for _, id := range []string{"following", "independent"} {
		if err := db.Create(&models.FilesBasesSetting{FilesBasesID: id, ConfigJsonData: `{"videoPreviewGeneration":"playback","videoPreviewInterval":10}`}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.SharedLibraryConfig{Module: "display", Revision: 1, Config: `{"videoPreviewGeneration":"manual","videoPreviewInterval":120}`}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LibraryConfigFollow{FilesBasesID: "following", Module: "display"}).Error; err != nil {
		t.Fatal(err)
	}
	config, err := videoPreviewLibraryConfig("following")
	if err != nil || config.Generation != "manual" || config.Interval != 120 {
		t.Fatalf("shared config not applied: %+v, %v", config, err)
	}
	config, err = videoPreviewLibraryConfig("independent")
	if err != nil || config.Generation != "playback" || config.Interval != 10 {
		t.Fatalf("independent config changed: %+v, %v", config, err)
	}
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r-shared", FilesBasesID: "following", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d-shared", ResourcesID: "r-shared", Src: src, DurationSeconds: 120})
	queued, err := (VideoMoments{}).QueueOnPlayback("d-shared")
	if err != nil || queued {
		t.Fatalf("manual shared config queued: %v, %v", queued, err)
	}
	db.Model(&models.SharedLibraryConfig{}).Where("module = ?", "display").Update("config", `{"videoPreviewGeneration":"playback","videoPreviewInterval":60}`)
	queued, err = (VideoMoments{}).QueueOnPlayback("d-shared")
	if err != nil || !queued {
		t.Fatalf("auto shared config did not queue: %v, %v", queued, err)
	}
	var preview models.VideoPreview
	db.First(&preview, "drama_series_id = ?", "d-shared")
	if preview.Interval != 60 {
		t.Fatalf("queue ignored shared interval: %+v", preview)
	}
}
func TestMomentCRUDAndPreviewPreservedAfterSourceReplacement(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	os.WriteFile(src, []byte("source"), 0600)
	db.Create(&models.Resources{ID: "r", Title: "测试视频", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src, DurationSeconds: 120})
	service := VideoMoments{}
	item, err := service.Save("", models.VideoMoment{DramaSeriesID: "d", Title: "片段", Start: 10, End: 20})
	if err != nil {
		t.Fatal(err)
	}
	item.Note = "修改备注"
	item.End = 0
	if _, err = service.Save(item.ID, *item); err != nil {
		t.Fatal(err)
	}
	var saved models.VideoMoment
	db.First(&saved, "id = ?", item.ID)
	if saved.End != 0 || saved.Note != "修改备注" {
		t.Fatal(saved)
	}
	if _, err = service.Queue("d", 30); err != nil {
		t.Fatal(err)
	}
	var preview models.VideoPreview
	db.First(&preview, "drama_series_id = ?", "d")
	preview.Status = "ready"
	preview.Completed = 1
	preview.Count = 1
	db.Save(&preview)
	os.MkdirAll(previewDir(preview), 0755)
	image := filepath.Join(previewDir(preview), "000000.jpg")
	os.WriteFile(image, []byte("old preview"), 0600)
	os.WriteFile(src, []byte("replacement video different size"), 0600)
	got, err := service.Queue("d", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceKey != preview.SourceKey || got.Status != "ready" || got.Interval != preview.Interval {
		t.Fatalf("source replacement changed existing preview: %+v", got)
	}
	list, err := service.List("r", "d", 1)
	if err != nil || len(list) != 1 || !list[0].Available || list[0].SourceChanged {
		t.Fatalf("bookmark disabled after replacement: %+v %v", list, err)
	}
	os.Remove(src)
	if path, err := service.Image("d", 0); err != nil || path != image {
		t.Fatalf("existing image unavailable after source removal: %s %v", path, err)
	}
	list, _ = service.List("r", "", 1)
	if len(list) != 1 || list[0].Available {
		t.Fatal("missing source should retain notes and report unavailable")
	}
	if err = service.DeletePreview("d"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(image); !os.IsNotExist(err) {
		t.Fatal("manual delete did not remove preview")
	}
	if err = service.Delete(item.ID); err != nil {
		t.Fatal(err)
	}
	list, _ = service.List("", "", 1)
	if len(list) != 0 {
		t.Fatal("delete failed")
	}
}
func TestPreviewCapacityEvictsOldestAndPreservesMoments(t *testing.T) {
	db := momentTestDB(t)
	now := time.Now()
	for i, id := range []string{"old", "new"} {
		v := models.VideoPreview{DramaSeriesID: id, SourceKey: id, Status: "ready", Width: 240, Interval: 30, UpdatedAt: now.Add(time.Duration(i) * time.Hour)}
		if err := db.Create(&v).Error; err != nil {
			t.Fatal(err)
		}
		os.MkdirAll(previewDir(v), 0755)
		f, err := os.Create(filepath.Join(previewDir(v), "000000.jpg"))
		if err != nil {
			t.Fatal(err)
		}
		if err = f.Truncate(40 * 1024 * 1024); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	db.Create(&models.VideoMoment{ID: "bookmark", DramaSeriesID: "old", Note: "保留收藏"})
	if err := (VideoMoments{}).SaveSetting(models.VideoPreviewSetting{MaxMB: 64, Width: 320}); err != nil {
		t.Fatal(err)
	}
	var previews []models.VideoPreview
	db.Find(&previews)
	if len(previews) != 1 || previews[0].DramaSeriesID != "new" {
		t.Fatalf("wrong eviction: %+v", previews)
	}
	var count int64
	db.Model(&models.VideoMoment{}).Count(&count)
	if count != 1 {
		t.Fatal("capacity eviction deleted bookmark")
	}
	if cacheBytes() > 64*1024*1024 {
		t.Fatal("cache remains above limit")
	}
	if err := (VideoMoments{}).SaveSetting(models.VideoPreviewSetting{MaxMB: 0, Width: 480}); err != nil {
		t.Fatal(err)
	}
	if err := (VideoMoments{}).SaveSetting(models.VideoPreviewSetting{MaxMB: 1, Width: 480}); err == nil {
		t.Fatal("accepted invalid budget")
	}
}
func TestPreviewFFmpegGeneratesConfiguredDimensions(t *testing.T) {
	// Discover bundled tools before switching into the isolated cache directory.
	toolsDir, err := filepath.Abs("../ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", toolsDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	db := momentTestDB(t)
	tool, err := (ff.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		t.Skip("ffmpeg unavailable:", err)
	}
	if _, err = (ff.FFmpeg{}).IsFFprobeAvailable(); err != nil {
		t.Skip("ffprobe unavailable:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	src := filepath.Join(t.TempDir(), "sample.mp4")
	cmd := ff.CreateCommandContext(ctx, tool, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "color=c=blue:s=320x180:d=2", "-c:v", "mpeg4", "-y", src)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, output)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src})
	_, _, key, err := momentSource("d")
	if err != nil {
		t.Fatal(err)
	}
	v := models.VideoPreview{DramaSeriesID: "d", SourceKey: key, Interval: 10, Width: 320, Status: "processing"}
	// An interrupted attempt may have left an empty frame; retry must replace it.
	if err := os.MkdirAll(previewDir(v), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(previewDir(v), "000000.jpg"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&v)
	previewQueue.paused = false
	if err = generatePreview(&v); err != nil {
		t.Fatal(err)
	}
	if v.Completed != 1 || v.Duration < 1.9 {
		t.Fatalf("wrong preview: %+v", v)
	}
	f, err := os.Open(filepath.Join(previewDir(v), "000000.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	image, err := jpeg.DecodeConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	if image.Width != 320 || image.Height != 180 {
		t.Fatalf("wrong dimensions %+v", image)
	}
	f.Close()
	// Simulate deleting a previously completed frame while a task is paused.
	framePath := filepath.Join(previewDir(v), "000000.jpg")
	if err := os.Remove(framePath); err != nil {
		t.Fatal(err)
	}
	if err := generatePreview(&v); err != nil {
		t.Fatal(err)
	}
	if !previewFileExists(framePath) || v.Completed != v.Count {
		t.Fatal("resume skipped a lost completed frame")
	}
	videoMomentCapture = captureVideoMomentCover
	moment, err := (VideoMoments{}).Save("", models.VideoMoment{DramaSeriesID: "d", Start: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	cover, err := (VideoMoments{}).Cover(moment.ID)
	if err != nil {
		t.Fatal(err)
	}
	coverFile, err := os.Open(cover)
	if err != nil {
		t.Fatal(err)
	}
	defer coverFile.Close()
	coverImage, err := jpeg.DecodeConfig(coverFile)
	if err != nil || coverImage.Width != 240 || coverImage.Height != 134 {
		t.Fatalf("invalid persistent cover: %+v %v", coverImage, err)
	}
}
func TestPreviewImageChecksOnlyRequestedFrame(t *testing.T) {
	db := momentTestDB(t)
	v := models.VideoPreview{DramaSeriesID: "partial", SourceKey: "partial", Status: "ready", Interval: 30, Width: 240, Count: 3, Completed: 3}
	if err := db.Create(&v).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(previewDir(v), 0755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(previewDir(v), "000000.jpg")
	if err := os.WriteFile(file, []byte("valid frame"), 0600); err != nil {
		t.Fatal(err)
	}
	service := VideoMoments{}
	status, err := service.Preview(v.DramaSeriesID)
	if err != nil || status.Status != "missing" {
		t.Fatalf("missing cache not detected: %+v %v", status, err)
	}
	path, err := service.Image(v.DramaSeriesID, 0)
	if err != nil || path != file {
		t.Fatalf("intact frame unavailable because of other missing frames: %s %v", path, err)
	}
	if _, err := service.Image(v.DramaSeriesID, 1); err == nil {
		t.Fatal("missing frame was served")
	}
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Image(v.DramaSeriesID, 0); err == nil {
		t.Fatal("empty frame was served")
	}
}

func TestMomentUsesCurrentPreciseDurationAndDoesNotRebuildOnEdit(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("new video"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src, DurationSeconds: 20})
	db.Create(&models.VideoPreview{DramaSeriesID: "d", SourceKey: "old", Status: "ready", Duration: 20})
	service := VideoMoments{}
	item, err := service.Save("", models.VideoMoment{DramaSeriesID: "d", Start: 120.5, End: 120.75})
	if err != nil {
		t.Fatal("stale or rounded duration rejected valid bookmark:", err)
	}
	if _, err = service.Save("", models.VideoMoment{DramaSeriesID: "d", Start: 121}); err == nil {
		t.Fatal("accepted time beyond actual duration")
	}
	if err = service.DeletePreview("d"); err != nil {
		t.Fatal(err)
	}
	item.Note = "edited"
	if _, err = service.Save(item.ID, *item); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.VideoPreview{}).Count(&count)
	if count != 0 {
		t.Fatal("editing recreated manually deleted preview")
	}
	for _, probeError := range []error{nil, errors.New("unreadable replacement")} {
		videoMomentDuration = func(string) (float64, error) { return 5, probeError }
		item.Note = "replacement note"
		saved, editErr := service.Save(item.ID, *item)
		if editErr != nil || saved.Note != item.Note || saved.SourceKey != item.SourceKey || saved.Start != item.Start || !saved.Available {
			t.Fatalf("text edit after replacement failed or altered snapshot: %+v %v", saved, editErr)
		}
		changed := *item
		changed.End += 1
		if _, editErr = service.Save(item.ID, changed); editErr == nil {
			t.Fatal("time change after shorter/unreadable replacement should fail")
		}
	}
	if err = os.Remove(src); err != nil {
		t.Fatal(err)
	}
	item.Note = "offline note"
	saved, err := service.Save(item.ID, *item)
	if err != nil || saved.Note != "offline note" {
		t.Fatalf("offline note edit failed: %+v %v", saved, err)
	}
	item.Start = 3
	if _, err = service.Save(item.ID, *item); err == nil {
		t.Fatal("offline time change should fail")
	}
}
func TestPreviewQueueRecoversInterruptedTaskAndRetriesInitialization(t *testing.T) {
	db := momentTestDB(t)
	db.Create(&models.VideoPreview{DramaSeriesID: "d", SourceKey: "old", Status: "processing"})
	original := videoMomentsDB
	broken := db.Session(&gorm.Session{NewDB: true})
	broken.Error = errors.New("temporary database error")
	videoMomentsDB = func() *gorm.DB { return broken }
	if _, err := (VideoMoments{}).QueueStatus(); err == nil {
		t.Fatal("initialization failure hidden")
	}
	videoMomentsDB = original
	status, err := (VideoMoments{}).QueueStatus()
	if err != nil || status["queued"] != int64(1) {
		t.Fatalf("queue recovery failed: %v %v", status, err)
	}
	var v models.VideoPreview
	db.First(&v)
	if v.Status != "queued" {
		t.Fatal("interrupted processing not recovered")
	}
	if err := (VideoMoments{}).DeletePreview("d"); err != nil {
		t.Fatal("paused recovered task cannot be deleted:", err)
	}
}

func TestPreviewConcurrentQueueAndDeletionGuards(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src})
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := (VideoMoments{}).Queue("d", 30); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	db.Model(&models.VideoPreview{}).Count(&count)
	if count != 1 {
		t.Fatalf("duplicate tasks: %d", count)
	}
	db.Model(&models.VideoPreview{}).Where("drama_series_id = ?", "d").Update("status", "processing")
	if err := (VideoMoments{}).DeletePreview("d"); err == nil {
		t.Fatal("deleted processing task")
	}
	previewQueue.Lock()
	previewQueue.running = true
	previewQueue.Unlock()
	err := (VideoMoments{}).QueueControl("clear")
	previewQueue.Lock()
	previewQueue.running = false
	previewQueue.Unlock()
	if err == nil {
		t.Fatal("cleared cache while worker active")
	}
	db.Model(&models.VideoPreview{}).Where("drama_series_id = ?", "d").Update("status", "queued")
	if err := (VideoMoments{}).DeletePreview("d"); err != nil {
		t.Fatal("paused queued deletion failed:", err)
	}
}

func TestPlaybackPreviewGenerationFollowsActualLibrary(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.FilesBasesSetting{FilesBasesID: "manual", ConfigJsonData: `{"videoPreviewGeneration":"manual"}`})
	db.Create(&models.FilesBasesSetting{FilesBasesID: "auto", ConfigJsonData: `{"videoPreviewGeneration":"playback","videoPreviewInterval":120}`})
	db.Create(&models.FilesBasesSetting{FilesBasesID: "old", ConfigJsonData: `{"sampleStatus":true}`})
	service := VideoMoments{}
	for _, library := range []string{"manual", "auto", "old", "unset"} {
		db.Create(&models.Resources{ID: library, FilesBasesID: library, Mode: datatype.E_resourceMode_Movies, Status: true})
		db.Create(&models.ResourcesDramaSeries{ID: library, ResourcesID: library, Src: src})
		initial, err := service.Preview(library)
		expectedInterval := 30
		if library == "auto" {
			expectedInterval = 120
		}
		if err != nil || initial.Interval != expectedInterval {
			t.Fatalf("wrong default interval: %+v %v", initial, err)
		}
		queued, err := service.QueueOnPlayback(library)
		if err != nil {
			t.Fatal(err)
		}
		if queued {
			v, err := service.Preview(library)
			if err != nil || v.Interval != expectedInterval {
				t.Fatalf("auto generation ignored default interval: %+v %v", v, err)
			}
		}
		if queued != (library != "manual") {
			t.Fatalf("wrong generation mode for %s: %v", library, queued)
		}
		queued, err = service.QueueOnPlayback(library)
		if err != nil || queued {
			t.Fatalf("repeated playback should not queue: %s %v %v", library, queued, err)
		}
	}
	if _, err := service.Save("", models.VideoMoment{DramaSeriesID: "manual", Start: 1}); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.VideoPreview{}).Where("drama_series_id = ?", "manual").Count(&count)
	if count != 0 {
		t.Fatal("saving bookmark bypassed manual generation setting")
	}
	if _, err := service.Queue("manual", 10); err != nil {
		t.Fatal("manual generation unavailable:", err)
	}
	if err := os.MkdirAll(filepath.Join(previewRoot, "old-source", "120-240"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(previewRoot, "old-source", "120-240", "000000.jpg"), []byte("old frame"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"ready", "failed", "processing", "queued"} {
		db.Model(&models.VideoPreview{}).Where("drama_series_id = ?", "auto").Updates(map[string]interface{}{"status": status, "source_key": "old-source", "completed": 1, "count": 1})
		queued, err := service.QueueOnPlayback("auto")
		if err != nil || queued {
			t.Fatalf("existing %s preview was requeued: %v", status, err)
		}
		var preview models.VideoPreview
		db.First(&preview, "drama_series_id = ?", "auto")
		if preview.SourceKey != "old-source" || preview.Status != status {
			t.Fatal("existing preview changed on playback")
		}
	}
}

func TestPreviewIntervalOptions(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	for _, interval := range []int{10, 30, 60, 120} {
		id := fmt.Sprintf("d%d", interval)
		db.Create(&models.ResourcesDramaSeries{ID: id, ResourcesID: "r", Src: src})
		v, err := (VideoMoments{}).Queue(id, interval)
		if err != nil || v.Interval != interval {
			t.Fatalf("interval %d failed: %+v %v", interval, v, err)
		}
	}
	if _, err := (VideoMoments{}).Queue("d10", 20); err == nil {
		t.Fatal("invalid interval accepted")
	}
}

func TestOpenPreviewCacheFolderUsesFixedPathAndPermission(t *testing.T) {
	momentTestDB(t)
	previousOpen, previousConfig := openVideoPreviewFolder, core.Config
	core.Config = &config.Config{}
	t.Cleanup(func() {
		openVideoPreviewFolder = previousOpen
		core.Config = previousConfig
	})
	var opened string
	openVideoPreviewFolder = func(path string) error { opened = path; return nil }
	core.Config.General.NotAllowServerOpenFile = true
	if err := (VideoMoments{}).OpenCacheFolder(); err == nil {
		t.Fatal("open ignored server permission")
	}
	if opened != "" {
		t.Fatal("folder opened despite denied permission")
	}
	core.Config.General.NotAllowServerOpenFile = false
	if err := (VideoMoments{}).OpenCacheFolder(); err != nil {
		t.Fatal(err)
	}
	expected, _ := filepath.Abs(previewRoot)
	if opened != expected {
		t.Fatalf("wrong folder: %s", opened)
	}
	info, err := os.Stat(expected)
	if err != nil || !info.IsDir() {
		t.Fatal("empty cache folder not created:", err)
	}
	openVideoPreviewFolder = func(string) error { return errors.New("open failed") }
	if err := (VideoMoments{}).OpenCacheFolder(); err == nil {
		t.Fatal("open error hidden")
	}
}

func TestBookmarkCoverSurvivesCacheDeletionAndSourceReplacement(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src})
	service := VideoMoments{}
	moment, err := service.Save("", models.VideoMoment{DramaSeriesID: "d", Start: 12, End: 20})
	if err != nil {
		t.Fatal(err)
	}
	cover, err := service.Cover(moment.ID)
	if err != nil || !strings.HasPrefix(cover, filepath.Clean(momentCoverRoot)) {
		t.Fatalf("cover not in database folder: %s %v", cover, err)
	}
	if _, err := service.Queue("d", 30); err != nil {
		t.Fatal(err)
	}
	if err := service.QueueControl("clear"); err != nil {
		t.Fatal(err)
	}
	if !previewFileExists(cover) {
		t.Fatal("cache clear removed persistent cover")
	}
	if err := os.WriteFile(src, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	videoMomentCapture = func(string, float64, int, string) error {
		t.Fatal("replacement regenerated existing cover")
		return nil
	}
	if _, err := service.Cover(moment.ID); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Cover(moment.ID); err != nil {
		t.Fatal("saved cover requires original video:", err)
	}
	if err := service.Delete(moment.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cover); !os.IsNotExist(err) {
		t.Fatal("bookmark deletion left its cover")
	}
}
func TestDeletedTimelineCacheCanRegenerate(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src})
	service := VideoMoments{}
	preview, err := service.Queue("d", 30)
	if err != nil {
		t.Fatal(err)
	}
	preview.Status, preview.Completed, preview.Count, preview.Duration = "ready", 2, 2, 60
	db.Save(&preview)
	if err := os.MkdirAll(previewDir(preview), 0755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := os.WriteFile(filepath.Join(previewDir(preview), fmt.Sprintf("%06d.jpg", i)), []byte("frame"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if v, err := service.Preview("d"); err != nil || v.Status != "ready" {
		t.Fatalf("valid cache changed: %+v %v", v, err)
	}
	if err := os.Remove(filepath.Join(previewDir(preview), "000001.jpg")); err != nil {
		t.Fatal(err)
	}
	v, err := service.Preview("d")
	if err != nil || v.Status != "missing" || v.Completed != 0 {
		t.Fatalf("missing frame not detected: %+v %v", v, err)
	}
	v, err = service.Queue("d", 120)
	if err != nil || v.Status != "queued" || v.Completed != 0 || v.Count != 0 || v.Interval != 120 {
		t.Fatalf("cache cannot regenerate: %+v %v", v, err)
	}
	// A whole directory deleted outside the app is also repaired in automatic mode.
	v.Status, v.Completed, v.Count = "ready", 1, 1
	db.Save(&v)
	if err := os.RemoveAll(previewRoot); err != nil {
		t.Fatal(err)
	}
	queued, err := service.QueueOnPlayback("d")
	if err != nil || !queued {
		t.Fatalf("automatic mode did not rebuild missing cache: %v %v", queued, err)
	}
}
func TestExistingBookmarkLazilyCreatesIndependentCover(t *testing.T) {
	db := momentTestDB(t)
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true})
	db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: src})
	item := models.VideoMoment{ID: "existing", DramaSeriesID: "d", Start: 4}
	db.Create(&item)
	file, err := (VideoMoments{}).Cover(item.ID)
	if err != nil || !previewFileExists(file) {
		t.Fatalf("existing bookmark cover not created: %s %v", file, err)
	}
}
