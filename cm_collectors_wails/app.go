package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// RequestServerShutdown 向服务器请求关闭服务器
func (a *App) RequestServerShutdown() (bool, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(a.url, "/")+"/api/app/shutdown", nil)
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Status bool   `json:"status"`
		Msg    string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	if !result.Status {
		return false, fmt.Errorf(result.Msg)
	}
	return true, nil
}

// OpenMultipleFilesDialog 打开文件选择对话框
// 前端通过 window.__WAILS__.Go.Main.OpenMultipleFilesDialog(options) 调用
func (a *App) OpenMultipleFilesDialog(title, name, pattern string) ([]string, error) {
	options := wailsRuntime.OpenDialogOptions{
		Title:   title,
		Filters: []wailsRuntime.FileFilter{{DisplayName: name, Pattern: pattern}},
	}
	paths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, options)
	if err != nil {
		return nil, err
	}
	// 如果用户取消，paths 可能为空切片，返回空切片表示用户取消
	if len(paths) == 0 {
		return []string{}, nil
	}
	for i, path := range paths {
		// 标准化路径，文件夹使用/分割
		paths[i] = filepath.ToSlash(filepath.Clean(path))
	}
	return paths, nil
}
