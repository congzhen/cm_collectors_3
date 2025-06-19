package core

import (
	"cm_collectors_server/cache"
	"cm_collectors_server/config"
	"fmt"
)

var cacheService cache.Cache

func initCache(config config.Cache) {
	c, err := cache.NewCache(config)
	if err != nil {
		panic(fmt.Errorf("get yamlConf error:%s", err))
	}
	cacheService = c
}

func Cache() cache.Cache {
	return cacheService
}
