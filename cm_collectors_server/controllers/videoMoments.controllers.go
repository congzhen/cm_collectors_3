package controllers

import (
	"cm_collectors_server/models"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

type VideoMoments struct{}

func (VideoMoments) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	v, e := (processors.VideoMoments{}).List(c.Query("resourceId"), c.Query("dramaSeriesId"), page)
	if ResError(c, e) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) ListLibrary(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	items, total, err := (processors.VideoMoments{}).ListLibrary(c.Query("filesBasesId"), c.Query("keyword"), page)
	if ResError(c, err) == nil {
		response.OkWithData(gin.H{"items": items, "total": total}, c)
	}
}
func (VideoMoments) Save(c *gin.Context) {
	var input models.VideoMoment
	if ParameterHandleShouldBindJSON(c, &input) != nil {
		return
	}
	v, e := (processors.VideoMoments{}).Save(c.Param("id"), input)
	if ResError(c, e) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) Delete(c *gin.Context) {
	if ResError(c, (processors.VideoMoments{}).Delete(c.Param("id"))) == nil {
		response.OkWithData(true, c)
	}
}
func (VideoMoments) Preview(c *gin.Context) {
	v, e := (processors.VideoMoments{}).Preview(c.Param("id"))
	if ResError(c, e) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) Queue(c *gin.Context) {
	var input struct {
		Interval int `json:"interval"`
	}
	if ParameterHandleShouldBindJSON(c, &input) != nil {
		return
	}
	v, e := (processors.VideoMoments{}).Queue(c.Param("id"), input.Interval)
	if ResError(c, e) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) Image(c *gin.Context) {
	i, e := strconv.Atoi(c.Param("index"))
	if e != nil {
		c.Status(404)
		return
	}
	path, e := (processors.VideoMoments{}).Image(c.Param("id"), i)
	if e != nil {
		c.Status(404)
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.File(path)
}
func (VideoMoments) QueueStatus(c *gin.Context) {
	v, err := (processors.VideoMoments{}).QueueStatus()
	if ResError(c, err) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) QueueControl(c *gin.Context) {
	if ResError(c, (processors.VideoMoments{}).QueueControl(c.Param("action"))) == nil {
		response.OkWithData(true, c)
	}
}

func (VideoMoments) DeletePreview(c *gin.Context) {
	if ResError(c, (processors.VideoMoments{}).DeletePreview(c.Param("id"))) == nil {
		response.OkWithData(true, c)
	}
}

func (VideoMoments) Setting(c *gin.Context) {
	v, e := (processors.VideoMoments{}).Setting()
	if ResError(c, e) == nil {
		response.OkWithData(v, c)
	}
}
func (VideoMoments) SaveSetting(c *gin.Context) {
	var v models.VideoPreviewSetting
	if ParameterHandleShouldBindJSON(c, &v) != nil {
		return
	}
	if ResError(c, (processors.VideoMoments{}).SaveSetting(v)) == nil {
		response.OkWithData(true, c)
	}
}

func (VideoMoments) QueueOnPlayback(c *gin.Context) {
	queued, err := (processors.VideoMoments{}).QueueOnPlayback(c.Param("id"))
	if ResError(c, err) == nil {
		response.OkWithData(queued, c)
	}
}

func (VideoMoments) OpenCacheFolder(c *gin.Context) {
	if ResError(c, (processors.VideoMoments{}).OpenCacheFolder()) == nil {
		response.OkWithData(true, c)
	}
}

func (VideoMoments) Cover(c *gin.Context) {
	path, err := (processors.VideoMoments{}).Cover(c.Param("id"))
	if err != nil {
		c.Status(404)
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.File(path)
}
