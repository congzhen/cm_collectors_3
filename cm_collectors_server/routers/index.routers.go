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
	routerGroup.GET("/updateSoftConfig", controllers.App{}.GetUpdateSoftConfig)

	routerGroup.GET("/app/data", controllers.App{}.Data)
	routerGroup.GET("/tag/data/:filesBasesId", controllers.Tag{}.TagData)
	routerGroup.GET("/tag/list/filesBasesId/:filesBasesId", controllers.Tag{}.TagList_FilesBasesId)
	routerGroup.GET("/tag/list/tagClassId/:tagClassId", controllers.Tag{}.TagList_TagClassId)
	routerGroup.GET("/tagClass/list/:filesBasesId", controllers.Tag{}.TagClassList)

	routerGroup.GET("/filesBases/info/:id", controllers.FilesBases{}.InfoDetails)
	routerGroup.GET("/filesBases/config/:id/:configType", controllers.FilesBases{}.Config)

	routerGroup.GET("/performer/info/:id", controllers.Performer{}.Info)
	routerGroup.POST("performer/basicList", controllers.Performer{}.BasicList)
	routerGroup.GET("/performer/dataList/ids/:ids", controllers.Performer{}.DataListByIds)
	routerGroup.GET("/performer/dataList/:performerBasesId/:fetchCount/:page/:limit", controllers.Performer{}.DataList)
	routerGroup.POST("/performer/list/top/preferred", controllers.Performer{}.ListTopPreferredPerformers)
	routerGroup.GET("/performer/recycleBin/:performerBasesId", controllers.Performer{}.RecycleBin)

	routerGroup.GET("/resource/info/:resourceId", controllers.Resource{}.Info)
	routerGroup.POST("/resource/dataList", controllers.Resource{}.DataList)
	routerGroup.POST("/resource/list/ids", controllers.Resource{}.ListIds)
	routerGroup.GET("/resource/sampleImages/:resourceId", controllers.Resource{}.SampleImages)
	routerGroup.GET("/resource/sampleData/:resourceId", controllers.Resource{}.SampleImageData)

	routerGroup.GET("/resourceDramaSeries/info/:dramaSeriesId", controllers.ResourceDramaSeries{}.Info)

	routerGroup.GET("play/video/info/:dramaSeriesId", controllers.Play{}.PlayVideoInfo)
	routerGroup.GET("play/open/resource/:resourceId", controllers.Play{}.PlayOpenResource)
	routerGroup.GET("play/open/dramaSeries/:dramaSeriesId", controllers.Play{}.PlayOpenDramaSeries)
	routerGroup.GET("play/open/resource/folder/:resourceId", controllers.Play{}.PlayOpenResourceFolder)
	routerGroup.GET("/play/update/:resourceId", controllers.Play{}.PlayUpdate)

	routerGroup.GET("/video/m3u8/:dramaSeriesId/:videoM3u8", controllers.Play{}.VideoM3u8)
	routerGroup.GET("/video/m3u8/stream/:dramaSeriesId/:start/:duration/:ts", middleware.VideoStreamRateLimitMiddleware(), controllers.Play{}.VideoM3u8StreamHLS)

	routerGroup.GET("/video/mp4/:dramaSeriesId/:videoMp4", middleware.VideoStreamRateLimitMiddleware(), controllers.Play{}.PlayVideoMP4)
	routerGroup.GET("/video/mp4/:dramaSeriesId/v.srt", controllers.Play{}.VideoSubtitle)
	routerGroup.GET("/video/mp4/:dramaSeriesId/v.ass", controllers.Play{}.VideoSubtitle)
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

	routerGroup.POST("database/cleanup", controllers.App{}.DatabaseCleanup)
	routerGroup.GET("database/dbBackupList", controllers.App{}.DBBackupList)
	routerGroup.DELETE("database/deleteDbBackup/:fileName", controllers.App{}.DeleteDbBackup)

	routerGroup.POST("resource/create", controllers.Resource{}.CreateResource)
	routerGroup.PUT("resource/update", controllers.Resource{}.UpdateResource)
	routerGroup.PUT("resource/update/performer", controllers.Resource{}.UpdateResourcePerformer)
	routerGroup.PUT("resource/update/tag", controllers.Resource{}.UpdateResourceTag)
	routerGroup.PUT("resource/batchSetTag", controllers.Resource{}.BatchSetTag)
	routerGroup.GET("resource/deleted/list", controllers.Resource{}.ListDeletedResource)
	routerGroup.DELETE("resource/delete/:resourceId", controllers.Resource{}.DeleteResource)
	routerGroup.POST("resourceDramaSeries/searchPath", controllers.ResourceDramaSeries{}.SearchPath)
	routerGroup.POST("resourceDramaSeries/replacePath", controllers.ResourceDramaSeries{}.ReplacePath)
	routerGroup.POST("filesBases/create", controllers.FilesBases{}.Create)
	routerGroup.PUT("filesBases/setData", controllers.FilesBases{}.SetFilesBases)
	routerGroup.PUT("filesBases/sort", controllers.FilesBases{}.Sort)
	routerGroup.PUT("filesBases/setConfig/filesBases", controllers.FilesBases{}.SetConfig_FilesBases)
	routerGroup.POST("performerBases/create", controllers.Performer{}.Create)
	routerGroup.PUT("performerBases/update", controllers.Performer{}.PerformerBasesUpdate)
	routerGroup.POST("performer/create", controllers.Performer{}.CreatePerformer)
	routerGroup.PUT("performer/update", controllers.Performer{}.UpdatePerformer)
	routerGroup.PUT("performer/updateStatus", controllers.Performer{}.UpdatePerformerStatus)
	routerGroup.POST("tag/create", controllers.Tag{}.CreateTag)
	routerGroup.PUT("tag/update", controllers.Tag{}.UpdateTag)
	routerGroup.DELETE("tag/delete/:tagId", controllers.Tag{}.DeleteTag)
	routerGroup.POST("tagClass/create", controllers.Tag{}.CreateTagClass)
	routerGroup.PUT("tagClass/update", controllers.Tag{}.UpdateTagClass)
	routerGroup.PUT("tag/update/sort", controllers.Tag{}.UpdateSort)
	routerGroup.POST("tag/import", controllers.Tag{}.ImportTag)

	routerGroup.POST("FFmpeg/getVideoThumbnails", controllers.FFmpeg{}.GetVideoThumbnails)

	routerGroup.POST("importData/scanDiskImportPaths", controllers.ImportData{}.ScanDiskImportPaths)
	routerGroup.POST("importData/scanDiskImportData", controllers.ImportData{}.ScanDiskImportData)
	routerGroup.POST("importData/updateScanDiskConfig", controllers.ImportData{}.UpdateScanDiskConfig)

	routerGroup.GET("scraper/configs", controllers.Scraper{}.ScraperConfigs)
	routerGroup.POST("scraper/updateConfig", controllers.ImportData{}.UpdateScraperConfig)
	routerGroup.POST("scraper/pretreatment", controllers.Scraper{}.Pretreatment)
	routerGroup.POST("scraper/scraperDataProcess", controllers.Scraper{}.ScraperDataProcess)

	routerGroup.POST("scraper/searchScraperPerformer", controllers.Scraper{}.SearchScraperPerformer)
	routerGroup.POST("scraper/scraperPerformerDataProcess", controllers.Scraper{}.ScraperPerformerDataProcess)
	routerGroup.POST("scraper/scraperOneResourceDataProcess", controllers.Scraper{}.ScraperOneResourceDataProcess)
	routerGroup.POST("scraper/scraperOnePerformerDataProcess", controllers.Scraper{}.ScraperOnePerformerDataProcess)

	routerGroup.GET("cronJobs/list", controllers.CronJobs{}.List)
	routerGroup.POST("cronJobs/create", controllers.CronJobs{}.Create)
	routerGroup.PUT("cronJobs/update", controllers.CronJobs{}.Update)
	routerGroup.DELETE("cronJobs/delete/:cronJobsId", controllers.CronJobs{}.Delete)

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
