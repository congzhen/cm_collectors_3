package processorsffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
)

// 支持的编解码器列表
var (
	CT_SupportedVideoCodecs = []string{"h264", "vp8", "vp9", "av1", "hevc"}
	CT_SupportedAudioCodecs = []string{"aac", "mp3", "vorbis", "opus", "pcm_s16le", "pcm_s24le"}
)

type VideoInfo struct {
	supportedVideoCodecs []string
	supportedAudioCodecs []string
}

// VideoFormatInfo 保存视频格式信息的结构体
type VideoFormatInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Profile   string `json:"profile"`
	} `json:"streams"`
}

func (v *VideoInfo) SetSupportedVideoCodecs(codecs []string) {
	v.supportedVideoCodecs = codecs
}
func (v VideoInfo) GetSupportedVideoCodecs() []string {
	if v.supportedVideoCodecs == nil || len(v.supportedVideoCodecs) == 0 {
		return CT_SupportedVideoCodecs
	}
	return v.supportedVideoCodecs
}
func (v *VideoInfo) SetSupportedAudioCodecs(codecs []string) {
	v.supportedAudioCodecs = codecs
}
func (v VideoInfo) GetSupportedAudioCodecs() []string {
	if v.supportedAudioCodecs == nil || len(v.supportedAudioCodecs) == 0 {
		return CT_SupportedAudioCodecs
	}
	return v.supportedAudioCodecs
}

// IsWebCompatible 检查视频是否与Web兼容
func (v VideoInfo) IsWebCompatible(formatInfo VideoFormatInfo) bool {
	webCompatible := true

	for _, stream := range formatInfo.Streams {
		// 对于视频流，检查编解码器是否为浏览器支持的格式
		if stream.CodecType == "video" {

			fmt.Println("######################### 视频编解码器：", v.GetSupportedVideoCodecs())
			fmt.Println("######################### 视频编码：", stream.CodecName)
			// 支持的视频编解码器包括 H.264 (所有profile), VP8, VP9, AV1, HEVC/H.265
			// 注意: HEVC支持有限，主要在Safari和Edge中
			if !slices.Contains(v.GetSupportedVideoCodecs(), stream.CodecName) {
				webCompatible = false
			}
		}
		// 对于音频流，检查编解码器是否为浏览器支持的格式
		if stream.CodecType == "audio" {
			fmt.Println("######################### 编解码器：", v.GetSupportedAudioCodecs())
			fmt.Println("######################### 音频编码：", stream.CodecName)
			// 支持的音频编解码器包括 AAC, MP3, Vorbis, Opus, PCM等
			if !slices.Contains(v.GetSupportedAudioCodecs(), stream.CodecName) {
				webCompatible = false
			}
		}
	}
	return webCompatible
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
