package processorsffmpeg

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
)

type FrameExtract struct{}

// ExtractFrameAt 提取视频指定时间点的一帧，返回 image.Image
// seconds: 从视频开始的秒数
func (f FrameExtract) ExtractFrameAt(videoPath string, seconds float64) (image.Image, error) {
	ffmpegPath, err := FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg不可用: %v", err)
	}

	cmd := createCommand(
		ffmpegPath,
		"-ss", fmt.Sprintf("%.3f", seconds),
		"-i", videoPath,
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", "5",
		"pipe:1",
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("提取帧失败 at %.3fs: %v, stderr: %s", seconds, err, stderrBuf.String())
	}

	img, err := jpeg.Decode(&stdoutBuf)
	if err != nil {
		return nil, fmt.Errorf("解码帧图像失败: %v", err)
	}
	return img, nil
}

// ExtractFramesAtPositions 按视频时长的百分比位置批量提取帧
// positions: 0.0~1.0 的比例列表，如 []float64{0.05, 0.15, ..., 0.95}
func (f FrameExtract) ExtractFramesAtPositions(videoPath string, duration float64, positions []float64) ([]image.Image, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("视频时长无效: %f", duration)
	}

	images := make([]image.Image, 0, len(positions))
	for _, pos := range positions {
		sec := duration * pos
		// 避免超出视频时长
		if sec >= duration {
			sec = duration * 0.99
		}
		img, err := f.ExtractFrameAt(videoPath, sec)
		if err != nil {
			return nil, fmt.Errorf("提取 %.0f%% 位置帧失败: %v", pos*100, err)
		}
		images = append(images, img)
	}
	return images, nil
}
