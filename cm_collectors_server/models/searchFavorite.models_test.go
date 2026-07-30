package models

import (
	"cm_collectors_server/datatype"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSearchFavoritePersistsStructuredSearchData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&SearchFavorite{}); err != nil {
		t.Fatalf("migrate search favorites: %v", err)
	}

	favorite := &SearchFavorite{
		ID:           "favorite-1",
		FilesBasesID: "library-1",
		SearchData: datatype.ReqParam_SearchData{
			SearchTextSlc: []string{"电影"},
			Country: datatype.I_searchGroup{
				Logic:   datatype.E_searchLogic("multiOr"),
				Options: []string{"中国", "日本"},
			},
			Tag: map[string]datatype.I_searchGroup{
				"class-1": {
					Logic:   datatype.E_searchLogic("multiAnd"),
					Options: []string{"tag-1", "tag-2"},
				},
			},
		},
		SchemaVersion: 1,
	}
	if err := (SearchFavorite{}).Create(db, favorite); err != nil {
		t.Fatalf("create search favorite: %v", err)
	}

	saved, err := (SearchFavorite{}).Get(db, favorite.ID)
	if err != nil {
		t.Fatalf("get search favorite: %v", err)
	}
	if saved.SearchData.Country.Logic != datatype.E_searchLogic("multiOr") ||
		len(saved.SearchData.Tag["class-1"].Options) != 2 {
		t.Fatalf("structured search data was not restored: %#v", saved.SearchData)
	}

	saved.SearchData.Year = datatype.I_searchGroup{
		Logic:   datatype.E_searchLogic("single"),
		Options: []string{"2026"},
	}
	if err := (SearchFavorite{}).Save(db, saved); err != nil {
		t.Fatalf("update search favorite: %v", err)
	}
	updated, err := (SearchFavorite{}).Get(db, favorite.ID)
	if err != nil || len(updated.SearchData.Year.Options) != 1 {
		t.Fatalf("search favorite update was not persisted: %#v, %v", updated, err)
	}
}

func TestSearchFavoriteListIsIsolatedByLibrary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&SearchFavorite{}); err != nil {
		t.Fatalf("migrate search favorites: %v", err)
	}
	records := []SearchFavorite{
		{ID: "one", FilesBasesID: "library-1", Sort: 1},
		{ID: "two", FilesBasesID: "library-1", Sort: 0},
		{ID: "three", FilesBasesID: "library-2", Sort: 0},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("create favorites: %v", err)
	}
	list, err := (SearchFavorite{}).List(db, "library-1")
	if err != nil || len(list) != 2 || list[0].ID != "two" {
		t.Fatalf("unexpected library favorites: %#v, %v", list, err)
	}
}
