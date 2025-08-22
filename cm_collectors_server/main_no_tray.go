//go:build !tray
// +build !tray

package main

import (
	"fmt"
)

// runWithTray 在不支持托盘的构建中以普通模式运行
func runWithTray() {
	fmt.Println("系统托盘功能未启用，以普通模式运行")
	serverInit(false)
}
