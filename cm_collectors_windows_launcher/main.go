package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func main() {
	// 获取当前可执行文件的路径
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("无法获取可执行文件路径: %v\n", err)
		return
	}

	// 构建wails应用的路径（与当前程序在同一目录）
	dir := filepath.Dir(execPath)

	var appPath = path.Join(dir, "CMCollectors3.exe")

	cmd := exec.Command(appPath, "-t", "-o")
	if err := cmd.Start(); err != nil {
		fmt.Printf("无法启动应用程序: %v\n", err)
	}
}
