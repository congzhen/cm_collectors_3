package processors

import (
	"cm_collectors_server/datatype"
	"testing"
)

func TestNormalizeSearchFavoriteDataFillsOldStructure(t *testing.T) {
	data := datatype.ReqParam_SearchData{
		SearchTextSlc: []string{" 电影 ", "电影", ""},
		Country: datatype.I_searchGroup{
			Logic:   datatype.E_searchLogic("unknown"),
			Options: []string{"中国", "中国", ""},
		},
	}
	normalizeSearchFavoriteData(&data)
	if len(data.SearchTextSlc) != 1 || data.SearchTextSlc[0] != "电影" {
		t.Fatalf("unexpected normalized search text: %#v", data.SearchTextSlc)
	}
	if data.Country.Logic != datatype.E_searchLogic("single") ||
		len(data.Country.Options) != 1 {
		t.Fatalf("unexpected normalized group: %#v", data.Country)
	}
	if data.Sort != datatype.E_searchSort("addTimeDesc") || data.Tag == nil {
		t.Fatalf("old search structure was not completed: %#v", data)
	}
}

func TestHasSearchFavoriteConditionsIgnoresSortOnly(t *testing.T) {
	data := datatype.ReqParam_SearchData{Sort: datatype.E_searchSort("hot")}
	normalizeSearchFavoriteData(&data)
	if hasSearchFavoriteConditions(data) {
		t.Fatal("sort-only state should not be treated as a saved search condition")
	}
	data.Year.Options = []string{"2026"}
	if !hasSearchFavoriteConditions(data) {
		t.Fatal("selected filter should be treated as a saved search condition")
	}
}
