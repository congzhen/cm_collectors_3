package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

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
