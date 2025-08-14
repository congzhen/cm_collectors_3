package processorsffmpeg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
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
	images, err := t.ExtractHighQualityKeyframes(videoPath, frameCount, 1)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("未能从视频中提取到关键帧")
	}
	return images[0], nil
}

// ExtractKeyframesAsBase64 从视频中提取指定数量的关键帧并转换为base64编码的JPEG图像
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的关键帧数量
//	level: 质量级别，决定提取帧的数量（实际提取帧数为frameCount*level）
//
// 返回值:
//
//	[]string: base64编码的关键帧图像数据切片，每个元素是一帧图像的base64编码字符串
//	error: 错误信息，如果提取成功则为nil
func (t KeyFrame) ExtractKeyframesAsBase64(videoPath string, frameCount int, level int) ([]string, error) {
	var images [][]byte
	var err error
	if level > 1 {
		images, err = t.ExtractHighQualityKeyframes(videoPath, frameCount, level)
	} else {
		images, err = t.ExtractKeyframes(videoPath, frameCount)
	}
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

// ExtractHighQualityKeyframes 提取高质量关键帧
// 通过提取比需求更多的帧，然后根据质量评估选择最佳的帧来提高关键帧质量
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的关键帧数量
//	level: 质量级别，决定提取帧的数量（实际提取帧数为frameCount*level）
//
// 返回值:
//
//	[][]byte: 提取的高质量关键帧图像数据切片
//	error: 错误信息，如果提取成功则为nil
func (t KeyFrame) ExtractHighQualityKeyframes(videoPath string, frameCount int, level int) ([][]byte, error) {
	if level < 1 {
		level = 1
	}
	images, err := t.ExtractKeyframes(videoPath, frameCount*level)
	if err != nil {
		return nil, err
	}
	// 如果需要多帧，从中选择指定数量的帧
	// 先评估所有帧的质量
	type frameScore struct {
		index int
		score float64
		data  []byte
		err   error
	}

	// 创建通道用于接收评估结果
	scoreChan := make(chan frameScore, len(images))

	// 启动协程并行评估所有帧
	for i, imgData := range images {
		go func(index int, data []byte) {
			score, err := t.evaluateFrame(data)
			if err != nil {
				// 如果评估失败，给一个较低的分数
				score = 0.1
			}
			scoreChan <- frameScore{
				index: index,
				score: score,
				data:  data,
				err:   err,
			}
		}(i, imgData)
	}

	// 收集所有评估结果
	var scoredFrames []frameScore
	for i := 0; i < len(images); i++ {
		scoredFrames = append(scoredFrames, <-scoreChan)
	}

	// 按分数排序（降序）
	for i := 0; i < len(scoredFrames)-1; i++ {
		for j := i + 1; j < len(scoredFrames); j++ {
			if scoredFrames[i].score < scoredFrames[j].score {
				scoredFrames[i], scoredFrames[j] = scoredFrames[j], scoredFrames[i]
			}
		}
	}

	// 选择前frameCount个帧
	var selectedImages [][]byte
	for i := 0; i < frameCount && i < len(scoredFrames); i++ {
		selectedImages = append(selectedImages, scoredFrames[i].data)
	}

	// 如果选择的帧数不够，用原始帧补充
	for len(selectedImages) < frameCount && len(images) > len(selectedImages) {
		selectedImages = append(selectedImages, images[len(selectedImages)])
	}

	return selectedImages, nil
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
	blackWhitePenalty := blackRatio + whiteRatio

	// 更严格的过滤条件
	if blackWhitePenalty > 0.8 { // 如果黑/白像素超过80%
		return -1, nil // 直接排除
	}

	// 对淡入淡出效果进行特殊处理
	if blackRatio > 0.5 || whiteRatio > 0.5 {
		// 不直接排除，但给予很低的分数
		blackWhitePenalty = math.Max(blackWhitePenalty, 0.7)
	}

	// 颜色丰富度得分 (0-1)
	colorScore := 1.0 - math.Min(1.0, blackWhitePenalty*1.8) // 调整系数以更敏感

	// 颜色变化得分 (0-1)
	varianceScore := math.Min(1.0, colorVariance/8000.0) // 调整分母使变化更敏感

	// 计算图像清晰度得分（模糊检测）
	sharpnessScore := t.calculateImageSharpness(img, sampleRate)

	// 增加对比度评估
	contrastScore := t.calculateContrast(img, sampleRate)

	// 增加边缘密度评估
	edgeDensityScore := t.calculateEdgeDensity(img, sampleRate)

	// 综合得分 (加入清晰度、对比度和边缘密度权重)
	// 现在考虑颜色、方差、清晰度、对比度和边缘密度五个因素
	score := (colorScore*0.15 + varianceScore*0.1 + sharpnessScore*0.4 + contrastScore*0.2 + edgeDensityScore*0.15)

	// 对于非常模糊的图像，直接排除
	if sharpnessScore < 0.05 {
		return -1, nil
	}

	// 对于颜色单调且不够清晰的图像，也排除
	if colorScore < 0.2 && sharpnessScore < 0.2 {
		return -1, nil
	}

	return score, nil
}

// calculateContrast 计算图像对比度得分
func (t KeyFrame) calculateContrast(img image.Image, sampleRate int) float64 {
	bounds := img.Bounds()

	// 计算直方图
	var histogram [256]int
	totalPixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, _ := img.At(x, y).RGBA()
			// 转换为8位灰度值
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			gray := uint8(0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8))

			histogram[gray]++
			totalPixels++
		}
	}

	// 计算累积直方图
	var cumHist [256]int
	cumHist[0] = histogram[0]
	for i := 1; i < 256; i++ {
		cumHist[i] = cumHist[i-1] + histogram[i]
	}

	// 找到5%和95%的分位点
	lowThresh := int(0.05 * float64(totalPixels))
	highThresh := int(0.95 * float64(totalPixels))

	var lowVal, highVal int
	for i := 0; i < 256; i++ {
		if cumHist[i] >= lowThresh && lowVal == 0 {
			lowVal = i
		}
		if cumHist[i] >= highThresh {
			highVal = i
		}
	}

	// 计算对比度
	contrast := float64(highVal-lowVal) / 255.0
	return math.Max(0.0, math.Min(1.0, contrast))
}

// calculateEdgeDensity 计算图像边缘密度得分
func (t KeyFrame) calculateEdgeDensity(img image.Image, sampleRate int) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width < 3 || height < 3 {
		return 0.5
	}

	// 使用Sobel算子计算边缘
	var edgePixels, totalPixels int

	// Sobel核:
	// Gx = [[-1, 0, 1], [-2, 0, 2], [-1, 0, 1]]
	// Gy = [[-1, -2, -1], [0, 0, 0], [1, 2, 1]]
	for y := bounds.Min.Y + sampleRate; y < bounds.Max.Y-sampleRate; y += sampleRate {
		for x := bounds.Min.X + sampleRate; x < bounds.Max.X-sampleRate; x += sampleRate {
			// 计算x方向梯度
			gx := (2 * t.getGrayValue(img, x+sampleRate, y)) +
				t.getGrayValue(img, x+sampleRate, y-sampleRate) +
				t.getGrayValue(img, x+sampleRate, y+sampleRate) -
				(2 * t.getGrayValue(img, x-sampleRate, y)) -
				t.getGrayValue(img, x-sampleRate, y-sampleRate) -
				t.getGrayValue(img, x-sampleRate, y+sampleRate)

			// 计算y方向梯度
			gy := (2 * t.getGrayValue(img, x, y+sampleRate)) +
				t.getGrayValue(img, x+sampleRate, y+sampleRate) +
				t.getGrayValue(img, x-sampleRate, y+sampleRate) -
				(2 * t.getGrayValue(img, x, y-sampleRate)) -
				t.getGrayValue(img, x+sampleRate, y-sampleRate) -
				t.getGrayValue(img, x-sampleRate, y-sampleRate)

			// 计算梯度幅值
			gradient := math.Sqrt(float64(gx*gx + gy*gy))

			// 如果梯度大于阈值，则认为是边缘
			if gradient > 30.0 { // 阈值可根据需要调整
				edgePixels++
			}

			totalPixels++
		}
	}

	if totalPixels == 0 {
		return 0.5
	}

	// 边缘密度得分
	edgeDensity := float64(edgePixels) / float64(totalPixels)

	// 将边缘密度映射到0-1范围
	// 假设理想的边缘密度在0.05-0.3之间
	if edgeDensity < 0.05 {
		return edgeDensity * 20.0 // 低于0.05时线性增加
	} else if edgeDensity > 0.3 {
		return math.Max(0.0, 1.0-(edgeDensity-0.3)*2.0) // 高于0.3时递减
	}

	// 0.05-0.3之间时，得分接近1
	return 1.0
}

// calculateImageSharpness 计算图像清晰度得分，用于检测模糊
func (t KeyFrame) calculateImageSharpness(img image.Image, sampleRate int) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width < 3 || height < 3 {
		return 0.5 // 太小的图像无法准确计算清晰度
	}

	// 转换为灰度图像并计算梯度
	var sobelV, sobelH float64
	count := 0

	// 使用Sobel算子计算图像梯度
	// Sobel核:
	// Gx = [[-1, 0, 1], [-2, 0, 2], [-1, 0, 1]]
	// Gy = [[-1, -2, -1], [0, 0, 0], [1, 2, 1]]
	for y := bounds.Min.Y + sampleRate; y < bounds.Max.Y-sampleRate; y += sampleRate {
		for x := bounds.Min.X + sampleRate; x < bounds.Max.X-sampleRate; x += sampleRate {
			// 计算x方向梯度
			gx := (2 * t.getGrayValue(img, x+sampleRate, y)) +
				t.getGrayValue(img, x+sampleRate, y-sampleRate) +
				t.getGrayValue(img, x+sampleRate, y+sampleRate) -
				(2 * t.getGrayValue(img, x-sampleRate, y)) -
				t.getGrayValue(img, x-sampleRate, y-sampleRate) -
				t.getGrayValue(img, x-sampleRate, y+sampleRate)

			// 计算y方向梯度
			gy := (2 * t.getGrayValue(img, x, y+sampleRate)) +
				t.getGrayValue(img, x+sampleRate, y+sampleRate) +
				t.getGrayValue(img, x-sampleRate, y+sampleRate) -
				(2 * t.getGrayValue(img, x, y-sampleRate)) -
				t.getGrayValue(img, x+sampleRate, y-sampleRate) -
				t.getGrayValue(img, x-sampleRate, y-sampleRate)

			// 累积梯度幅值
			sobelV += math.Abs(float64(gx))
			sobelH += math.Abs(float64(gy))
			count++
		}
	}

	if count == 0 {
		return 0.5
	}

	// 计算平均梯度幅值
	avgGradient := (sobelV + sobelH) / float64(count*255) // 归一化到0-1范围

	// 使用拉普拉斯算子作为补充检测
	laplacianScore := t.calculateLaplacianVariance(img, sampleRate)

	// 综合两个指标
	combinedScore := 0.7*avgGradient + 0.3*laplacianScore

	// 保证得分在0-1之间
	return math.Max(0.0, math.Min(1.0, combinedScore))
}

// calculateLaplacianVariance 计算拉普拉斯方差作为模糊检测的补充方法
func (t KeyFrame) calculateLaplacianVariance(img image.Image, sampleRate int) float64 {
	bounds := img.Bounds()
	var sum, sumSq float64
	count := 0

	// 使用拉普拉斯算子计算每个像素的二阶导数
	// 简化的3x3拉普拉斯核 [[0,1,0],[1,-4,1],[0,1,0]]
	for y := bounds.Min.Y + sampleRate; y < bounds.Max.Y-sampleRate; y += sampleRate {
		for x := bounds.Min.X + sampleRate; x < bounds.Max.X-sampleRate; x += sampleRate {
			// 获取当前像素及其邻居的灰度值
			center := t.getGrayValue(img, x, y)
			up := t.getGrayValue(img, x, y-sampleRate)
			down := t.getGrayValue(img, x, y+sampleRate)
			left := t.getGrayValue(img, x-sampleRate, y)
			right := t.getGrayValue(img, x+sampleRate, y)

			// 应用拉普拉斯算子: 4*center - up - down - left - right
			laplacian := math.Abs(float64(4*center - up - down - left - right))

			sum += laplacian
			sumSq += laplacian * laplacian
			count++
		}
	}

	if count == 0 {
		return 0.5
	}

	// 计算方差作为清晰度指标
	mean := sum / float64(count)
	variance := sumSq/float64(count) - mean*mean

	// 将方差映射到0-1范围
	// 这里使用不同的缩放因子以适应拉普拉斯方差的范围
	sharpness := 1.0 - math.Exp(-variance/5000.0)
	return math.Max(0.0, math.Min(1.0, sharpness))
}

// getGrayValue 获取指定坐标的灰度值
func (t KeyFrame) getGrayValue(img image.Image, x, y int) uint8 {
	r, g, b, _ := img.At(x, y).RGBA()
	// 转换为8位值
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	// 使用标准的RGB转灰度公式
	gray := uint8(0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8))
	return gray
}
