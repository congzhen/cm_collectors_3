package selfupdate

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		n    int
	}{{"3.10.0", "3.9.9", 1}, {"v3.3.0", "3.3.0", 0}, {"3.2.0", "3.3.0", -1}} {
		n, e := Compare(tt.a, tt.b)
		if e != nil || n != tt.n {
			t.Fatalf("%+v: %d %v", tt, n, e)
		}
	}
	for _, v := range []string{"", "3.x.0", "3.3.0-beta", "03.3.0", "3.3.0.1", "999999999999999999999999999.0.0"} {
		if _, e := Compare(v, "3.3.0"); e == nil {
			t.Fatalf("accepted %q", v)
		}
	}
}
func TestRejectUnsafeArchiveNames(t *testing.T) {
	for _, v := range []string{"../escape", "/absolute", "C:/file", "a\\b", "a//b", "a/../b", "a:stream", "a/NUL.txt", "a/file.", "a/file "} {
		if safeName(v) {
			t.Fatalf("accepted %q", v)
		}
	}
	for _, v := range []string{"folder/server", "folder/", "中文.txt"} {
		if !safeName(v) {
			t.Fatalf("rejected %q", v)
		}
	}
}

type zipEntry struct {
	name string
	data []byte
}

func archiveForTest(t *testing.T, job string, entries []zipEntry) Asset {
	t.Helper()
	f, e := os.Create(filepath.Join(job, "package.zip"))
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	for _, v := range entries {
		w, e := z.Create(v.name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(v.data); e != nil {
			t.Fatal(e)
		}
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(job, "package.zip"))
	if e != nil {
		t.Fatal(e)
	}
	h := sha256.Sum256(b)
	return Asset{Size: int64(len(b)), SHA256: hex.EncodeToString(h[:])}
}
func binaryForTest(t *testing.T) []byte {
	t.Helper()
	p, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestExtractRejectsTraversalAndDuplicates(t *testing.T) {
	for _, entries := range [][]zipEntry{
		{{serverName(), []byte("bad")}, {"../escape", []byte("x")}},
		{{serverName(), []byte("bad")}, {"asset.txt", []byte("x")}, {"ASSET.txt", []byte("y")}},
		{{serverName(), []byte("not an executable")}},
	} {
		dir := t.TempDir()
		archiveForTest(t, dir, entries)
		if _, e := extract(filepath.Join(dir, "package.zip"), filepath.Join(dir, "stage")); e == nil {
			t.Fatal("accepted unsafe archive")
		}
	}
}
func TestChecksumDetectsCorruption(t *testing.T) {
	dir := t.TempDir()
	a := archiveForTest(t, dir, []zipEntry{{"file", []byte("data")}})
	name := filepath.Join(dir, "package.zip")
	if e := verifyArchive(name, a); e != nil {
		t.Fatal(e)
	}
	f, _ := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	_, _ = f.Write([]byte("x"))
	f.Close()
	if verifyArchive(name, a) == nil {
		t.Fatal("corruption accepted")
	}
}
func TestApplyPreservesDataAndOverwritesPackageFiles(t *testing.T) {
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-test")
	if e := os.MkdirAll(job, 0700); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{serverName(), "config.yaml", "db/media.db", "scraper/custom.json"} {
		p := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("original"), 0644)
	}
	a := archiveForTest(t, job, []zipEntry{{"release/" + serverName(), binaryForTest(t)}, {"release/config.yaml", []byte("bad")}, {"release/db/media.db", []byte("bad")}, {"release/scraper/custom.json", []byte("bad")}, {"release/asset.txt", []byte("new")}})
	if e := apply(Plan{Root: root, Job: job, Attempt: job, Asset: a}); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"config.yaml", "db/media.db", "scraper/custom.json"} {
		b, _ := os.ReadFile(filepath.Join(root, name))
		if string(b) != "original" {
			t.Fatalf("overwrote %s", name)
		}
	}
	b, _ := os.ReadFile(filepath.Join(root, "asset.txt"))
	if string(b) != "new" {
		t.Fatal("missing new file")
	}
	b, _ = os.ReadFile(filepath.Join(job, "original", serverName()))
	if string(b) != "original" {
		t.Fatal("missing original backup")
	}
}
func TestApplyRollsBackPartialReplacement(t *testing.T) {
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-test")
	os.MkdirAll(job, 0700)
	os.WriteFile(filepath.Join(root, serverName()), []byte("old-server"), 0755)
	os.Mkdir(filepath.Join(root, "blocked"), 0755)
	a := archiveForTest(t, job, []zipEntry{{serverName(), binaryForTest(t)}, {"new.txt", []byte("new")}, {"blocked", []byte("cannot replace directory")}})
	if e := apply(Plan{Root: root, Job: job, Attempt: job, Asset: a}); e == nil {
		t.Fatal("expected replacement error")
	}
	b, _ := os.ReadFile(filepath.Join(root, serverName()))
	if string(b) != "old-server" {
		t.Fatal("old server not restored")
	}
	if _, e := os.Stat(filepath.Join(root, "new.txt")); !os.IsNotExist(e) {
		t.Fatal("new file not rolled back")
	}
}
func TestConfiguredDataDirectoryProtected(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "custom-data")
	if !protected("custom-data/a", filepath.Join(data, "a"), []string{data}) {
		t.Fatal("custom data not protected")
	}
	if protected("custom-data-2/a", filepath.Join(root, "custom-data-2", "a"), []string{data}) {
		t.Fatal("prefix confusion")
	}
	config := filepath.Join(root, "video_caller", "config.json")
	os.MkdirAll(filepath.Dir(config), 0755)
	os.WriteFile(config, []byte("user settings"), 0644)
	if !protected("video_caller/config.json", config, nil) {
		t.Fatal("nested auxiliary config not protected")
	}
}
func TestRejectNonGithubSources(t *testing.T) {
	for _, v := range []string{"http://github.com/a", "https://github.com.evil.test/a", "https://github.com:443/a", "https://user@github.com/a", "https://127.0.0.1/a"} {
		if githubURL(v, true) == nil {
			t.Fatalf("accepted %s", v)
		}
	}
}

func TestApplyPreservesBundledHelperExecutePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable permissions")
	}
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-mode")
	os.MkdirAll(job, 0700)
	helper := filepath.Join(root, "ffmpeg")
	if e := os.WriteFile(helper, []byte("old helper"), 0755); e != nil {
		t.Fatal(e)
	}
	a := archiveForTest(t, job, []zipEntry{{serverName(), binaryForTest(t)}, {"ffmpeg", []byte("updated helper")}})
	if e := apply(Plan{Root: root, Job: job, Attempt: job, Asset: a}); e != nil {
		t.Fatal(e)
	}
	st, e := os.Stat(helper)
	if e != nil {
		t.Fatal(e)
	}
	if st.Mode().Perm()&0111 != 0111 {
		t.Fatalf("lost execute permission: %v", st.Mode())
	}
}
