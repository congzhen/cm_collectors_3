package controllers

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	"cm_collectors_server/response"
	"github.com/gin-gonic/gin"
)

type ResourceGroups struct{}
type resourceGroupMutation struct {
	FilesBasesID string   `json:"filesBasesId" binding:"required,max=20"`
	GroupID      string   `json:"groupId" binding:"max=20"`
	ResourceIDs  []string `json:"resourceIds" binding:"max=1000,dive,required,max=20"`
}

func (ResourceGroups) Info(c *gin.Context) {
	data, err := (models.ResourceGroups{}).Info(core.DBS(), c.Param("filesBasesId"), c.Param("groupId"))
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(data, c)
}
func (ResourceGroups) Summaries(c *gin.Context) {
	var par struct {
		FilesBasesID string   `json:"filesBasesId" binding:"required,max=20"`
		GroupIDs     []string `json:"groupIds" binding:"max=1000,dive,required,max=20"`
		Search       string   `json:"search" binding:"max=200"`
		Page         int      `json:"page"`
		Limit        int      `json:"limit"`
	}
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	data, total, err := (models.ResourceGroups{}).Summaries(core.DBS(), par.FilesBasesID, par.GroupIDs, par.Search, par.Page, par.Limit)
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(gin.H{"dataList": data, "total": total}, c)
}
func (ResourceGroups) Mutate(c *gin.Context) {
	var par resourceGroupMutation
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	var err error
	id := par.GroupID
	switch c.Param("action") {
	case "create":
		id, err = (models.ResourceGroups{}).Create(core.DBS(), par.FilesBasesID, par.ResourceIDs)
	case "add":
		err = (models.ResourceGroups{}).Add(core.DBS(), par.FilesBasesID, id, par.ResourceIDs)
	case "remove":
		err = (models.ResourceGroups{}).Remove(core.DBS(), par.FilesBasesID, id, par.ResourceIDs)
	case "reorder":
		err = (models.ResourceGroups{}).Reorder(core.DBS(), par.FilesBasesID, id, par.ResourceIDs)
	case "dissolve":
		err = (models.ResourceGroups{}).Dissolve(core.DBS(), par.FilesBasesID, id)
	default:
		c.AbortWithStatus(404)
		return
	}
	if ResError(c, err) != nil {
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}
func (ResourceGroups) SetDisplay(c *gin.Context) {
	var par struct {
		FilesBasesID string `json:"filesBasesId" binding:"required,max=20"`
		Enabled      *bool  `json:"enabled" binding:"required"`
	}
	if ParameterHandleShouldBindJSON(c, &par) != nil {
		return
	}
	if err := models.SetResourceGroupDisplay(core.DBS(), par.FilesBasesID, *par.Enabled); ResError(c, err) != nil {
		return
	}
	response.OkWithData(true, c)
}
