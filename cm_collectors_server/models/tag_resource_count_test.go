package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestTagDataListFillsActiveResourceCounts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Tag{}, &Resources{}, &ResourcesTags{}); err != nil {
		t.Fatal(err)
	}

	tags := []Tag{
		{ID: "tag-1", TagClassID: "class-1", Name: "标签一", Status: true},
		{ID: "tag-2", TagClassID: "class-1", Name: "标签二", Status: true},
		{ID: "tag-3", TagClassID: "class-2", Name: "其他分类", Status: true},
	}
	if err := db.Create(&tags).Error; err != nil {
		t.Fatal(err)
	}

	resources := []Resources{
		{ID: "resource-1", FilesBasesID: "files-1", Title: "资源一", Status: true},
		{ID: "resource-2", FilesBasesID: "files-1", Title: "资源二", Status: true},
		{ID: "resource-3", FilesBasesID: "files-1", Title: "已停用", Status: false},
	}
	if err := db.Create(&resources).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&Resources{}).Where("id = ?", "resource-3").Update("status", false).Error; err != nil {
		t.Fatal(err)
	}

	relations := []ResourcesTags{
		{ID: "relation-1", ResourcesID: "resource-1", TagID: "tag-1"},
		{ID: "relation-2", ResourcesID: "resource-2", TagID: "tag-1"},
		{ID: "relation-3", ResourcesID: "resource-3", TagID: "tag-1"},
		{ID: "relation-4", ResourcesID: "resource-2", TagID: "tag-2"},
		{ID: "relation-5", ResourcesID: "resource-1", TagID: "tag-3"},
	}
	if err := db.Create(&relations).Error; err != nil {
		t.Fatal(err)
	}

	list, err := (Tag{}).DataListByTagClassIds(db, []string{"class-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(*list) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(*list))
	}

	counts := make(map[string]int64, len(*list))
	for _, tag := range *list {
		counts[tag.ID] = tag.ResourceCount
	}
	if counts["tag-1"] != 2 {
		t.Fatalf("expected tag-1 count 2, got %d", counts["tag-1"])
	}
	if counts["tag-2"] != 1 {
		t.Fatalf("expected tag-2 count 1, got %d", counts["tag-2"])
	}

	listWithoutCounts, err := (Tag{}).DataListByTagClassIds(db, []string{"class-1"}, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range *listWithoutCounts {
		if tag.ResourceCount != 0 {
			t.Fatalf("expected resource count to be skipped for %s, got %d", tag.ID, tag.ResourceCount)
		}
	}

	var resource Resources
	if err := db.Preload("Tags").Where("id = ?", "resource-1").First(&resource).Error; err != nil {
		t.Fatalf("preloading resource tags should ignore computed resourceCount: %v", err)
	}
}
