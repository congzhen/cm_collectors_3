//go:build tray
// +build tray

package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"cm_collectors_server/core"
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

	// 检查是否有自动启动应用程序的参数
	args := os.Args[1:]
	autoOpenApp := false
	for _, arg := range args {
		if arg == "--open-app" || arg == "-o" {
			autoOpenApp = true
			break
		}
	}
	// 如果需要自动打开应用程序，则在服务器启动后执行
	if autoOpenApp && !core.Config.General.WindowsStartNotRunApp {
		go func() {
			// 等待服务器准备好接受连接
			waitForServerReady(serverAddr)
			// 自动触发打开应用程序事件
			menu.OpenAppClicked()
		}()
	}
}

// waitForServerReady 等待服务器准备好接受连接
func waitForServerReady(serverAddr string) {
	// 最多等待30秒
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Println("等待服务器启动超时")
			return
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
			if err == nil {
				conn.Close()
				fmt.Println("服务器已准备就绪")
				return
			}
		}
	}
}

// onTrayExit 系统托盘退出时的回调函数
func onTrayExit() {
	fmt.Println("系统托盘退出")
	os.Exit(0)
}
