package models

import (
	"cm_collectors_server/datatype"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openVideoTranscodeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&VideoTranscodeTask{}); err != nil {
		t.Fatalf("migrate video transcode task: %v", err)
	}
	return db
}

func TestVideoTranscodeMigrationContainsMediaSnapshotColumns(t *testing.T) {
	db := openVideoTranscodeTestDB(t)
	for _, column := range []string{
		"source_duration",
		"source_video_codec",
		"output_duration",
		"output_video_codec",
		"processed_seconds",
	} {
		if !db.Migrator().HasColumn(&VideoTranscodeTask{}, column) {
			t.Fatalf("migration did not create column %s", column)
		}
	}
}

func TestFirstQueuedMatchesVisibleTopToBottomOrder(t *testing.T) {
	db := openVideoTranscodeTestDB(t)
	older := datatype.CustomTime(time.Date(2026, 7, 30, 10, 0, 0, 0, time.Local))
	newer := datatype.CustomTime(time.Date(2026, 7, 30, 11, 0, 0, 0, time.Local))
	tasks := []VideoTranscodeTask{
		{ID: "task-older", Status: VideoTranscodeStatusQueued, CreatedAt: &older},
		{ID: "task-newer", Status: VideoTranscodeStatusQueued, CreatedAt: &newer},
	}
	if err := db.Create(&tasks).Error; err != nil {
		t.Fatalf("create queue: %v", err)
	}
	list, err := (VideoTranscodeTask{}).List(db)
	if err != nil {
		t.Fatalf("list queue: %v", err)
	}
	next, err := (VideoTranscodeTask{}).FirstQueued(db)
	if err != nil {
		t.Fatalf("read next task: %v", err)
	}
	if len(list) == 0 || next.ID != list[0].ID || next.ID != "task-newer" {
		t.Fatalf("next task %q does not match first visible row %q", next.ID, list[0].ID)
	}
}
