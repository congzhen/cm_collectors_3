package core

import (
	"cm_collectors_server/config"
)

var (
	Config *config.Config
)

func Init() {
	Config = initConf()
	initCache(Config.Cache)
	logInit(Config.System.LogFilePath, Config.System.LogLevel)
	initGorm()
}
