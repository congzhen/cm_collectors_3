package core

import (
	"cm_collectors_server/config"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func initConf() *config.Config {
	const configFile = "server_config.yaml"
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
