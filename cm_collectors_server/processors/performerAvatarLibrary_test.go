package processors

import (
	"cm_collectors_server/config"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

const avatarLibraryTestData = `{
	"Information":{"TotalNum":"3","TotalSize":"100","Timestamp":"123"},
	"Content":{
		"Low":{"测试演员.jpg":"测试演员.jpg?t=1","Test Actor.jpg":"测试演员.jpg?t=1"},
		"High":{"测试演员.jpg":"AI-Fix-测试演员.jpg?t=2","测试演员-1.jpg":"测试演员-1.jpg?t=3"}
	}
}`

func TestBuildAvatarLibraryIndexPreservesQualityOrderAndAliases(t *testing.T) {
	data := []byte(avatarLibraryTestData)
	index, err := buildAvatarLibraryIndex(data, performerAvatarMetadata{ActiveBaseURL: avatarLibraryDefaultBaseURL})
	if err != nil {
		t.Fatal(err)
	}
	candidates := matchAvatarCandidates(index, models.Performer{Name: "不存在", AliasName: "Test Actor，测试演员"})
	if len(candidates) != 3 {
		t.Fatalf("expected 3 unique candidates, got %d", len(candidates))
	}
	if candidates[0].FileName != "测试演员-1.jpg?t=3" {
		t.Fatalf("expected highest ranked candidate first, got %s", candidates[0].FileName)
	}
	if candidates[1].FileName != "AI-Fix-测试演员.jpg?t=2" {
		t.Fatalf("expected AI candidate second, got %s", candidates[1].FileName)
	}
}

func TestSplitAvatarAliasesSupportsMixedSeparatorsAndKeepsFullName(t *testing.T) {
	aliases := splitAvatarAliases("Rui Hiduki，测试演员;藤咲舞／新妻优香　栗衣みい")
	expected := []string{
		"Rui Hiduki，测试演员;藤咲舞／新妻优香　栗衣みい",
		"Rui",
		"Hiduki",
		"测试演员",
		"藤咲舞",
		"新妻优香",
		"栗衣みい",
	}
	if len(aliases) != len(expected) {
		t.Fatalf("expected %d aliases, got %d: %#v", len(expected), len(aliases), aliases)
	}
	for index, alias := range expected {
		if aliases[index] != alias {
			t.Fatalf("expected alias %d to be %q, got %q", index, alias, aliases[index])
		}
	}
}

func TestMatchAvatarCandidatesUsesScriptConversionOnlyAsFallback(t *testing.T) {
	exactCandidate := PerformerAvatarCandidate{ID: "exact", Rank: 1}
	exactAliasCandidate := PerformerAvatarCandidate{ID: "exact-alias", Rank: 3}
	convertedCandidate := PerformerAvatarCandidate{ID: "converted", Rank: 2}
	index := &performerAvatarIndex{
		ByName: map[string][]PerformerAvatarCandidate{
			normalizeAvatarName("藤咲まい"): {exactCandidate},
			normalizeAvatarName("别名演员"): {exactAliasCandidate},
			normalizeAvatarName("灘坂舞"):  {convertedCandidate},
		},
	}

	candidates := matchAvatarCandidates(index, models.Performer{
		Name:      "藤咲まい",
		AliasName: "别名演员，滩坂舞",
	})
	if len(candidates) != 1 || candidates[0].ID != exactCandidate.ID {
		t.Fatalf("expected the exact match to win, got %#v", candidates)
	}

	candidates = matchAvatarCandidates(index, models.Performer{
		Name:      "不存在",
		AliasName: "滩坂舞",
	})
	if len(candidates) != 1 || candidates[0].ID != convertedCandidate.ID {
		t.Fatalf("expected simplified/traditional fallback match, got %#v", candidates)
	}
}

func TestAvatarBatchProgressTracksConcurrentWork(t *testing.T) {
	const batchID = "progress-test"
	avatarBatchProgress.Delete(batchID)
	tracker := startAvatarBatchProgress(batchID, PerformerAvatarBatchPreview{
		Total: 6, Unmatched: 2, SkippedExisting: 1,
	})
	setAvatarBatchActorRunning(tracker, models.Performer{ID: "actor-1", Name: "演员一"})

	progress, ok := (PerformerAvatarLibrary{}).BatchProgress(batchID)
	if !ok {
		t.Fatal("expected running batch progress")
	}
	if progress.Failures == nil || progress.CurrentActors == nil {
		t.Fatalf("progress collections must serialize as arrays: %#v", progress)
	}
	if progress.Completed != 3 || len(progress.CurrentActors) != 1 || progress.CurrentActors[0] != "演员一" {
		t.Fatalf("unexpected running progress: %#v", progress)
	}

	completeAvatarBatchActor(tracker, models.Performer{ID: "actor-1", Name: "演员一"}, nil)
	progress, ok = (PerformerAvatarLibrary{}).BatchProgress(batchID)
	if !ok {
		t.Fatal("expected completed batch progress")
	}
	if progress.Completed != 4 || progress.Success != 1 || len(progress.CurrentActors) != 0 {
		t.Fatalf("unexpected completed progress: %#v", progress)
	}

	finishAvatarBatchProgress(batchID, tracker)
	progress, ok = (PerformerAvatarLibrary{}).BatchProgress(batchID)
	if !ok {
		t.Fatal("expected final batch progress")
	}
	if !progress.Done || progress.Completed != progress.Total {
		t.Fatalf("unexpected final progress: %#v", progress)
	}
	avatarBatchProgress.Delete(batchID)
}

func TestDownloadAvatarLibraryFileRetriesRateLimit(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requestCount.Add(1) < 3 {
			writer.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = writer.Write([]byte("ok"))
	}))
	defer server.Close()

	data, err := downloadAvatarLibraryFileWithRetry(&http.Client{}, server.URL, 1024, 3)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ok" || requestCount.Load() != 3 {
		t.Fatalf("unexpected retry result: data=%q requests=%d", data, requestCount.Load())
	}
}

func TestUpdateAvatarLibraryDataFileFromCustomSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/Filetree.json" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(avatarLibraryTestData))
	}))
	defer server.Close()

	originalConfig := core.Config
	core.Config = &config.Config{}
	testRoot := t.TempDir()
	core.Config.System.FilePath = filepath.Join(testRoot, "db")
	core.Config.PerformerAvatarLibrary.CachePath = filepath.Join(testRoot, "runtime", "cache", "gfriends")
	core.Config.PerformerAvatarLibrary.CustomBaseURL = server.URL
	core.Config.PerformerAvatarLibrary.DefaultStrategy = string(datatype.PerformerAvatarStrategyRecommended)
	defer func() {
		core.Config = originalConfig
		avatarLibraryRuntime.Lock()
		avatarLibraryRuntime.Index = nil
		avatarLibraryRuntime.Updating = false
		avatarLibraryRuntime.Unlock()
	}()
	avatarLibraryRuntime.Lock()
	avatarLibraryRuntime.Index = nil
	avatarLibraryRuntime.Updating = false
	avatarLibraryRuntime.Unlock()

	library := PerformerAvatarLibrary{}
	status, err := library.UpdateDataFile()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready || status.TotalNum != "3" || status.ActiveBaseURL != server.URL+"/" {
		t.Fatalf("unexpected status: %#v", status)
	}
	if _, err := os.Stat(filepath.Join(core.Config.PerformerAvatarLibrary.CachePath, "Filetree.json")); err != nil {
		t.Fatalf("data file was not saved: %v", err)
	}
}

func TestSelectAvatarCandidateStrategies(t *testing.T) {
	candidates := []PerformerAvatarCandidate{
		{ID: "ai", AIFixed: true, Rank: 3},
		{ID: "original", AIFixed: false, Rank: 2},
	}
	tests := []struct {
		strategy datatype.PerformerAvatarStrategy
		wantID   string
	}{
		{datatype.PerformerAvatarStrategyRecommended, "ai"},
		{datatype.PerformerAvatarStrategyOriginal, "original"},
		{datatype.PerformerAvatarStrategyAIFix, "ai"},
	}
	for _, test := range tests {
		got, ok := selectAvatarCandidate(candidates, test.strategy)
		if !ok || got.ID != test.wantID {
			t.Fatalf("strategy %s: expected %s, got %#v", test.strategy, test.wantID, got)
		}
	}
}
