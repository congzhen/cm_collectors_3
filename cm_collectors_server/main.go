package main

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/routers"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	//初始化核心
	core.Init()
	//初始化项目数据库数据库
	dbInitErr := models.AutoDatabase(core.DBS())
	if dbInitErr != nil {
		fmt.Println(dbInitErr)
		return
	}

	// 禁用控制台颜色
	// gin.DisableConsoleColor()
	// 设置运行模式 release | debug
	gin.SetMode(core.Config.System.Env)
	// 日志记录到文件。
	logFile, _ := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	// 将日志写入文件和控制台
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	router := gin.Default()

	//初始化路由
	router = routers.InitRouter(router)

	serverAddr := fmt.Sprintf("%s:%d", core.Config.System.ServerHost, core.Config.System.Port)

	//运行http服务(方法1)
	//router.Run(serverAddr)

	//运行http服务(方法2)
	runServer(serverAddr, router, 10)
}

// runServer 启动HTTP服务器并监听指定地址。
// 参数:
// - serverAddr: 服务器监听的地址。
// - router: gin框架的路由器指针。
// - timeout: 服务器关闭时的超时时间。
func runServer(serverAddr string, router *gin.Engine, timeout time.Duration) {
	// 创建HTTP服务器
	server := &http.Server{
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
	// 优雅重启
	gracefulShutdown(server, timeout)
}

// gracefulShutdown 处理优雅重启逻辑
// gracefulShutdown 函数用于优雅地关闭HTTP服务器。
// 它接受一个http.Server实例和一个超时时间作为参数。
// 当接收到终止信号时，该函数将在指定的超时时间内完成对服务器的关闭操作。
func gracefulShutdown(server *http.Server, timeout time.Duration) {
	// 创建一个通道来接收信号
	signalChan := make(chan os.Signal, 1)
	// 监听系统中断和终止信号
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-signalChan
	fmt.Printf("Received signal: %v\n", sig)

	// 设置超时时间
	timeoutContext, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(timeoutContext); err != nil {
		fmt.Printf("Server shutdown failed: %v\n", err)
	}

	fmt.Println("Server exiting")
}
