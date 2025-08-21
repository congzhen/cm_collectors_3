package tray

import (
	"fmt"
	"strings"

	"github.com/getlantern/systray"
)

// TrayMenu 托盘菜单结构
type TrayMenu struct {
	OpenWebUI *systray.MenuItem
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

// UpdateServerAddr 更新服务器地址
func (tm *TrayMenu) UpdateServerAddr(serverAddr string) {
	ServerAddr = serverAddr
}
