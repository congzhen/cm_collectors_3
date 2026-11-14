package processorsffmpeg

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"
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
	return t.ExtractThumbnailPosterContext(context.Background(), videoPath)
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
	for len(data) > 0 {
		start := bytes.Index(data, []byte{0xff, 0xd8})
		if start < 0 {
			break
		}
		data = data[start:]
		end := bytes.Index(data[2:], []byte{0xff, 0xd9})
		if end < 0 {
			break
		}
		end += 4
		images = append(images, data[:end])
		data = data[end:]
	}
	return images
}

var posterSlots = make(chan struct{}, 1)

// Import posters retain representative selection across 100 frames. Scale before
// buffering the candidates, not after exporting an original-size JPEG.
func (t Thumbnail) ExtractThumbnailPosterContext(parent context.Context, videoPath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case posterSlots <- struct{}{}:
		defer func() { <-posterSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	tool, err := FFmpeg{}.isToolAvailableContext(ctx, "ffmpeg")
	if err != nil {
		return nil, err
	}
	cmd := createCommandContext(ctx, tool, "-hide_banner", "-loglevel", "error", "-nostdin", "-threads", "1", "-i", videoPath,
		"-vf", "scale=w='min(400,iw)':h='min(600,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2,thumbnail=100", "-filter_threads", "1", "-frames:v", "1", "-f", "image2pipe", "-vcodec", "mjpeg", "-threads", "1", "-q:v", "2", "pipe:1")
	cmd.WaitDelay = 2 * time.Second
	stdout := &boundedOutput{limit: 8 * 1024 * 1024, cancel: cancel}
	stderr := &boundedOutput{limit: 16 * 1024}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if stdout.exceeded {
		return nil, fmt.Errorf("封面输出超过 8 MiB")
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("封面生成超时或已取消: %w", ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("FFmpeg执行失败: %w; %s", err, stderr.String())
	}
	images := t.splitMJPEGStream(stdout.Bytes())
	if len(images) == 0 {
		return nil, fmt.Errorf("未生成完整封面")
	}
	return images[0], nil
}

// Drain excess diagnostic output without retaining it; cancel oversized image output.
type boundedOutput struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
	cancel   context.CancelFunc
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.Len()
	if n > remaining {
		b.exceeded = true
		if b.cancel != nil {
			b.cancel()
		}
	}
	if remaining > 0 {
		keep := n
		if keep > remaining {
			keep = remaining
		}
		_, _ = b.buffer.Write(p[:keep])
	}
	return n, nil
}

// Do not embed bytes.Buffer: its ReaderFrom would bypass Write in io.Copy.
func (b *boundedOutput) Len() int       { return b.buffer.Len() }
func (b *boundedOutput) Bytes() []byte  { return b.buffer.Bytes() }
func (b *boundedOutput) String() string { return b.buffer.String() }
