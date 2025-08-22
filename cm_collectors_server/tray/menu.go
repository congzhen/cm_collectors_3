package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/getlantern/systray"
)

// TrayMenu 托盘菜单结构
type TrayMenu struct {
	OpenWebUI *systray.MenuItem
	OpenApp   *systray.MenuItem
	Quit      *systray.MenuItem
}

// ServerAddr 服务器地址
var ServerAddr string

// CreateTrayMenu 创建托盘菜单
func CreateTrayMenu(iconData []byte, serverAddr string) *TrayMenu {
	// 设置托盘图标和信息
	systray.SetIcon(iconData)
	systray.SetTitle("CM Collectors")
	systray.SetTooltip("CM Collectors Server")

	// 保存服务器地址
	ServerAddr = serverAddr

	// 创建菜单项
	menu := &TrayMenu{
		OpenWebUI: systray.AddMenuItem("打开网页界面", "在浏览器中打开管理界面"),
		OpenApp:   systray.AddMenuItem("打开应用程序", "启动桌面应用程序"),
		Quit:      systray.AddMenuItem("退出", "停止服务器并退出程序"),
	}

	systray.AddSeparator()

	return menu
}

// HandleEvents 处理托盘菜单事件
func (tm *TrayMenu) HandleEvents(shutdownFunc func()) {
	go func() {
		for {
			select {
			case <-tm.OpenWebUI.ClickedCh:
				// 将0.0.0.0转换为localhost并添加端口
				displayAddr := strings.Replace(ServerAddr, "0.0.0.0", "localhost", 1)
				OpenBrowser(fmt.Sprintf("http://%s", displayAddr))
			case <-tm.OpenApp.ClickedCh:
				tm.OpenAppClicked()
			case <-tm.Quit.ClickedCh:
				fmt.Println("正在退出...")
				// 调用关闭函数
				if shutdownFunc != nil {
					shutdownFunc()
				}
				// 发出退出信号
				systray.Quit()
				return
			}
		}
	}()
}

// OpenAppClicked 处理打开应用程序的逻辑
func (tm *TrayMenu) OpenAppClicked() {
	// 获取当前可执行文件的路径
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("无法获取可执行文件路径: %v\n", err)
		return
	}

	// 构建wails应用的路径（与当前程序在同一目录）
	dir := filepath.Dir(execPath)
	var wailsApp string
	if runtime.GOOS == "windows" {
		wailsApp = filepath.Join(dir, "cm_collectors_wails.exe")
	} else {
		wailsApp = filepath.Join(dir, "cm_collectors_wails")
	}

	// 将0.0.0.0转换为127.0.0.1
	appServerAddr := strings.Replace(ServerAddr, "0.0.0.0", "127.0.0.1", 1)
	// 确保URL有http://前缀
	if !strings.HasPrefix(appServerAddr, "http://") && !strings.HasPrefix(appServerAddr, "https://") {
		appServerAddr = "http://" + appServerAddr
	}

	// 启动wails应用并传递服务器地址参数
	cmd := exec.Command(wailsApp, fmt.Sprintf("-url=%s", appServerAddr))
	if err := cmd.Start(); err != nil {
		fmt.Printf("无法启动应用程序: %v\n", err)
	} else {
		fmt.Println("应用程序已启动，服务器地址:", appServerAddr)
	}
}

// UpdateServerAddr 更新服务器地址
func (tm *TrayMenu) UpdateServerAddr(serverAddr string) {
	ServerAddr = serverAddr
}
