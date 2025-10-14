package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type ImportData struct {
}

func (ImportData) ScanDiskImportPaths(c *gin.Context) {
	var par datatype.ReqParam_ImportData_ScanDisk_ImportPaths
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	nonExistingSrcPaths, err := processors.ImportData{}.ScanDiskImportPaths(par.FilesBasesId, par.Config)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(nonExistingSrcPaths, c)
}
func (ImportData) ScanDiskImportData(c *gin.Context) {
	var par datatype.ReqParam_ImportData_ScanDisk_ImportData
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.ImportData{}.ScanDiskImportData(par.FilesBasesId, par.FilePath, par.Config)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (ImportData) UpdateScanDiskConfig(c *gin.Context) {
	var par datatype.ReqParam_ImportData_UpdateScanDiskConfig
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.ImportData{}.UpdateScanDiskConfig(par.FilesBasesId, par.ConfigJson)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
