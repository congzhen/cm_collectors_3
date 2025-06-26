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
	router.Static("/api/performerFace", "./db/performerFace")
	router.Static("/api/resCoverPoster", "./db/resCoverPoster")
	publicRouter(router)
	AdminRouter(router)
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
	routerGroup.GET("/tag/data/:filesBasesId", controllers.Tag{}.TagData)
	routerGroup.GET("/filesBases/info/:id", controllers.FilesBases{}.InfoDetails)
	routerGroup.GET("/performer/dataList/:performerBasesId/:fetchCount/:page/:limit", controllers.Performer{}.DataList)
	routerGroup.POST("/performer/list/top/preferred", controllers.Performer{}.ListTopPreferredPerformers)
	routerGroup.POST("/resource/dataList", controllers.Resource{}.DataList)
}

func AdminRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")
	routerGroup.PUT("performerBases/update", controllers.Performer{}.PerformerBasesUpdate)
}
