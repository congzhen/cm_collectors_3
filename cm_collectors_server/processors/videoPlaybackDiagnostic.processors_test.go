package processors

import (
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"encoding/json"
	"strings"
	"testing"
)

func mustVideoFormatInfo(t *testing.T, value string) processorsffmpeg.VideoFormatInfo {
	t.Helper()
	var info processorsffmpeg.VideoFormatInfo
	if err := json.Unmarshal([]byte(value), &info); err != nil {
		t.Fatalf("解析测试媒体信息失败: %v", err)
	}
	return info
}

func TestPlaybackCompatibilityRisksNormalMP4(t *testing.T) {
	info := mustVideoFormatInfo(t, `{
		"format":{"format_name":"mov,mp4,m4a,3gp,3g2,mj2"},
		"streams":[
			{"codec_type":"video","codec_name":"h264","profile":"High","pix_fmt":"yuv420p","width":1920,"height":1080,"bits_per_raw_sample":"8"},
			{"codec_type":"audio","codec_name":"aac","sample_rate":"48000","channels":2,"channel_layout":"stereo"}
		]
	}`)

	if risks := playbackCompatibilityRisks(info, true); len(risks) != 0 {
		t.Fatalf("普通 H.264/AAC MP4 不应产生日志风险，实际为: %v", risks)
	}
	if hasFFmpegWarning("ffmpeg version 7.1\nframe=25 fps=0.0") {
		t.Fatal("普通 FFmpeg 信息输出不应被识别为异常")
	}
}

func TestPlaybackCompatibilityRisksDetectsVideoAndAudioIssues(t *testing.T) {
	info := mustVideoFormatInfo(t, `{
		"format":{"format_name":"matroska,webm"},
		"streams":[
			{"codec_type":"video","codec_name":"h264","profile":"High 4:4:4 Predictive","pix_fmt":"yuv444p","bits_per_raw_sample":"10"},
			{"codec_type":"audio","codec_name":"flac","channels":6,"channel_layout":"5.1"}
		]
	}`)

	risks := strings.Join(playbackCompatibilityRisks(info, true), "\n")
	for _, expected := range []string{"不是 MP4/MOV", "yuv444p", "High 4:4:4 Predictive", "10 bit", "6 个声道"} {
		if !strings.Contains(risks, expected) {
			t.Errorf("风险日志缺少 %q，实际为: %s", expected, risks)
		}
	}
}

func TestLimitedDiagnosticBuffer(t *testing.T) {
	buffer := &limitedDiagnosticBuffer{limit: 8}
	input := []byte("1234567890")
	n, err := buffer.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("写入结果异常: n=%d err=%v", n, err)
	}
	if output := buffer.String(); !strings.Contains(output, "12345678") || !strings.Contains(output, "已截断") {
		t.Fatalf("未正确限制 stderr: %q", output)
	}
}
