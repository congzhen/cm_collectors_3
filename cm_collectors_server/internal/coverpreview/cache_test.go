package coverpreview

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestReuseRefreshInvalidationAndRestart(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "video.mp4")
	os.WriteFile(source, []byte("video"), 0600)
	c := Cache{Root: filepath.Join(root, "cache")}
	calls := 0
	generate := func() ([]string, error) { calls++; return []string{fmt.Sprint(calls)}, nil }
	get := func(count int, refresh bool) {
		t.Helper()
		if _, err := c.Get(source, count, refresh, generate); err != nil {
			t.Fatal(err)
		}
	}
	get(32, false)
	get(32, false)
	if calls != 1 {
		t.Fatal(calls)
	}
	c = Cache{Root: c.Root}
	get(32, false)
	if calls != 1 {
		t.Fatal("disk cache not reused")
	}
	get(32, true)
	if calls != 2 {
		t.Fatal("refresh reused cache")
	}
	get(16, false)
	if calls != 3 {
		t.Fatal("frame count not invalidated")
	}
	os.WriteFile(source, []byte("changed video"), 0600)
	get(16, false)
	if calls != 4 {
		t.Fatal("source not invalidated")
	}
	files, _ := os.ReadDir(c.Root)
	if len(files) != 1 {
		t.Fatal("same resource uses multiple slots")
	}
}

func TestFiveMostRecentResourcesAndUnrelatedFiles(t *testing.T) {
	root := t.TempDir()
	c := Cache{Root: filepath.Join(root, "cache")}
	calls := 0
	generate := func() ([]string, error) { calls++; return []string{"image"}, nil }
	sources := []string{}
	for i := 0; i < 6; i++ {
		p := filepath.Join(root, fmt.Sprint(i))
		os.WriteFile(p, []byte("video"), 0600)
		sources = append(sources, p)
	}
	for _, p := range sources[:5] {
		c.Get(p, 32, false, generate)
	}
	files, _ := os.ReadDir(c.Root)
	past := time.Now().Add(-time.Hour)
	for _, f := range files {
		os.Chtimes(filepath.Join(c.Root, f.Name()), past, past)
	}
	c.Get(sources[0], 32, false, generate)
	os.WriteFile(filepath.Join(c.Root, "keep.txt"), []byte("unrelated"), 0600)
	c.Get(sources[5], 32, false, generate)
	files, _ = os.ReadDir(c.Root)
	if len(files) != 6 {
		t.Fatalf("want five entries and unrelated file, got %d", len(files))
	}
	c.Get(sources[0], 32, false, generate)
	if calls != 6 {
		t.Fatal("recently read entry evicted")
	}
	if _, err := os.Stat(filepath.Join(c.Root, "keep.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestFailureAndCorruption(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "video")
	os.WriteFile(source, []byte("x"), 0600)
	c := Cache{Root: filepath.Join(root, "cache")}
	calls := 0
	good := func() ([]string, error) { calls++; return []string{"old"}, nil }
	c.Get(source, 1, false, good)
	_, err := c.Get(source, 1, true, func() ([]string, error) { return nil, fmt.Errorf("ffmpeg failure") })
	if err == nil {
		t.Fatal("lost error")
	}
	c.Get(source, 1, false, good)
	if calls != 1 {
		t.Fatal("failed refresh destroyed cache")
	}
	files, _ := os.ReadDir(c.Root)
	os.WriteFile(filepath.Join(c.Root, files[0].Name()), []byte("broken"), 0600)
	c.Get(source, 1, false, good)
	if calls != 2 {
		t.Fatal("corrupt cache not regenerated")
	}
	c.Root = source // Cache write failure must not prevent returning generated images.
	images, err := c.Get(source, 1, false, good)
	if err != nil || len(images) != 1 {
		t.Fatal(images, err)
	}
}

func TestConcurrentRepeatedOpen(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "video")
	os.WriteFile(source, []byte("x"), 0600)
	c := Cache{Root: filepath.Join(root, "cache")}
	calls := 0
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Get(source, 32, false, func() ([]string, error) { calls++; return []string{"image"}, nil })
		}()
	}
	wg.Wait()
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestSettingsAndClear(t *testing.T) {
	root := t.TempDir()
	c := Cache{Root: filepath.Join(root, "cache")}
	setting, err := c.Settings()
	if err != nil || setting.MaxResources != 5 {
		t.Fatal(setting, err)
	}
	for i := 0; i < 5; i++ {
		source := filepath.Join(root, fmt.Sprint(i))
		os.WriteFile(source, []byte("video"), 0600)
		c.Get(source, 32, false, func() ([]string, error) { return []string{"image"}, nil })
	}
	if err = c.SaveSettings(Settings{MaxResources: 2}); err != nil {
		t.Fatal(err)
	}
	status, err := c.Status()
	if err != nil || status.Count != 2 {
		t.Fatal(status, err)
	}
	if err = c.SaveSettings(Settings{MaxResources: 2}); err != nil {
		t.Fatal("overwrite settings", err)
	}
	c = Cache{Root: c.Root}
	setting, err = c.Settings()
	if err != nil || setting.MaxResources != 2 {
		t.Fatal("setting not persisted", setting, err)
	}
	if c.SaveSettings(Settings{MaxResources: 0}) == nil {
		t.Fatal("accepted invalid limit")
	}
	if c.SaveSettings(Settings{MaxResources: 101}) == nil {
		t.Fatal("accepted invalid limit")
	}
	os.WriteFile(filepath.Join(c.Root, "keep.txt"), []byte("not a preview"), 0600)
	if err = c.Clear(); err != nil {
		t.Fatal(err)
	}
	status, err = c.Status()
	if err != nil || status.Count != 0 || status.Bytes != 0 {
		t.Fatal(status, err)
	}
	setting, err = c.Settings()
	if err != nil || setting.MaxResources != 2 {
		t.Fatal("clear removed settings")
	}
	if _, err = os.Stat(filepath.Join(c.Root, "keep.txt")); err != nil {
		t.Fatal("clear removed unrelated file")
	}
	for i := 0; i < 4; i++ {
		source := filepath.Join(root, fmt.Sprint(i))
		c.Get(source, 32, false, func() ([]string, error) { return []string{"image"}, nil })
	}
	status, _ = c.Status()
	if status.Count != 2 {
		t.Fatal("new generation ignored limit", status)
	}
}

func TestClearWaitsForInFlightGeneration(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "video")
	os.WriteFile(source, []byte("video"), 0600)
	c := Cache{Root: filepath.Join(root, "cache")}
	started := make(chan struct{})
	release := make(chan struct{})
	generated := make(chan struct{})
	go func() {
		defer close(generated)
		c.Get(source, 32, false, func() ([]string, error) { close(started); <-release; return []string{"image"}, nil })
	}()
	<-started
	cleared := make(chan error, 1)
	go func() { cleared <- c.Clear() }()
	close(release)
	<-generated
	if err := <-cleared; err != nil {
		t.Fatal(err)
	}
	status, err := c.Status()
	if err != nil || status.Count != 0 {
		t.Fatal("generation recreated cleared cache", status, err)
	}
}

func TestSaveLimitDuringGeneration(t *testing.T) {
	root := t.TempDir()
	c := Cache{Root: filepath.Join(root, "cache")}
	for i := 0; i < 3; i++ {
		source := filepath.Join(root, fmt.Sprint(i))
		os.WriteFile(source, []byte("video"), 0600)
		c.Get(source, 32, false, func() ([]string, error) { return []string{"image"}, nil })
	}
	source := filepath.Join(root, "new")
	os.WriteFile(source, []byte("new video"), 0600)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Get(source, 32, false, func() ([]string, error) { close(started); <-release; return []string{"image"}, nil })
	}()
	<-started
	saved := make(chan error, 1)
	go func() { saved <- c.SaveSettings(Settings{MaxResources: 1}) }()
	close(release)
	<-done
	if err := <-saved; err != nil {
		t.Fatal(err)
	}
	status, err := c.Status()
	if err != nil || status.Count != 1 {
		t.Fatal(status, err)
	}
	setting, err := c.Settings()
	if err != nil || setting.MaxResources != 1 {
		t.Fatal(setting, err)
	}
}
