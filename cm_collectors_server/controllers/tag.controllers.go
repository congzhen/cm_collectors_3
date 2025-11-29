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
func (Tag) TagList_FilesBasesId(c *gin.Context) {
	filesBasesId := c.Param("filesBasesId")
	tagData, err := processors.Tag{}.TagData(filesBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(tagData.Tag, c)
}
func (Tag) TagList_TagClassId(c *gin.Context) {
	tagClassId := c.Param("tagClassId")
	tagList, err := processors.Tag{}.TagListByTagClassId(tagClassId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(tagList, c)
}
func (Tag) TagClassList(c *gin.Context) {
	filesBasesId := c.Param("filesBasesId")
	tagClassList, err := processors.TagClass{}.DataListByFilesBasesId(filesBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(tagClassList, c)
}
func (Tag) CreateTag(c *gin.Context) {
	var par datatype.ReqParam_Tag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	_, err := processors.Tag{}.Create(&par)
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

func (Tag) DeleteTag(c *gin.Context) {
	tagID := c.Param("tagId")
	err := processors.Tag{}.DeleteTag(tagID)
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
	_, err := processors.TagClass{}.Create(&par)
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
func (Tag) ImportTag(c *gin.Context) {
	var par datatype.ReqParam_ImportTag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Tag{}.ImportTag(par.FilesBasesID, par.ImportData)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
