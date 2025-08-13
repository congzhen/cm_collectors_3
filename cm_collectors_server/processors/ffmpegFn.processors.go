package processors

import (
	"bytes"
	"fmt"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type FFmpeg struct {
}

// IsFFmpegAvailable 检查系统中FFmpeg是否可用
//
// 该函数会根据不同的操作系统查找ffmpeg可执行文件：
//   - 在Windows系统中，首先检查程序根目录下的ffmpeg文件夹，
//     如果未找到，则检查系统PATH环境变量中的ffmpeg.exe或ffmpeg
//   - 在Unix-like系统中，直接检查系统PATH环境变量中的ffmpeg
//
// 返回值:
//
//	string: 找到的ffmpeg可执行文件的完整路径
//	error: 如果未找到ffmpeg或无法执行，则返回错误信息
func (f FFmpeg) IsFFmpegAvailable() (string, error) {
	// 在不同系统中检查ffmpeg
	var cmd *exec.Cmd
	var ffmpegPath string

	// 根据操作系统选择命令
	switch runtime.GOOS {
	case "windows":
		// Windows系统优先检查当前程序根目录的ffmpeg文件夹
		exePath, err := os.Executable()
		if err == nil {
			// 获取程序根目录
			rootDir := filepath.Dir(filepath.Dir(exePath))
			// 检查ffmpeg文件夹下的ffmpeg.exe
			ffmpegLocalPath := filepath.Join(rootDir, "ffmpeg", "ffmpeg.exe")
			if _, err := os.Stat(ffmpegLocalPath); err == nil {
				// 本地ffmpeg存在
				ffmpegPath = ffmpegLocalPath
				cmd = exec.Command(ffmpegPath, "-version")
				break
			}
		}

		// 如果本地没有，检查系统PATH中的ffmpeg.exe
		ffmpegPath, err = exec.LookPath("ffmpeg.exe")
		if err != nil {
			// 尝试不带.exe后缀的版本
			ffmpegPath, err = exec.LookPath("ffmpeg")
			if err != nil {
				return "", fmt.Errorf("在Windows系统中未找到FFmpeg: %v", err)
			}
		}
		cmd = exec.Command(ffmpegPath, "-version")
	default:
		// Unix-like系统 (Linux, macOS等)
		var err error
		ffmpegPath, err = exec.LookPath("ffmpeg")
		if err != nil {
			return "", fmt.Errorf("在Unix-like系统中未找到FFmpeg: %v", err)
		}
		cmd = exec.Command(ffmpegPath, "-version")
	}

	// 尝试运行命令
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("FFmpeg无法执行: %v", err)
	}

	return ffmpegPath, nil
}

// ExtractKeyframePoster 从视频中提取关键帧作为海报，返回图像字节数据
func (f FFmpeg) ExtractKeyframePoster(videoPath string) ([]byte, error) {
	// 检查FFmpeg是否可用
	ffmpegPath, err := f.IsFFmpegAvailable()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg不可用: %v", err)
	}

	// 构造FFmpeg命令，提取多个关键帧到内存
	// 使用复杂的FFmpeg命令直接输出到stdout
	cmd := exec.Command(
		ffmpegPath,
		"-i", videoPath, // 输入视频文件
		"-vf", "select=eq(pict_type\\,I)", // 只选择关键帧
		"-vframes", "5", // 最多提取5帧
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

	// 从输出中提取最佳帧
	bestFrameData, err := f.selectMeaningfulFrameFromData(imagesData)
	if err != nil {
		return nil, fmt.Errorf("选择有意义的帧失败: %v", err)
	}

	return bestFrameData, nil
}

// selectMeaningfulFrameFromData 从FFmpeg输出的数据中选择最有意义的帧
func (f FFmpeg) selectMeaningfulFrameFromData(data []byte) ([]byte, error) {
	// FFmpeg输出的多张图片连接在一起，需要分割处理
	images := f.splitMJPEGStream(data)

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
		score, err := f.evaluateFrame(imgData)
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
func (f FFmpeg) splitMJPEGStream(data []byte) [][]byte {
	var images [][]byte
	var start int

	// JPEG文件以0xFFD8开始，以0xFFD9结束
	for i := 0; i < len(data)-1; i++ {
		// 查找JPEG开始标记
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			start = i
			// 查找JPEG结束标记
			for j := i + 2; j < len(data)-1; j++ {
				if data[j] == 0xFF && data[j+1] == 0xD9 {
					// 找到完整JPEG图像
					images = append(images, data[start:j+2])
					i = j + 1 // 继续查找下一个图像
					break
				}
			}
		}
	}

	return images
}

// evaluateFrame 评估帧的质量分数
func (f FFmpeg) evaluateFrame(frameData []byte) (float64, error) {
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
