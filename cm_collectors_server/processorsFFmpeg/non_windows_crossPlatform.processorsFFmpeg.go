//go:build !windows
// +build !windows

package processorsffmpeg

import (
	"os/exec"
)

// setSysProcAttr 在非Windows平台下为空实现
func setSysProcAttr(cmd *exec.Cmd) {
	// 非Windows平台不需要特殊处理
}
