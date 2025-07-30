package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type FilesCL struct{}

func (FilesCL) FilesList_Image(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	list, err := processors.FilesCL{}.FilesListByDramaSeriesId(dramaSeriesId, processors.FilesCL_Image)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(list, c)
}

func (FilesCL) Files_Image(c *gin.Context) {
	dramaSeriesId := c.Param("dramaSeriesId")
	fileNameBase64 := c.Param("fileNameBase64")
	fileName, err := UrlDecode(fileNameBase64)
	if err != nil {
		response.FailWithMessage("Base64解码失败: "+err.Error(), c)
		return
	}
	fmt.Println("-----------------------", fileName)
	data, err := processors.FilesCL{}.FilesImage(dramaSeriesId, fileName)
	if err := ResError(c, err); err != nil {
		return
	}
	// 根据文件扩展名设置Content-Type
	contentType := "image/jpeg" // 默认类型
	ext := filepath.Ext(fileName)
	switch strings.ToLower(ext) {
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
	// 返回图片数据
	c.Data(200, contentType, data)
}
