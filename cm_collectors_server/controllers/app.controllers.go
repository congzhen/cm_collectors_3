package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type App struct{}

func (App) Data(c *gin.Context) {
	appData, err := processors.App{}.InitData()
	if err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "服务器错误",
		})
		return
	}
	response.OkWithData(appData, c)
}
