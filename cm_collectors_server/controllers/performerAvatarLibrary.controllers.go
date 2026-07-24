package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PerformerAvatarLibrary struct{}

func (PerformerAvatarLibrary) Status(c *gin.Context) {
	response.OkWithData(processors.PerformerAvatarLibrary{}.Status(), c)
}

func (PerformerAvatarLibrary) UpdateDataFile(c *gin.Context) {
	status, err := processors.PerformerAvatarLibrary{}.UpdateDataFile()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(status, c)
}

func (PerformerAvatarLibrary) ClearImageCache(c *gin.Context) {
	result, err := processors.PerformerAvatarLibrary{}.ClearImageCache()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (PerformerAvatarLibrary) Candidates(c *gin.Context) {
	performerID := c.Param("performerId")
	candidates, err := processors.PerformerAvatarLibrary{}.Candidates(performerID)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(candidates, c)
}

func (PerformerAvatarLibrary) BatchActors(c *gin.Context) {
	var page, limit int
	GetUrlParameter_Param(c, "page", &page)
	GetUrlParameter_Param(c, "limit", &limit)
	actors, err := processors.PerformerAvatarLibrary{}.BatchActors(
		c.Param("performerBasesId"), page, limit, c.Query("search"), c.Query("photoFilter"),
	)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(actors, c)
}

func (PerformerAvatarLibrary) BatchActorIDs(c *gin.Context) {
	ids, err := processors.PerformerAvatarLibrary{}.BatchActorIDs(
		c.Param("performerBasesId"), c.Query("search"), c.Query("photoFilter"),
	)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(ids, c)
}

func (PerformerAvatarLibrary) PreviewImage(c *gin.Context) {
	data, contentType, err := processors.PerformerAvatarLibrary{}.CandidateImage(c.Param("performerId"), c.Param("candidateId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadGateway, response.Response{
			Status:     false,
			StatusCode: response.Error,
			Msg:        err.Error(),
			Data:       false,
		})
		return
	}
	c.Header("Cache-Control", "private, max-age=86400")
	c.Data(200, contentType, data)
}

func (PerformerAvatarLibrary) Apply(c *gin.Context) {
	var par datatype.ReqParam_PerformerAvatarApply
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := (processors.PerformerAvatarLibrary{}).Apply(par.PerformerID, par.CandidateID, par.Overwrite)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (PerformerAvatarLibrary) BatchPreview(c *gin.Context) {
	var par datatype.ReqParam_PerformerAvatarBatch
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	result, err := processors.PerformerAvatarLibrary{}.BatchPreview(par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (PerformerAvatarLibrary) BatchApply(c *gin.Context) {
	var par datatype.ReqParam_PerformerAvatarBatch
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	result, err := processors.PerformerAvatarLibrary{}.StartBatchApply(par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(result, c)
}

func (PerformerAvatarLibrary) BatchProgress(c *gin.Context) {
	progress, ok := (processors.PerformerAvatarLibrary{}).BatchProgress(c.Param("batchId"))
	if !ok {
		response.FailWithMessage("批量头像任务不存在或结果已过期", c)
		return
	}
	response.OkWithData(progress, c)
}
