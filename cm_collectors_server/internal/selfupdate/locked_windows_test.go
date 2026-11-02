//go:build windows

package selfupdate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestLockedPackageFileRestoresEarlierReplacements(t *testing.T) {
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-locked")
	os.MkdirAll(job, 0700)
	os.WriteFile(filepath.Join(root, serverName()), []byte("old-server"), 0755)
	locked := filepath.Join(root, "locked.txt")
	os.WriteFile(locked, []byte("old-file"), 0644)
	ptr, _ := windows.UTF16PtrFromString(locked)
	h, e := windows.CreateFile(ptr, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CloseHandle(h)
	a := archiveForTest(t, job, []zipEntry{{serverName(), binaryForTest(t)}, {"locked.txt", []byte("new")}})
	p := Plan{Root: root, Job: job, Attempt: job, Asset: a}
	if e := apply(p); e == nil || !strings.Contains(e.Error(), "locked.txt") {
		t.Fatalf("expected a replacement error naming the locked file, got %v", e)
	}
	data, _ := os.ReadFile(filepath.Join(root, serverName()))
	if string(data) != "old-server" {
		t.Fatal("earlier replacement not restored")
	}
	data, _ = os.ReadFile(locked)
	if string(data) != "old-file" {
		t.Fatal("locked file changed")
	}
	if !canRestartOriginal(p) {
		t.Fatal("unchanged locked file should not prevent restarting original server")
	}
}

func TestUnchangedLockedClientDoesNotPreventInstallation(t *testing.T) {
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-unchanged")
	os.MkdirAll(job, 0700)
	os.WriteFile(filepath.Join(root, serverName()), []byte("old-server"), 0755)
	client := filepath.Join(root, "cm_collectors_wails.exe")
	os.WriteFile(client, []byte("unchanged client"), 0644)
	ptr, _ := windows.UTF16PtrFromString(client)
	h, e := windows.CreateFile(ptr, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CloseHandle(h)
	a := archiveForTest(t, job, []zipEntry{{serverName(), binaryForTest(t)}, {"cm_collectors_wails.exe", []byte("unchanged client")}, {"new.txt", []byte("installed")}})
	p := Plan{Root: root, Job: job, Attempt: job, Asset: a}
	if e := apply(p); e != nil {
		t.Fatalf("unchanged locked client prevented installation: %v", e)
	}
	if _, e := os.Stat(filepath.Join(job, "original", "cm_collectors_wails.exe")); !os.IsNotExist(e) {
		t.Fatal("unchanged client should not be backed up or replaced")
	}
	if data, e := os.ReadFile(filepath.Join(root, "new.txt")); e != nil || string(data) != "installed" {
		t.Fatalf("remaining package files were not installed: %s %v", data, e)
	}
}
