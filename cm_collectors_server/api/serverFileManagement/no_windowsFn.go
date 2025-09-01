//go:build !windows
// +build !windows

package serverfilemanagement

import "os"

// isWindowsDirectoryJunction 在非Windows平台上始终返回false
func isWindowsDirectoryJunction(fullPath string, fileInfo os.FileInfo) bool {
	// 非Windows平台没有Junction点概念
	return false
}
