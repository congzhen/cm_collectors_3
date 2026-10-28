package models

import (
	"testing"

	"gorm.io/gorm"
)

func seedResourceMoments(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, r := range []Resources{{ID: "a", FilesBasesID: "lib"}, {ID: "b", FilesBasesID: "lib"}, {ID: "c", FilesBasesID: "other"}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
		for _, suffix := range []string{"1", "2"} {
			if err := db.Create(&VideoMoment{ID: r.ID + suffix, ResourceID: r.ID, DramaSeriesID: "missing-file", Note: "keep until resource is deleted"}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestPermanentResourceDeletionRemovesOnlyRelatedMoments(t *testing.T) {
	db := sidebarTestDB(t)
	seedResourceMoments(t, db)
	if err := (Resources{}).DeleteById(db, "a"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&VideoMoment{}).Count(&count).Error; err != nil || count != 4 {
		t.Fatalf("single delete left %d moments: %v", count, err)
	}
	if err := db.Model(&VideoMoment{}).Where("resource_id = ?", "a").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("deleted resource retained moments: %d %v", count, err)
	}
	if err := (Resources{}).DeleteByFilesBasesID(db, "lib"); err != nil {
		t.Fatal(err)
	}
	var remaining []VideoMoment
	if err := db.Find(&remaining).Error; err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 2 || remaining[0].ResourceID != "c" || remaining[1].ResourceID != "c" {
		t.Fatalf("library delete affected unrelated library: %+v", remaining)
	}
}

func TestFailedResourceDeletionRestoresMoments(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "library"}[batch], func(t *testing.T) {
			db := sidebarTestDB(t)
			seedResourceMoments(t, db)
			if err := db.Exec(`CREATE TRIGGER reject_resource_delete BEFORE DELETE ON resources BEGIN SELECT RAISE(ABORT, 'delete blocked'); END`).Error; err != nil {
				t.Fatal(err)
			}
			var err error
			if batch {
				err = (Resources{}).DeleteByFilesBasesID(db, "lib")
			} else {
				err = (Resources{}).DeleteById(db, "a")
			}
			if err == nil {
				t.Fatal("expected deletion failure")
			}
			var moments, resources int64
			db.Model(&VideoMoment{}).Count(&moments)
			db.Model(&Resources{}).Count(&resources)
			if moments != 6 || resources != 3 {
				t.Fatalf("rollback lost records: resources=%d moments=%d", resources, moments)
			}
		})
	}
}
