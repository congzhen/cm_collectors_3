package processorsffmpeg

import (
	"encoding/json"
	"fmt"
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

// VideoDefinition 清晰度类型定义
type VideoDefinition string

const (
	Definition8K                 VideoDefinition = "8K"
	Definition4K                 VideoDefinition = "4K"
	Definition2K                 VideoDefinition = "2K"
	Definition1080P              VideoDefinition = "1080P"
	Definition720P               VideoDefinition = "720P"
	DefinitionHighDefinition     VideoDefinition = "HighDefinition"
	DefinitionStandardDefinition VideoDefinition = "StandardDefinition"
	DefinitionEmpty              VideoDefinition = ""
)

// VideoFormatInfo 保存视频格式信息的结构体
type VideoFormatInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Profile   string `json:"profile"`
		Width     int    `json:"width,omitempty"`
		Height    int    `json:"height,omitempty"`
		Duration  string `json:"duration,omitempty"`
		BitRate   string `json:"bit_rate,omitempty"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration,omitempty"`
		Size     string `json:"size,omitempty"`
		BitRate  string `json:"bit_rate,omitempty"`
	} `json:"format"`
}

// VideoBasicInfo 保存视频基本信息的结构体
type VideoBasicInfo struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
	Size     string `json:"size"`
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
	cmd := createCommand(ffprobePath, "-v", "quiet", "-print_format", "json", "-show_streams", src)
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

// GetVideoBasicInfo 获取视频基本信息（宽度、高度等）
func (v VideoInfo) GetVideoBasicInfo(src string) (VideoBasicInfo, error) {
	var basicInfo VideoBasicInfo

	// 获取视频格式信息
	formatInfo, err := v.GetVideoFormatInfo(src)
	if err != nil {
		return basicInfo, fmt.Errorf("无法获取视频格式信息: %v", err)
	}

	// 查找视频流以获取宽度和高度
	for _, stream := range formatInfo.Streams {
		if stream.CodecType == "video" {
			basicInfo.Width = stream.Width
			basicInfo.Height = stream.Height
			basicInfo.Duration = stream.Duration
			basicInfo.BitRate = stream.BitRate
			break
		}
	}

	// 如果视频流中没有时长信息，则从format中获取
	if basicInfo.Duration == "" {
		basicInfo.Duration = formatInfo.Format.Duration
	}

	// 获取文件大小和总比特率
	basicInfo.Size = formatInfo.Format.Size
	basicInfo.BitRate = formatInfo.Format.BitRate

	return basicInfo, nil
}

// GetVideoDefinition 根据视频的宽度和高度确定视频清晰度
func (v VideoInfo) GetVideoDefinition(width, height int) VideoDefinition {
	// 获取最小值作为主要尺寸，以更好地反映视频的清晰度
	// 这样可以确保竖版视频（如手机拍摄）也能正确识别
	minDimension := width
	if height < minDimension {
		minDimension = height
	}

	// 根据最小尺寸判断清晰度
	// 标准基于常见的视频分辨率:
	// 8K: >= 4320 (例如 7680×4320)
	// 4K: >= 2160 (例如 3840×2160, 4096×2160)
	// 2K: >= 1440 (例如 2560×1440)
	// 1080P: >= 1080 (例如 1920×1080)
	// 720P: >= 720 (例如 1280×720)
	// HighDefinition: >= 480 (例如 854×480)
	// StandardDefinition: < 480 (例如 640×480, 480×360等)
	switch {
	case minDimension >= 4320:
		return Definition8K
	case minDimension >= 2160:
		return Definition4K
	case minDimension >= 1440:
		return Definition2K
	case minDimension >= 1080:
		return Definition1080P
	case minDimension >= 720:
		return Definition720P
	case minDimension >= 480:
		return DefinitionHighDefinition
	case minDimension > 0:
		return DefinitionStandardDefinition
	default:
		return DefinitionEmpty
	}
}
