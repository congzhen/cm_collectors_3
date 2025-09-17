package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config 配置结构体
type Config struct {
	PlayerPath string `json:"palyerPath"`
}

// readConfig 读取配置文件并解析到结构体
func ReadConfig(filename string) (*Config, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 JSON 数据到结构体
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}
