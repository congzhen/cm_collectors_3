package processors

import (
	"bytes"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestEffectiveSeriesSortMode(t *testing.T) {
	tests := []struct {
		name   string
		config datatype.Config_ScanDisk
		want   datatype.SeriesSortMode
	}{
		{name: "new mode", config: datatype.Config_ScanDisk{FolderToSeriesSortMode: datatype.SeriesSortModeSizeDesc}, want: datatype.SeriesSortModeSizeDesc},
		{name: "legacy enabled", config: datatype.Config_ScanDisk{FolderToSeriesSort: true}, want: datatype.SeriesSortModeNameAsc},
		{name: "legacy disabled", config: datatype.Config_ScanDisk{}, want: datatype.SeriesSortModeKeep},
		{name: "invalid mode", config: datatype.Config_ScanDisk{FolderToSeriesSortMode: "invalid"}, want: datatype.SeriesSortModeKeep},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := effectiveSeriesSortMode(test.config); got != test.want {
				t.Fatalf("effectiveSeriesSortMode() = %q, want %q", got, test.want)
			}
		})
	}
}

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

func TestImportNfoTextVariants(t *testing.T) {
	for _, input := range []string{
		`<movie><title lang="zh">标题</title></movie>`,
		`<movie><title>标题</title><title>备用标题</title></movie>`,
	} {
		root, err := parseNfoXMLToMap([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		warnings := []string{}
		service := ImportData{warnings: &warnings}
		par := datatype.ReqParam_Resource{}
		err = service.NfoExecData("", "", root, datatype.Config_ScanDisk_Nfo{Roots: []string{"movie"}, Titles: []string{"title"}}, &par)
		if err != nil || par.Resource.Title != "标题" || len(warnings) != 0 {
			t.Fatalf("unexpected NFO result: %+v %v %v", par.Resource, warnings, err)
		}
	}
}
func TestImportNfoInvalidTextKeepsOriginalTitle(t *testing.T) {
	root, err := parseNfoXMLToMap([]byte(`<movie><title><unexpected>nested</unexpected></title></movie>`))
	if err != nil {
		t.Fatal(err)
	}
	warnings := []string{}
	service := ImportData{warnings: &warnings}
	par := datatype.ReqParam_Resource{Resource: datatype.ReqParam_ResourceBase{Title: "filename"}}
	err = service.NfoExecData("", "", root, datatype.Config_ScanDisk_Nfo{Roots: []string{"movie"}, Titles: []string{"title"}}, &par)
	if err != nil || par.Resource.Title != "filename" || len(warnings) != 1 {
		t.Fatalf("lost fallback title or warning: %+v %v %v", par.Resource, warnings, err)
	}
}

func TestImportLocalPosterPreservesOriginalAndRejectsInvalidImage(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "movie.mp4")
	poster := filepath.Join(dir, "movie.jpg")
	var b bytes.Buffer
	jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 800, 600)), nil)
	if err := os.WriteFile(poster, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	config := datatype.Config_ScanDisk{CoverPosterSuffixName: []string{"jpg"}, CoverPosterType: -1}
	data, w, h, err := (ImportData{}).GetCoverPosterBase64(video, config)
	if err != nil || w != 800 || h != 600 {
		t.Fatal(w, h, err)
	}
	decoded, err := utils.Base64ToBytes(data, true)
	if err != nil || !bytes.Equal(decoded, b.Bytes()) {
		t.Fatal("original bytes were changed", err)
	}
	if err := os.WriteFile(poster, []byte("broken image"), 0600); err != nil {
		t.Fatal(err)
	}
	data, _, _, err = (ImportData{}).GetCoverPosterBase64(video, config)
	if err == nil || data != "" {
		t.Fatal("invalid poster was not rejected")
	}
}

func TestImportPreservesExistingPosterOnFallback(t *testing.T) {
	existing := models.Resources{Subtitle: "manual subtitle", CoverPosterMode: 2, CoverPosterWidth: 600, CoverPosterHeight: 900, Definition: "4K"}
	par := datatype.ReqParam_Resource{Resource: datatype.ReqParam_ResourceBase{CoverPosterMode: -1}}
	preserveImportMetadata(&par, &existing)
	if par.Resource.CoverPosterWidth != 600 || par.Resource.CoverPosterHeight != 900 || par.Resource.CoverPosterMode != 2 || par.Resource.Definition != "4K" || par.Resource.Subtitle != "manual subtitle" {
		t.Fatalf("lost retained metadata: %+v", par.Resource)
	}
	par.PhotoBase64 = "replacement"
	par.Resource.CoverPosterWidth = 400
	par.Resource.CoverPosterHeight = 600
	par.Resource.CoverPosterMode = 1
	par.Resource.Definition = "1080P"
	preserveImportMetadata(&par, &existing)
	if par.Resource.CoverPosterWidth != 400 || par.Resource.CoverPosterHeight != 600 || par.Resource.CoverPosterMode != 1 || par.Resource.Definition != "1080P" {
		t.Fatal("new metadata overwritten")
	}
}

func TestImportCommittedErrorsBecomeWarnings(t *testing.T) {
	failure := errors.New("detail query failed")
	_, committed := committedResourceResult(nil, failure)
	if !errors.Is(committed, failure) {
		t.Fatal("error cause lost")
	}
	for _, test := range []struct {
		name        string
		err         error
		saved       bool
		wantWarning bool
	}{
		{"committed resource", committed, false, true},
		{"saved episode sort failure", failure, true, true},
		{"cancel after episode save", context.Canceled, true, true},
		{"write failed", failure, false, false},
		{"cancel before save", context.Canceled, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			warnings := []string{}
			service := ImportData{warnings: &warnings}
			got := service.finishImportError(test.err, test.saved)
			if test.wantWarning {
				if got != nil || len(warnings) != 1 {
					t.Fatal(got, warnings)
				}
			} else if got != test.err || len(warnings) != 0 {
				t.Fatal(got, warnings)
			}
		})
	}
}

func TestImportCanceledBeforeProcessing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	warnings, err := (ImportData{}).ScanDiskImportDataResult(ctx, "unused", "unused", datatype.Config_ScanDisk{})
	if !errors.Is(err, context.Canceled) || len(warnings) != 0 {
		t.Fatal(err, warnings)
	}
}
