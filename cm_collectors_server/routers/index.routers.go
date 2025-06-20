package routers

import (
	"cm_collectors_server/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
*初始化路由
 */
func InitRouter(router *gin.Engine) *gin.Engine {
	//设置信任ip
	router.SetTrustedProxies([]string{"127.0.0.1", "0.0.0.0"})
	publicRouter(router)
	router.NoRoute(handleNoRoute)
	return router
}

func handleNoRoute(c *gin.Context) {
	// 返回404错误
	c.JSON(http.StatusNotFound, gin.H{"error": "404 Not Found"})
}

func publicRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")
	routerGroup.GET("/app/data", controllers.App{}.Data)
}
