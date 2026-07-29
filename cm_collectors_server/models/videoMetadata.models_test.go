package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestCronJobMultipleFilesBasesPreload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&FilesBases{}, &CronJobs{}, &CronJobsFilesBases{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	filesBases := []FilesBases{
		{ID: "files-a", Name: "A"},
		{ID: "files-b", Name: "B"},
	}
	if err := db.Create(&filesBases).Error; err != nil {
		t.Fatalf("create files bases: %v", err)
	}
	job := CronJobs{ID: "cron-1", ScopeMode: VideoMetadataScopeSelected, JobsType: "videoMetadata"}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("create cron job: %v", err)
	}
	if err := (CronJobsFilesBases{}).Replace(db, job.ID, []string{"files-a", "files-b"}); err != nil {
		t.Fatalf("replace scope: %v", err)
	}
	loaded, err := (CronJobs{}).Info(db, job.ID)
	if err != nil {
		t.Fatalf("load cron job: %v", err)
	}
	if len(loaded.FilesBasesList) != 2 {
		t.Fatalf("expected two files bases, got %#v", loaded.FilesBasesList)
	}
}

func TestFreshDatabaseIncludesVideoMetadataSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := DB_Init(db); err != nil {
		t.Fatalf("initialize fresh database: %v", err)
	}
	for _, table := range []any{
		&ResourcesVideoMetadata{},
		&VideoMetadataSetting{},
		&VideoMetadataSettingFilesBases{},
		&VideoMetadataBatchTask{},
		&CronJobsFilesBases{},
	} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("missing migrated table for %T", table)
		}
	}
	setting, err := (VideoMetadataSetting{}).Ensure(db)
	if err != nil {
		t.Fatalf("load default setting: %v", err)
	}
	if !setting.CollectOnNewOrChanged || !setting.CollectOnDetailOrPlay ||
		setting.CollectOnList || setting.IdleBackfillEnabled {
		t.Fatalf("unexpected safe defaults: %#v", setting)
	}
}

func TestDeleteDramaSeriesAlsoDeletesVideoMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&ResourcesDramaSeries{}, &ResourcesVideoMetadata{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	ds := ResourcesDramaSeries{ID: "ds-1", ResourcesID: "resource-1", Src: "video.mp4"}
	if err := db.Create(&ds).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	if err := db.Create(&ResourcesVideoMetadata{
		DramaSeriesID: ds.ID, ProbeStatus: VideoMetadataStatusSuccess,
	}).Error; err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if err := (ResourcesDramaSeries{}).DeleteByResourcesID(db, ds.ResourcesID); err != nil {
		t.Fatalf("delete drama series: %v", err)
	}
	var count int64
	if err := db.Model(&ResourcesVideoMetadata{}).Count(&count).Error; err != nil {
		t.Fatalf("count metadata: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected metadata to be deleted, count=%d", count)
	}
}
