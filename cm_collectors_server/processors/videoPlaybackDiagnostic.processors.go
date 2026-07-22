package processors

import (
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const playbackDiagnosticLogPath = "./playback_error.log"

var playbackDiagnosticLogMu sync.Mutex
var playbackDiagnosticLastWrite = make(map[string]time.Time)

type playbackDiagnosticRecord struct {
	Time          string                            `json:"time"`
	Event         string                            `json:"event"`
	MediaID       string                            `json:"media_id,omitempty"`
	FileName      string                            `json:"file_name,omitempty"`
	PlaybackMode  string                            `json:"playback_mode,omitempty"`
	UserAgent     string                            `json:"user_agent,omitempty"`
	Risks         []string                          `json:"risks,omitempty"`
	MediaInfo     *processorsffmpeg.VideoFormatInfo `json:"media_info,omitempty"`
	SegmentStart  float64                           `json:"segment_start,omitempty"`
	SegmentLength float64                           `json:"segment_length,omitempty"`
	FFmpegVersion string                            `json:"ffmpeg_version,omitempty"`
	FFmpegCommand []string                          `json:"ffmpeg_command,omitempty"`
	Error         string                            `json:"error,omitempty"`
	FFmpegStderr  string                            `json:"ffmpeg_stderr,omitempty"`
}

func writePlaybackDiagnostic(record playbackDiagnosticRecord) {
	record.Time = time.Now().Format(time.RFC3339)

	playbackDiagnosticLogMu.Lock()
	defer playbackDiagnosticLogMu.Unlock()

	file, err := os.OpenFile(playbackDiagnosticLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("无法写入播放异常诊断日志 %s: %v", playbackDiagnosticLogPath, err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(record); err != nil {
		log.Printf("无法编码播放异常诊断日志: %v", err)
	}
}

func writePlaybackDiagnosticRateLimited(record playbackDiagnosticRecord, interval time.Duration) {
	keyContent := strings.Join([]string{record.Event, record.MediaID, record.Error, record.FFmpegStderr}, "|")
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(keyContent)))
	now := time.Now()

	playbackDiagnosticLogMu.Lock()
	lastWrite, exists := playbackDiagnosticLastWrite[key]
	if exists && now.Sub(lastWrite) < interval {
		playbackDiagnosticLogMu.Unlock()
		return
	}
	playbackDiagnosticLastWrite[key] = now
	playbackDiagnosticLogMu.Unlock()

	writePlaybackDiagnostic(record)
}

func playbackCompatibilityRisks(formatInfo processorsffmpeg.VideoFormatInfo, directPlay bool) []string {
	var risks []string
	formatName := strings.ToLower(formatInfo.Format.FormatName)
	if directPlay && !strings.Contains(formatName, "mp4") && !strings.Contains(formatName, "mov") {
		risks = append(risks, fmt.Sprintf("直接播放的容器不是 MP4/MOV，实际格式为 %q，但响应类型固定为 video/mp4", formatInfo.Format.FormatName))
	}

	videoStreams := 0
	audioStreams := 0
	for _, stream := range formatInfo.Streams {
		switch stream.CodecType {
		case "video":
			videoStreams++
			if stream.CodecName == "h264" {
				pixelFormat := strings.ToLower(stream.PixelFormat)
				if pixelFormat != "" && pixelFormat != "yuv420p" && pixelFormat != "yuvj420p" {
					risks = append(risks, fmt.Sprintf("H.264 像素格式 %q 可能不受浏览器硬件解码支持", stream.PixelFormat))
				}
				profile := strings.ToLower(stream.Profile)
				if strings.Contains(profile, "4:4:4") || strings.Contains(profile, "4:2:2") || strings.Contains(profile, "high 10") {
					risks = append(risks, fmt.Sprintf("H.264 Profile %q 兼容性有限", stream.Profile))
				}
			}
			if bits, err := strconv.Atoi(stream.BitsPerRawSample); err == nil && bits > 8 {
				risks = append(risks, fmt.Sprintf("视频位深为 %d bit，部分浏览器或设备可能无法解码", bits))
			}
		case "audio":
			audioStreams++
			if stream.Channels > 2 {
				risks = append(risks, fmt.Sprintf("音频包含 %d 个声道，部分浏览器或设备可能无法输出声音", stream.Channels))
			}
			if directPlay && strings.Contains(formatName, "mp4") && stream.CodecName != "aac" && stream.CodecName != "mp3" {
				risks = append(risks, fmt.Sprintf("MP4 中的音频编码 %q 可能无法被浏览器直接播放", stream.CodecName))
			}
		}
	}

	if videoStreams == 0 {
		risks = append(risks, "未检测到视频流")
	}
	if audioStreams == 0 {
		risks = append(risks, "未检测到音频流")
	}
	return risks
}

func sanitizedMediaInfo(formatInfo processorsffmpeg.VideoFormatInfo) processorsffmpeg.VideoFormatInfo {
	formatInfo.Format.Filename = filepath.Base(formatInfo.Format.Filename)
	return formatInfo
}

func playbackFFmpegVersion() string {
	ffmpegPath, err := (processorsffmpeg.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return "不可用: " + err.Error()
	}
	version, err := (processorsffmpeg.FFmpeg{}).GetToolVersion(ffmpegPath)
	if err != nil {
		return "读取失败: " + err.Error()
	}
	return version
}

func sanitizedFFmpegCommand(args []string, sourcePath string) []string {
	result := append([]string(nil), args...)
	for i := range result {
		if result[i] == sourcePath {
			result[i] = filepath.Base(sourcePath)
		}
		if i == 0 {
			result[i] = filepath.Base(result[i])
		}
	}
	return result
}

func hasFFmpegWarning(stderr string) bool {
	lower := strings.ToLower(stderr)
	markers := []string{
		"warning", "error", "failed", "invalid", "corrupt", "unsupported",
		"non-monoton", "deprecated", "concealing", "damaged", "missing picture",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

type limitedDiagnosticBuffer struct {
	data      []byte
	limit     int
	truncated bool
}

func (b *limitedDiagnosticBuffer) Write(p []byte) (int, error) {
	originalLength := len(p)
	remaining := b.limit - len(b.data)
	if remaining <= 0 {
		b.truncated = true
		return originalLength, nil
	}
	if len(p) > remaining {
		b.data = append(b.data, p[:remaining]...)
		b.truncated = true
	} else {
		b.data = append(b.data, p...)
	}
	return originalLength, nil
}

func (b *limitedDiagnosticBuffer) String() string {
	text := strings.TrimSpace(string(b.data))
	if b.truncated {
		text += "\n[stderr 已截断，仅保留前 64KB]"
	}
	return text
}
