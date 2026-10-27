package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

const UpdateExitCode = 75

type Plan struct {
	Root, Job, Workdir, Executable, Version, HealthURL string
	HealthToken                                        string
	Attempt                                            string
	Args, Protected                                    []string
	PID                                                int
	Asset                                              Asset
}
type replacement struct {
	Rel     string
	Existed bool
}

func loadPlan(name string) (Plan, error) {
	var p Plan
	if e := readJSON(name, &p); e != nil {
		return p, e
	}
	if !filepath.IsAbs(p.Root) || !within(filepath.Join(p.Root, ".cm-updates"), p.Job) || p.Job == filepath.Join(p.Root, ".cm-updates") || filepath.Dir(name) != filepath.Join(p.Root, ".cm-updates") || p.PID <= 0 || p.HealthToken == "" || !within(p.Job, p.Attempt) || p.Attempt == p.Job {
		return p, fmt.Errorf("无效的安装计划")
	}
	return p, nil
}
func setInstallResult(p Plan, phase string, err error) {
	name := filepath.Join(p.Root, ".cm-updates", "state.json")
	var s State
	if readJSON(name, &s) != nil || s.Attempt != p.Attempt {
		return
	}
	s.Phase = phase
	s.Error = ""
	if err != nil {
		s.Error = err.Error()
	}
	if e := writeJSON(name, s); e != nil {
		fmt.Fprintln(os.Stderr, "保存更新结果失败:", e)
	}
}

// HandleMode runs before any database initialization or port handling.
func HandleMode() bool {
	if len(os.Args) >= 3 && os.Args[1] == "--apply-update" {
		p, e := loadPlan(os.Args[2])
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			return true
		}
		e = os.WriteFile(filepath.Join(p.Attempt, "helper-ready"), []byte("ready"), 0600)
		if e == nil {
			deadline := time.Now().Add(2 * time.Minute)
			for processAlive(p.PID) && time.Now().Before(deadline) {
				time.Sleep(200 * time.Millisecond)
			}
			if processAlive(p.PID) {
				e = fmt.Errorf("旧进程未退出，未替换文件")
			}
		}
		attempted := false
		if e == nil {
			if _, e = os.Stat(filepath.Join(p.Attempt, "commit")); e == nil {
				attempted = true
				e = apply(p)
			}
		}
		if e != nil {
			setInstallResult(p, "failed", e)
			fmt.Fprintln(os.Stderr, e)
			if attempted && canRestartOriginal(p) {
				if cmd, startErr := startUpdated(p); startErr == nil {
					_ = cmd.Process.Release()
				} else {
					fmt.Fprintln(os.Stderr, startErr)
				}
			}
			return true
		}
		cmd, e := startUpdated(p)
		if e != nil {
			cause := e
			if restoreErr := restoreBeforeStart(p); restoreErr != nil {
				cause = fmt.Errorf("启动失败 (%v)，恢复失败: %w", e, restoreErr)
			} else {
				if old, startErr := startUpdated(p); startErr == nil {
					_ = old.Process.Release()
				} else {
					cause = fmt.Errorf("新旧程序均无法启动: %v / %w", e, startErr)
				}
			}
			setInstallResult(p, "failed", cause)
			return true
		}
		e = waitHealthy(p)
		_ = cmd.Process.Release()
		if e != nil {
			setInstallResult(p, "failed", fmt.Errorf("启动验证失败，备份已保留，请查看更新器日志: %w", e))
		} else {
			setInstallResult(p, "installed", nil)
		}
		return true
	}
	if len(os.Args) >= 2 && os.Args[1] == "--supervise" {
		os.Exit(supervise())
		return true
	}
	// Never migrate a database using a package left half-replaced by power loss.
	if e := checkInterruptedInstall(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	return false
}
func checkInterruptedInstall() error {
	exe, e := os.Executable()
	if e != nil {
		return e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return e
	}
	store := filepath.Join(filepath.Dir(exe), ".cm-updates")
	return checkInstallState(store)
}
func checkInstallState(store string) error {
	var s State
	if readJSON(filepath.Join(store, "state.json"), &s) != nil || (s.Phase != "installing" && s.Phase != "failed") || s.Attempt == "" {
		return nil
	}
	p, e := loadPlan(filepath.Join(store, "plan.json"))
	if e != nil {
		return e
	}
	if p.Attempt != s.Attempt {
		return nil
	}
	if _, e := os.Stat(filepath.Join(p.Attempt, "restore-failed")); e == nil {
		return fmt.Errorf("上次更新文件恢复未完成，请根据 %s 中的 journal.json 和 original 目录恢复文件", p.Attempt)
	}
	if _, e := os.Stat(filepath.Join(p.Attempt, "journal.json")); e != nil {
		return nil
	}
	if _, e := os.Stat(filepath.Join(p.Attempt, "files-applied")); e == nil {
		return nil
	}
	if _, e := os.Stat(filepath.Join(p.Attempt, "files-restored")); e == nil {
		return nil
	}
	return fmt.Errorf("上次更新在替换文件时中断，已停止启动以保护数据库；请根据 %s 中的 journal.json 和 original 目录恢复文件", p.Attempt)
}
func apply(p Plan) error {
	if e := verifyArchive(filepath.Join(p.Job, "package.zip"), p.Asset); e != nil {
		return e
	}
	stage, e := os.MkdirTemp(p.Job, "install-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	files, e := extract(filepath.Join(p.Job, "package.zip"), stage)
	if e != nil {
		return e
	}
	changes := []replacement{}
	rollback := func(cause error) error {
		if err := restoreChanges(p, changes); err != nil {
			return fmt.Errorf("安装失败 (%v)，文件恢复失败: %w", cause, err)
		}
		return cause
	}

	for _, rel := range files {
		dst := filepath.Join(p.Root, rel)
		if protected(rel, dst, p.Protected) {
			continue
		}
		dst, e = safeDestination(p.Root, rel)
		if e != nil {
			return rollback(e)
		}
		st, e := os.Stat(dst)
		exists := e == nil
		if e != nil && !os.IsNotExist(e) {
			return rollback(e)
		}
		if exists {
			if !st.Mode().IsRegular() {
				return rollback(fmt.Errorf("目标不是普通文件: %s", rel))
			}
			if e := copyFile(dst, filepath.Join(p.Attempt, "original", rel), st.Mode().Perm()); e != nil {
				return rollback(e)
			}
			// Windows-created ZIPs often omit Unix execute bits on bundled helpers.
			if runtime.GOOS != "windows" {
				source := filepath.Join(stage, rel)
				staged, err := os.Stat(source)
				if err != nil {
					return rollback(err)
				}
				if err = os.Chmod(source, staged.Mode().Perm()|(st.Mode().Perm()&0111)); err != nil {
					return rollback(err)
				}
			}
		}
		if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return rollback(e)
		}
		changes = append(changes, replacement{rel, exists})
		if e := writeJSON(filepath.Join(p.Attempt, "journal.json"), changes); e != nil {
			changes = changes[:len(changes)-1]
			return rollback(e)
		}
		if e := replaceFile(filepath.Join(stage, rel), dst); e != nil {
			changes = changes[:len(changes)-1]
			return rollback(e)
		}
	}
	if e := os.WriteFile(filepath.Join(p.Attempt, "files-applied"), []byte("applied"), 0600); e != nil {
		return rollback(e)
	}
	return nil
}
func restoreChanges(p Plan, changes []replacement) error {
	var first error
	for i := len(changes) - 1; i >= 0; i-- {
		v := changes[i]
		dst, e := safeDestination(p.Root, v.Rel)
		if e == nil {
			if v.Existed {
				e = replaceFile(filepath.Join(p.Attempt, "original", v.Rel), dst)
			} else {
				e = os.Remove(dst)
				if os.IsNotExist(e) {
					e = nil
				}
			}
		}
		if e != nil && first == nil {
			first = fmt.Errorf("恢复 %s: %w", v.Rel, e)
		}
	}
	if first != nil {
		_ = os.WriteFile(filepath.Join(p.Attempt, "restore-failed"), []byte(first.Error()), 0600)
		return first
	}
	return os.WriteFile(filepath.Join(p.Attempt, "files-restored"), []byte("restored"), 0600)
}
func restoreBeforeStart(p Plan) error {
	var changes []replacement
	if e := readJSON(filepath.Join(p.Attempt, "journal.json"), &changes); e != nil {
		return e
	}
	return restoreChanges(p, changes)
}
func startUpdated(p Plan) (*exec.Cmd, error) {
	cmd := exec.Command(p.Executable, p.Args...)
	cmd.Dir = p.Workdir
	cmd.Env = append(os.Environ(), "CM_UPDATE_TOKEN="+p.HealthToken)
	prepareDetached(cmd)
	log, e := os.OpenFile(filepath.Join(p.Job, "new-server.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return nil, e
	}
	defer log.Close()
	cmd.Stdout = log
	cmd.Stderr = log
	if e = cmd.Start(); e != nil {
		return nil, e
	}
	return cmd, nil
}
func waitHealthy(p Plan) error {
	c := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		r, e := c.Get(p.HealthURL)
		if e == nil {
			var h struct {
				Status  string `json:"status"`
				Version string `json:"version"`
				Token   string `json:"updateToken"`
			}
			e = json.NewDecoder(r.Body).Decode(&h)
			r.Body.Close()
			if e == nil && r.StatusCode == 200 && h.Status == "ok" && h.Token == p.HealthToken {
				if n, e := Compare(h.Version, p.Version); e == nil && n == 0 {
					return nil
				}
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("新版本未在两分钟内通过版本及数据库健康检查；不自动回退可能已迁移的数据库")
}

// The immutable supervisor lives outside /app; /app is a persistent application volume.
func canRestartOriginal(p Plan) bool {
	_, e := os.Stat(filepath.Join(p.Attempt, "restore-failed"))
	return os.IsNotExist(e)
}
func supervise() int {
	root := "/app"
	exe := filepath.Join(root, serverName())
	if _, e := os.Stat(exe); os.IsNotExist(e) {
		if e := copyFile("/opt/cm-seed/cm_collectors_server", exe, 0755); e != nil {
			fmt.Fprintln(os.Stderr, e)
			return 1
		}
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	var pending *Plan
	for {
		cmd := exec.Command(exe)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), "CM_UPDATE_SUPERVISED=1", "CM_CONTAINER=1")
		if pending != nil {
			cmd.Env = append(cmd.Env, "CM_UPDATE_TOKEN="+pending.HealthToken)
		}
		if e := cmd.Start(); e != nil {
			fmt.Fprintln(os.Stderr, e)
			if pending != nil {
				p := *pending
				pending = nil
				if restoreErr := restoreBeforeStart(p); restoreErr == nil {
					setInstallResult(p, "failed", e)
					continue
				} else {
					setInstallResult(p, "failed", restoreErr)
				}
			}
			return 1
		}
		if pending != nil {
			p := *pending
			go func() {
				e := waitHealthy(p)
				if e != nil {
					setInstallResult(p, "failed", e)
				} else {
					setInstallResult(p, "installed", nil)
				}
			}()
			pending = nil
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case sig := <-signals:
			_ = cmd.Process.Signal(sig)
			<-done
			return 0
		case <-done:
			if cmd.ProcessState.ExitCode() != UpdateExitCode {
				return cmd.ProcessState.ExitCode()
			}
		}
		p, e := loadPlan(filepath.Join(root, ".cm-updates", "plan.json"))
		if e == nil {
			_, e = os.Stat(filepath.Join(p.Attempt, "commit"))
		}
		if e == nil {
			e = apply(p)
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			if p.Root != "" {
				setInstallResult(p, "failed", e)
			}
			if p.Root != "" && canRestartOriginal(p) {
				continue
			}
			return 1
		}
		pending = &p
	}
}
