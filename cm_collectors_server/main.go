package main

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/routers"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/gin-gonic/gin"
)

//go:embed html
var htmlFS embed.FS

//go:embed ico/favicon.ico
var icon []byte

// 全局服务器变量，用于优雅关闭
var server *http.Server

func main() {
	// 检查是否有命令行参数来决定是否启用系统托盘
	args := os.Args[1:]
	trayMode := false
	for _, arg := range args {
		if arg == "--tray" || arg == "-t" {
			trayMode = true
			break
		}
	}

	if trayMode {
		fmt.Println("以系统托盘模式运行")
		// 以系统托盘模式运行
		systray.Run(onTrayReady, onTrayExit)
	} else {
		fmt.Println("以服务器模式运行")
		// 正常模式运行
		serverInit(false)
	}
}

// onTrayReady 系统托盘准备就绪时的回调函数
func onTrayReady() {
	// 设置托盘图标
	systray.SetIcon(icon)
	systray.SetTitle("CM Collectors")
	systray.SetTooltip("CM Collectors Server")

	// 添加菜单项
	openWebUI := systray.AddMenuItem("打开网页界面", "在浏览器中打开管理界面")
	systray.AddSeparator()
	quit := systray.AddMenuItem("退出", "停止服务器并退出程序")

	// 启动服务器
	serverAddr := startServerInBackground()

	// 启动协程处理菜单事件
	go func() {
		for {
			select {
			case <-openWebUI.ClickedCh:
				// 将0.0.0.0转换为localhost并添加端口
				displayAddr := strings.Replace(serverAddr, "0.0.0.0", "localhost", 1)
				openBrowser(fmt.Sprintf("http://%s", displayAddr))
			case <-quit.ClickedCh:
				fmt.Println("正在退出...")
				// 优雅关闭服务器
				shutdownServer()
				// 退出托盘
				systray.Quit()
				// 退出程序
				os.Exit(0)
				return
			}
		}
	}()
}

// onTrayExit 系统托盘退出时的回调函数
func onTrayExit() {
	fmt.Println("系统托盘退出")
}

// startServerInBackground 在后台启动服务器
func startServerInBackground() string {
	// 在新的goroutine中启动服务器
	go serverInit(true)

	// 等待配置初始化完成，获取服务器地址
	// 使用轮询方式等待core.Config初始化完成
	for i := 0; i < 100; i++ {
		if core.Config != nil && core.Config.System.ServerHost != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	serverAddr := fmt.Sprintf("%s:%d", core.Config.System.ServerHost, core.Config.System.Port)
	fmt.Printf("服务器地址: %s\n", serverAddr)
	return serverAddr
}

// serverInit 初始化服务器配置和组件
// 该函数负责初始化核心配置、数据库、路由等服务器运行所需组件
func serverInit(trayMode bool) {
	//初始化核心
	core.Init()
	//初始化项目数据库数据库
	dbInitErr := models.DB_Init(core.DBS())
	if dbInitErr != nil {
		fmt.Println(dbInitErr)
		return
	}

	// 禁用控制台颜色
	// gin.DisableConsoleColor()
	// 设置运行模式 release | debug
	gin.SetMode(core.Config.System.Env)
	// 日志记录到文件。
	//logFile, _ := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	// 将日志写入文件和控制台
	//gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	router := gin.Default()

	//初始化路由
	router = routers.InitRouter(router, &htmlFS)

	serverAddr := fmt.Sprintf("%s:%d", core.Config.System.ServerHost, core.Config.System.Port)

	//运行http服务(方法1)
	//router.Run(serverAddr)

	//运行http服务(方法2)
	runServer(serverAddr, router, trayMode)
}

// runServer 启动HTTP服务器并监听指定地址。
// 参数:
// - serverAddr: 服务器监听的地址。
// - router: gin框架的路由器指针。
// - trayMode: 是否以系统托盘模式运行。
func runServer(serverAddr string, router *gin.Engine, trayMode bool) {
	// 创建HTTP服务器
	server = &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 启动HTTP服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	fmt.Println("Server started: ", serverAddr)

	// 如果不是托盘模式，则等待中断信号
	if !trayMode {
		// 优雅重启
		gracefulShutdown(server)
	}
}

// openBrowser 在默认浏览器中打开网页
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("无法打开浏览器: %v\n", err)
	}
}

// shutdownServer 优雅关闭服务器
func shutdownServer() {
	if server != nil {
		// 创建一个5秒的超时上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 关闭服务器
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("服务器关闭错误: %v\n", err)
		} else {
			fmt.Println("服务器已关闭")
		}
	}
}

// gracefulShutdown 处理优雅重启逻辑
// gracefulShutdown 函数用于优雅地关闭HTTP服务器。
// 它接受一个http.Server实例作为参数。
// 当接收到终止信号时，该函数将在指定的超时时间内完成对服务器的关闭操作。
func gracefulShutdown(server *http.Server) {
	// 创建一个通道来接收信号
	signalChan := make(chan os.Signal, 1)
	// 监听系统中断和终止信号
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-signalChan
	fmt.Printf("Received signal: %v\n", sig)

	// 设置5秒超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown failed: %v\n", err)
	}

	fmt.Println("Server exiting")
}
