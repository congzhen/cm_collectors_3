package processorsffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
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
	hasVideoStream := false
	hasAudioStream := false
	videoCompatible := true
	audioCompatible := true

	for _, stream := range formatInfo.Streams {
		// 对于视频流，检查编解码器是否为浏览器支持的格式
		if stream.CodecType == "video" {
			hasVideoStream = true
			fmt.Println("######################### 视频编码：", stream.CodecName)
			// 支持的视频编解码器包括 H.264 (所有profile), VP8, VP9, AV1, HEVC/H.265
			// 注意: HEVC支持有限，主要在Safari和Edge中
			if stream.CodecName != "h264" && stream.CodecName != "vp8" &&
				stream.CodecName != "vp9" && stream.CodecName != "av1" {
				videoCompatible = false
			}
			// 移除了对H.264 High Profile的限制，因为现代浏览器都支持
		}
		// 对于音频流，检查编解码器是否为浏览器支持的格式
		if stream.CodecType == "audio" {
			hasAudioStream = true
			fmt.Println("######################### 音频编码：", stream.CodecName)
			// 支持的音频编解码器包括 AAC, MP3, Vorbis, Opus, PCM等
			if stream.CodecName != "aac" && stream.CodecName != "mp3" &&
				stream.CodecName != "vorbis" && stream.CodecName != "opus" &&
				stream.CodecName != "pcm_s16le" && stream.CodecName != "pcm_s24le" {
				audioCompatible = false
			}
		}
	}

	// 确保至少有一个视频流和一个音频流，并且都兼容
	return hasVideoStream && hasAudioStream && videoCompatible && audioCompatible
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
