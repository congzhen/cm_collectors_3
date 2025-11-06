//go:build linux
// +build linux

package tool

import (
	"fmt"
	"os/exec"
	"strings"
)

// Linux实现
func getDefaultMediaPlayer() (string, error) {
	// 尝试使用xdg-mime获取默认视频播放器
	cmd := exec.Command("xdg-mime", "query", "default", "video/mp4")
	output, err := cmd.Output()
	if err != nil {
		// 尝试其他视频格式
		cmd = exec.Command("xdg-mime", "query", "default", "video/x-msvideo")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("无法查询默认视频播放器: %v", err)
		}
	}

	desktopFile := strings.TrimSpace(string(output))

	// 尝试从desktop文件获取执行路径
	cmd = exec.Command("grep", "Exec=", "/usr/share/applications/"+desktopFile)
	if output, err = cmd.Output(); err != nil {
		cmd = exec.Command("grep", "Exec=", "$HOME/.local/share/applications/"+desktopFile)
		output, err = cmd.Output()
		if err != nil {
			// 如果无法找到desktop文件，尝试直接使用常见的播放器
			return findCommonPlayer()
		}
	}

	execLine := strings.TrimSpace(string(output))
	// 简单解析Exec行，提取可执行文件路径
	if strings.HasPrefix(execLine, "Exec=") {
		execCmd := strings.Split(execLine[5:], " ")[0]
		// 处理可能的环境变量
		if strings.Contains(execCmd, "/") {
			return execCmd, nil
		}
		// 查找可执行文件路径
		pathCmd := exec.Command("which", execCmd)
		if pathOutput, err := pathCmd.Output(); err == nil {
			return strings.TrimSpace(string(pathOutput)), nil
		}
	}

	return findCommonPlayer()
}

// 查找常见的视频播放器
func findCommonPlayer() (string, error) {
	commonPlayers := []string{
		"vlc",
		"mpv",
		"mplayer",
		"totem",
	}

	for _, player := range commonPlayers {
		cmd := exec.Command("which", player)
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	return "", fmt.Errorf("未找到常见的视频播放器")
}
