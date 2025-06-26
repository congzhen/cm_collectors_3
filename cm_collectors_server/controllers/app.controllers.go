package controllers

import (
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
