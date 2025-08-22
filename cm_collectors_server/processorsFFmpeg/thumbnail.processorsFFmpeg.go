package processorsffmpeg

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

type Thumbnail struct{}

// ExtractThumbnailPoster 从视频中提取一张关键帧作为海报图像
// 参数:
//
//	videoPath: 视频文件的路径
//
// 返回值:
//
//	[]byte: JPEG格式的图像数据
//	error: 错误信息，如果提取成功则为nil，如果未能提取到关键帧则返回错误
func (t Thumbnail) ExtractThumbnailPoster(videoPath string) ([]byte, error) {
	images, err := t.ExtractThumbnail(videoPath, 1, 100)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("未能从视频中提取到关键帧")
	}
	return images[0], nil
}

// ExtractThumbnailAsBase64 从视频中提取指定数量的缩略图并以base64编码格式返回
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的帧数量
//	level: 质量级别，决定提取帧的数量（实际提取帧数为frameCount*level）值越大速度越慢，1为不赛选高质量
//
// 返回值:
//
//	[]string: 包含base64编码图像数据的字符串切片，每个元素都带有"data:image/jpeg;base64,"前缀
//	error: 错误信息，如果提取成功则为nil
func (t Thumbnail) ExtractThumbnailAsBase64(videoPath string, frameCount int, level int) ([]string, error) {
	var images [][]byte
	var err error
	if level > 1 {
		images, err = t.ExtractHighQualityThumbnail(videoPath, frameCount, level)
	} else {
		images, err = t.ExtractThumbnail(videoPath, frameCount, 100)
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

// ExtractHighQualityThumbnail 从视频中提取高质量缩略图
// 通过提取比需求更多的帧，然后根据质量评估选择最佳的帧来提高缩略图质量
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的缩略图数量
//	level: 质量级别，决定提取帧的数量（实际提取帧数为frameCount*level）
//
// 返回值:
//
//	[][]byte: 提取的高质量缩略图图像数据切片
//	error: 错误信息，如果提取成功则为nil
func (t Thumbnail) ExtractHighQualityThumbnail(videoPath string, frameCount int, level int) ([][]byte, error) {
	if level < 1 {
		level = 1
	}
	images, err := t.ExtractThumbnail(videoPath, frameCount*level, 50)
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
			score, err := KeyFrame{}.evaluateFrame(data)
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

// ExtractThumbnail 从视频中提取指定数量的缩略图
// 参数:
//
//	videoPath: 视频文件的路径
//	frameCount: 需要提取的帧数量
//	thumbnailFilter: 缩略图过滤器参数，用于控制帧选择算法的敏感度
//
// 返回值:
//
//	[][]byte: 包含JPEG格式图像数据的二维字节切片
//	error: 错误信息，如果提取成功则为nil
func (t Thumbnail) ExtractThumbnail(videoPath string, frameCount, thumbnailFilter int) ([][]byte, error) {
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
	cmd := createCommand(
		ffmpegPath,
		"-i", videoPath, // 输入视频文件
		"-vf", fmt.Sprintf("thumbnail=%d", thumbnailFilter), // 分析每100帧选择一个最具代表性的帧
		//"-vf", "select=eq(pict_type\\,I)", // 只选择关键帧
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
func (t Thumbnail) splitMJPEGStream(data []byte) [][]byte {
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
