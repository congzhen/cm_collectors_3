//go:build windows
// +build windows

package processorsffmpeg

import (
	"os/exec"
	"syscall"
)

// 为了确保跨平台编译，使用syscall包

// setSysProcAttr 设置Windows平台下隐藏窗口的属性
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
