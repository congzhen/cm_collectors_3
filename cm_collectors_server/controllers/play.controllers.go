package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Play struct{}

func (Play) PlayUpdate(c *gin.Context) {
	resourceId := c.Param("resourceId")
	dramaSeriesId := c.Query("dramaSeriesId")
	err := processors.Play{}.PlayUpdate(resourceId, dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
func (Play) PlayVideoInfo(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	info, err := processors.Play{}.PlayVideoInfo(dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Play) PlayOpenResource(c *gin.Context) {
	resourceId := c.Param("resourceId")
	dramaSeriesId := c.Query("dramaSeriesId")
	err := processors.Play{}.PlayOpenResource(resourceId, dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
func (Play) PlayOpenDramaSeries(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	err := processors.Play{}.PlayOpenDramaSeries(dramaSeriesId)
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
	needEncoding := true
	if c.Query("playCloud") == "true" { // 获取参数
		needEncoding = false
	}
	processors.Video{}.VideoMP4Stream(c, dramaSeriesId, needEncoding)
}

func (Play) VideoSubtitle(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	if err := (processors.VideoSubtitle{}).GetVideoSubtitle(c, dramaSeriesId); err != nil {
		logrus.Errorf("Video subtitle failed: %v; Api:%s", err, c.Request.URL.String())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to load subtitle"})
	}
}

func (Play) VideoM3u8(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	// 原有静态m3u8文件处理
	m3u8Bytes, err := processors.Video{}.GetVideoM3u8(dramaSeriesId)
	if err := ResError(c, err); err != nil {
		return
	}
	// 设置正确的CORS头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Accept, Range")

	c.Data(200, "application/x-mpegURL", m3u8Bytes)
}

func (Play) VideoM3u8StreamHLS(c *gin.Context) {
	// 设置正确的CORS头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Accept, Range")

	dramaSeriesId := c.Param("dramaSeriesId")
	var start float64
	GetUrlParameter_Param(c, "start", &start)
	var duration float64
	GetUrlParameter_Param(c, "duration", &duration)
	fmt.Printf("处理HLS视频流请求: dramaSeriesId=%s, start=%f, duration=%f\n", dramaSeriesId, start, duration)
	// Media readers use a child context which is cancelled even on normal
	// cleanup. Only the original client context represents a disconnected client.
	requestContext := c.Request.Context()
	err := processors.Video{}.VideoM3u8StreamHLS(c, dramaSeriesId, start, duration)
	finishVideoM3u8Stream(c, requestContext, err)
}

func finishVideoM3u8Stream(c *gin.Context, requestContext context.Context, err error) {
	if requestContext.Err() != nil || errors.Is(err, context.Canceled) {
		return
	}
	if err != nil && c.Writer.Written() {
		// Never append a JSON error to an already streaming MPEG-TS response.
		logrus.Errorf("HLS stream failed: %v; Api:%s", err, c.Request.URL.String())
		c.Abort()
		return
	}
	if err != nil {
		c.Writer.Header().Del("Content-Type")
	}
	if err := ResError(c, err); err != nil {
		fmt.Printf("处理HLS视频流时出错: %v\n", err)
		return
	}
	return
}
