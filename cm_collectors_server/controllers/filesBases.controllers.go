package controllers

import (
	"cm_collectors_server/datatype"
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

func (FilesBases) SetFilesBases(c *gin.Context) {
	var par datatype.ReqParam_SetFilesBases
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.FilesBases{}.SetFilesBases(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (FilesBases) Create(c *gin.Context) {
	var par datatype.ReqParam_CreateFilesBases
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	id, err := processors.FilesBases{}.Create(par.Name, par.MainPerformerBasesId, par.RelatedPerformerBasesIds)
	if err := ResError(c, err); err != nil {
		return
	}
	info, err := processors.FilesBases{}.InfoById(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (FilesBases) Sort(c *gin.Context) {
	var par datatype.ReqParam_FilesBasesSort
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.FilesBases{}.Sort(par.SortData)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (FilesBases) Config(c *gin.Context) {
	id := c.Param("id")
	configType := c.Param("configType")
	configStr, err := processors.FilesBases{}.ConfigById(id, configType)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(configStr, c)
}
