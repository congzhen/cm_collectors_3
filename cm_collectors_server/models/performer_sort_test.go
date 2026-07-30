package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPerformerDataListSortsByResourceCountBeforePagination(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&Performer{}, &Resources{}, &ResourcesPerformers{}, &ResourcesDirectors{}, &FilesRelatedPerformerBases{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	performers := []Performer{
		{ID: "performer-a", PerformerBasesID: "performer-base", Name: "A", Status: true},
		{ID: "performer-b", PerformerBasesID: "performer-base", Name: "B", Status: true},
		{ID: "performer-c", PerformerBasesID: "performer-base", Name: "C", Status: true},
	}
	resources := []Resources{
		{ID: "resource-1", FilesBasesID: "files-base", Title: "R1", Status: true},
		{ID: "resource-2", FilesBasesID: "files-base", Title: "R2", Status: true},
		{ID: "resource-3", FilesBasesID: "files-base", Title: "R3", Status: true},
		{ID: "resource-4", FilesBasesID: "unrelated-files-base", Title: "R4", Status: true},
	}
	resourcePerformers := []ResourcesPerformers{
		{ID: "rp-1", ResourcesID: "resource-1", PerformerID: "performer-a"},
		{ID: "rp-2", ResourcesID: "resource-2", PerformerID: "performer-a"},
		{ID: "rp-3", ResourcesID: "resource-3", PerformerID: "performer-b"},
		{ID: "rp-4", ResourcesID: "resource-4", PerformerID: "performer-c"},
	}
	// The same performer/resource pair also appears as a director. UNION must
	// keep it from being counted twice.
	resourceDirectors := []ResourcesDirectors{
		{ID: "rd-1", ResourcesID: "resource-1", PerformerID: "performer-a"},
	}
	relatedBases := []FilesRelatedPerformerBases{
		{ID: "related-1", FilesBasesID: "files-base", PerformerBasesID: "performer-base", Main: true},
	}
	for _, data := range []any{&performers, &resources, &resourcePerformers, &resourceDirectors, &relatedBases} {
		if err := db.Create(data).Error; err != nil {
			t.Fatalf("seed test database: %v", err)
		}
	}

	firstPage, total, err := (Performer{}).DataList(
		db, "performer-base", true, 1, 2, "", "", "", "", PerformerSortResourceCountDesc, "files-base",
	)
	if err != nil {
		t.Fatalf("query first page: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	assertPerformerCounts(t, *firstPage, []string{"performer-a", "performer-b"}, []int64{2, 1})

	secondPage, _, err := (Performer{}).DataList(
		db, "performer-base", false, 2, 2, "", "", "", "", PerformerSortResourceCountDesc, "files-base",
	)
	if err != nil {
		t.Fatalf("query second page: %v", err)
	}
	assertPerformerCounts(t, *secondPage, []string{"performer-c"}, []int64{0})

	ascending, _, err := (Performer{}).DataList(
		db, "performer-base", false, 1, 3, "", "", "", "", PerformerSortResourceCountAsc, "files-base",
	)
	if err != nil {
		t.Fatalf("query ascending order: %v", err)
	}
	assertPerformerCounts(t, *ascending, []string{"performer-c", "performer-b", "performer-a"}, []int64{0, 1, 2})

	allRelatedFilesBases, _, err := (Performer{}).DataList(
		db, "performer-base", false, 1, 3, "", "", "", "", PerformerSortResourceCountDesc, "",
	)
	if err != nil {
		t.Fatalf("query related file bases: %v", err)
	}
	assertPerformerCounts(t, *allRelatedFilesBases, []string{"performer-a", "performer-b", "performer-c"}, []int64{2, 1, 0})
}

func TestPerformerDataListSortsByNameBeforePagination(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&Performer{}, &Resources{}, &ResourcesPerformers{}, &ResourcesDirectors{}, &FilesRelatedPerformerBases{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	performers := []Performer{
		{ID: "performer-charlie", PerformerBasesID: "performer-base", Name: "Charlie", Status: true},
		{ID: "performer-alice", PerformerBasesID: "performer-base", Name: "Alice", Status: true},
		{ID: "performer-bob", PerformerBasesID: "performer-base", Name: "Bob", Status: true},
	}
	if err := db.Create(&performers).Error; err != nil {
		t.Fatalf("seed performers: %v", err)
	}

	firstPage, total, err := (Performer{}).DataList(
		db, "performer-base", true, 1, 2, "", "", "", "", PerformerSortNameAsc, "",
	)
	if err != nil {
		t.Fatalf("query ascending first page: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	assertPerformerNames(t, *firstPage, []string{"Alice", "Bob"})

	secondPage, _, err := (Performer{}).DataList(
		db, "performer-base", false, 2, 2, "", "", "", "", PerformerSortNameAsc, "",
	)
	if err != nil {
		t.Fatalf("query ascending second page: %v", err)
	}
	assertPerformerNames(t, *secondPage, []string{"Charlie"})

	descending, _, err := (Performer{}).DataList(
		db, "performer-base", false, 1, 3, "", "", "", "", PerformerSortNameDesc, "",
	)
	if err != nil {
		t.Fatalf("query descending order: %v", err)
	}
	assertPerformerNames(t, *descending, []string{"Charlie", "Bob", "Alice"})
}

func assertPerformerCounts(t *testing.T, actual []Performer, ids []string, counts []int64) {
	t.Helper()
	if len(actual) != len(ids) {
		t.Fatalf("expected %d performers, got %d", len(ids), len(actual))
	}
	for i := range ids {
		if actual[i].ID != ids[i] || actual[i].ResourceCount != counts[i] {
			t.Fatalf("position %d: expected %s count %d, got %s count %d", i, ids[i], counts[i], actual[i].ID, actual[i].ResourceCount)
		}
	}
}

func assertPerformerNames(t *testing.T, actual []Performer, names []string) {
	t.Helper()
	if len(actual) != len(names) {
		t.Fatalf("expected %d performers, got %d", len(names), len(actual))
	}
	for i := range names {
		if actual[i].Name != names[i] {
			t.Fatalf("position %d: expected %s, got %s", i, names[i], actual[i].Name)
		}
	}
}
