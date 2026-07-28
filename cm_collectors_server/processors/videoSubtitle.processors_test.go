package processors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLocalizedSubtitleFileCaseInsensitiveExtension(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "Episode01.mkv")
	subtitlePath := filepath.Join(dir, "Episode01.ZH.SRT")

	if err := os.WriteFile(videoPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(subtitlePath, []byte("1\n00:00:00,000 --> 00:00:01,000\n字幕\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	gotPath, gotContentType, err := (VideoSubtitle{}).findLocalizedSubtitleFile(videoPath, "zh", "")
	if err != nil {
		t.Fatalf("findLocalizedSubtitleFile() error = %v", err)
	}
	if gotPath != subtitlePath {
		t.Fatalf("findLocalizedSubtitleFile() path = %q, want %q", gotPath, subtitlePath)
	}
	if gotContentType != "text/plain; charset=utf-8" {
		t.Fatalf("findLocalizedSubtitleFile() content type = %q", gotContentType)
	}
}

func TestFindLocalizedSubtitleFilePreservesLanguagePriority(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "Episode02.mp4")
	defaultSubtitlePath := filepath.Join(dir, "Episode02.srt")
	localizedSubtitlePath := filepath.Join(dir, "Episode02.en.vtt")

	for path, content := range map[string][]byte{
		videoPath:             nil,
		defaultSubtitlePath:   []byte("default"),
		localizedSubtitlePath: []byte("WEBVTT"),
	} {
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	gotPath, gotContentType, err := (VideoSubtitle{}).findLocalizedSubtitleFile(videoPath, "en", "")
	if err != nil {
		t.Fatalf("findLocalizedSubtitleFile() error = %v", err)
	}
	if gotPath != localizedSubtitlePath {
		t.Fatalf("findLocalizedSubtitleFile() path = %q, want %q", gotPath, localizedSubtitlePath)
	}
	if gotContentType != "text/vtt; charset=utf-8" {
		t.Fatalf("findLocalizedSubtitleFile() content type = %q", gotContentType)
	}
}
