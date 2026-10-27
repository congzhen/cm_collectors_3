//go:build windows

package selfupdate

import (
	"os"
	"path/filepath"
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
	if e := apply(p); e == nil {
		t.Fatal("locked file should prevent installation")
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
