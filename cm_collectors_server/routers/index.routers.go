package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
*初始化路由
 */
func InitRouter(router *gin.Engine) *gin.Engine {
	//设置信任ip
	router.SetTrustedProxies([]string{"127.0.0.1", "0.0.0.0"})
	router.NoRoute(HandleNoRoute)
	return router
}

func HandleNoRoute(c *gin.Context) {
	// 返回404错误
	c.JSON(http.StatusNotFound, gin.H{"error": "404 Not Found"})
}
