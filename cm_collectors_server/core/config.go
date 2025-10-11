package core

import (
	"cm_collectors_server/config"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const configFile = "config.yaml"

// getDefaultConfig 返回默认配置
func getDefaultConfig() *config.Config {
	return &config.Config{
		General: config.General{
			LogoName:               "CM File Collectors",
			IsAdminLogin:           false,
			AdminPassword:          "",
			IsAutoCreateM3u8:       true,
			Language:               "zhCn",
			NotAllowServerOpenFile: false,
			VideoRateLimit: config.VideoRateLimit{
				Enabled:           false,
				RequestsPerSecond: 5,  // 每秒5个请求
				Burst:             10, // 桶容量为10
			},
		},
		Sqlite3: config.Sqlite3{
			Path:        "./db/cm_collectors.db",
			LoggerLevel: "info",
		},
		System: config.System{
			ServerHost:          "0.0.0.0",
			Port:                12345,
			Database:            "sqlite3",
			FilePath:            "./db/",
			Env:                 "debug",
			ResponseMsgLanguage: "en",
			LogFilePath:         "./err.log",
			LogLevel:            "Error",
			UpdateSoftConfig:    "https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/updateConfig/updateSoftConfig.json",
		},
		Scraper: config.Scraper{
			Headless:  true,
			LogPath:   "./scraper.log",
			LogStatus: true,
			VisitHome: false,
			Timeout:   30,
		},
		Play: config.Play{
			PlayVideoFormats: []string{"h264", "vp8", "vp9", "av1", "hevc"},
			PlayAudioFormats: []string{"aac", "opus", "mp3", "vorbis", "pcm_s16le", "pcm_s24le"},
		},
		Jwt: config.Jwt{
			ExpiresAt:      86400,
			Issuer:         "cmCollectors",
			Audience:       "cmCollectors",
			Subject:        "cmCollectors",
			PrivateKeyPath: "./cert/jwtPrivate.key",
			PublicKeyPath:  "./cert/jwtPublic.key",
		},
		Mysql: config.Mysql{
			Host:         "127.0.0.1",
			Port:         3306,
			Db:           "cm_collectors",
			User:         "root",
			Password:     "123456",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
			MaxIdleConns: 3,
			MaxOpenConns: 20,
			LoggerLevel:  "info",
		},
		Cache: config.Cache{
			Mode: "freeCache",
			Redis: config.RedisCache{
				Host:             "127.0.0.1",
				Port:             6379,
				Password:         "xxxx",
				Db:               0,
				DefaultExpireSec: 3600,
				MinIdleConns:     5,
				PoolSize:         20,
				ConnMaxIdleTime:  240,
			},
			FreeCache: config.FreeCache{
				MaxMemoryMB:      100,
				DefaultExpireSec: 3600,
			},
		},
		ServerFileManagement: config.ServerFileManagement{
			RootPath: []string{},
		},
	}
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	defaultConfig := getDefaultConfig()
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入默认配置文件失败: %v", err)
	}

	log.Println("默认配置文件创建成功")
	return nil
}

func initConf() *config.Config {
	c := &config.Config{}
	yamlConf, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("无法读取配置文件: %v，将创建默认配置文件", err)
		err = createDefaultConfig()
		if err != nil {
			panic(fmt.Errorf("创建默认配置文件失败:%s", err))
		}

		// 重新读取配置文件
		yamlConf, err = os.ReadFile(configFile)
		if err != nil {
			panic(fmt.Errorf("重新读取配置文件失败:%s", err))
		}
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
