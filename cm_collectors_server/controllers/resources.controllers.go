package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"net/http"

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

func (Resource) SampleImages(c *gin.Context) {
	resourceId := c.Param("resourceId")
	imagePath := c.Query("q")
	dataList, err := processors.Resources{}.SampleImages(resourceId, imagePath)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}
func (Resource) SampleImageData(c *gin.Context) {
	resourceId := c.Param("resourceId")
	imagePath := c.Query("q")
	if imagePath == "" {
		response.FailWithCode(5, c)
		return
	}
	ext, imageBytes, err := processors.Resources{}.SampleImageBytes(resourceId, imagePath)
	if err := ResError(c, err); err != nil {
		return
	}
	// 设置正确的Content-Type头
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".bmp":
		contentType = "image/bmp"
	case ".svg":
		contentType = "image/svg+xml"
	}
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, imageBytes)
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
	info, err := processors.Resources{}.UpdateResource(&par, true)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Resource) UpdateResourceTag(c *gin.Context) {
	var par datatype.ReqParam_ResourceTag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Resources{}.UpdateResourceTag(par.ResourceID, par.Tags)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Resource) BatchSetTag(c *gin.Context) {
	var par datatype.ReqParam_BatchSetTag
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Resources{}.BatchSetTag(par.Mode, par.ResourceIDS, par.Tags)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Resource) DeleteResource(c *gin.Context) {
	resourceId := c.Param("resourceId")
	err := processors.Resources{}.DeleteResource(resourceId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

type ResourceDramaSeries struct{}

func (ResourceDramaSeries) SearchPath(c *gin.Context) {
	var par datatype.ReqParam_ResourceDramaSeries_SearchPath
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	dataList, err := processors.ResourcesDramaSeries{}.SearchPath(par.FilesBasesIds, par.SearchPath)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}

func (ResourceDramaSeries) ReplacePath(c *gin.Context) {
	var par datatype.ReqParam_ResourceDramaSeries_ReplacePath
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	dataList, err := processors.ResourcesDramaSeries{}.ReplacePath(par.FilesBasesIds, par.SearchPath, par.ReplacePath)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}
