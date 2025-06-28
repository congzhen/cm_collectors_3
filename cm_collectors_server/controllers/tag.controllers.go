package controllers

import (
	"cm_collectors_server/datatype"
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
func (Tag) CreateTag(c *gin.Context) {
	var par datatype.ReqParam_Tag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Tag{}.Create(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
func (Tag) UpdateTag(c *gin.Context) {
	var par datatype.ReqParam_Tag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Tag{}.Update(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Tag) CreateTagClass(c *gin.Context) {
	var par datatype.ReqParam_TagClass
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.TagClass{}.Create(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Tag) UpdateTagClass(c *gin.Context) {
	var par datatype.ReqParam_TagClass
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.TagClass{}.Update(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Tag) UpdateSort(c *gin.Context) {
	var par datatype.ReqParam_UpdateTagDataSort
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Tag{}.TagDataUpdateSort(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
