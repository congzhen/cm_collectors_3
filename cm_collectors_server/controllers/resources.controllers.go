package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type Resource struct{}

func (Resource) Info(c *gin.Context) {
	resourceId := c.Param("resourceId")
	resourceInfo, err := processors.Resources{}.Info(resourceId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(resourceInfo, c)
}

func (Resource) DataList(c *gin.Context) {
	var par datatype.ReqParam_ResourcesList
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	dataList, total, err := processors.Resources{}.DataList(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	resDataList := datatype.ResDataList{
		DataList: dataList,
		Total:    total,
	}
	response.OkWithData(resDataList, c)
}

func (Resource) CreateResource(c *gin.Context) {
	var par datatype.ReqParam_Resource
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Resources{}.CreateResource(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Resource) UpdateResource(c *gin.Context) {
	var par datatype.ReqParam_Resource
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Resources{}.UpdateResource(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}

func (Resource) DeleteResource(c *gin.Context) {
	resourceId := c.Param("resourceId")
	err := processors.Resources{}.DeleteResource(resourceId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
