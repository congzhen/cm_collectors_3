package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
	url string
}

// NewApp creates a new App application struct
func NewApp() *App {
	url := parseURLFromArgs()
	return &App{
		url: url,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// parseURLFromArgs 解析命令行参数获取URL
func parseURLFromArgs() string {
	args := os.Args[1:]

	// 默认URL
	defaultURL := "http://127.0.0.1:12345"

	// 遍历参数查找 -url 参数
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-url" && i+1 < len(args) {
			return args[i+1]
		} else if strings.HasPrefix(arg, "-url=") {
			return strings.TrimPrefix(arg, "-url=")
		}
	}

	if defaultURL != "" && !strings.HasPrefix(defaultURL, "http://") && !strings.HasPrefix(defaultURL, "https://") {
		defaultURL = "http://" + defaultURL
	}

	return defaultURL
}

// GetURL returns the URL to be used in iframe
func (a *App) GetURL() string {
	return a.url
}
