package cmscraper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger 日志记录器结构
type Logger struct {
	filePath   string          // 日志文件路径
	file       *os.File        // 日志文件句柄
	enabled    bool            // 是否启用日志
	buffer     strings.Builder // 日志缓冲区
	bufferLock sync.Mutex      // 缓冲区锁，保护缓冲区的并发访问
}

// 全局日志实例
var defaultLogger *Logger

// 初始化默认日志记录器
func init() {
	defaultLogger = &Logger{
		filePath: "scraper.log",
		enabled:  false,
	}
}

// SetLogEnabled 设置日志是否启用
func (l *Logger) SetLogEnabled(enabled bool) {
	l.enabled = enabled
}

// SetLogFilePath 设置日志文件路径
func (l *Logger) SetLogFilePath(path string) error {
	l.filePath = path

	// 确保日志目录存在
	dir := filepath.Dir(l.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return nil
}

// openFile 打开日志文件
func (l *Logger) openFile() error {
	if l.file != nil {
		return nil
	}

	// 确保目录存在
	dir := filepath.Dir(l.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 以追加模式打开文件
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	return nil
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	// 写入缓冲区中剩余的日志
	l.Flush()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// log 添加日志到缓冲区
func (l *Logger) log(level string, format string, args ...interface{}) {
	if !l.enabled {
		return
	}

	// 保护缓冲区的并发访问
	l.bufferLock.Lock()
	defer l.bufferLock.Unlock()

	// 格式化日志消息
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// 添加到缓冲区
	l.buffer.WriteString(logEntry)

	// 同时输出到控制台
	fmt.Print(logEntry)
}

// Flush 将缓冲区中的日志统一写入文件
func (l *Logger) Flush() error {
	if !l.enabled {
		// 清空缓冲区
		l.bufferLock.Lock()
		l.buffer.Reset()
		l.bufferLock.Unlock()
		return nil
	}

	l.bufferLock.Lock()
	defer l.bufferLock.Unlock()

	// 如果缓冲区为空，直接返回
	if l.buffer.Len() == 0 {
		return nil
	}

	// 打开日志文件（如果尚未打开）
	if err := l.openFile(); err != nil {
		fmt.Printf("无法打开日志文件: %v\n", err)
		// 清空缓冲区
		l.buffer.Reset()
		return err
	}

	// 写入文件
	if l.file != nil {
		_, err := l.file.WriteString(l.buffer.String())
		if err != nil {
			fmt.Printf("写入日志文件失败: %v\n", err)
			// 清空缓冲区
			l.buffer.Reset()
			return err
		}
		l.file.Sync() // 确保写入磁盘
	}

	// 清空缓冲区
	l.buffer.Reset()

	return nil
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

// 全局日志函数

// SetGlobalLogEnabled 设置全局日志启用状态
func SetGlobalLogEnabled(enabled bool) {
	defaultLogger.SetLogEnabled(enabled)
}

// SetGlobalLogFilePath 设置全局日志文件路径
func SetGlobalLogFilePath(path string) error {
	return defaultLogger.SetLogFilePath(path)
}

// LogInfo 记录全局信息日志
func LogInfo(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// LogError 记录全局错误日志
func LogError(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// LogDebug 记录全局调试日志
func LogDebug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// LogWarn 记录全局警告日志
func LogWarn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// FlushGlobalLogger 刷新全局日志记录器，将缓冲区内容写入文件
func FlushGlobalLogger() error {
	return defaultLogger.Flush()
}

// CloseGlobalLogger 关闭全局日志记录器
func CloseGlobalLogger() error {
	return defaultLogger.Close()
}
