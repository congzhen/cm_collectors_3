package processorsffmpeg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"image/jpeg"
	"math"
	"os/exec"
)

type KeyFrame struct{}

// ExtractKeyframePoster 从视频中提取关键帧并选择最有意义的一帧作为海报图像
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的关键帧数量
//
// 返回值:
//
//	[]byte: 最佳关键帧图像的字节数据
//	error: 错误信息，如果提取成功则为nil
func (t KeyFrame) ExtractKeyframePoster(videoPath string, frameCount int) ([]byte, error) {
	images, err := t.ExtractKeyframes(videoPath, frameCount)
	if err != nil {
		return nil, err
	}
	// 从输出中提取最佳帧
	bestFrameData, err := t.selectMeaningfulFrameFromData(images)
	if err != nil {
		return nil, fmt.Errorf("选择有意义的帧失败: %v", err)
	}
	return bestFrameData, nil
}

// ExtractKeyframesAsBase64 从视频中提取指定数量的关键帧并转换为base64编码的JPEG图像
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的关键帧数量
//
// 返回值:
//
//	[]string: base64编码的关键帧图像数据切片，每个元素是一帧图像的base64编码字符串
//	error: 错误信息，如果提取成功则为nil
func (t KeyFrame) ExtractKeyframesAsBase64(videoPath string, frameCount int) ([]string, error) {
	images, err := t.ExtractKeyframes(videoPath, frameCount)
	if err != nil {
		return nil, err
	}
	// 将图像转换为base64编码
	var base64Frames []string
	for _, imgData := range images {
		// 将图像数据编码为base64
		base64Str := base64.StdEncoding.EncodeToString(imgData)
		// 添加data URI前缀
		base64Frames = append(base64Frames, "data:image/jpeg;base64,"+base64Str)
	}

	return base64Frames, nil
}

// ExtractKeyframes 从视频中提取指定数量的关键帧
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的关键帧数量
//
// 返回值:
//
//	[][]byte: 提取的关键帧图像数据切片，每个元素是一帧图像的字节数据
//	error: 错误信息，如果提取成功则为nil
func (t KeyFrame) ExtractKeyframes(videoPath string, frameCount int) ([][]byte, error) {
	// 检查FFmpeg是否可用
	ffmpegPath, err := FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg不可用: %v", err)
	}

	// 限制帧数在合理范围内
	if frameCount <= 0 {
		frameCount = 1
	} else if frameCount > 20 {
		frameCount = 20 // 限制最大帧数为20
	}

	// 构造FFmpeg命令，提取指定数量的关键帧到内存
	cmd := exec.Command(
		ffmpegPath,
		"-i", videoPath, // 输入视频文件
		"-vf", "select=eq(pict_type\\,I)", // 只选择关键帧
		"-vframes", fmt.Sprintf("%d", frameCount), // 提取指定数量的帧
		"-f", "image2pipe", // 输出到管道
		"-vcodec", "mjpeg", // JPEG编码
		"-q:v", "2", // 高质量JPEG
		"-vsync", "vfr", // 变帧率同步
		"pipe:1", // 输出到stdout
	)

	// 执行命令并获取输出
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg执行失败: %v, stderr: %s", err, stderrBuf.String())
	}

	// 解析输出中的图像
	imagesData := stdoutBuf.Bytes()

	// 分割MJPEG流
	images := t.splitMJPEGStream(imagesData)

	// 检查是否成功提取到图像
	if len(images) == 0 {
		return nil, fmt.Errorf("未能从视频中提取到关键帧，stderr: %s", stderrBuf.String())
	}

	return images, nil
}

// selectMeaningfulFrameFromData 从FFmpeg输出的数据中选择最有意义的帧
func (t KeyFrame) selectMeaningfulFrameFromData(images [][]byte) ([]byte, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("未提取到任何关键帧")
	}

	// 如果只有一张图片，直接返回
	if len(images) == 1 {
		return images[0], nil
	}

	// 选择最佳图片
	var bestImage []byte
	bestScore := -1.0

	for _, imgData := range images {
		score, err := t.evaluateFrame(imgData)
		if err != nil {
			continue
		}

		if score > bestScore {
			bestScore = score
			bestImage = imgData
		}
	}

	if bestImage == nil {
		// 如果没有找到合适的帧，返回第一帧
		return images[0], nil
	}

	return bestImage, nil
}

// splitMJPEGStream 分割MJPEG流中的各个JPEG图像
func (t KeyFrame) splitMJPEGStream(data []byte) [][]byte {
	var images [][]byte

	// JPEG文件以0xFFD8开始，以0xFFD9结束
	i := 0
	for i < len(data)-1 {
		// 查找JPEG开始标记
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			start := i
			// 从开始标记后查找结束标记
			for j := i + 2; j < len(data)-1; j++ {
				if data[j] == 0xFF && data[j+1] == 0xD9 {
					// 找到完整JPEG图像
					images = append(images, data[start:j+2])
					i = j + 2 // 继续查找下一个图像
					break
				}
			}
			// 如果找到了JPEG结尾
			if len(images) > 0 && len(images[len(images)-1]) == len(data[start:][:len(images[len(images)-1])]) {
				// 已经处理了这部分数据
			} else {
				// 没有找到匹配的结尾，移动到下一个字节
				i++
			}
		} else {
			i++
		}
	}

	return images
}

// evaluateFrame 评估帧的质量分数
func (t KeyFrame) evaluateFrame(frameData []byte) (float64, error) {
	img, err := jpeg.Decode(bytes.NewReader(frameData))
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	totalPixels := width * height

	// 采样检查，避免处理太慢
	sampleRate := 1
	if totalPixels > 1000000 { // 如果超过100万像素
		sampleRate = int(math.Sqrt(float64(totalPixels) / 1000000))
	}

	var (
		blackPixels   int
		whitePixels   int
		totalSampled  int
		colorVariance float64
	)

	// 计算平均颜色
	var avgR, avgG, avgB float64
	var pixels []color.RGBA

	// 采样分析图像
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, _ := img.At(x, y).RGBA()
			// 转换为8位值
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			pixels = append(pixels, color.RGBA{r8, g8, b8, 255})

			// 计算平均值
			avgR += float64(r8)
			avgG += float64(g8)
			avgB += float64(b8)

			// 检查是否为纯黑或纯白
			if r8 < 30 && g8 < 30 && b8 < 30 {
				blackPixels++
			} else if r8 > 225 && g8 > 225 && b8 > 225 {
				whitePixels++
			}

			totalSampled++
		}
	}

	avgR /= float64(totalSampled)
	avgG /= float64(totalSampled)
	avgB /= float64(totalSampled)

	// 计算颜色方差
	for _, pixel := range pixels {
		dr := float64(pixel.R) - avgR
		dg := float64(pixel.G) - avgG
		db := float64(pixel.B) - avgB
		colorVariance += (dr*dr + dg*dg + db*db) / 3
	}
	colorVariance /= float64(totalSampled)

	// 计算黑色和白色像素比例
	blackRatio := float64(blackPixels) / float64(totalSampled)
	whiteRatio := float64(whitePixels) / float64(totalSampled)

	// 评估分数：
	// 1. 惩罚过多的黑/白像素
	// 2. 奖励适度的颜色变化
	// 3. 惩罚过于单调的图像
	blackWhitePenalty := blackRatio + whiteRatio
	if blackWhitePenalty > 0.6 { // 如果黑/白像素超过60%
		return -1, nil // 直接排除
	}

	// 颜色丰富度得分 (0-1)
	colorScore := 1.0 - math.Min(1.0, blackWhitePenalty*2)

	// 颜色变化得分 (0-1)
	varianceScore := math.Min(1.0, colorVariance/10000.0)

	// 综合得分
	score := (colorScore*0.6 + varianceScore*0.4)

	return score, nil
}
