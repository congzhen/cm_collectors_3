package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
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

// ScaleImage 将图像数据按指定最大宽度进行缩放，并根据质量级别选择不同的采样算法。
// 参数：
//   - data: 原始图像的字节数据
//   - maxWidth: 图像缩放后的最大宽度，若小于等于0则不进行缩放
//   - level: 缩放质量级别，1表示低质量（快速），2表示中等质量，3表示高质量（较慢）
//
// 返回值：
//   - []byte: 缩放后图像的字节数据
//   - error: 如果在解码、缩放或编码过程中发生错误，则返回相应的错误信息
func ScaleImage(data []byte, maxWidth int, level int) ([]byte, error) {
	if maxWidth <= 0 {
		return data, nil
	}
	// 将字节数据解码为图像
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// 获取原始图像的边界
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// 如果原始宽度小于等于指定宽度，直接返回原数据
	if srcWidth <= maxWidth {
		return data, nil
	}

	// 计算缩放后的尺寸
	ratio := float64(maxWidth) / float64(srcWidth)
	newWidth := maxWidth
	newHeight := int(float64(srcHeight) * ratio)

	// 创建新的RGBA图像
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 根据质量级别选择不同的采样器
	var scaler draw.Scaler
	switch level {
	case 1:
		// 低质量，最快的速度
		scaler = draw.NearestNeighbor
	case 2:
		// 中等质量
		scaler = draw.ApproxBiLinear
	case 3:
		// 高质量，较慢的速度
		scaler = draw.CatmullRom
	default:
		// 默认使用中等质量
		scaler = draw.ApproxBiLinear
	}

	// 使用选定的采样器进行缩放
	scaler.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	// 创建缓冲区用于存储编码后的图像
	buf := new(bytes.Buffer)

	// 根据原始图像格式进行编码
	// 这里假设使用JPEG格式，实际应用中可能需要根据文件扩展名判断
	err = jpeg.Encode(buf, dst, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
