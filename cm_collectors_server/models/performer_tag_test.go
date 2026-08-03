package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupPerformerTagTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetupJoinTable(&Performer{}, "Tags", &PerformersTags{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&Performer{}, &PerformerTagClass{}, &PerformerTag{}, &PerformersTags{},
		&Resources{}, &ResourcesPerformers{}, &ResourcesDirectors{}, &FilesRelatedPerformerBases{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPerformerTagsDataSetAndFilter(t *testing.T) {
	db := setupPerformerTagTestDB(t)
	class := PerformerTagClass{ID: "class-1", PerformerBasesID: "base-1", Name: "风格", Status: true}
	tags := []PerformerTag{
		{ID: "tag-1", PerformerTagClassID: class.ID, Name: "甜美", Status: true},
		{ID: "tag-2", PerformerTagClassID: class.ID, Name: "短发", Status: true},
	}
	performers := []Performer{
		{ID: "performer-1", PerformerBasesID: "base-1", Name: "A", Status: true},
		{ID: "performer-2", PerformerBasesID: "base-1", Name: "B", Status: true},
		{ID: "performer-3", PerformerBasesID: "base-1", Name: "C", Status: true},
	}
	if err := db.Create(&class).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&tags).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&performers).Error; err != nil {
		t.Fatal(err)
	}
	if err := (PerformersTags{}).Set(db, performers[0].ID, []string{tags[0].ID, tags[1].ID}); err != nil {
		t.Fatal(err)
	}
	if err := (PerformersTags{}).Set(db, performers[1].ID, []string{tags[0].ID}); err != nil {
		t.Fatal(err)
	}
	if err := (PerformersTags{}).Set(db, performers[2].ID, []string{tags[1].ID}); err != nil {
		t.Fatal(err)
	}

	data, err := (PerformerTag{}).Data(db, "base-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Tags) != 2 || data.Tags[0].PerformerCount != 2 || data.Tags[1].PerformerCount != 2 {
		t.Fatalf("unexpected tag counts: %#v", data.Tags)
	}

	anyList, anyTotal, err := (Performer{}).DataList(db, "base-1", true, 1, 20, "", "", "", "", PerformerSortCreatedAtDesc, "", PerformerListTagFilter{TagIDs: []string{"tag-1", "tag-2"}, MatchMode: "any"})
	if err != nil {
		t.Fatal(err)
	}
	if anyTotal != 3 || len(*anyList) != 3 {
		t.Fatalf("any filter returned total=%d list=%d", anyTotal, len(*anyList))
	}

	allList, allTotal, err := (Performer{}).DataList(db, "base-1", true, 1, 20, "", "", "", "", PerformerSortCreatedAtDesc, "", PerformerListTagFilter{TagIDs: []string{"tag-1", "tag-2"}, MatchMode: "all"})
	if err != nil {
		t.Fatal(err)
	}
	if allTotal != 1 || len(*allList) != 1 || (*allList)[0].ID != performers[0].ID {
		t.Fatalf("all filter returned total=%d list=%#v", allTotal, *allList)
	}
	if len((*allList)[0].Tags) != 2 {
		t.Fatalf("expected preloaded tags, got %#v", (*allList)[0].Tags)
	}
	if err := db.Model(&Performer{}).Where("id = ?", performers[1].ID).Update("status", false).Error; err != nil {
		t.Fatal(err)
	}
	activeData, err := (PerformerTag{}).Data(db, "base-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if activeData.Tags[0].PerformerCount != 1 {
		t.Fatalf("disabled performers must not be counted: %#v", activeData.Tags)
	}

	if err := (PerformersTags{}).Set(db, performers[0].ID, []string{tags[1].ID}); err != nil {
		t.Fatal(err)
	}
	info, err := (Performer{}).InfoByID(db, performers[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Tags) != 1 || info.Tags[0].ID != tags[1].ID {
		t.Fatalf("set should replace performer tags: %#v", info.Tags)
	}
}
