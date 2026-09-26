package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestMetadataRulesFilterUnknownCorruptAndOverrides(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.AutoMigrate(&models.FilesBases{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.FilesBases{ID: "library-1", Name: "测试库"}).Error; err != nil {
		t.Fatal(err)
	}
	setting := (models.VideoMetadataSetting{}).Default()
	setting.ExcludedExtensions = "PDF, .zip， Zip"
	normalized, err := models.NormalizeMetadataExtensions(setting.ExcludedExtensions)
	if err != nil {
		t.Fatal(err)
	}
	if normalized != "pdf,zip" {
		t.Fatalf("normalization: %s", normalized)
	}
	setting.ExcludedExtensions = normalized
	if err := db.Create(&setting).Error; err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		id, src, classification, status, reason string
		excluded                                bool
	}{
		{"image", "cover.JPG", "", "failed", "system", false},
		{"url", "https://host/poster.png?token=abc#x", "", "failed", "system", false},
		{"html", "page.html", "", "failed", "system", false},
		{"pdf", "notes.PDF", "", "failed", "custom", false},
		{"broken", "broken.mp4", "", "failed", "", false},
		{"missing", "missing.mkv", "", "failed", "", false},
		{"partial", "video.mkv.td", "", "failed", "", false},
		{"unknown", "video.unknown", "", "failed", "", false},
		{"none", "video", "", "failed", "", false},
		{"hash", "folder#name/video.mp4", "", "failed", "", false},
		{"force", "special.jpg", "video", "failed", "", false},
		{"manual-data", "manual.jpg", "", "manual", "", false},
		{"manual-exclude", "movie.mp4", "nonvideo", "failed", "manual", true},
		{"legacy", "old.mp4", "", "failed", "legacy", true},
	}
	for _, c := range cases {
		ds := models.ResourcesDramaSeries{ID: c.id, ResourcesID: "resource-1", Src: c.src, VideoMetadataClassification: c.classification, VideoMetadataExcluded: c.excluded}
		if err := db.Create(&ds).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: c.id, ProbeStatus: c.status}).Error; err != nil {
			t.Fatal(err)
		}
	}
	reason, err := models.MetadataExclusionSQL(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		var got string
		err := db.Table("resourcesDramaSeries ds").Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id=ds.id").Where("ds.id = ?", c.id).Select(reason).Scan(&got).Error
		if err != nil {
			t.Fatal(err)
		}
		if got != c.reason {
			t.Errorf("%s got %q want %q", c.id, got, c.reason)
		}
	}
	list, err := findVideoMetadataFailures(db, VideoMetadataFailureQuery{Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 7 {
		t.Fatalf("want 7 real/unknown failures, got %d", list.Total)
	}
	excluded, err := findVideoMetadataFailures(db, VideoMetadataFailureQuery{Excluded: true, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if excluded.Total != 6 || len(excluded.DataList) != 2 {
		t.Fatalf("exclusion pagination: %+v", excluded)
	}
	stats, err := findVideoMetadataStats(db)
	if err != nil || len(stats) != 1 {
		t.Fatalf("stats: %+v %v", stats, err)
	}
	if stats[0].Failed != 7 || stats[0].Excluded != 6 || stats[0].Total != 8 || stats[0].Manual != 1 {
		t.Fatalf("stats disagree with lists: %+v", stats[0])
	}
	// 删除自动规则后恢复资格，但人工、旧来源未知排除仍然保留。
	disabled := false
	setting.AutoExcludeNonVideo = &disabled
	setting.ExcludedExtensions = ""
	if err := (models.VideoMetadataSetting{}).Save(db, &setting); err != nil {
		t.Fatal(err)
	}
	list, err = findVideoMetadataFailures(db, VideoMetadataFailureQuery{Excluded: true, Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 {
		t.Fatalf("manual and legacy must remain: %d", list.Total)
	}
}

func TestMetadataCleanupBatchesAndRuleRemoval(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	setting := (models.VideoMetadataSetting{}).Default()
	db.Create(&setting)
	for i := 0; i < 405; i++ {
		id := fmt.Sprintf("img-%04d", i)
		if err := db.Create(&models.ResourcesDramaSeries{ID: id, ResourcesID: "resource-1", Src: id + ".jpg", DurationProbeStatus: "failed"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	db.Create(&models.ResourcesDramaSeries{ID: "manual", ResourcesID: "resource-1", Src: "manual.jpg", DurationSeconds: 20})
	db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: "manual", ProbeStatus: "manual"})
	progress := []int{}
	n, err := cleanMetadataRules(db, models.VideoMetadataScopeAll, nil, func(n int) error { progress = append(progress, n); return nil })
	if err != nil || n != 405 || len(progress) < 3 {
		t.Fatalf("cleanup %d progress %v err %v", n, progress, err)
	}
	var count int64
	db.Model(&models.ResourcesDramaSeries{}).Count(&count)
	if count != 406 {
		t.Fatal("cleanup deleted resources")
	}
	var manual models.ResourcesVideoMetadata
	db.First(&manual, "drama_series_id = ?", "manual")
	if manual.ProbeStatus != "manual" {
		t.Fatal("manual metadata was removed")
	}
	disabled := false
	setting.AutoExcludeNonVideo = &disabled
	if err := (models.VideoMetadataSetting{}).Save(db, &setting); err != nil {
		t.Fatal(err)
	}
	reason, err := models.MetadataExclusionSQL(db)
	if err != nil {
		t.Fatal(err)
	}
	db.Table("resourcesDramaSeries ds").Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id=ds.id").Where("(" + reason + ") = ''").Count(&count)
	if count != 406 {
		t.Fatalf("rule-based exclusion remained sticky: %d", count)
	}
}

func TestMetadataExtensionValidation(t *testing.T) {
	for _, s := range []string{"*.jpg", "a/b", "jpg' OR 1=1", "."} {
		if _, err := models.NormalizeMetadataExtensions(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
	if got, err := models.NormalizeMetadataExtensions(" .JPG, png；jpg\n PDF "); err != nil || got != "jpg,pdf,png" {
		t.Fatalf("%s %v", got, err)
	}
}

func TestMetadataProbeResultGuard(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	disabled := false
	setting := (models.VideoMetadataSetting{}).Default()
	setting.AutoExcludeNonVideo = &disabled
	if err := db.Create(&setting).Error; err != nil {
		t.Fatal(err)
	}
	ds := models.ResourcesDramaSeries{ID: "guard", ResourcesID: "resource-1", Src: "video.jpg", VideoMetadataExcluded: true, VideoMetadataClassification: "system"}
	db.Create(&ds)
	db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: ds.ID, ProbeStatus: "processing"})
	allowed, err := metadataCanSaveProbe(db, ds)
	if err != nil || !allowed {
		t.Fatalf("eligible probe: %v %v", allowed, err)
	}
	var current models.ResourcesDramaSeries
	db.First(&current, "id = ?", ds.ID)
	if current.VideoMetadataExcluded {
		t.Fatal("automatic exclusion flag remains after rule removal")
	}
	for _, status := range []string{"manual", "stale"} {
		db.Model(&models.ResourcesVideoMetadata{}).Where("drama_series_id = ?", ds.ID).Update("probe_status", status)
		if allowed, err := metadataCanSaveProbe(db, ds); err != nil || allowed {
			t.Fatalf("overwrites %s: %v %v", status, allowed, err)
		}
	}
	db.Model(&models.ResourcesVideoMetadata{}).Where("drama_series_id = ?", ds.ID).Update("probe_status", "processing")
	db.Model(&models.ResourcesDramaSeries{}).Where("id = ?", ds.ID).Update("video_metadata_classification", "nonvideo")
	if allowed, err := metadataCanSaveProbe(db, ds); err != nil || allowed {
		t.Fatalf("overwrites manual exclusion: %v %v", allowed, err)
	}
	db.Model(&models.ResourcesDramaSeries{}).Where("id = ?", ds.ID).Update("video_metadata_classification", "video")
	db.Delete(&models.ResourcesVideoMetadata{}, "drama_series_id = ?", ds.ID)
	if allowed, err := metadataCanSaveProbe(db, ds); err != nil || allowed {
		t.Fatalf("recreates cleaned metadata: %v %v", allowed, err)
	}
}

func TestMetadataCreateAndEditRespectDisabledSystemRules(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	disabled := false
	setting := (models.VideoMetadataSetting{}).Default()
	setting.AutoExcludeNonVideo = &disabled
	setting.ExcludedExtensions = "pdf"
	if err := db.Create(&setting).Error; err != nil {
		t.Fatal(err)
	}
	ds, err := (ResourcesDramaSeries{}).Create(db, "resource-1", "special.jpg", 0)
	if err != nil || ds.VideoMetadataExcluded {
		t.Fatalf("system switch ignored: %+v %v", ds, err)
	}
	if err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1", []datatype.ReqParam_resourceDramaSeries_Base{{ID: ds.ID, Src: "notes.pdf"}}); err != nil {
		t.Fatal(err)
	}
	db.First(ds, "id = ?", ds.ID)
	if !ds.VideoMetadataExcluded {
		t.Fatal("custom rule ignored on path edit")
	}
}

func TestMetadataRuleChangeDuringProbeCanResume(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	ds := models.ResourcesDramaSeries{ID: "changed", ResourcesID: "resource-1", Src: "cover.jpg", VideoMetadataClassification: "system"}
	db.Create(&ds)
	db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: ds.ID, ProbeStatus: "processing"})
	allowed, err := metadataCanSaveProbe(db, ds)
	if allowed || err != nil {
		t.Fatalf("excluded in-flight result allowed: %v %v", allowed, err)
	}
	var metadata models.ResourcesVideoMetadata
	db.First(&metadata, "drama_series_id = ?", ds.ID)
	if metadata.ProbeStatus != "stale" {
		t.Fatalf("stuck processing: %s", metadata.ProbeStatus)
	}
}

func TestMetadataRefreshSkipsRulesAndManualData(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	for _, row := range []struct{ id, src, status string }{{"video", "movie.mp4", "success"}, {"image", "cover.jpg", "success"}, {"manual", "manual.mp4", "manual"}} {
		db.Create(&models.ResourcesDramaSeries{ID: row.id, ResourcesID: "resource-1", Src: row.src})
		db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: row.id, ProbeStatus: row.status})
	}
	if err := markMetadataScopeStale(db, models.VideoMetadataScopeSelected, []string{"library-1"}); err != nil {
		t.Fatal(err)
	}
	for id, expected := range map[string]string{"video": "stale", "image": "success", "manual": "manual"} {
		var metadata models.ResourcesVideoMetadata
		db.First(&metadata, "drama_series_id = ?", id)
		if metadata.ProbeStatus != expected {
			t.Fatalf("%s status %s", id, metadata.ProbeStatus)
		}
	}
}

func TestMetadataNewStatusOverridesLegacyFailure(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.AutoMigrate(&models.FilesBases{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.FilesBases{ID: "library-1", Name: "test"})
	for _, status := range []string{"manual", "success", "processing", "legacy"} {
		if err := db.Create(&models.ResourcesDramaSeries{ID: status, ResourcesID: "resource-1", Src: status + ".mp4", DurationProbeStatus: "failed"}).Error; err != nil {
			t.Fatal(err)
		}
		if status != "legacy" {
			if err := db.Create(&models.ResourcesVideoMetadata{DramaSeriesID: status, ProbeStatus: status, MetadataVersion: CurrentVideoMetadataVersion}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	list, err := findVideoMetadataFailures(db, VideoMetadataFailureQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 1 || list.DataList[0].DramaSeriesID != "legacy" {
		t.Fatalf("old failure overrides newer state: %+v", list)
	}
	stats, err := findVideoMetadataStats(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 || stats[0].Failed != 1 || stats[0].Manual != 1 || stats[0].Completed != 1 || stats[0].Processing != 1 {
		t.Fatalf("inconsistent stats: %+v", stats)
	}
}

func TestMetadataRulesMaximumExtensions(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	exts := []string{}
	for i := 0; i < 100; i++ {
		exts = append(exts, fmt.Sprintf("custom_%d", i))
	}
	normalized, err := models.NormalizeMetadataExtensions(strings.Join(exts, ","))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := models.NormalizeMetadataExtensions(normalized + ",overflow"); err == nil {
		t.Fatal("accepted more than 100 suffixes")
	}
	setting := (models.VideoMetadataSetting{}).Default()
	setting.ExcludedExtensions = normalized
	if err := db.Create(&setting).Error; err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{"file.CUSTOM_99", "file.customX99", "file.custom_99.mp4"} {
		if err := db.Create(&models.ResourcesDramaSeries{ID: src, Src: src, ResourcesID: "resource-1"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	reason, err := models.MetadataExclusionSQL(db)
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Table("resourcesDramaSeries ds").Joins("LEFT JOIN resources_video_metadata vm ON vm.drama_series_id=ds.id").Where("(" + reason + ") <> ''").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("suffix boundary/wildcard match: %d", count)
	}
}

func TestMetadataCleanupStopScopeAndRepeat(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.Create(&models.Resources{ID: "other", FilesBasesID: "library-2", Mode: datatype.E_resourceMode_Movies}).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 206; i++ {
		resource := "resource-1"
		if i == 205 {
			resource = "other"
		}
		if err := db.Create(&models.ResourcesDramaSeries{ID: fmt.Sprintf("row-%03d", i), ResourcesID: resource, Src: "cover.jpg", DurationProbeStatus: "failed"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	stop := errors.New("rules changed")
	n, err := cleanMetadataRules(db, models.VideoMetadataScopeSelected, []string{"library-1"}, func(done int) error {
		if done >= 200 {
			return stop
		}
		return nil
	})
	if n != 200 || !errors.Is(err, stop) {
		t.Fatalf("stop boundary: %d %v", n, err)
	}
	n, err = cleanMetadataRules(db, models.VideoMetadataScopeSelected, []string{"library-1"}, nil)
	if n != 5 || err != nil {
		t.Fatalf("resume: %d %v", n, err)
	}
	n, err = cleanMetadataRules(db, models.VideoMetadataScopeSelected, []string{"library-1"}, nil)
	if n != 0 || err != nil {
		t.Fatalf("repeat: %d %v", n, err)
	}
	var other models.ResourcesDramaSeries
	if err := db.First(&other, "id = ?", "row-205").Error; err != nil {
		t.Fatal(err)
	}
	if other.DurationProbeStatus != "failed" || other.VideoMetadataExcluded {
		t.Fatal("modified unselected library")
	}
}

func TestMetadataBulkRejectsInvalidRequests(t *testing.T) {
	ids := []string{}
	for i := 0; i < 201; i++ {
		ids = append(ids, fmt.Sprint(i))
	}
	for _, request := range []MetadataBulkRequest{{Action: "retry"}, {IDs: ids, Action: "retry"}, {IDs: []string{"one"}, Action: "delete"}} {
		if _, err := (VideoMetadata{}).Bulk(request); err == nil {
			t.Fatalf("invalid bulk accepted: %+v", request)
		}
	}
}
