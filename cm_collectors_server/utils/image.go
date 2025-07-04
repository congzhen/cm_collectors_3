package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SaveBase64AsImage(base64Str, filePath string) error {
	// 定义支持的 MIME 类型
	validMimeTypes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
	}

	// 查找匹配的 MIME 类型并移除前缀
	var replaced string
	for _, mimeType := range validMimeTypes {
		if strings.HasPrefix(base64Str, mimeType) {
			replaced = strings.Replace(base64Str, mimeType, "", 1)
			break
		}
	}

	// 如果没有匹配的 MIME 类型，返回错误
	if replaced == "" {
		return errors.New("unsupported or invalid base64 string")
	}

	// 解码 Base64 数据
	data, err := base64.StdEncoding.DecodeString(replaced)
	if err != nil {
		return err
	}

	// 检查文件夹是否存在，不存在则创建
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 存储为文件
	err = os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
