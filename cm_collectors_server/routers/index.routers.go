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
	routerGroup.GET("/tag/list/:filesBasesId", controllers.Tag{}.TagList)
	routerGroup.GET("/tagClass/list/:filesBasesId", controllers.Tag{}.TagClassList)
	routerGroup.GET("/filesBases/info/:id", controllers.FilesBases{}.InfoDetails)
	routerGroup.POST("performer/basicList", controllers.Performer{}.BasicList_Performer)
	routerGroup.GET("/performer/dataList/:performerBasesId/:fetchCount/:page/:limit", controllers.Performer{}.DataList)
	routerGroup.POST("/performer/list/top/preferred", controllers.Performer{}.ListTopPreferredPerformers)
	routerGroup.GET("/performer/recycleBin/:performerBasesId", controllers.Performer{}.RecycleBin)
	routerGroup.POST("/resource/dataList", controllers.Resource{}.DataList)
}

func AdminRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")
	routerGroup.PUT("filesBases/setData", controllers.FilesBases{}.SetFilesBases)
	routerGroup.PUT("performerBases/update", controllers.Performer{}.PerformerBasesUpdate)
	routerGroup.POST("performer/create", controllers.Performer{}.CreatePerformer)
	routerGroup.PUT("performer/update", controllers.Performer{}.UpdatePerformer)
	routerGroup.PUT("performer/updateStatus", controllers.Performer{}.UpdatePerformerStatus)
	routerGroup.POST("tag/create", controllers.Tag{}.CreateTag)
	routerGroup.PUT("tag/update", controllers.Tag{}.UpdateTag)
	routerGroup.POST("tagClass/create", controllers.Tag{}.CreateTagClass)
	routerGroup.PUT("tagClass/update", controllers.Tag{}.UpdateTagClass)
	routerGroup.PUT("tag/update/sort", controllers.Tag{}.UpdateSort)
}
