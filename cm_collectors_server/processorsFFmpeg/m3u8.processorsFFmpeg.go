package processorsffmpeg

import (
	"fmt"
	"math"
	"os"
	"os/exec"
)

type M3U8 struct {
}

// PlayVideoM3u8 流式传输HLS视频片段
func (v M3U8) PlayVideoM3u8(videoSrc string, start, duration float64, checkPath, transcode bool) (*exec.Cmd, error) {
	// 使用FFmpeg生成视频片段
	ffmpegPath, err := FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg不可用: %v", err)
	}
	if checkPath {
		// 检查文件是否存在
		if _, err := os.Stat(videoSrc); os.IsNotExist(err) {
			return nil, fmt.Errorf("视频文件不存在: %s", videoSrc)
		}
	}
	var cmd *exec.Cmd
	if transcode {
		cmd = createCommand(
			ffmpegPath,
			"-ss", fmt.Sprintf("%.6f", start), // seek到指定时间点
			"-i", videoSrc, // 输入文件
			"-t", fmt.Sprintf("%.6f", duration), // 指定时长
			"-c:v", "libx264", // 视频编码
			"-c:a", "aac", // 音频编码
			"-f", "mpegts", // HLS格式输出
			"-bsf:v", "h264_mp4toannexb", // H.264比特流过滤器
			"-bsf:a", "aac_adtstoasc", // AAC音频过滤器
			"-preset", "ultrafast", // 编码速度优化
			"-tune", "zerolatency", // 零延迟优化
			"-output_ts_offset", fmt.Sprintf("%.6f", start), // 设置时间戳偏移
			"-threads", "0", // 使用所有可用线程
			"pipe:1", // 输出到标准输出
		)
	} else {
		cmd = createCommand(
			ffmpegPath,
			"-ss", fmt.Sprintf("%.6f", start), // seek到指定时间点
			"-i", videoSrc, // 输入文件
			"-t", fmt.Sprintf("%.6f", duration), // 指定时长
			"-c", "copy", // 直接复制流，不重新编码
			"-f", "mpegts", // HLS格式输出
			"-output_ts_offset", fmt.Sprintf("%.6f", start), // 设置时间戳偏移
			"pipe:1", // 输出到标准输出
		)
	}

	return cmd, nil
}

// GenerateM3u8File 生成m3u8文件并返回内容
func (v M3U8) GenerateM3u8File(dramaSeriesId string, duration float64) ([]byte, error) {
	// 使用固定时间间隔分割视频，提高性能
	segmentDuration := 10.0 // 每个片段10秒
	var keyframes []float64

	// 生成固定时间间隔的关键帧点
	currentTime := 0.0
	for currentTime < duration {
		keyframes = append(keyframes, currentTime)
		currentTime += segmentDuration
	}

	// 确保包含视频结尾
	if len(keyframes) == 0 || keyframes[len(keyframes)-1] < duration {
		keyframes = append(keyframes, duration)
	}

	// 构建片段信息
	type fragment struct {
		Duration float64
		Start    float64
	}

	var fragments []fragment

	// 为每两个相邻时间点之间创建一个片段
	for i := 0; i < len(keyframes)-1; i++ {
		startTime := keyframes[i]
		endTime := keyframes[i+1]
		segDuration := endTime - startTime

		// 确保片段时长有效
		if segDuration > 0 && segDuration <= duration {
			fragments = append(fragments, fragment{
				Duration: v.numberDecimal(segDuration),
				Start:    v.numberDecimal(startTime),
			})
		}
	}

	// 计算最大片段时长
	maxDuration := segmentDuration

	// 生成m3u8内容
	m3u8Data := "#EXTM3U\n"
	m3u8Data += "#EXT-X-VERSION:3\n"
	m3u8Data += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	m3u8Data += fmt.Sprintf("#EXT-X-TARGETDURATION:%.0f\n", math.Ceil(maxDuration))
	m3u8Data += "#EXT-X-MEDIA-SEQUENCE:0\n"

	for i, f := range fragments {
		m3u8Data += fmt.Sprintf("#EXTINF:%.3f,\n", f.Duration)
		m3u8Data += fmt.Sprintf("/api/video/m3u8/stream/%s/%.3f/%.3f/%d.ts\n", dramaSeriesId, f.Start, f.Duration, i+1)
	}

	m3u8Data += "#EXT-X-ENDLIST\n"

	// 打印调试信息
	fmt.Printf("使用固定时间间隔生成m3u8: 片段数=%d, 视频总时长=%.2f秒\n", len(fragments), duration)

	return []byte(m3u8Data), nil
}

// numberDecimal 保留小数位数
func (v M3U8) numberDecimal(value float64) float64 {
	return math.Round(value*1000) / 1000
}
