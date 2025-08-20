package core

import (
	"cm_collectors_server/config"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const configFile = "config.yaml"

func initConf() *config.Config {
	c := &config.Config{}
	yamlConf, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Errorf("get yamlConf error:%s", err))
	}
	err = yaml.Unmarshal(yamlConf, c)
	if err != nil {
		log.Fatalf("config Init Unmarshal: %v", err)
	}
	log.Println("config yamlFile load Init success.")
	return c
}

// SaveConfig 将当前配置保存到YAML文件
func SaveConfig() error {
	data, err := yaml.Marshal(Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Println("配置文件保存成功")
	return nil
}

// GetConfig 获取全局配置实例
func GetConfig() *config.Config {
	return Config
}
