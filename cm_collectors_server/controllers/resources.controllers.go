package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type Resource struct{}

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
