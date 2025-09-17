package main

import (
	"cm_collectors_video_caller/tool"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var playerPath string
	// 读取同目录下的 config.json 文件
	config, err := tool.ReadConfig("config.json")

	// 检查是否应该使用配置的播放器路径
	useConfigPlayer := err == nil && config.PlayerPath != ""
	if useConfigPlayer {
		// 检查配置中的播放器路径是否存在
		if _, statErr := os.Stat(config.PlayerPath); os.IsNotExist(statErr) {
			useConfigPlayer = false
		}
	}

	if useConfigPlayer {
		playerPath = config.PlayerPath
		fmt.Println("使用自定义播放器路径:", playerPath)
	} else {
		playerPath, err = tool.GetDefaultMediaPlayer()
		if err != nil {
			log.Fatalf("无法获取默认播放器路径: %v", err)
		}
		fmt.Println("使用系统默认播放器:", playerPath)
	}

	// 最后检查playerPath是否存在
	if _, err := os.Stat(playerPath); os.IsNotExist(err) {
		log.Fatalf("播放器路径不存在: %v", err)
	}

	// 从命令行参数获取视频播放地址
	args := os.Args[1:]
	if len(args) > 0 {
		videoPath := args[0]
		prefix := "cmcollectorsvideoplay://"
		videoPath = videoPath[len(prefix)+1:]
		// 修复链接中可能缺少的冒号
		if strings.HasPrefix(videoPath, "http//") {
			videoPath = "http://" + videoPath[len("http//"):]
		} else if strings.HasPrefix(videoPath, "https//") {
			videoPath = "https://" + videoPath[len("https//"):]
		}

		fmt.Println("播放视频:", videoPath)
		cmd := exec.Command(playerPath, videoPath)
		err = cmd.Start()
		if err != nil {
			fmt.Printf("无法启动播放器: %v\n", err)
			return
		}
		fmt.Println("已启动播放器，请等待视频播放完成...")
	}
}
