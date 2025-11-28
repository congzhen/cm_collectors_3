package datatype

import "cm_collectors_server/config"

type ResDataList struct {
	DataList any   `json:"dataList"  `
	Total    int64 `json:"total"  `
}

type E_resourceMode string

const (
	E_resourceMode_Movies    E_resourceMode = "movies"
	E_resourceMode_Comic     E_resourceMode = "comic"
	E_resourceMode_Atlas     E_resourceMode = "atlas"
	E_resourceMode_Files     E_resourceMode = "files"
	E_resourceMode_VideoLink E_resourceMode = "videoLink"
	E_resourceMode_NetDisk   E_resourceMode = "netDisk"
)

type E_cronJobsType string

const (
	E_cronJobsType_Import           E_cronJobsType = "import"
	E_cronJobsType_ScraperResource  E_cronJobsType = "scraperResource"
	E_cronJobsType_ScraperPerformer E_cronJobsType = "scraperPerformer"
	E_cronJobsType_Clear            E_cronJobsType = "clear"
)

type App_Config struct {
	LogoName     string `json:"logoName"`
	IsAdminLogin bool   `json:"isAdminLogin"`
}

type App_Config_Scraper struct {
	UseBrowserPath bool   `json:"useBrowserPath"`
	BrowserPath    string `json:"browserPath"`
}

type App_SystemConfig struct {
	App_Config
	AdminPassword                string                `json:"adminPassword"`
	IsAutoCreateM3u8             bool                  `json:"isAutoCreateM3u8"`
	Language                     string                `json:"language"`
	NotAllowServerOpenFile       bool                  `json:"notAllowServerOpenFile"`
	PlayVideoFormats             []string              `json:"playVideoFormats"`
	PlayAudioFormats             []string              `json:"playAudioFormats"`
	VideoRateLimit               config.VideoRateLimit `json:"videoRateLimit"`
	Scraper                      App_Config_Scraper    `json:"scraper"`
	TaryMenu                     []config.TaryMenu     `json:"taryMenu"`
	ServerFileManagementRootPath []string              `json:"serverFileManagementRootPath"`
}

// 用户类型
type UserType string

const (
	// 普通用户
	ENUM_UserType_Normal UserType = "normal"
	// 管理员用户
	ENUM_UserType_Admin UserType = "admin"
)

type DatabaseCleanupClearItem string

const (
	ENUM_DatabaseCleanupClearItem_Resource               DatabaseCleanupClearItem = "resource"
	ENUM_DatabaseCleanupClearItem_Performer              DatabaseCleanupClearItem = "performer"
	ENUM_DatabaseCleanupClearItem_Tags                   DatabaseCleanupClearItem = "tags"
	ENUM_DatabaseCleanupClearItem_TagClass               DatabaseCleanupClearItem = "tagClass"
	ENUM_DatabaseCleanupClearItem_FileDatabaseConfig     DatabaseCleanupClearItem = "fileDatabaseConfig"
	ENUM_DatabaseCleanupClearItem_ImportConfig           DatabaseCleanupClearItem = "importConfig"
	ENUM_DatabaseCleanupClearItem_ResourceScraperConfig  DatabaseCleanupClearItem = "resourceScraperConfig"
	ENUM_DatabaseCleanupClearItem_PerformerScraperConfig DatabaseCleanupClearItem = "performerScraperConfig"
	ENUM_DatabaseCleanupClearItem_GeneralConfig          DatabaseCleanupClearItem = "generalConfig"
	ENUM_DatabaseCleanupClearItem_CronJobs               DatabaseCleanupClearItem = "cronJobs"
)
