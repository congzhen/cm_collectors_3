package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newDramaSeriesSyncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Resources{},
		&models.ResourcesDramaSeries{},
		&models.ResourcesVideoMetadata{},
		&models.VideoFingerprint{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	resource := models.Resources{
		ID:           "resource-1",
		FilesBasesID: "library-1",
		Mode:         datatype.E_resourceMode_Movies,
		Title:        "test",
	}
	if err := db.Create(&resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	return db
}

func TestSetResourcesDramaSeriesReorderPreservesRelatedData(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	old := []models.ResourcesDramaSeries{
		{ID: "ds-1", ResourcesID: "resource-1", Src: "one.mp4", Sort: 0, DurationSeconds: 10},
		{ID: "ds-2", ResourcesID: "resource-1", Src: "two.mp4", Sort: 1, DurationSeconds: 20},
	}
	if err := db.Create(&old).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := db.Create(&models.ResourcesVideoMetadata{
		DramaSeriesID: "ds-1", ProbeStatus: models.VideoMetadataStatusSuccess,
		MetadataVersion: CurrentVideoMetadataVersion, Width: 1920, Height: 1080,
	}).Error; err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if err := db.Create(&models.VideoFingerprint{
		ID: "vf-1", DramaSeriesID: "ds-1", ResourcesID: "resource-1",
		FilesBasesID: "library-1", Src: "one.mp4", Status: models.VideoFingerprintStatus_Done,
	}).Error; err != nil {
		t.Fatalf("create fingerprint: %v", err)
	}

	err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1",
		[]datatype.ReqParam_resourceDramaSeries_Base{
			{ID: "ds-2", Src: "two.mp4"},
			{ID: "ds-1", Src: "one.mp4"},
		})
	if err != nil {
		t.Fatalf("reorder drama series: %v", err)
	}

	list, err := (models.ResourcesDramaSeries{}).ListByResourceID(db, "resource-1")
	if err != nil || len(*list) != 2 {
		t.Fatalf("load reordered drama series: %#v, %v", list, err)
	}
	if (*list)[0].ID != "ds-2" || (*list)[1].ID != "ds-1" {
		t.Fatalf("unexpected reorder result: %#v", *list)
	}
	metadata, err := (models.ResourcesVideoMetadata{}).Get(db, "ds-1")
	if err != nil || metadata.ProbeStatus != models.VideoMetadataStatusSuccess || metadata.Width != 1920 {
		t.Fatalf("metadata should be preserved: %#v, %v", metadata, err)
	}
	fingerprint, err := (models.VideoFingerprint{}).GetByDramaSeriesID(db, "ds-1")
	if err != nil || fingerprint.Status != models.VideoFingerprintStatus_Done {
		t.Fatalf("fingerprint should be preserved: %#v, %v", fingerprint, err)
	}
}

func TestSetResourcesDramaSeriesPathChangeMarksStale(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.Create(&models.ResourcesDramaSeries{
		ID: "ds-1", ResourcesID: "resource-1", Src: "old.mp4", DurationSeconds: 10,
	}).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := db.Create(&models.ResourcesVideoMetadata{
		DramaSeriesID: "ds-1", ProbeStatus: models.VideoMetadataStatusSuccess,
		MetadataVersion: CurrentVideoMetadataVersion, Width: 1920, Height: 1080,
	}).Error; err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if err := db.Create(&models.VideoFingerprint{
		ID: "vf-old", DramaSeriesID: "ds-1", ResourcesID: "resource-1",
		FilesBasesID: "library-1", Src: "old.mp4", Status: models.VideoFingerprintStatus_Done,
	}).Error; err != nil {
		t.Fatalf("create fingerprint: %v", err)
	}

	err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1",
		[]datatype.ReqParam_resourceDramaSeries_Base{{ID: "ds-1", Src: "new.mp4"}})
	if err != nil {
		t.Fatalf("change path: %v", err)
	}
	ds, err := (models.ResourcesDramaSeries{}).Info(db, "ds-1")
	if err != nil || ds.Src != "new.mp4" || ds.DurationSeconds != 10 {
		t.Fatalf("identity and last duration should be preserved: %#v, %v", ds, err)
	}
	metadata, err := (models.ResourcesVideoMetadata{}).Get(db, "ds-1")
	if err != nil || metadata.ProbeStatus != models.VideoMetadataStatusStale ||
		metadata.MetadataVersion != 0 || metadata.Width != 1920 {
		t.Fatalf("last metadata should be retained but stale: %#v, %v", metadata, err)
	}
	fingerprint, err := (models.VideoFingerprint{}).GetByDramaSeriesID(db, "ds-1")
	if err != nil || fingerprint.Src != "new.mp4" ||
		fingerprint.Status != models.VideoFingerprintStatus_Pending {
		t.Fatalf("fingerprint should be reset for the new path: %#v, %v", fingerprint, err)
	}
}

func TestSetResourcesDramaSeriesSyncsFingerprintForModeAndLibraryChanges(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.Create(&models.ResourcesDramaSeries{
		ID: "ds-1", ResourcesID: "resource-1", Src: "one.mp4",
	}).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := db.Create(&models.VideoFingerprint{
		ID: "vf-1", DramaSeriesID: "ds-1", ResourcesID: "resource-1",
		FilesBasesID: "library-1", Src: "one.mp4", Status: models.VideoFingerprintStatus_Done,
	}).Error; err != nil {
		t.Fatalf("create fingerprint: %v", err)
	}
	submitted := []datatype.ReqParam_resourceDramaSeries_Base{{ID: "ds-1", Src: "one.mp4"}}

	if err := db.Model(&models.Resources{}).Where("id = ?", "resource-1").
		Update("mode", datatype.E_resourceMode_Comic).Error; err != nil {
		t.Fatalf("change resource mode: %v", err)
	}
	if err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1", submitted); err != nil {
		t.Fatalf("sync non-video resource: %v", err)
	}
	var count int64
	if err := db.Model(&models.VideoFingerprint{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("non-video resource should not retain fingerprints, count=%d, err=%v", count, err)
	}

	if err := db.Model(&models.Resources{}).Where("id = ?", "resource-1").
		Updates(map[string]interface{}{
			"mode":          datatype.E_resourceMode_Movies,
			"filesBases_id": "library-2",
		}).Error; err != nil {
		t.Fatalf("restore video resource: %v", err)
	}
	if err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1", submitted); err != nil {
		t.Fatalf("sync restored video resource: %v", err)
	}
	fingerprint, err := (models.VideoFingerprint{}).GetByDramaSeriesID(db, "ds-1")
	if err != nil || fingerprint.FilesBasesID != "library-2" ||
		fingerprint.Status != models.VideoFingerprintStatus_Pending {
		t.Fatalf("video resource should receive a pending fingerprint: %#v, %v", fingerprint, err)
	}

	// 已存在的完成指纹在资源仅移动文件库时应保留计算结果，只修正文件库归属。
	fingerprint.Status = models.VideoFingerprintStatus_Done
	if err := db.Model(&models.VideoFingerprint{}).Where("id = ?", fingerprint.ID).
		Update("status", fingerprint.Status).Error; err != nil {
		t.Fatalf("mark fingerprint done: %v", err)
	}
	if err := db.Model(&models.Resources{}).Where("id = ?", "resource-1").
		Update("filesBases_id", "library-3").Error; err != nil {
		t.Fatalf("move resource library: %v", err)
	}
	if err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1", submitted); err != nil {
		t.Fatalf("sync moved resource: %v", err)
	}
	fingerprint, err = (models.VideoFingerprint{}).GetByDramaSeriesID(db, "ds-1")
	if err != nil || fingerprint.FilesBasesID != "library-3" ||
		fingerprint.Status != models.VideoFingerprintStatus_Done {
		t.Fatalf("library move should preserve computed fingerprint: %#v, %v", fingerprint, err)
	}
}

func TestSetResourcesDramaSeriesEmptyListClearsExisting(t *testing.T) {
	db := newDramaSeriesSyncTestDB(t)
	if err := db.Create(&models.ResourcesDramaSeries{
		ID: "ds-1", ResourcesID: "resource-1", Src: "old.mp4",
	}).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := (ResourcesDramaSeries{}).SetResourcesDramaSeries(db, "resource-1", nil); err != nil {
		t.Fatalf("clear drama series: %v", err)
	}
	var count int64
	if err := db.Model(&models.ResourcesDramaSeries{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("expected all drama series deleted, count=%d, err=%v", count, err)
	}
}
