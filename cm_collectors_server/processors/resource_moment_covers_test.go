package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func TestMomentUpdateDetectsConcurrentDeletionAndAcceptsUnchangedRow(t *testing.T) {
	db := momentTestDB(t)
	if err := db.Create(&models.VideoMoment{ID: "exists", Title: "unchanged"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := momentUpdateResult(&gorm.DB{RowsAffected: 0}, "exists"); err != nil {
		t.Fatalf("zero affected rows must not reject an existing bookmark: %v", err)
	}
	if err := momentUpdateResult(&gorm.DB{RowsAffected: 0}, "deleted"); err == nil {
		t.Fatal("reported a successful update for a deleted bookmark")
	}
}

func TestMomentEditRejectsBookmarkDeletedDuringDurationProbe(t *testing.T) {
	db := momentTestDB(t)
	if err := os.WriteFile("video.mp4", []byte("test source"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, value := range []interface{}{
		&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true},
		&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: "video.mp4"},
		&models.VideoMoment{ID: "m", ResourceID: "r", DramaSeriesID: "d", Title: "old", Start: 1, End: 3},
	} {
		if err := db.Create(value).Error; err != nil {
			t.Fatal(err)
		}
	}
	probed := false
	videoMomentDuration = func(string) (float64, error) {
		probed = true
		return 120, db.Where("id = ?", "m").Delete(&models.VideoMoment{}).Error
	}
	if _, err := (VideoMoments{}).Save("m", models.VideoMoment{Title: "edited", Start: 2, End: 4}); err == nil {
		t.Fatal("editing a concurrently deleted bookmark reported success")
	}
	if !probed {
		t.Fatal("test did not reach the concurrent deletion during probe")
	}
}

func TestMomentSaveRejectsResourceDeletedDuringDurationProbe(t *testing.T) {
	db := momentTestDB(t)
	if err := db.AutoMigrate(&models.ResourcePlayDaily{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("video.mp4", []byte("test source"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Resources{ID: "r", Mode: datatype.E_resourceMode_Movies, Status: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.ResourcesDramaSeries{ID: "d", ResourcesID: "r", Src: "video.mp4"}).Error; err != nil {
		t.Fatal(err)
	}
	videoMomentDuration = func(string) (float64, error) {
		// Simulate deletion after the initial source lookup but before insertion.
		// The expensive probe must not hold the resource write lock.
		if !resourceMomentWriteMutex.TryLock() {
			t.Fatal("duration probe holds the resource mutation lock")
		}
		resourceMomentWriteMutex.Unlock()
		return 120, (models.Resources{}).DeleteById(db, "r")
	}
	if _, err := (VideoMoments{}).Save("", models.VideoMoment{DramaSeriesID: "d", Title: "late bookmark", Start: 1, End: 3}); err == nil {
		t.Fatal("saved a bookmark for a permanently deleted resource")
	}
	var count int64
	if err := db.Model(&models.VideoMoment{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("orphan bookmark retained: %d %v", count, err)
	}
}

func TestDeletedResourceMomentCoverCleanupIsTargetedAndIdempotent(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, id := range []string{"deleted-1", "deleted-2", "retained"} {
		if err := os.MkdirAll(momentCoverDir(id), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(momentCoverDir(id), "cover.jpg"), []byte("cover"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := deleteMomentCovers([]string{"deleted-1", "deleted-2"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"deleted-1", "deleted-2"} {
		if _, err := os.Stat(momentCoverDir(id)); !os.IsNotExist(err) {
			t.Fatalf("deleted moment cover still exists: %s %v", id, err)
		}
	}
	if _, err := os.Stat(filepath.Join(momentCoverDir("retained"), "cover.jpg")); err != nil {
		t.Fatalf("unrelated moment cover removed: %v", err)
	}
}
