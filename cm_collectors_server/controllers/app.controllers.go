package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

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
