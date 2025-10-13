package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"cm_collectors_server/utils"
	"errors"

	"github.com/gin-gonic/gin"
)

type Scraper struct {
}

func (Scraper) ScraperConfigs(c *gin.Context) {
	// 获取所有配置文件
	configs, err := utils.GetFilesByExtensions([]string{"./scraper"}, []string{".json"}, false)
	if err := ResError(c, err); err != nil {
		return
	}
	for i := range configs {
		configs[i] = utils.GetFileNameFromPath(configs[i], false)
	}
	response.OkWithData(configs, c)
}
func (Scraper) Pretreatment(c *gin.Context) {
	var par datatype.ReqParam_Scraper
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	pendingFilePaths, err := processors.Scraper{}.Pretreatment(par.FilesBasesId, par.Config)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(pendingFilePaths, c)
}
func (Scraper) ScraperDataProcess(c *gin.Context) {
	var par datatype.ReqParam_ScraperProcess
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Scraper{}.ScraperDataProcess(par.FilesBasesId, par.FilePath, par.Config)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(true, c)
}

func (Scraper) SearchScraperPerformer(c *gin.Context) {
	var par datatype.ReqParam_SearchScraperPerformer
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	datalist, err := processors.Performer{}.SearchLastScraperUpdateTime(par.PerformerBasesId, par.LastScraperUpdateTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(datalist, c)
}

func (Scraper) ScraperPerformerDataProcess(c *gin.Context) {
	var par datatype.ReqParam_ScraperPerformerDataProcess
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Scraper{}.ScraperPerformerDataProcess(&par)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(true, c)
}

func (Scraper) ScraperOneResourceDataProcess(c *gin.Context) {
	var par datatype.ReqParam_ScraperOneResourceDataProcess
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Scraper{}.ScraperOneResourceDataProcess(&par)
	if err != nil {
		if errors.Is(err, errorMessage.Err_No_Config_ScanDisk) {
			ResError(c, err)
		} else {
			response.FailWithMessage(err.Error(), c)
		}
		return
	}
	response.OkWithData(info, c)
}
