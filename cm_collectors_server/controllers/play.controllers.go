package controllers

import (
	"cm_collectors_server/processors"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/response"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Play struct{}

func (Play) PlayOpenResource(c *gin.Context) {
	resourceId := c.Param("resourceId")
	dramaSeriesId := c.Query("dramaSeriesId")
	err := processors.Play{}.PlayOpenResource(resourceId, dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Play) PlayOpenResourceFolder(c *gin.Context) {
	resourceId := c.Param("resourceId")
	err := processors.Play{}.PlayOpenResourceFolder(resourceId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Play) PlayVideoMP4(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	processors.Video{}.VideoMP4Stream(c, dramaSeriesId)
}

func (Play) VideoSubtitle(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	processors.VideoSubtitle{}.GetVideoSubtitle(c, dramaSeriesId)
}

func (Play) VideoM3u8(c *gin.Context) {
	// 创建一个具有超时的上下文，设置为30分钟
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
	defer cancel()

	// 将超时上下文应用到gin上下文中
	c.Request = c.Request.WithContext(ctx)

	dramaSeriesId := c.Param("dramaSeriesId")
	m3u8Bytes, err := processors.Video{}.GetVideoM3u8(dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	c.Data(200, "application/x-mpegURL", m3u8Bytes)
}

func (Play) PlayVideoHLS(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	startStr := c.Param("start")
	durationStr := c.Param("duration")

	start, err := strconv.ParseFloat(startStr, 32)
	if err != nil {
		ResError(c, fmt.Errorf("无效的开始时间: %v", err))
		return
	}

	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		ResError(c, fmt.Errorf("无效的持续时间: %v", err))
		return
	}

	// 设置正确的MPEG-TS响应头
	c.Header("Content-Type", "video/MP2T")

	// 获取视频源
	dramaSeries, err := processors.ResourcesDramaSeries{}.Info(dramaSeriesId)
	if err != nil {
		ResError(c, fmt.Errorf("无法获取视频信息: %v", err))
		return
	}

	// 使用FFmpeg生成视频片段
	ffmpegPath, err := processorsffmpeg.FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		ResError(c, fmt.Errorf("FFmpeg不可用: %v", err))
		return
	}

	// 构建FFmpeg命令参数
	cmd := exec.Command(
		ffmpegPath,
		"-ss", fmt.Sprintf("%.3f", start),
		"-i", dramaSeries.Src,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264", // MPEG-2视频编码，TVBox支持良好
		"-c:a", "aac", // MP2音频编码，TVBox兼容性好
		"-f", "mpegts",
		"-preset", "ultrafast", // 编码速度优化
		"-crf", "23", // 视频质量控制
		"-b:a", "128k", // 音频比特率
		"-b:v", "2000k", // 视频比特率
		"-b:a", "128k", // 音频比特率
		"-threads", "0", // 使用所有可用线程
		"pipe:1",
	)
	/*
		cmd := exec.Command(
			ffmpegPath,
			"-ss", fmt.Sprintf("%.3f", start),
			"-i", dramaSeries.Src,
			"-t", fmt.Sprintf("%.3f", duration),
			"-c:v", "mpeg2video", // MPEG-2视频编码，TVBox支持良好
			"-c:a", "mp2", // MP2音频编码，TVBox兼容性好
			"-f", "mpegts",
			"-b:v", "2000k", // 视频比特率
			"-b:a", "128k", // 音频比特率
			"pipe:1",
		)
	*/

	// 直接将输出连接到响应
	cmd.Stdout = c.Writer

	// 执行命令
	if err := cmd.Run(); err != nil {
		// 忽略客户端断开连接的错误
		if !strings.Contains(err.Error(), "broken pipe") &&
			!strings.Contains(err.Error(), "连接被对方关闭") &&
			!strings.Contains(err.Error(), "The pipe has been ended") {
			fmt.Printf("FFmpeg执行错误: %v\n", err)
		}
	}
}
