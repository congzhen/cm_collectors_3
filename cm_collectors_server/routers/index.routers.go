package routers

import (
	"cm_collectors_server/controllers"
	"cm_collectors_server/middleware"
	"embed"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/*
*初始化路由
 */
func InitRouter(router *gin.Engine, htmlFS *embed.FS) *gin.Engine {
	//设置信任ip
	router.SetTrustedProxies([]string{"127.0.0.1", "0.0.0.0"})

	embeddedFolderCreateStaticRoutes := EmbeddedFolderCreateStaticRoutes{
		Router:     router,
		Efs:        *htmlFS,
		FolderName: "html",
	}
	if err := embeddedFolderCreateStaticRoutes.Register(); err != nil {
		log.Fatalf("注册静态文件失败: %v", err)
	}
	router.GET("/", func(c *gin.Context) {
		data, _ := htmlFS.ReadFile("html/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.Static("/api/performerFace", "./db/performerFace")
	router.Static("/api/resCoverPoster", "./db/resCoverPoster")
	SFMRouter(router)
	publicRouter(router)
	AdminRouter(router)
	router.NoRoute(handleNoRoute)
	return router
}

func handleNoRoute(c *gin.Context) {
	// 对于非API路由，返回index.html让Vue Router处理前端路由
	if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.Redirect(http.StatusFound, "/")
		return
	}
	// API路由返回404错误
	c.JSON(http.StatusNotFound, gin.H{"error": "404 Not Found"})
}

func publicRouter(router *gin.Engine) {
	routerGroup := router.Group("/api")

	routerGroup.POST("/login/admin", controllers.Login{}.Admin)

	routerGroup.GET("/app/data", controllers.App{}.Data)
	routerGroup.GET("/tag/data/:filesBasesId", controllers.Tag{}.TagData)
	routerGroup.GET("/tag/list/filesBasesId/:filesBasesId", controllers.Tag{}.TagList_FilesBasesId)
	routerGroup.GET("/tag/list/tagClassId/:tagClassId", controllers.Tag{}.TagList_TagClassId)
	routerGroup.GET("/tagClass/list/:filesBasesId", controllers.Tag{}.TagClassList)

	routerGroup.GET("/filesBases/info/:id", controllers.FilesBases{}.InfoDetails)
	routerGroup.GET("/filesBases/config/:id/:configType", controllers.FilesBases{}.Config)

	routerGroup.POST("performer/basicList", controllers.Performer{}.BasicList)
	routerGroup.GET("/performer/dataList/:performerBasesId/:fetchCount/:page/:limit", controllers.Performer{}.DataList)
	routerGroup.POST("/performer/list/top/preferred", controllers.Performer{}.ListTopPreferredPerformers)
	routerGroup.GET("/performer/recycleBin/:performerBasesId", controllers.Performer{}.RecycleBin)

	routerGroup.GET("/resource/info/:resourceId", controllers.Resource{}.Info)
	routerGroup.POST("/resource/dataList", controllers.Resource{}.DataList)

	routerGroup.GET("play/open/resource/:resourceId", controllers.Play{}.PlayOpenResource)
	routerGroup.GET("play/open/resource/folder/:resourceId", controllers.Play{}.PlayOpenResourceFolder)

	routerGroup.GET("/video/mp4/:dramaSeriesId/v.mp4", controllers.Play{}.PlayVideoMP4)
	routerGroup.GET("/video/m3u8/:dramaSeriesId/v.m3u8", controllers.Play{}.VideoM3u8)
	routerGroup.GET("hls_video/:dramaSeriesId/:start/:duration", controllers.Play{}.PlayVideoHLS)
	routerGroup.GET("/video/subtitle/:dramaSeriesId", controllers.Play{}.VideoSubtitle)

	routerGroup.GET("files/list/image/:dramaSeriesId", controllers.FilesCL{}.FilesList_Image)
	routerGroup.GET("files/image/:dramaSeriesId/:fileNameBase64", controllers.FilesCL{}.Files_Image) //Query("thumbWidth")缩率图宽度,Query("thumbLevel") 缩率图质量

	// 添加TVBox相关路由
	routerGroup.GET("/tvbox/home", controllers.TVBox{}.Home)
	routerGroup.GET("/tvbox/sites/videos", controllers.TVBox{}.Videos)
}

func AdminRouter(router *gin.Engine) {
	routerGroup := router.Group("/api", middleware.AdminLoginApiMiddleware())
	routerGroup.GET("/app/getConfig", controllers.App{}.GetConfig)
	routerGroup.PUT("/app/setConfig", controllers.App{}.SetConfig)

	routerGroup.POST("resource/create", controllers.Resource{}.CreateResource)
	routerGroup.PUT("resource/update", controllers.Resource{}.UpdateResource)
	routerGroup.DELETE("resource/delete/:resourceId", controllers.Resource{}.DeleteResource)
	routerGroup.POST("resourceDramaSeries/searchPath", controllers.ResourceDramaSeries{}.SearchPath)
	routerGroup.POST("resourceDramaSeries/replacePath", controllers.ResourceDramaSeries{}.ReplacePath)
	routerGroup.POST("filesBases/create", controllers.FilesBases{}.Create)
	routerGroup.PUT("filesBases/setData", controllers.FilesBases{}.SetFilesBases)
	routerGroup.PUT("filesBases/sort", controllers.FilesBases{}.Sort)
	routerGroup.POST("performerBases/create", controllers.Performer{}.Create)
	routerGroup.PUT("performerBases/update", controllers.Performer{}.PerformerBasesUpdate)
	routerGroup.POST("performer/create", controllers.Performer{}.CreatePerformer)
	routerGroup.PUT("performer/update", controllers.Performer{}.UpdatePerformer)
	routerGroup.PUT("performer/updateStatus", controllers.Performer{}.UpdatePerformerStatus)
	routerGroup.POST("tag/create", controllers.Tag{}.CreateTag)
	routerGroup.PUT("tag/update", controllers.Tag{}.UpdateTag)
	routerGroup.POST("tagClass/create", controllers.Tag{}.CreateTagClass)
	routerGroup.PUT("tagClass/update", controllers.Tag{}.UpdateTagClass)
	routerGroup.PUT("tag/update/sort", controllers.Tag{}.UpdateSort)

	routerGroup.POST("FFmpeg/getVideoThumbnails", controllers.FFmpeg{}.GetVideoThumbnails)

	routerGroup.POST("importData/scanDiskImportPaths", controllers.ImportData{}.ScanDiskImportPaths)
	routerGroup.POST("importData/scanDiskImportData", controllers.ImportData{}.ScanDiskImportData)
}
func SFMRouter(router *gin.Engine) {
	routerGroup := router.Group("/api", middleware.AdminLoginApiMiddleware())
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
