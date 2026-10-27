package selfupdate

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if HandleMode() {
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// These subprocesses only run inside temporary test directories, never the application.
func TestUpdaterChild(t *testing.T) {
	mode := os.Getenv("CM_TEST_UPDATE_CHILD")
	if mode == "" {
		return
	}
	if mode == "old" {
		for i := 0; i < 200; i++ {
			if _, e := os.Stat("stop-old"); e == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		return
	}
	if mode == "new" {
		os.WriteFile("new.pid", []byte(strconv.Itoa(os.Getpid())), 0600)
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": Version, "updateToken": os.Getenv("CM_UPDATE_TOKEN")})
		})
		_ = http.ListenAndServe(os.Getenv("CM_TEST_UPDATE_ADDR"), nil)
	}
}
func waitTestFile(t *testing.T, name string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, e := os.Stat(name); e == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", name)
}
func TestUpdaterSubprocessReplacesAndRestarts(t *testing.T) {
	root := t.TempDir()
	job := filepath.Join(root, ".cm-updates", "job-integration")
	attempt := filepath.Join(job, "attempt")
	os.MkdirAll(attempt, 0700)
	current, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	target := filepath.Join(root, serverName())
	if e := copyFile(current, target, 0755); e != nil {
		t.Fatal(e)
	}
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	addr := listener.Addr().String()
	listener.Close()
	old := exec.Command(current, "-test.run=^TestUpdaterChild$")
	old.Dir = root
	old.Env = append(os.Environ(), "CM_TEST_UPDATE_CHILD=old")
	if e := old.Start(); e != nil {
		t.Fatal(e)
	}
	defer old.Process.Kill()
	a := archiveForTest(t, job, []zipEntry{{serverName(), binaryForTest(t)}, {"release-marker.txt", []byte("installed")}})
	p := Plan{Root: root, Job: job, Attempt: attempt, Workdir: root, Executable: target, Version: Version, PID: old.Process.Pid, Asset: a, HealthToken: "test-update", HealthURL: "http://" + addr + "/health", Args: []string{"-test.run=^TestUpdaterChild$"}}
	planFile := filepath.Join(root, ".cm-updates", "plan.json")
	if e := writeJSON(planFile, p); e != nil {
		t.Fatal(e)
	}
	if e := writeJSON(filepath.Join(root, ".cm-updates", "state.json"), State{Phase: "installing", Version: Version, Job: job, Attempt: attempt}); e != nil {
		t.Fatal(e)
	}
	updater := exec.Command(current, "--apply-update", planFile)
	updater.Dir = root
	updater.Env = append(os.Environ(), "CM_TEST_UPDATE_CHILD=new", "CM_TEST_UPDATE_ADDR="+addr)
	if e := updater.Start(); e != nil {
		t.Fatal(e)
	}
	defer updater.Process.Kill()
	defer func() {
		if b, e := os.ReadFile(filepath.Join(root, "new.pid")); e == nil {
			pid, _ := strconv.Atoi(string(b))
			if p, e := os.FindProcess(pid); e == nil {
				_ = p.Kill()
				_, _ = p.Wait()
			}
		}
	}()
	waitTestFile(t, filepath.Join(attempt, "helper-ready"))
	if _, e := os.Stat(filepath.Join(root, "release-marker.txt")); !os.IsNotExist(e) {
		t.Fatal("installed before old process exited")
	}
	os.WriteFile(filepath.Join(attempt, "commit"), []byte("yes"), 0600)
	os.WriteFile(filepath.Join(root, "stop-old"), []byte("yes"), 0600)
	if e := old.Wait(); e != nil {
		t.Fatal(e)
	}
	done := make(chan error, 1)
	go func() { done <- updater.Wait() }()
	select {
	case e := <-done:
		if e != nil {
			t.Fatal(e)
		}
	case <-time.After(60 * time.Second):
		log, _ := os.ReadFile(filepath.Join(job, "new-server.log"))
		t.Fatalf("updater did not finish; new server log: %s", log)
	}
	var state State
	if e := readJSON(filepath.Join(root, ".cm-updates", "state.json"), &state); e != nil {
		t.Fatal(e)
	}
	if state.Phase != "installed" {
		t.Fatalf("unexpected state: %+v", state)
	}
	if _, e := os.Stat(filepath.Join(root, "release-marker.txt")); e != nil {
		t.Fatal(e)
	}
}
