package processorsffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type M3U8 struct {
}

// PlayVideoM3u8 流式传输HLS视频片段
// PlayVideoM3u8 流式传输HLS视频片段
func (v M3U8) PlayVideoM3u8(videoSrc string, start, duration float32) (*exec.Cmd, io.ReadCloser, error) {
	// 使用FFmpeg生成视频片段
	ffmpegPath, err := FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		return nil, nil, fmt.Errorf("FFmpeg不可用: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(videoSrc); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("视频文件不存在: %s", videoSrc)
	}

	// 构建FFmpeg命令参数
	// 使用mpegts格式直接输出到管道
	cmd := createCommand(
		ffmpegPath,
		"-ss", fmt.Sprintf("%.3f", start),
		"-i", videoSrc,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c", "copy", // 复制所有流，不做重新编码
		"-f", "mpegts", // 使用mpegts格式
		"-mpegts_flags", "+initial_discontinuity",
		"-bsf:v", "h264_mp4toannexb",
		"pipe:1",
	)

	// 获取stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("无法创建stdout管道: %v", err)
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("无法启动FFmpeg: %v", err)
	}

	return cmd, stdout, nil
}

// createM3u8File 创建m3u8文件并返回内容
func (v M3U8) CreateM3u8File(dramaSeriesId string, videoSrc string, m3u8Path string) ([]byte, error) {
	// 使用FFmpeg获取关键帧信息
	ffprobePath, err := FFmpeg{}.IsFFprobeAvailable()
	if err != nil {
		return nil, fmt.Errorf("ffprobe不可用: %v", err)
	}

	// 获取关键帧时间戳
	cmd := createCommand(
		ffprobePath,
		"-v", "error",
		"-skip_frame", "nokey",
		"-select_streams", "v:0",
		"-show_entries", "frame=pkt_pts_time",
		"-of", "csv=print_section=0",
		videoSrc,
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("获取关键帧信息失败: %v, stderr: %s", err, stderrBuf.String())
	}

	// 解析关键帧时间戳
	keyframesStr := strings.Split(strings.TrimSpace(stdoutBuf.String()), "\n")
	var keyframes []float64
	for _, k := range keyframesStr {
		trimmed := strings.TrimSpace(k)
		if matched, _ := regexp.MatchString(`^\d+(\.\d+)?$`, trimmed); matched {
			if kf, err := strconv.ParseFloat(trimmed, 64); err == nil {
				keyframes = append(keyframes, kf)
			}
		}
	}

	// 融合关键帧数据
	fusedKeyframes, maxDuration, addDuration := v.keyframesFusion(keyframes)

	// 获取视频元数据
	cmdMeta := createCommand(
		ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoSrc,
	)

	var metaStdout, metaStderr bytes.Buffer
	cmdMeta.Stdout = &metaStdout
	cmdMeta.Stderr = &metaStderr

	err = cmdMeta.Run()
	if err != nil {
		return nil, fmt.Errorf("获取视频元数据失败: %v, stderr: %s", err, metaStderr.String())
	}

	var metaData map[string]interface{}
	if err := json.Unmarshal(metaStdout.Bytes(), &metaData); err != nil {
		return nil, fmt.Errorf("解析视频元数据失败: %v", err)
	}

	formatData, ok := metaData["format"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无法获取视频时长信息")
	}

	durationStr, ok := formatData["duration"].(string)
	if !ok {
		return nil, fmt.Errorf("无法获取视频时长信息")
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return nil, fmt.Errorf("解析视频时长失败: %v", err)
	}

	// 构建片段信息
	type fragment struct {
		Duration float64
		Start    float64
	}

	var fragments []fragment
	for i, k := range fusedKeyframes {
		var fragDuration float64
		if i == len(fusedKeyframes)-1 {
			fragDuration = v.numberDecimal(duration - k + addDuration)
		} else {
			fragDuration = v.numberDecimal(fusedKeyframes[i+1] - k + addDuration)
		}
		fragments = append(fragments, fragment{
			Duration: fragDuration,
			Start:    k,
		})
	}

	// 生成m3u8内容
	m3u8Data := "#EXTM3U\n"
	m3u8Data += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	m3u8Data += fmt.Sprintf("#EXT-X-TARGETDURATION:%.0f\n", maxDuration+addDuration)

	for _, f := range fragments {
		m3u8Data += fmt.Sprintf("#EXTINF:%.3f,\n", f.Duration)
		m3u8Data += fmt.Sprintf("/api/hls_video/%s/%.3f/%.3f\n", dramaSeriesId, f.Start, f.Duration)
	}

	m3u8Data += "#EXT-X-ENDLIST\n"

	// 写入文件
	err = os.WriteFile(m3u8Path, []byte(m3u8Data), 0644)
	if err != nil {
		return nil, err
	}

	return []byte(m3u8Data), nil
}

// keyframesFusion 融合关键帧数据
func (v M3U8) keyframesFusion(keyframes []float64) ([]float64, float64, float64) {
	if len(keyframes) == 0 {
		return []float64{0}, 10, 1
	}

	// 简化实现，实际项目中可能需要更复杂的算法
	maxDuration := 0.0
	for i := 0; i < len(keyframes)-1; i++ {
		duration := keyframes[i+1] - keyframes[i]
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	// 如果只有一帧或最后一段比较长
	if len(keyframes) > 0 {
		lastDuration := 10.0 // 默认值
		if len(keyframes) > 1 {
			lastDuration = keyframes[len(keyframes)-1] - keyframes[len(keyframes)-2]
		}
		if lastDuration > maxDuration {
			maxDuration = lastDuration
		}
	}

	return keyframes, maxDuration, 1.0
}

// numberDecimal 保留小数位数
func (v M3U8) numberDecimal(value float64) float64 {
	return math.Round(value*1000) / 1000
}
