package controllers

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type App struct{}

func (App) Data(c *gin.Context) {
	appData, err := processors.App{}.InitData()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(appData, c)
}

func (App) GetConfig(c *gin.Context) {
	config := processors.App{}.GetConfig()
	response.OkWithData(config, c)
}

func (App) SetConfig(c *gin.Context) {
	var par datatype.App_SystemConfig
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.App{}.SetConfig(par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (App) GetUpdateSoftConfig(c *gin.Context) {
	updateURL := core.Config.System.UpdateSoftConfig
	// 发起HTTP GET请求
	resp, err := http.Get(updateURL)
	if err != nil {
		response.FailWithMessage("请求更新配置失败: "+err.Error(), c)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.FailWithMessage("读取更新配置失败: "+err.Error(), c)
		return
	}

	// 直接返回JSON数据
	response.OkWithData(string(body), c)
}
