package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
)

type App struct {
	AppConfig      datatype.App_Config      `json:"appConfig"`
	FilesBases     *[]models.FilesBases     `json:"filesBases"`
	PerformerBases *[]models.PerformerBases `json:"performerBases"`
}

func (App) InitData() (*App, error) {
	filesBases, err := FilesBases{}.DataList()
	if err != nil {
		return nil, err
	}
	performerBases, err := PerformerBases{}.DataList()
	if err != nil {
		return nil, err
	}
	return &App{
		AppConfig: datatype.App_Config{
			LogoName:     core.Config.General.LogoName,
			IsAdminLogin: core.Config.General.IsAdminLogin,
		},
		FilesBases:     filesBases,
		PerformerBases: performerBases,
	}, nil
}

func (App) GetConfig() datatype.App_SystemConfig {
	config := datatype.App_SystemConfig{
		App_Config: datatype.App_Config{
			LogoName:     core.Config.General.LogoName,
			IsAdminLogin: core.Config.General.IsAdminLogin,
		},
		AdminPassword:          "",
		IsAutoCreateM3u8:       core.Config.General.IsAutoCreateM3u8,
		Language:               core.Config.General.Language,
		NotAllowServerOpenFile: core.Config.General.NotAllowServerOpenFile,
		PlayVideoFormats:       core.Config.Play.PlayVideoFormats,
		PlayAudioFormats:       core.Config.Play.PlayAudioFormats,
		VideoRateLimit:         core.Config.General.VideoRateLimit,
		Scraper: datatype.App_Config_Scraper{
			BrowserPath:    core.Config.Scraper.BrowserPath,
			UseBrowserPath: core.Config.Scraper.UseBrowserPath,
		},
		TaryMenu:                     core.Config.TaryMenu,
		ServerFileManagementRootPath: core.GetConfig_ServerFileManagementRootPath(),
	}
	return config
}

func (App) SetConfig(config datatype.App_SystemConfig) error {
	core.Config.General.LogoName = config.LogoName
	core.Config.General.IsAdminLogin = config.IsAdminLogin
	if config.AdminPassword != "" {
		core.Config.General.AdminPassword = config.AdminPassword
	}
	core.Config.General.IsAutoCreateM3u8 = config.IsAutoCreateM3u8
	core.Config.General.Language = config.Language
	core.Config.General.NotAllowServerOpenFile = config.NotAllowServerOpenFile
	core.Config.Play.PlayVideoFormats = config.PlayVideoFormats
	core.Config.Play.PlayAudioFormats = config.PlayAudioFormats
	core.Config.General.VideoRateLimit = config.VideoRateLimit
	core.Config.Scraper.BrowserPath = config.Scraper.BrowserPath
	core.Config.Scraper.UseBrowserPath = config.Scraper.UseBrowserPath
	core.Config.TaryMenu = config.TaryMenu
	core.Config.ServerFileManagement.RootPath = config.ServerFileManagementRootPath
	return core.SaveConfig()
}
