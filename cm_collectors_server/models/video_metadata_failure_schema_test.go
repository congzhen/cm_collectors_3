package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateVideoMetadataFailureManagementSchemaDoesNotRewriteHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Exec(`CREATE TABLE resourcesDramaSeries (
		id char(20) PRIMARY KEY,
		resources_id char(20),
		type varchar(50),
		src text,
		sort integer DEFAULT 0,
		durationSeconds integer DEFAULT 0,
		durationProbeStatus varchar(20),
		durationProbeTime datetime,
		m3u8BuilderTime datetime,
		m3u8BuilderStatus tinyint DEFAULT 0
	)`).Error; err != nil {
		t.Fatalf("create legacy drama series table: %v", err)
	}
	legacy := ResourcesDramaSeries{
		ID: "legacy-image", Src: "poster.jpg", DurationSeconds: 99,
	}
	if err := db.Exec(
		"INSERT INTO resourcesDramaSeries (id, src, durationSeconds) VALUES (?, ?, ?)",
		legacy.ID, legacy.Src, legacy.DurationSeconds,
	).Error; err != nil {
		t.Fatalf("create legacy row: %v", err)
	}

	if err := migrateVideoMetadataFailureManagementSchema(db); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	if !db.Migrator().HasColumn(&ResourcesDramaSeries{}, "VideoMetadataExcluded") {
		t.Fatal("video_metadata_excluded column was not created")
	}
	if !db.Migrator().HasColumn(&ResourcesVideoMetadata{}, "MetadataSource") {
		t.Fatal("metadata_source column was not created")
	}

	var after ResourcesDramaSeries
	if err := db.First(&after, "id = ?", legacy.ID).Error; err != nil {
		t.Fatalf("load legacy row: %v", err)
	}
	if after.VideoMetadataExcluded || after.DurationSeconds != 99 || after.Src != legacy.Src {
		t.Fatalf("schema migration rewrote historical data: %#v", after)
	}
}

func TestExclusionMigrationPreservesLegacyAndDisabledSetting(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"CREATE TABLE resourcesDramaSeries (id char(20) PRIMARY KEY, src text, video_metadata_excluded integer DEFAULT 0)",
		"INSERT INTO resourcesDramaSeries (id,src,video_metadata_excluded) VALUES ('legacy','old.mp4',1)",
		"CREATE TABLE video_metadata_settings (id char(20) PRIMARY KEY)",
		"INSERT INTO video_metadata_settings (id) VALUES ('default')",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := db.AutoMigrate(&ResourcesDramaSeries{}, &VideoMetadataSetting{}); err != nil {
			t.Fatal(err)
		}
	}
	var ds ResourcesDramaSeries
	if err := db.First(&ds, "id = ?", "legacy").Error; err != nil {
		t.Fatal(err)
	}
	if !ds.VideoMetadataExcluded || ds.VideoMetadataClassification != "" {
		t.Fatal("legacy manual intent lost")
	}
	setting, err := MetadataRules(db)
	if err != nil {
		t.Fatal(err)
	}
	if MetadataSourceRule(setting, "cover.jpg") != "system" {
		t.Fatal("legacy default not enabled")
	}
	disabled := false
	setting.AutoExcludeNonVideo = &disabled
	setting.ExcludedExtensions = "pdf"
	if err := (VideoMetadataSetting{}).Save(db, &setting); err != nil {
		t.Fatal(err)
	}
	reloaded, err := MetadataRules(db)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.AutoExcludeNonVideo == nil || *reloaded.AutoExcludeNonVideo || MetadataSourceRule(reloaded, "cover.jpg") != "" || MetadataSourceRule(reloaded, "book.pdf") != "custom" {
		t.Fatal("saved false/custom setting lost")
	}
}
