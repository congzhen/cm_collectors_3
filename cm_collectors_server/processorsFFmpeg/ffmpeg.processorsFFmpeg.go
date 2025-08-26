package processorsffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type FFmpeg struct{}

// 创建一个通用的命令执行函数，自动处理Windows平台的窗口隐藏
func createCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	// 在Windows上隐藏控制台窗口
	if runtime.GOOS == "windows" {
		setSysProcAttr(cmd)
	}
	return cmd
}

// 创建一个通用的带上下文的命令执行函数，自动处理Windows平台的窗口隐藏
func createCommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	// 在Windows上隐藏控制台窗口
	if runtime.GOOS == "windows" {
		setSysProcAttr(cmd)
	}
	return cmd
}

// IsFFmpegAvailable 检查系统中FFmpeg是否可用
//
// 该函数会根据不同的操作系统查找ffmpeg可执行文件：
//   - 在Windows系统中，首先检查程序根目录下的ffmpeg文件夹，
//     如果未找到，则检查系统PATH环境变量中的ffmpeg.exe或ffmpeg
//   - 在Unix-like系统中，直接检查系统PATH环境变量中的ffmpeg
//
// 返回值:
//
//	string: 找到的ffmpeg可执行文件的完整路径
//	error: 如果未找到ffmpeg或无法执行，则返回错误信息
func (f FFmpeg) IsFFmpegAvailable() (string, error) {
	return f.isToolAvailable("ffmpeg")
}

// IsFFprobeAvailable 检查系统中FFprobe是否可用
//
// 该函数会根据不同的操作系统查找ffprobe可执行文件：
//   - 在Windows系统中，首先检查程序根目录下的ffprobe文件夹，
//     如果未找到，则检查系统PATH环境变量中的ffprobe.exe或ffprobe
//   - 在Unix-like系统中，直接检查系统PATH环境变量中的ffprobe
//
// 返回值:
//
//	string: 找到的ffprobe可执行文件的完整路径
//	error: 如果未找到ffprobe或无法执行，则返回错误信息
func (f FFmpeg) IsFFprobeAvailable() (string, error) {
	return f.isToolAvailable("ffprobe")
}

// 检查指定工具是否可用的通用方法
func (f FFmpeg) isToolAvailable(toolName string) (string, error) {
	var cmd *exec.Cmd
	var toolPath string

	// 根据操作系统选择命令
	switch runtime.GOOS {
	case "windows":
		// Windows系统优先检查当前程序根目录的工具文件夹
		localToolPath := filepath.Join("./", "ffmpeg", toolName+".exe")
		if _, err := os.Stat(localToolPath); err == nil {
			// 本地工具存在
			toolPath = localToolPath
			cmd = createCommand(toolPath, "-version")
			break
		}

		// 如果本地没有，检查系统PATH中的可执行文件
		toolPath, err := exec.LookPath(toolName + ".exe")
		if err != nil {
			// 尝试不带.exe后缀的版本
			toolPath, err = exec.LookPath(toolName)
			if err != nil {
				return "", fmt.Errorf("在Windows系统中未找到%s: %v", toolName, err)
			}
		}
		cmd = createCommand(toolPath, "-version")
	default:
		// Unix-like系统 (Linux, macOS等)
		var err error
		toolPath, err = exec.LookPath(toolName)
		if err != nil {
			return "", fmt.Errorf("在Unix-like系统中未找到%s: %v", toolName, err)
		}
		cmd = createCommand(toolPath, "-version")
	}

	// 尝试运行命令
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s无法执行: %v", toolName, err)
	}

	return toolPath, nil
}
