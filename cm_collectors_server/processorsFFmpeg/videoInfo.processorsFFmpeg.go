package processorsffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type VideoInfo struct{}

// VideoFormatInfo 保存视频格式信息的结构体
type VideoFormatInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Profile   string `json:"profile"`
	} `json:"streams"`
}

// IsWebCompatible 检查视频是否与Web兼容
func (v VideoInfo) IsWebCompatible(formatInfo VideoFormatInfo) bool {
	for _, stream := range formatInfo.Streams {
		// 对于视频流，检查编解码器是否为浏览器支持的格式 (H.264, VP8, VP9, AV1)
		if stream.CodecType == "video" {
			if stream.CodecName != "h264" && stream.CodecName != "vp8" && stream.CodecName != "vp9" && stream.CodecName != "av1" {
				return false
			}
			// 如果是H.264，检查profile是否被广泛支持
			if stream.CodecName == "h264" && strings.Contains(stream.Profile, "High") {
				// High profile可能不被所有设备支持
				return false
			}
		}
		// 对于音频流，检查编解码器是否为浏览器支持的格式 (AAC, MP3, Vorbis, Opus)
		if stream.CodecType == "audio" {
			if stream.CodecName != "aac" && stream.CodecName != "mp3" && stream.CodecName != "vorbis" && stream.CodecName != "opus" {
				return false
			}
		}
	}
	return true
}

// GetVideoFormatInfo 使用ffprobe获取视频格式信息
func (v VideoInfo) GetVideoFormatInfo(src string) (VideoFormatInfo, error) {
	var formatInfo VideoFormatInfo

	// 检查ffprobe是否可用
	ffprobePath, err := FFmpeg{}.IsFFprobeAvailable()
	if err != nil {
		return formatInfo, fmt.Errorf("ffprobe不可用: %v", err)
	}

	// 使用ffprobe获取视频信息
	cmd := exec.Command(ffprobePath, "-v", "quiet", "-print_format", "json", "-show_streams", src)
	output, err := cmd.Output()
	if err != nil {
		return formatInfo, fmt.Errorf("无法获取视频信息: %v", err)
	}

	// 解析JSON输出
	err = json.Unmarshal(output, &formatInfo)
	if err != nil {
		return formatInfo, fmt.Errorf("无法解析视频信息: %v", err)
	}

	return formatInfo, nil
}
