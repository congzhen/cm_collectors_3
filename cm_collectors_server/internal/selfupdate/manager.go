package selfupdate

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"cm_collectors_server/internal/maintenance"
)

type State struct {
	Attempt    string `json:"attempt,omitempty"`
	Phase      string `json:"phase"`
	Version    string `json:"version"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Error      string `json:"error,omitempty"`
	Backup     string `json:"backup,omitempty"`
	Job        string `json:"job,omitempty"`
	Asset      Asset  `json:"asset"`
}
type Hooks struct {
	ConfigURL func() string
	Backup    func(string) error
	Protected func() []string
	Shutdown  func()
	HealthURL string
}
type Manager struct {
	mu          sync.Mutex
	checkMu     sync.Mutex
	root, store string
	state       State
	busy        bool
	cancel      context.CancelFunc
	hooks       Hooks
	cached      CheckResult
	checked     time.Time
	checkErr    error
}

var Default *Manager

func New(h Hooks) (*Manager, error) {
	exe, e := os.Executable()
	if e != nil {
		return nil, e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return nil, e
	}
	root := filepath.Dir(exe)
	m := &Manager{root: root, store: filepath.Join(root, ".cm-updates"), hooks: h, state: State{Phase: "idle"}}
	if e := os.MkdirAll(m.store, 0700); e != nil {
		return nil, e
	}
	if e := readJSON(filepath.Join(m.store, "state.json"), &m.state); e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	if m.state.Phase == "downloading" || m.state.Phase == "waiting" || m.state.Phase == "backingUp" {
		m.state.Phase = "failed"
		m.state.Error = "上次更新准备被中断，请重新下载"
		if e := m.save(); e != nil {
			return nil, e
		}
	}
	if m.state.Phase == "installing" && os.Getenv("CM_UPDATE_TOKEN") != filepath.Base(m.state.Job) {
		m.state.Phase = "failed"
		m.state.Error = "上次安装未完成，请检查更新日志及备份后重新下载"
		if e := m.save(); e != nil {
			return nil, e
		}
	}
	return m, nil
}
func (m *Manager) save() error { return writeJSON(filepath.Join(m.store, "state.json"), m.state) }
func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Only the external updater writes this phase. Other states are owned in memory.
	if !m.busy && m.state.Phase == "installing" {
		var s State
		if readJSON(filepath.Join(m.store, "state.json"), &s) == nil {
			m.state = s
		}
	}
	return m.state
}
func (m *Manager) Check(ctx context.Context, manual bool) (CheckResult, error) {
	m.checkMu.Lock()
	defer m.checkMu.Unlock()
	ttl := 15 * time.Minute
	if m.checkErr != nil {
		ttl = time.Minute
	}
	if manual {
		ttl = 5 * time.Second
	}
	if !m.checked.IsZero() && time.Since(m.checked) < ttl {
		return m.cached, m.checkErr
	}
	m.cached, m.checkErr = Check(ctx, m.hooks.ConfigURL())
	m.checked = time.Now()
	return m.cached, m.checkErr
}
func (m *Manager) Download(version string) error {
	m.checkMu.Lock()
	info := m.cached
	m.checkMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy || m.state.Phase == "installing" {
		return fmt.Errorf("已有更新任务正在运行")
	}
	if !info.HasUpdate || info.Version != version {
		return fmt.Errorf("请先重新检测版本")
	}
	job, e := os.MkdirTemp(m.store, "job-")
	if e != nil {
		return e
	}
	previous := m.state
	m.state = State{Phase: "downloading", Version: version, Job: job}
	if e = m.save(); e != nil {
		m.state = previous
		_ = os.Remove(job)
		return e
	}
	m.busy = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		a, e := Resolve(ctx, info.Manifest)
		if e == nil {
			m.mu.Lock()
			m.state.Asset = a
			m.state.Total = a.Size
			m.mu.Unlock()
			e = m.download(ctx, a, job)
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		m.busy = false
		if e != nil {
			m.state.Phase = "failed"
			m.state.Error = e.Error()
		} else {
			m.state.Phase = "ready"
		}
		if e := m.save(); e != nil {
			m.state.Phase = "failed"
			m.state.Error = e.Error()
		}
	}()
	return nil
}

type progressWriter struct {
	m *Manager
	w io.Writer
}

func (p progressWriter) Write(b []byte) (int, error) {
	n, e := p.w.Write(b)
	p.m.mu.Lock()
	p.m.state.Downloaded += int64(n)
	p.m.mu.Unlock()
	return n, e
}
func (m *Manager) download(ctx context.Context, a Asset, job string) error {
	if e := requireSpace(job, uint64(a.Size)+(64<<20)); e != nil {
		return e
	}
	r, e := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if e != nil {
		return e
	}
	r.Header.Set("User-Agent", "CM-Collectors-Updater")
	resp, e := client(30 * time.Minute).Do(r)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	name := filepath.Join(job, "package.zip")
	f, e := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = io.Copy(progressWriter{m, f}, io.LimitReader(resp.Body, a.Size+1))
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	if e = verifyArchive(name, a); e != nil {
		return e
	}
	_, e = extract(name, filepath.Join(job, "checked"))
	_ = os.RemoveAll(filepath.Join(job, "checked"))
	return e
}
func (m *Manager) Install(version, job string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy || m.state.Phase != "ready" {
		return fmt.Errorf("更新包尚未准备好")
	}
	if version != m.state.Version || job != m.state.Job {
		return fmt.Errorf("待安装的更新包已变化，请重新打开更新窗口确认")
	}
	if n, e := Compare(m.state.Version, Version); e != nil || n <= 0 {
		return fmt.Errorf("待安装版本不高于当前版本，请重新检测更新")
	}
	if !within(m.store, m.state.Job) || m.state.Job == m.store {
		return fmt.Errorf("无效的更新暂存目录")
	}
	if os.Getenv("INVOCATION_ID") != "" || os.Getenv("NOTIFY_SOCKET") != "" {
		return fmt.Errorf("当前由 systemd 管理，请通过服务管理器更新；自动安装支持解压直接运行及新版 Docker 镜像")
	}
	if _, e := os.Stat("/.dockerenv"); e == nil && os.Getenv("CM_UPDATE_SUPERVISED") != "1" {
		return fmt.Errorf("请先升级到带更新启动器的 Docker 镜像")
	}
	ctx, cancel := context.WithCancel(context.Background())
	previous := m.state
	m.cancel = cancel
	m.busy = true
	m.state.Phase = "waiting"
	m.state.Error = ""
	if e := m.save(); e != nil {
		m.state = previous
		m.busy = false
		m.cancel = nil
		cancel()
		return e
	}
	go func() { time.Sleep(300 * time.Millisecond); m.install(ctx) }()
	return nil
}
func (m *Manager) CancelWait() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state.Phase != "waiting" || m.cancel == nil {
		return fmt.Errorf("当前阶段无法取消")
	}
	m.cancel()
	return nil
}
func (m *Manager) install(ctx context.Context) {
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		defer cancel()
	}
	e := maintenance.Drain(ctx)
	if e == nil {
		m.mu.Lock()
		e = ctx.Err()
		if e == nil {
			m.state.Phase = "backingUp"
			e = m.save()
		}
		m.mu.Unlock()
	}
	if e == nil {
		backup := filepath.Join(m.state.Job, "backup")
		e = os.MkdirAll(backup, 0700)
		if e == nil {
			e = m.hooks.Backup(backup)
		}
		m.mu.Lock()
		m.state.Backup = backup
		m.mu.Unlock()
	}
	if e == nil {
		e = m.prepareInstall()
	}
	if e == nil {
		e = os.WriteFile(filepath.Join(m.state.Attempt, "commit"), []byte("install"), 0600)
	}
	if e != nil {
		maintenance.Resume()
		m.mu.Lock()
		defer m.mu.Unlock()
		m.busy = false
		m.cancel = nil
		m.state.Phase = "ready"
		m.state.Error = e.Error()
		if e == context.Canceled {
			m.state.Error = ""
		}
		_ = m.save()
		return
	}
	m.hooks.Shutdown()
}
func (m *Manager) prepareInstall() error {
	wd, e := os.Getwd()
	if e != nil {
		return e
	}
	exe, e := os.Executable()
	if e != nil {
		return e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return e
	}
	if filepath.Base(exe) != serverName() {
		return fmt.Errorf("自动安装需要使用发行包中的标准程序名 %s", serverName())
	}
	p := Plan{Root: m.root, Job: m.state.Job, Workdir: wd, Executable: exe, Args: os.Args[1:], PID: os.Getpid(), Version: m.state.Version, Asset: m.state.Asset, HealthURL: m.hooks.HealthURL, Protected: m.hooks.Protected(), HealthToken: filepath.Base(m.state.Job)}
	if e := verifyArchive(filepath.Join(p.Job, "package.zip"), p.Asset); e != nil {
		return e
	}
	if e := preflightInstall(p); e != nil {
		return e
	}

	p.Attempt, e = os.MkdirTemp(p.Job, "attempt-")
	if e != nil {
		return e
	}
	m.mu.Lock()
	m.state.Attempt = p.Attempt
	m.mu.Unlock()
	if e := writeJSON(filepath.Join(m.store, "plan.json"), p); e != nil {
		return e
	}
	m.mu.Lock()
	m.state.Phase = "installing"
	e = m.save()
	m.mu.Unlock()
	if e != nil {
		return e
	}
	if os.Getenv("CM_UPDATE_SUPERVISED") == "1" {
		return nil
	}
	helper := filepath.Join(p.Attempt, "updater")
	if filepath.Ext(exe) == ".exe" {
		helper += ".exe"
	}
	if e := copyFile(exe, helper, 0700); e != nil {
		return e
	}
	cmd := exec.Command(helper, "--apply-update", filepath.Join(m.store, "plan.json"))
	cmd.Dir = wd
	prepareDetached(cmd)
	log, e := os.OpenFile(filepath.Join(p.Attempt, "updater.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer log.Close()
	cmd.Stdout = log
	cmd.Stderr = log
	if e := cmd.Start(); e != nil {
		return e
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, e := os.Stat(filepath.Join(p.Attempt, "helper-ready")); e == nil {
			_ = cmd.Process.Release()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	return fmt.Errorf("更新器未能启动，已保留当前程序")
}

// Inspect the already verified archive instead of extracting a third full copy.
func preflightInstall(p Plan) error {
	r, e := zip.OpenReader(filepath.Join(p.Job, "package.zip"))
	if e != nil {
		return e
	}
	defer r.Close()
	entries, total, e := inspectArchive(r)
	if e != nil {
		return e
	}
	needed := total + (64 << 20)
	checkedDirs := map[string]bool{}
	for _, entry := range entries {
		rel := entry.rel
		target := filepath.Join(p.Root, rel)
		if protected(rel, target, p.Protected) {
			if rel == serverName() {
				return fmt.Errorf("程序目录与数据保护目录重叠，请分开数据目录后重试")
			}
			continue
		}
		target, e = safeDestination(p.Root, rel)
		if e != nil {
			return e
		}
		if st, e := os.Stat(target); e == nil {
			if !st.Mode().IsRegular() {
				return fmt.Errorf("目标不是普通文件: %s", rel)
			}
			needed += uint64(st.Size())
		} else if !os.IsNotExist(e) {
			return e
		}
		dir := filepath.Dir(target)
		for {
			if st, e := os.Stat(dir); e == nil {
				if !st.IsDir() {
					return fmt.Errorf("目标父路径不是目录: %s", dir)
				}
				break
			} else if !os.IsNotExist(e) {
				return e
			}
			next := filepath.Dir(dir)
			if next == dir {
				return fmt.Errorf("无法检查安装目录")
			}
			dir = next
		}
		if checkedDirs[dir] {
			continue
		}
		f, e := os.CreateTemp(dir, ".cm-write-test-")
		if e != nil {
			return fmt.Errorf("安装目录不可写: %w", e)
		}
		name := f.Name()
		ce := f.Close()
		re := os.Remove(name)
		if ce != nil {
			return ce
		}
		if re != nil {
			return re
		}
		checkedDirs[dir] = true
	}
	if os.Getenv("CM_UPDATE_SUPERVISED") != "1" {
		if st, e := os.Stat(p.Executable); e == nil {
			needed += uint64(st.Size())
		}
	}
	return requireSpace(p.Root, needed)
}
