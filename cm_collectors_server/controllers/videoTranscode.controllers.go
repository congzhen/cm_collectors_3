package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type VideoTranscode struct{}

func (VideoTranscode) List(c *gin.Context) {
	data, err := (processors.VideoTranscode{}).List()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoTranscode) Capabilities(c *gin.Context) {
	response.OkWithData((processors.VideoTranscode{}).Capabilities(), c)
}

func (VideoTranscode) Add(c *gin.Context) {
	var request processors.VideoTranscodeAddRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	data, err := (processors.VideoTranscode{}).Add(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoTranscode) UpdateConfig(c *gin.Context) {
	var request processors.VideoTranscodeUpdateConfigRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	if err := (processors.VideoTranscode{}).UpdateConfig(request); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoTranscode) Start(c *gin.Context) {
	var request processors.VideoTranscodeIDsRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	if err := (processors.VideoTranscode{}).Start(request); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoTranscode) ResetBatch(c *gin.Context) {
	var request processors.VideoTranscodeIDsRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	result, err := (processors.VideoTranscode{}).ResetBatch(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (VideoTranscode) Pause(c *gin.Context) {
	(processors.VideoTranscode{}).Pause()
	response.OkWithData(true, c)
}

func (VideoTranscode) Resume(c *gin.Context) {
	(processors.VideoTranscode{}).Resume()
	response.OkWithData(true, c)
}

func (VideoTranscode) QueueStatus(c *gin.Context) {
	response.OkWithData((processors.VideoTranscode{}).QueueStatus(), c)
}

func (VideoTranscode) Cancel(c *gin.Context) {
	if err := (processors.VideoTranscode{}).Cancel(c.Param("id")); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoTranscode) Delete(c *gin.Context) {
	if err := (processors.VideoTranscode{}).Delete(c.Param("id")); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoTranscode) DeleteBatch(c *gin.Context) {
	var request processors.VideoTranscodeIDsRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	count, err := (processors.VideoTranscode{}).DeleteBatch(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(count, c)
}
