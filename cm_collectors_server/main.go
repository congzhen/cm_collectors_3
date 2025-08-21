package main

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/routers"
	"cm_collectors_server/tray"
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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

	router := gin.Default()

	//初始化路由
	router = routers.InitRouter(router, &htmlFS)

	serverAddr := fmt.Sprintf("%s:%d", core.Config.System.ServerHost, core.Config.System.Port)

	// 检查端口是否被占用，如果被占用则尝试释放
	if isPortInUse(core.Config.System.Port) {
		fmt.Printf("端口 %d 已被占用，尝试释放...\n", core.Config.System.Port)
		if killProcessUsingPort(core.Config.System.Port) {
			fmt.Printf("已尝试释放端口 %d，等待进程结束...\n", core.Config.System.Port)
			time.Sleep(2 * time.Second) // 等待进程完全结束
		} else {
			fmt.Printf("无法释放端口 %d\n", core.Config.System.Port)
		}
	}

	//运行http服务(方法1)
	//router.Run(serverAddr)

	//运行http服务(方法2)
	runServer(serverAddr, router, trayMode)
}

// isPortInUse 检查端口是否被占用
func isPortInUse(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// killProcessUsingPort 查找并终止使用指定端口的进程
func killProcessUsingPort(port int) bool {
	switch runtime.GOOS {
	case "windows":
		return killProcessOnWindows(port)
	case "linux", "darwin":
		return killProcessOnUnix(port)
	default:
		fmt.Printf("不支持的操作系统: %s\n", runtime.GOOS)
		return false
	}
}

// killProcessOnWindows 在Windows上查找并终止使用指定端口的进程
func killProcessOnWindows(port int) bool {
	// 使用netstat查找占用端口的进程
	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("执行netstat命令失败: %v\n", err)
		return false
	}

	// 解析输出查找占用端口的进程ID
	lines := string(output)
	var pid string
	for _, line := range splitLines(lines) {
		if len(line) > 0 && contains(line, fmt.Sprintf(":%d", port)) && contains(line, "LISTENING") {
			fields := splitFields(line)
			if len(fields) > 0 {
				pid = fields[len(fields)-1]
				break
			}
		}
	}

	if pid != "" {
		// 终止进程
		fmt.Printf("尝试终止占用端口 %d 的进程 (PID: %s)\n", port, pid)
		killCmd := exec.Command("taskkill", "/PID", pid, "/F")
		if err := killCmd.Run(); err != nil {
			fmt.Printf("终止进程失败: %v\n", err)
			return false
		}
		fmt.Printf("成功终止进程 PID: %s\n", pid)
		return true
	}

	fmt.Printf("未找到占用端口 %d 的进程\n", port)
	return false
}

// killProcessOnUnix 在Unix-like系统上查找并终止使用指定端口的进程
func killProcessOnUnix(port int) bool {
	// 使用lsof查找占用端口的进程
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("执行lsof命令失败: %v\n", err)
		return false
	}

	pid := string(output)
	pid = trimWhitespace(pid)

	if pid != "" {
		// 终止进程
		fmt.Printf("尝试终止占用端口 %d 的进程 (PID: %s)\n", port, pid)
		killCmd := exec.Command("kill", "-9", pid)
		if err := killCmd.Run(); err != nil {
			fmt.Printf("终止进程失败: %v\n", err)
			return false
		}
		fmt.Printf("成功终止进程 PID: %s\n", pid)
		return true
	}

	fmt.Printf("未找到占用端口 %d 的进程\n", port)
	return false
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

// 辅助函数
func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else if char != '\r' {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitFields(s string) []string {
	fields := []string{}
	current := ""
	for _, char := range s {
		if char == ' ' || char == '\t' {
			if current != "" {
				fields = append(fields, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		fields = append(fields, current)
	}
	return fields
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					index(s, substr) >= 0)))
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimWhitespace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
