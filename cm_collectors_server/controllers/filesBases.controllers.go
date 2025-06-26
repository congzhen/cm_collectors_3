package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type FilesBases struct{}

func (FilesBases) InfoDetails(c *gin.Context) {
	id := c.Param("id")
	info, err := processors.FilesBases{}.InfoDetailsById(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
