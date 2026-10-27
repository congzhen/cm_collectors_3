//go:build !windows

package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func requireSpace(dir string, bytes uint64) error {
	var s syscall.Statfs_t
	if e := syscall.Statfs(dir, &s); e != nil {
		return e
	}
	if uint64(s.Bavail)*uint64(s.Bsize) < bytes {
		return fmt.Errorf("更新磁盘空间不足：至少还需要 %d MB", bytes>>20)
	}
	return nil
}

func prepareDetached(cmd *exec.Cmd)     { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func processAlive(pid int) bool         { e := syscall.Kill(pid, 0); return e == nil || e == syscall.EPERM }
func replaceFile(src, dst string) error { return os.Rename(src, dst) }
