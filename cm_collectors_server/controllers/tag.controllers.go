package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type Tag struct{}

func (Tag) TagData(c *gin.Context) {
	filesBasesId := c.Param("filesBasesId")
	tagData, err := processors.Tag{}.TagData(filesBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(tagData, c)
}
