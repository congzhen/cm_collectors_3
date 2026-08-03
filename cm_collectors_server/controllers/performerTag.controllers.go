package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PerformerTag struct{}

func (PerformerTag) Data(c *gin.Context) {
	includeDisabled, _ := strconv.ParseBool(c.DefaultQuery("includeDisabled", "false"))
	data, err := processors.PerformerTag{}.Data(c.Param("performerBasesId"), includeDisabled)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (PerformerTag) CreateClass(c *gin.Context) {
	var par datatype.ReqParam_PerformerTagClass
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	data, err := processors.PerformerTag{}.CreateClass(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (PerformerTag) UpdateClass(c *gin.Context) {
	var par datatype.ReqParam_PerformerTagClass
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.PerformerTag{}.UpdateClass(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (PerformerTag) DeleteClass(c *gin.Context) {
	err := processors.PerformerTag{}.DeleteClass(c.Param("id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (PerformerTag) CreateTag(c *gin.Context) {
	var par datatype.ReqParam_PerformerTag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	data, err := processors.PerformerTag{}.CreateTag(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(data, c)
}

func (PerformerTag) UpdateTag(c *gin.Context) {
	var par datatype.ReqParam_PerformerTag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.PerformerTag{}.UpdateTag(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (PerformerTag) DeleteTag(c *gin.Context) {
	err := processors.PerformerTag{}.DeleteTag(c.Param("id"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (PerformerTag) UpdateSort(c *gin.Context) {
	var par datatype.ReqParam_PerformerTagSort
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.PerformerTag{}.UpdateSort(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
