package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type VideoMetadata struct{}

func (VideoMetadata) CleanupPreview(c *gin.Context) {
	data, err := (processors.VideoMetadata{}).CleanupPreview()
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(data, c)
}
func (VideoMetadata) CleanupStatus(c *gin.Context) {
	response.OkWithData((processors.VideoMetadata{}).CleanupStatus(), c)
}
func (VideoMetadata) StartCleanup(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}
	if ParameterHandleShouldBindJSON(c, &request) != nil {
		return
	}
	if err := (processors.VideoMetadata{}).StartCleanup(request.Token); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}
func (VideoMetadata) Bulk(c *gin.Context) {
	var request processors.MetadataBulkRequest
	if ParameterHandleShouldBindJSON(c, &request) != nil {
		return
	}
	data, err := (processors.VideoMetadata{}).Bulk(request)
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Setting(c *gin.Context) {
	data, err := (processors.VideoMetadata{}).Setting()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) SaveSetting(c *gin.Context) {
	var request processors.VideoMetadataSettingData
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	data, err := (processors.VideoMetadata{}).SaveSetting(&request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Stats(c *gin.Context) {
	data, err := (processors.VideoMetadata{}).Stats()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Info(c *gin.Context) {
	data, err := (processors.VideoMetadata{}).MetadataInfo(c.Param("dramaSeriesId"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Failures(c *gin.Context) {
	var request processors.VideoMetadataFailureQuery
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	data, err := (processors.VideoMetadata{}).Failures(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) RetryFailure(c *gin.Context) {
	if err := (processors.VideoMetadata{}).RetryFailure(c.Param("dramaSeriesId")); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoMetadata) SetClassification(c *gin.Context) {
	var request processors.VideoMetadataClassificationRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	if err := (processors.VideoMetadata{}).SetClassification(request); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoMetadata) SaveManual(c *gin.Context) {
	var request processors.VideoMetadataManualRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	data, err := (processors.VideoMetadata{}).SaveManual(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Run(c *gin.Context) {
	var request processors.VideoMetadataRunRequest
	if err := ParameterHandleShouldBindJSON(c, &request); err != nil {
		return
	}
	data, err := (processors.VideoMetadata{}).StartBatch(request)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) TaskStatus(c *gin.Context) {
	data, err := (processors.VideoMetadata{}).BatchStatus()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (VideoMetadata) Pause(c *gin.Context) {
	if err := (processors.VideoMetadata{}).PauseBatch(); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoMetadata) Resume(c *gin.Context) {
	if err := (processors.VideoMetadata{}).ResumeBatch(); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoMetadata) Stop(c *gin.Context) {
	if err := (processors.VideoMetadata{}).StopBatch(); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}
