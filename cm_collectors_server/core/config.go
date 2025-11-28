package core

import (
	"cm_collectors_server/config"
	"fmt"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

const configFile = "config.yaml"

// 定义始终优先使用配置文件值的字段列表（即使为零值也不使用默认值覆盖）
var configFilePriorityFields = map[string]bool{
	"General.IsAutoCreateM3u8": true,
	"Scraper.LogStatus":        true,
	"Scraper.Headless":         true,
	"Scraper.VisitHome":        true,
}

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
			UseBrowserPath: false,
			BrowserPath:    "",
			Headless:       true,
			LogPath:        "./scraper.log",
			LogStatus:      true,
			VisitHome:      false,
			Timeout:        30,
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
			RootPath: []string{"/", "A:\\", "B:\\", "C:\\", "D:\\", "E:\\", "F:\\", "G:\\", "H:\\", "I:\\", "J:\\", "K:\\", "L:\\", "M:\\", "N:\\", "O:\\", "P:\\", "Q:\\", "R:\\", "S:\\", "T:\\", "U:\\", "V:\\", "W:\\", "X:\\", "Y:\\", "Z:\\"},
		},
		TaryMenu: []config.TaryMenu{},
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

// mergeWithDefaults 使用默认配置填充未设置的字段
func mergeWithDefaults(defaultConfig, userConfig *config.Config) {
	mergeRecursive(reflect.ValueOf(defaultConfig).Elem(), reflect.ValueOf(userConfig).Elem(), "")
}

// mergeRecursive 递归合并两个结构体
func mergeRecursive(defaultVal, userVal reflect.Value, path string) {
	switch userVal.Kind() {
	case reflect.Struct:
		for i := 0; i < userVal.NumField(); i++ {
			field := userVal.Field(i)
			defaultField := defaultVal.Field(i)
			fieldName := userVal.Type().Field(i).Name
			currentPath := path
			if currentPath == "" {
				currentPath = fieldName
			} else {
				currentPath = path + "." + fieldName
			}

			// 检查字段是否可以设置
			if field.CanSet() {
				// 如果是结构体则递归处理
				if field.Kind() == reflect.Struct {
					mergeRecursive(defaultField, field, currentPath)
				} else {
					// 检查是否是配置文件优先字段
					_, isConfigPriority := configFilePriorityFields[currentPath]

					// 如果是配置文件优先字段，永远不使用默认值覆盖
					if isConfigPriority {
						// 不做任何操作，保持配置文件中的值，即使它是零值
					} else if isZeroValue(field) && !isZeroValue(defaultField) {
						// 如果用户配置中的字段是零值，而默认配置中的字段不是零值，则使用默认值
						field.Set(defaultField)
					}
				}
			}
		}
	case reflect.Slice:
		// 对于切片，如果用户配置为空则使用默认值
		if userVal.Len() == 0 && defaultVal.Len() > 0 {
			userVal.Set(defaultVal)
		}
		// 其他类型不需要特殊处理
	}
}

// isZeroValue 检查值是否为零值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan:
		return v.IsNil()
	case reflect.Struct:
		// 对于结构体，检查所有字段是否都是零值
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// 对于其他类型，使用reflect.Zero比较
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func initConf() *config.Config {
	defaultConfig := getDefaultConfig()
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

	// 先将配置解析到空结构体中
	err = yaml.Unmarshal(yamlConf, c)
	if err != nil {
		log.Fatalf("config Init Unmarshal: %v", err)
	}

	// 使用默认配置填充未设置的字段
	mergeWithDefaults(defaultConfig, c)

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

// 恢复config默认值
func ResetConfig() {
	defaultConfig := getDefaultConfig()
	Config = defaultConfig
	SaveConfig()
}

// GetConfig 获取全局配置实例
func GetConfig() *config.Config {
	return Config
}

func GetConfig_ServerFileManagementRootPath() []string {
	if Config.ServerFileManagement.RootPath == nil || len(Config.ServerFileManagement.RootPath) == 0 {
		defaultConfig := getDefaultConfig()
		return defaultConfig.ServerFileManagement.RootPath
	}
	return Config.ServerFileManagement.RootPath
}
