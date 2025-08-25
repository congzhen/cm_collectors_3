package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"context"
	"fmt"
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

	var start float64
	GetUrlParameter_Param(c, "start", &start)
	var duration float64
	GetUrlParameter_Param(c, "duration", &duration)
	cmd, err := processors.Video{}.PlayVideoM3u8(dramaSeriesId, start, duration)
	if err := ResError(c, err); err != nil {
		return
	}
	// 设置正确的MPEG-TS响应头
	c.Header("Content-Type", "video/MP2T")

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
