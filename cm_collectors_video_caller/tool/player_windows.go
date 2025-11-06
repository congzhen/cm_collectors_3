//go:build windows
// +build windows

package tool

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Windows实现
func getDefaultMediaPlayer() (string, error) {
	// 首先查找.mp4文件的默认程序
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.mp4\UserChoice`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	// 获取ProgId
	progId, _, err := k.GetStringValue("ProgId")
	if err != nil {
		return "", err
	}

	// 根据ProgId查找实际的程序路径
	// 尝试在HKEY_CLASSES_ROOT中查找命令
	path := progId + `\shell\open\command`
	k, err = registry.OpenKey(registry.CLASSES_ROOT, path, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	// 获取默认值（通常是程序路径）
	command, _, err := k.GetStringValue("")
	if err != nil || command == "" {
		return "", fmt.Errorf("无法获取命令行")
	}

	// 使用正则表达式提取可执行文件路径
	// 匹配 "C:\Program Files\...\program.exe" 或 C:\Program Files\...\program.exe
	re := regexp.MustCompile(`(?i)"?([a-z]:\\[^"]*?\.exe)"?`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// 如果正则表达式没有匹配，尝试手动解析
	// 移除开头的引号（如果有的话）
	command = strings.TrimSpace(command)
	if strings.HasPrefix(command, "\"") {
		// 找到第一个引号之后到第二个引号之间的内容
		endQuote := strings.Index(command[1:], "\"")
		if endQuote > 0 {
			return command[1 : endQuote+1], nil
		}
	} else {
		// 没有引号的情况下，提取.exe之前的路径
		exeIndex := strings.Index(strings.ToLower(command), ".exe")
		if exeIndex > 0 {
			// 找到.exe之前最后一个空格作为起点
			startIndex := strings.LastIndex(command[:exeIndex], " ")
			if startIndex < 0 {
				startIndex = 0
			} else {
				startIndex++
			}
			return command[startIndex : exeIndex+4], nil
		}
	}

	return "", fmt.Errorf("无法解析命令行: %s", command)
}
