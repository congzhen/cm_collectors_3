package controllers

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/internal/importbatch"
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
	nonExistingSrcPaths, err := processors.ImportData{}.ScanDiskImportPaths(par.FilesBasesId, par.Config, true)
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
	if par.BatchID != "" && !importbatch.Global.Touch(par.BatchID, par.FilesBasesId) {
		response.FailWithCodeMsg(409, "导入批次已过期，请重新开始", c)
		return
	}
	warnings, err := processors.ImportData{}.ScanDiskImportDataResult(c.Request.Context(), par.FilesBasesId, par.FilePath, par.Config)
	if err := ResError(c, err); err != nil {
		return
	}
	if par.BatchID == "" {
		response.OkWithData(true, c)
	} else {
		response.OkWithData(gin.H{"warnings": warnings}, c)
	}
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

func (ImportData) Batch(c *gin.Context) {
	var par struct {
		Action  string `json:"action"`
		BatchID string `json:"batchId"`
		Library string `json:"filesBases_id"`
	}
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	switch par.Action {
	case "begin":
		if par.Library == "" {
			response.FailWithMessage("文件库不能为空", c)
			return
		}
		id := core.GenerateUniqueID()
		importbatch.Global.Begin(id, par.Library)
		response.OkWithData(id, c)
	case "heartbeat":
		if !importbatch.Global.Touch(par.BatchID, par.Library) {
			response.FailWithMessage("导入批次已过期", c)
			return
		}
		response.OkWithData(true, c)
	case "end":
		if importbatch.Global.Touch(par.BatchID, par.Library) {
			importbatch.Global.End(par.BatchID)
		}
		response.OkWithData(true, c)
	default:
		response.FailWithMessage("未知批次操作", c)
	}
}
