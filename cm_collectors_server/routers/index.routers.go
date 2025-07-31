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
	SFMRouter(router)
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
	routerGroup.GET("/tag/list/filesBasesId/:filesBasesId", controllers.Tag{}.TagList_FilesBasesId)
	routerGroup.GET("/tag/list/tagClassId/:tagClassId", controllers.Tag{}.TagList_TagClassId)
	routerGroup.GET("/tagClass/list/:filesBasesId", controllers.Tag{}.TagClassList)
	routerGroup.GET("/filesBases/info/:id", controllers.FilesBases{}.InfoDetails)
	routerGroup.POST("performer/basicList", controllers.Performer{}.BasicList)
	routerGroup.GET("/performer/dataList/:performerBasesId/:fetchCount/:page/:limit", controllers.Performer{}.DataList)
	routerGroup.POST("/performer/list/top/preferred", controllers.Performer{}.ListTopPreferredPerformers)
	routerGroup.GET("/performer/recycleBin/:performerBasesId", controllers.Performer{}.RecycleBin)
	routerGroup.GET("/resource/info/:resourceId", controllers.Resource{}.Info)
	routerGroup.POST("/resource/dataList", controllers.Resource{}.DataList)

	routerGroup.GET("play/open/resource/:resourceId", controllers.Play{}.PlayOpenResource)
	routerGroup.GET("play/open/resource/folder/:resourceId", controllers.Play{}.PlayOpenResourceFolder)

	routerGroup.GET("/video/mp4/:dramaSeriesId", controllers.Play{}.PlayVideoMP4)
	routerGroup.GET("/video/subtitle/:dramaSeriesId", controllers.Play{}.VideoSubtitle)

	routerGroup.GET("files/list/image/:dramaSeriesId", controllers.FilesCL{}.FilesList_Image)
	routerGroup.GET("files/image/:dramaSeriesId/:fileNameBase64", controllers.FilesCL{}.Files_Image) //Query("thumbWidth")缩率图宽度,Query("thumbLevel") 缩率图质量
}

func AdminRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")
	routerGroup.POST("resource/create", controllers.Resource{}.CreateResource)
	routerGroup.PUT("resource/update", controllers.Resource{}.UpdateResource)
	routerGroup.DELETE("resource/delete/:resourceId", controllers.Resource{}.DeleteResource)
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
func SFMRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")
	routerGroup.POST("sfm/pathDir", controllers.ServerFileManagement{}.PathDir)
	routerGroup.POST("sfm/searchFiles", controllers.ServerFileManagement{}.SearchFiles)
	routerGroup.POST("sfm/createFile", controllers.ServerFileManagement{}.CreateFile)
	routerGroup.POST("sfm/createFolder", controllers.ServerFileManagement{}.CreateFolder)
	routerGroup.POST("sfm/openFile", controllers.ServerFileManagement{}.OpenFile)
	routerGroup.POST("sfm/saveFile", controllers.ServerFileManagement{}.SaveFile)
	routerGroup.POST("sfm/renameFile", controllers.ServerFileManagement{}.RenameFile)
	routerGroup.POST("sfm/uploadFile", controllers.ServerFileManagement{}.UploadFile)
	routerGroup.GET("sfm/downloadFile", controllers.ServerFileManagement{}.DownloadFile)
	routerGroup.POST("sfm/copyFile", controllers.ServerFileManagement{}.CopyFile)
	routerGroup.POST("sfm/moveFile", controllers.ServerFileManagement{}.MoveFile)
	routerGroup.POST("sfm/compressFile", controllers.ServerFileManagement{}.CompressFile)
	routerGroup.POST("sfm/permissionsFile", controllers.ServerFileManagement{}.PermissionsFile)
	routerGroup.POST("/sfm/unCompressFile", controllers.ServerFileManagement{}.UnCompressFile)
	routerGroup.POST("sfm/deleteFile", controllers.ServerFileManagement{}.DeleteFile)
}
