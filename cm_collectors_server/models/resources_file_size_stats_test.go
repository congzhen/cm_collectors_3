package models

import (
	"cm_collectors_server/datatype"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newResourceFileSizeStatsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&Resources{},
		&ResourcesDramaSeries{},
		&ResourcesVideoMetadata{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	return db
}

func TestResourcesFileSizeStatsUsesAllFilteredVideoResources(t *testing.T) {
	db := newResourceFileSizeStatsTestDB(t)
	resources := []Resources{
		{ID: "movie-match", FilesBasesID: "library-1", Mode: datatype.E_resourceMode_Movies, Title: "匹配电影"},
		{ID: "movie-other", FilesBasesID: "library-1", Mode: datatype.E_resourceMode_Movies, Title: "其他电影"},
		{ID: "comic-match", FilesBasesID: "library-1", Mode: datatype.E_resourceMode_Comic, Title: "匹配漫画"},
		{ID: "link-match", FilesBasesID: "library-1", Mode: datatype.E_resourceMode_VideoLink, Title: "匹配链接"},
		{ID: "other-library", FilesBasesID: "library-2", Mode: datatype.E_resourceMode_Movies, Title: "匹配异库"},
	}
	if err := db.Create(&resources).Error; err != nil {
		t.Fatalf("create resources: %v", err)
	}
	series := []ResourcesDramaSeries{
		{ID: "match-success", ResourcesID: "movie-match", Src: "one.mp4"},
		{ID: "match-stale", ResourcesID: "movie-match", Src: "two.mp4"},
		{ID: "match-missing", ResourcesID: "movie-match", Src: "three.mp4"},
		{ID: "other-success", ResourcesID: "movie-other", Src: "four.mp4"},
		{ID: "comic-success", ResourcesID: "comic-match", Src: "five.jpg"},
		{ID: "link-success", ResourcesID: "link-match", Src: "https://example.com/video.mp4"},
		{ID: "library-success", ResourcesID: "other-library", Src: "six.mp4"},
	}
	if err := db.Create(&series).Error; err != nil {
		t.Fatalf("create drama series: %v", err)
	}
	metadata := []ResourcesVideoMetadata{
		{DramaSeriesID: "match-success", ProbeStatus: VideoMetadataStatusSuccess, MetadataVersion: 2, FileSize: 100},
		{DramaSeriesID: "match-stale", ProbeStatus: VideoMetadataStatusStale, MetadataVersion: 0, FileSize: 200},
		{DramaSeriesID: "other-success", ProbeStatus: VideoMetadataStatusSuccess, MetadataVersion: 2, FileSize: 400},
		{DramaSeriesID: "comic-success", ProbeStatus: VideoMetadataStatusSuccess, MetadataVersion: 2, FileSize: 500},
		{DramaSeriesID: "link-success", ProbeStatus: VideoMetadataStatusSuccess, MetadataVersion: 2, FileSize: 700},
		{DramaSeriesID: "library-success", ProbeStatus: VideoMetadataStatusSuccess, MetadataVersion: 2, FileSize: 600},
	}
	if err := db.Create(&metadata).Error; err != nil {
		t.Fatalf("create metadata: %v", err)
	}

	par := &datatype.ReqParam_ResourceFileSizeStats{
		FilesBasesId: "library-1",
		SearchData: datatype.ReqParam_SearchData{
			SearchTextSlc: []string{"匹配"},
		},
	}
	stats, err := (Resources{}).FileSizeStats(db, par, 2)
	if err != nil {
		t.Fatalf("aggregate file size stats: %v", err)
	}
	if stats.TotalFiles != 4 || stats.CountedFiles != 2 || stats.UncountedFiles != 2 || stats.TotalSize != 800 {
		t.Fatalf("unexpected file size stats: %#v", stats)
	}
}

func TestResourcesFileSizeStatsReturnsZerosForNoMatches(t *testing.T) {
	db := newResourceFileSizeStatsTestDB(t)
	par := &datatype.ReqParam_ResourceFileSizeStats{
		FilesBasesId: "missing-library",
	}
	stats, err := (Resources{}).FileSizeStats(db, par, 2)
	if err != nil {
		t.Fatalf("aggregate empty file size stats: %v", err)
	}
	if stats.TotalFiles != 0 || stats.CountedFiles != 0 || stats.UncountedFiles != 0 || stats.TotalSize != 0 {
		t.Fatalf("expected zero stats: %#v", stats)
	}
}
