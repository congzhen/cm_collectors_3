package controllers

import (
	"cm_collectors_server/models"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AiTag struct{}

type ReqAiTagFilesBasesSettings struct {
	Items []processors.AiTagFilesBasesSetting `json:"items"`
}

func (AiTag) Setting(c *gin.Context) {
	info, err := processors.AiTag{}.Setting()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (AiTag) SaveSetting(c *gin.Context) {
	var par models.AiTagSetting
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.AiTag{}.SaveSetting(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (AiTag) FilesBases(c *gin.Context) {
	list, err := processors.AiTag{}.FilesBasesSettings()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(list, c)
}

func (AiTag) SaveFilesBases(c *gin.Context) {
	var par ReqAiTagFilesBasesSettings
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.AiTag{}.SaveFilesBasesSettings(par.Items)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (AiTag) Stats(c *gin.Context) {
	stats, err := processors.AiTag{}.Stats(c.Query("files_bases_id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(stats, c)
}

func (AiTag) Records(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := processors.AiTag{}.Records(c.Query("files_bases_id"), c.Query("status"), page, limit)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(map[string]interface{}{"dataList": list, "total": total}, c)
}

func (AiTag) RunOnce(c *gin.Context) {
	result, err := processors.AiTag{}.RunOnce(c.Query("files_bases_id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (AiTag) ResetFailed(c *gin.Context) {
	reset, err := processors.AiTag{}.ResetFailed(c.Query("files_bases_id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(map[string]int64{"reset": reset}, c)
}

func (AiTag) ResetProcessing(c *gin.Context) {
	reset, err := processors.AiTag{}.ResetProcessing(c.Query("files_bases_id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(map[string]int64{"reset": reset}, c)
}

func (AiTag) Pause(c *gin.Context) {
	info, err := processors.AiTag{}.Pause()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (AiTag) Resume(c *gin.Context) {
	info, err := processors.AiTag{}.Resume()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (AiTag) Rescan(c *gin.Context) {
	err := processors.AiTag{}.Rescan(c.Query("files_bases_id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (AiTag) TestConnection(c *gin.Context) {
	var par models.AiTagSetting
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	result, err := processors.AiTag{}.TestConnection(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (AiTag) TestService(c *gin.Context) {
	err := processors.AiTag{}.TestService()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
