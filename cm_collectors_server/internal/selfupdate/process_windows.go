//go:build windows

package selfupdate

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
)

func requireSpace(dir string, bytes uint64) error {
	p, e := windows.UTF16PtrFromString(dir)
	if e != nil {
		return e
	}
	var free, total, unused uint64
	if e = windows.GetDiskFreeSpaceEx(p, &free, &total, &unused); e != nil {
		return e
	}
	if free < bytes {
		return fmt.Errorf("更新磁盘空间不足：至少还需要 %d MB", bytes>>20)
	}
	return nil
}

func prepareDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x00000200}
}
func processAlive(pid int) bool {
	h, e := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if e != nil {
		return false
	}
	defer windows.CloseHandle(h)
	s, e := windows.WaitForSingleObject(h, 0)
	return e == nil && s == uint32(windows.WAIT_TIMEOUT)
}
func replaceFile(src, dst string) error {
	s, e := windows.UTF16PtrFromString(src)
	if e != nil {
		return e
	}
	d, e := windows.UTF16PtrFromString(dst)
	if e != nil {
		return e
	}
	return windows.MoveFileEx(s, d, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
