package controllers

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"github.com/gin-gonic/gin"
)

type CategoryBrowse struct{}

func (CategoryBrowse) Summaries(c *gin.Context) {
	var par models.CategoryBrowseRequest
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	data, err := (models.CategoryBrowse{}).Summaries(core.DBS(), par)
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(data, c)
}
func (CategoryBrowse) Members(c *gin.Context) {
	var par models.CategoryBrowseRequest
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	data, total, err := (models.CategoryBrowse{}).Members(core.DBS(), par)
	if ResError(c, err) != nil {
		return
	}
	processors.VideoMetadata{}.TriggerForResources(data)
	response.OkWithData(gin.H{"dataList": data, "total": total}, c)
}
func (CategoryBrowse) Settings(c *gin.Context) {
	var par struct {
		FilesBasesID string `json:"filesBasesId" binding:"required,max=20"`
		models.CategoryBrowseSettings
	}
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	if ResError(c, models.SetCategoryBrowseSettings(core.DBS(), par.FilesBasesID, par.CategoryBrowseSettings)) != nil {
		return
	}
	response.OkWithData(true, c)
}
