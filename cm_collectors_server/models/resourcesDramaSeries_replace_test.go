package models

import (
	"cm_collectors_server/datatype"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestReplacePathInvalidatesMetadataAndRebuildsFingerprint(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&Resources{},
		&ResourcesDramaSeries{},
		&ResourcesVideoMetadata{},
		&VideoFingerprint{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	resource := Resources{
		ID: "resource-1", FilesBasesID: "library-1",
		Mode: datatype.E_resourceMode_Movies, Title: "test",
	}
	if err := db.Create(&resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if err := db.Create(&ResourcesDramaSeries{
		ID: "ds-1", ResourcesID: resource.ID, Src: `D:\old\one.mp4`,
	}).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := db.Create(&ResourcesVideoMetadata{
		DramaSeriesID: "ds-1", ProbeStatus: VideoMetadataStatusFailed,
		MetadataVersion: 1, RetryCount: 3, ErrorCode: "io_error", ErrorMessage: "old",
	}).Error; err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if err := db.Create(&VideoFingerprint{
		ID: "vf-old", DramaSeriesID: "ds-1", ResourcesID: resource.ID,
		FilesBasesID: resource.FilesBasesID, Src: `D:\old\one.mp4`,
		Status: VideoFingerprintStatus_Done,
	}).Error; err != nil {
		t.Fatalf("create fingerprint: %v", err)
	}

	replaced, err := (ResourcesDramaSeries{}).ReplacePath(
		db, []string{resource.FilesBasesID}, `D:\old`, `E:\new`, func() string { return "vf-new" },
	)
	if err != nil {
		t.Fatalf("replace path: %v", err)
	}
	if len(*replaced) != 1 || (*replaced)[0].Src != `E:\new\one.mp4` {
		t.Fatalf("unexpected replace result: %#v", replaced)
	}
	ds, err := (ResourcesDramaSeries{}).Info(db, "ds-1")
	if err != nil || ds.Src != `E:\new\one.mp4` {
		t.Fatalf("drama series path not updated: %#v, %v", ds, err)
	}
	metadata, err := (ResourcesVideoMetadata{}).Get(db, "ds-1")
	if err != nil || metadata.ProbeStatus != VideoMetadataStatusStale ||
		metadata.MetadataVersion != 0 || metadata.RetryCount != 0 ||
		metadata.ErrorCode != "" || metadata.ErrorMessage != "" {
		t.Fatalf("metadata should be stale with retry state cleared: %#v, %v", metadata, err)
	}
	fingerprint, err := (VideoFingerprint{}).GetByDramaSeriesID(db, "ds-1")
	if err != nil || fingerprint.ID != "vf-new" ||
		fingerprint.Src != `E:\new\one.mp4` ||
		fingerprint.Status != VideoFingerprintStatus_Pending {
		t.Fatalf("fingerprint should be rebuilt for new path: %#v, %v", fingerprint, err)
	}
}
