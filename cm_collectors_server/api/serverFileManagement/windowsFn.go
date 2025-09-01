//go:build windows
// +build windows

package serverfilemanagement

import (
	"os"
	"runtime"

	"golang.org/x/sys/windows"
)

// 检查Windows上的目录Junction点
func isWindowsDirectoryJunction(fullPath string, fileInfo os.FileInfo) bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// 使用Windows API检查是否为Junction点
	// 获取文件属性
	utf16Path, err := windows.UTF16PtrFromString(fullPath)
	if err != nil {
		return false
	}

	attributes, err := windows.GetFileAttributes(utf16Path)
	if err != nil {
		return false
	}

	// 检查是否为重解析点且是目录
	const (
		FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400
		FILE_ATTRIBUTE_DIRECTORY     = 0x00000010
	)

	isReparsePoint := (attributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0
	isDirectory := (attributes & FILE_ATTRIBUTE_DIRECTORY) != 0

	// Junction点是重解析点且是目录
	if isReparsePoint && isDirectory {
		return true
	}

	return false
}
