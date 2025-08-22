//go:build tray
// +build tray

package main

import (
	"fmt"
	"os"

	"cm_collectors_server/tray"

	"github.com/getlantern/systray"
)

// runWithTray 启动带系统托盘的应用
func runWithTray() {
	systray.Run(onTrayReady, onTrayExit)
}

// onTrayReady 系统托盘准备就绪时的回调函数
func onTrayReady() {
	// 启动服务器
	serverAddr := startServerInBackground()

	// 创建托盘菜单
	menu := tray.CreateTrayMenu(icon, serverAddr)

	// 处理菜单事件
	menu.HandleEvents(shutdownServer)
}

// onTrayExit 系统托盘退出时的回调函数
func onTrayExit() {
	fmt.Println("系统托盘退出")
	os.Exit(0)
}
