package models

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestImportLookupIsLibraryScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	for _, sql := range []string{
		"CREATE TABLE resources (id TEXT PRIMARY KEY, filesBases_id TEXT)",
		"CREATE TABLE resourcesDramaSeries (id TEXT PRIMARY KEY, resources_id TEXT, src TEXT)",
		"INSERT INTO resources VALUES ('ra','a'),('rb','b')",
		"INSERT INTO resourcesDramaSeries VALUES ('da','ra','same.mp4')",
	} {
		if err = db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	list, err := (ResourcesDramaSeries{}).ListBySrcInLibrary(db, "b", "same.mp4")
	if err != nil || len(*list) != 0 {
		t.Fatalf("another library matched: %v %v", list, err)
	}
	if err = db.Exec("INSERT INTO resourcesDramaSeries VALUES ('db','rb','same.mp4')").Error; err != nil {
		t.Fatal(err)
	}
	list, err = (ResourcesDramaSeries{}).ListBySrcInLibrary(db, "b", "same.mp4")
	if err != nil || len(*list) != 1 || (*list)[0].ResourcesID != "rb" {
		t.Fatalf("wrong target: %v %v", list, err)
	}
	if _, err = (ResourcesDramaSeries{}).ListBySrcInLibrary(db, "", "same.mp4"); err == nil {
		t.Fatal("unscoped lookup allowed")
	}
}
