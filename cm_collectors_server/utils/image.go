package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ImageToBase64 将图片文件转换为base64编码的字符串
// 参数:
//   - imagePath: 图片文件的路径
//
// 返回值:
//   - string: base64编码的图片数据，格式为data:image/jpeg;base64,xxx
//   - error: 错误信息，如果转换成功则为nil
func ImageToBase64(imagePath string) (string, error) {
	content, err := ImageToBytes(imagePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}
	return ImageBytesToBase64(content)
}

// ImageToBytes 将指定路径的图片文件读取为字节切片
// 参数:
//
//	imagePath - 图片文件的路径
//
// 返回值:
//
//	[]byte - 图片文件的字节数据
//	error - 读取过程中发生的错误，如果文件不存在则返回文件不存在错误
func ImageToBytes(imagePath string) ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", imagePath)
	}
	// 读取图片文件内容
	return os.ReadFile(imagePath)
}

// ImageBytesToBase64 将图片字节数据转换为base64编码的字符串
// 参数:
//   - imageBytes: 图片的字节数据
//
// 返回值:
//   - string: base64编码的图片数据，格式为data:image/jpeg;base64,xxx
//   - error: 错误信息，如果转换成功则为nil
func ImageBytesToBase64(imageBytes []byte) (string, error) {
	// 检查输入数据是否为空
	if len(imageBytes) == 0 {
		return "", errors.New("图片数据为空")
	}

	// 自动检测MIME类型
	mimeType := http.DetectContentType(imageBytes[:min(len(imageBytes), 512)])

	// 如果无法检测到MIME类型，默认使用jpeg
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	// 编码为 Base64 字符串
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	// 组合成完整的 Data URL 格式
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// GetImageDimensionsFromBytes 从图片字节数据获取图片的宽度和高度
// 参数:
//   - imageBytes: 图片的字节数据
//
// 返回值:
//   - int: 图片宽度
//   - int: 图片高度
//   - error: 错误信息，如果获取成功则为nil
func GetImageDimensionsFromBytes(imageBytes []byte) (int, int, error) {
	if len(imageBytes) == 0 {
		return 0, 0, errors.New("图片数据为空")
	}

	// 使用bytes.Reader来读取图片数据
	reader := bytes.NewReader(imageBytes)

	// 解码图片配置信息（只解码图片尺寸，不解码全部像素数据，提高性能）
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, fmt.Errorf("解码图片配置失败: %v", err)
	}

	return config.Width, config.Height, nil
}

// ResizeImageByMaxWidth 将图像数据按指定最大宽度进行缩放，保持宽高比
// 参数：
//   - data: 原始图像的字节数据
//   - maxWidth: 图像缩放后的最大宽度，若小于等于0则不进行缩放
//
// 返回值：
//   - []byte: 缩放后图像的字节数据
//   - int: 缩放后图像的宽度
//   - int: 缩放后图像的高度
//   - error: 如果在解码、缩放或编码过程中发生错误，则返回相应的错误信息
func ResizeImageByMaxWidth(data []byte, maxWidth int) ([]byte, int, int, error) {
	if maxWidth <= 0 {
		// 如果最大宽度无效，则不进行缩放
		img, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, 0, 0, err
		}
		return data, img.Width, img.Height, nil
	}

	// 将字节数据解码为图像配置以获取原始尺寸
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, err
	}

	srcWidth := config.Width
	srcHeight := config.Height

	// 如果原始宽度小于等于指定宽度，直接返回原数据和尺寸
	if srcWidth <= maxWidth {
		return data, srcWidth, srcHeight, nil
	}

	// 计算缩放后的尺寸
	ratio := float64(maxWidth) / float64(srcWidth)
	newWidth := maxWidth
	newHeight := int(float64(srcHeight) * ratio)

	// 使用现有的 ScaleImage 函数进行缩放，使用中等质量
	scaledData, err := ScaleImage(data, maxWidth, 2)
	if err != nil {
		return nil, 0, 0, err
	}

	return scaledData, newWidth, newHeight, nil
}

func SaveBase64AsImage(base64Str, filePath string, hasMimeType bool) error {
	// 定义支持的 MIME 类型
	validMimeTypes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
	}
	var replaced string
	if hasMimeType {
		// 查找匹配的 MIME 类型并移除前缀
		for _, mimeType := range validMimeTypes {
			if strings.HasPrefix(base64Str, mimeType) {
				replaced = strings.Replace(base64Str, mimeType, "", 1)
				break
			}
		}
	} else {
		replaced = base64Str

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

// ResizeAndCropImage 将图像数据按指定的宽度和高度进行裁剪和缩放
// 如果源图像尺寸小于目标尺寸，将会放大；如果大于目标尺寸，将会缩小
// 当源图像的宽高比与目标尺寸不匹配时，会以中心为基准裁剪多余部分，确保最终图像完全符合指定尺寸
// 参数：
//   - data: 原始图像的字节数据
//   - targetWidth: 目标图像宽度
//   - targetHeight: 目标图像高度
//
// 返回值：
//   - []byte: 调整尺寸后图像的字节数据
//   - error: 如果在解码、裁剪或编码过程中发生错误，则返回相应的错误信息
func ResizeAndCropImage(data []byte, targetWidth, targetHeight int) ([]byte, error) {
	if targetWidth <= 0 || targetHeight <= 0 {
		return nil, errors.New("目标宽度和高度必须大于0")
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

	// 计算目标比例
	targetRatio := float64(targetWidth) / float64(targetHeight)
	srcRatio := float64(srcWidth) / float64(srcHeight)

	// 计算裁剪后的尺寸
	var cropWidth, cropHeight int
	var cropX, cropY int

	if srcRatio > targetRatio {
		// 源图片更宽，需要裁剪宽度
		cropHeight = srcHeight
		cropWidth = int(float64(srcHeight) * targetRatio)
		cropX = (srcWidth - cropWidth) / 2
		cropY = 0
	} else {
		// 源图片更高，需要裁剪高度
		cropWidth = srcWidth
		cropHeight = int(float64(srcWidth) / targetRatio)
		cropX = 0
		cropY = (srcHeight - cropHeight) / 2
	}

	// 创建裁剪后的图像
	cropRect := image.Rect(0, 0, cropWidth, cropHeight)
	cropImg := image.NewRGBA(cropRect)
	draw.Draw(cropImg, cropRect, img, image.Point{cropX, cropY}, draw.Src)

	// 创建缩放后的图像
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// 使用中等质量进行缩放
	scaler := draw.ApproxBiLinear
	scaler.Scale(dst, dst.Bounds(), cropImg, cropImg.Bounds(), draw.Over, nil)

	// 创建缓冲区用于存储编码后的图像
	buf := new(bytes.Buffer)

	// 编码为JPEG格式
	err = jpeg.Encode(buf, dst, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
