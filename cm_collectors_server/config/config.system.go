package config

import "fmt"

type System struct {
	ServerHost          string `yaml:"serverHost"`
	Port                int    `yaml:"port"`
	Database            string `yaml:"database"`
	FilePath            string `yaml:"filePath"`
	Env                 string `yaml:"Env"`
	ResponseMsgLanguage string `yaml:"responseMsgLanguage"`
	LogFilePath         string `yaml:"logFilePath"`
	LogLevel            string `yaml:"logLevel"`
}

/*获取服务器地址*/
func (s System) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.ServerHost, s.Port)
}

/*获取gin的运行模式*/
func (s System) GetGinMode() string {
	if s.Env == "dev" {
		return "debug"
	} else {
		return "release"
	}
}
