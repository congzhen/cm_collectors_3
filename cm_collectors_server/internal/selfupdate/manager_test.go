package selfupdate

import (
	"cm_collectors_server/internal/maintenance"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func readyManager(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	store := filepath.Join(root, ".cm-updates")
	job := filepath.Join(store, "job-test")
	if e := os.MkdirAll(job, 0700); e != nil {
		t.Fatal(e)
	}
	return &Manager{root: root, store: store, state: State{Phase: "ready", Version: "999.0.0", Job: job}}
}
func TestStateDoesNotReloadStaleDiskAfterPersistenceFailure(t *testing.T) {
	m := readyManager(t)
	if e := writeJSON(filepath.Join(m.store, "state.json"), State{Phase: "downloading"}); e != nil {
		t.Fatal(e)
	}
	m.state.Phase = "failed"
	m.state.Error = "disk full"
	if s := m.State(); s.Phase != "failed" || s.Error != "disk full" {
		t.Fatalf("lost in-memory failure: %+v", s)
	}
}
func TestInstallRejectsChangedPackageAndOlderVersion(t *testing.T) {
	m := readyManager(t)
	if e := m.Install(m.state.Version, m.state.Job+"-stale"); e == nil {
		t.Fatal("accepted a different download")
	}
	for _, v := range []string{Version, "0.0.1", "invalid"} {
		m.state.Version = v
		if e := m.Install(v, m.state.Job); e == nil {
			t.Fatalf("accepted non-upgrade %s", v)
		}
		if m.busy || maintenance.Paused() {
			t.Fatal("rejected request started maintenance")
		}
	}
}
func TestInstallSaveFailurePreservesReadyPackage(t *testing.T) {
	t.Setenv("CM_UPDATE_SUPERVISED", "1")
	t.Setenv("INVOCATION_ID", "")
	t.Setenv("NOTIFY_SOCKET", "")
	m := readyManager(t)
	if e := os.Mkdir(filepath.Join(m.store, "state.json.tmp"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := m.Install(m.state.Version, m.state.Job); e == nil {
		t.Fatal("expected state write failure")
	}
	if m.State().Phase != "ready" || m.busy || m.cancel != nil {
		t.Fatal("failed save left an active installation")
	}
}
func TestDownloadSaveFailurePreservesPreviousPackage(t *testing.T) {
	m := readyManager(t)
	before := m.state
	m.cached = CheckResult{Manifest: Manifest{Version: "999.1.0"}, HasUpdate: true}
	os.Mkdir(filepath.Join(m.store, "state.json.tmp"), 0700)
	if e := m.Download("999.1.0"); e == nil {
		t.Fatal("expected state write failure")
	}
	if m.state != before || m.busy {
		t.Fatal("failed download start replaced the prepared package")
	}
}
func TestBackupFailureResumesBusinessAndReleasesContext(t *testing.T) {
	maintenance.Resume()
	defer maintenance.Resume()
	m := readyManager(t)
	m.busy = true
	m.state.Phase = "waiting"
	m.hooks.Backup = func(string) error { return errors.New("backup failed") }
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.install(ctx)
	if m.state.Phase != "ready" || m.state.Error != "backup failed" || m.busy {
		t.Fatalf("bad recovery: %+v", m.state)
	}
	if !maintenance.Begin() {
		t.Fatal("business requests still blocked")
	}
	maintenance.End()
	if ctx.Err() != context.Canceled {
		t.Fatal("installation context was not released")
	}
}
func TestFailedRollbackRemainsDetectedAfterResultWasSaved(t *testing.T) {
	m := readyManager(t)
	attempt := filepath.Join(m.state.Job, "attempt-test")
	os.Mkdir(attempt, 0700)
	p := Plan{Root: m.root, Job: m.state.Job, Attempt: attempt, PID: os.Getpid(), HealthToken: "test"}
	if e := writeJSON(filepath.Join(m.store, "plan.json"), p); e != nil {
		t.Fatal(e)
	}
	m.state.Phase = "failed"
	m.state.Attempt = attempt
	if e := m.save(); e != nil {
		t.Fatal(e)
	}
	writeJSON(filepath.Join(attempt, "journal.json"), []replacement{{Rel: serverName(), Existed: true}})
	if e := checkInstallState(m.store); e == nil {
		t.Fatal("failed phase bypassed an incomplete replacement")
	}
	os.WriteFile(filepath.Join(attempt, "files-applied"), []byte("yes"), 0600)
	os.WriteFile(filepath.Join(attempt, "restore-failed"), []byte("locked"), 0600)
	if e := checkInstallState(m.store); e == nil {
		t.Fatal("applied marker hid a failed restoration")
	}
	os.Remove(filepath.Join(attempt, "restore-failed"))
	os.WriteFile(filepath.Join(attempt, "files-restored"), []byte("yes"), 0600)
	if e := checkInstallState(m.store); e != nil {
		t.Fatal(e)
	}
}
func TestPreflightChecksArchiveWithoutExtracting(t *testing.T) {
	m := readyManager(t)
	archiveForTest(t, m.state.Job, []zipEntry{{"release/" + serverName(), []byte("metadata inspection only")}, {"release/assets/app.js", []byte("new")}})
	p := Plan{Root: m.root, Job: m.state.Job}
	if e := preflightInstall(p); e != nil {
		t.Fatal(e)
	}
	entries, e := os.ReadDir(m.state.Job)
	if e != nil {
		t.Fatal(e)
	}
	if len(entries) != 1 || entries[0].Name() != "package.zip" {
		t.Fatal("preflight extracted an unnecessary package copy")
	}
	os.WriteFile(filepath.Join(m.root, "assets"), []byte("not a directory"), 0644)
	if e := preflightInstall(p); e == nil {
		t.Fatal("preflight accepted a blocked parent path")
	}
}
