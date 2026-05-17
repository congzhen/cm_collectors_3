package processors

import (
	"cm_collectors_server/models"
	"testing"
)

func TestNormalizeSeriesName(t *testing.T) {
	tests := map[string]string{
		"Show.Name.S01E02":  "show name",
		"连续剧 第12集":          "连续剧",
		"Drama-Part-03":     "drama",
		"Movie 2024":        "movie",
		"Series [1080p] 01": "series",
	}

	for input, expected := range tests {
		if actual := normalizeSeriesName(input); actual != expected {
			t.Fatalf("normalizeSeriesName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestIsSimilarSeriesName(t *testing.T) {
	target := normalizeSeriesName("My Show 第02集")

	if !isSimilarSeriesName(target, "My Show 第01集") {
		t.Fatal("expected episode names to match")
	}
	if !isSimilarSeriesName(target, "My Show - 03") {
		t.Fatal("expected simple numbered suffix to match")
	}
	if isSimilarSeriesName(target, "Other Show 第01集") {
		t.Fatal("expected different titles not to match")
	}
}

func TestIsSimilarSeriesFileNameLeadingEpisodePrefix(t *testing.T) {
	tests := [][2]string{
		{"01xxxxxxx", "02xxxxxxx"},
		{"01.xxxxxxx", "02.xxxxxxx"},
		{"01-短剧", "02.短剧"},
	}

	for _, test := range tests {
		if !isSimilarSeriesFileName(test[0], test[1]) {
			t.Fatalf("expected %q and %q to match", test[0], test[1])
		}
	}

	if isSimilarSeriesFileName("01xxxxxxx", "02yyyyyyy") {
		t.Fatal("expected different leading-number suffixes not to match")
	}
}

func TestFilterSimilarNameDramaSeriesSimpleEpisodeNames(t *testing.T) {
	existing := []models.DramaSeriesWithResource{
		{
			ResourcesID: "resource-1",
			Src:         "D:/video/show/01.mp4",
			Title:       "01",
		},
	}

	matched := ImportData{}.filterSimilarNameDramaSeries(&existing, "02")
	if len(*matched) != 1 || (*matched)[0].ResourcesID != "resource-1" {
		t.Fatalf("expected simple numbered filenames to match existing resource, got %#v", matched)
	}
}

func TestFilterSimilarNameDramaSeriesLeadingEpisodePrefix(t *testing.T) {
	existing := []models.DramaSeriesWithResource{
		{
			ResourcesID: "resource-1",
			Src:         "D:/video/show/01.xxxxxxx.mp4",
			Title:       "01.xxxxxxx",
		},
	}

	matched := ImportData{}.filterSimilarNameDramaSeries(&existing, "02.xxxxxxx")
	if len(*matched) != 1 || (*matched)[0].ResourcesID != "resource-1" {
		t.Fatalf("expected leading numbered filenames to match existing resource, got %#v", matched)
	}
}

func TestFindBestMatchedNfoPathByTruncatedName(t *testing.T) {
	nfos := []string{
		"D:/video/高橋ほなみ_MIDV_893_1080P_4_5_2024_系列：未知系列_被我恨得要死的变态老师逼着我一次又一次的射精…我连.nfo",
	}

	matched := findBestMatchedNfoPath(nfos, "高橋ほなみ_MIDV_893_1080P_4_5_2024_系列：未知系列_被我恨得要死的变态老师逼着我一次又一次的射精…我连学校里的求救都打不通")
	if matched != nfos[0] {
		t.Fatalf("expected truncated nfo name to match by issue number, got %q", matched)
	}
}

func TestParseNfoXMLToMapRepairsBareAmpersand(t *testing.T) {
	root, err := parseNfoXMLToMap([]byte(`<?xml version="1.0" encoding="UTF-8"?><movie><title>Tom & Jerry</title></movie>`))
	if err != nil {
		t.Fatalf("expected repaired XML to parse, got %v", err)
	}

	movie, ok := root["movie"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected movie root, got %#v", root["movie"])
	}
	if movie["title"] != "Tom & Jerry" {
		t.Fatalf("expected title to survive repair, got %#v", movie["title"])
	}
}
