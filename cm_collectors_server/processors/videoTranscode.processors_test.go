package processors

import (
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateVideoTranscodeConfig(t *testing.T) {
	valid := DefaultVideoTranscodeConfig()
	if err := validateVideoTranscodeConfig(valid); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}

	copyWithResize := valid
	copyWithResize.VideoCodec = "copy"
	copyWithResize.ResolutionHeight = 1080
	if err := validateVideoTranscodeConfig(copyWithResize); err == nil {
		t.Fatal("stream copy with resize should be rejected")
	}

	invalidContainer := valid
	invalidContainer.Container = "avi"
	if err := validateVideoTranscodeConfig(invalidContainer); err == nil {
		t.Fatal("unsupported container should be rejected")
	}
}

func TestVideoTranscodePathsStayBesideSource(t *testing.T) {
	task := &models.VideoTranscodeTask{
		ID:         "task-1",
		SourcePath: filepath.Join("D:", "videos", "movie.mkv"),
	}
	config := DefaultVideoTranscodeConfig()
	temp, output, backup := videoTranscodePaths(task, config)

	if filepath.Dir(temp) != filepath.Dir(task.SourcePath) ||
		filepath.Dir(output) != filepath.Dir(task.SourcePath) ||
		filepath.Dir(backup) != filepath.Dir(task.SourcePath) {
		t.Fatalf("all replacement files must stay beside source: %q, %q, %q", temp, output, backup)
	}
	if filepath.Ext(output) != ".mp4" || !strings.HasSuffix(temp, ".mp4.part") {
		t.Fatalf("unexpected output paths: %q, %q", output, temp)
	}
	nextTemp, _, nextBackup := videoTranscodePaths(task, config)
	if nextTemp == temp || nextBackup == backup {
		t.Fatal("each run must use unique temporary and backup paths")
	}
}

func TestCountVideoTranscodeSubtitleStreams(t *testing.T) {
	info := processorsffmpeg.VideoFormatInfo{
		Streams: []processorsffmpeg.VideoStreamInfo{
			{CodecType: "video"},
			{CodecType: "subtitle"},
			{CodecType: "audio"},
			{CodecType: "subtitle"},
		},
	}
	if got := countVideoTranscodeStreams(info, "subtitle"); got != 2 {
		t.Fatalf("expected two subtitle streams, got %d", got)
	}
}

func TestHardwareVideoEncoderSelection(t *testing.T) {
	if got := hardwareVideoEncoder("h264", "nvenc", "libx264"); got != "h264_nvenc" {
		t.Fatalf("unexpected NVENC encoder: %s", got)
	}
	if got := hardwareVideoEncoder("h265", "qsv", "libx265"); got != "hevc_qsv" {
		t.Fatalf("unexpected QSV encoder: %s", got)
	}
	if got := hardwareVideoEncoder("h264", "", "libx264"); got != "libx264" {
		t.Fatalf("CPU fallback should be preserved: %s", got)
	}
}

func TestValidateVideoTranscodeRejectsGPUCopy(t *testing.T) {
	config := DefaultVideoTranscodeConfig()
	config.VideoCodec = "copy"
	config.GPUEncoder = "nvenc"
	if err := validateVideoTranscodeConfig(config); err == nil {
		t.Fatal("GPU encoder must be rejected when video stream is copied")
	}
}

func TestVideoTranscodeTemporaryPathValidation(t *testing.T) {
	if !isVideoTranscodeTemporaryPath(`D:\videos\.movie.cmtranscode-task-1-run.mp4.part`, "task-1") {
		t.Fatal("owned temporary path should be accepted")
	}
	if isVideoTranscodeTemporaryPath(`D:\videos\.movie.cmtranscode-task-2-run.mp4.part`, "task-1") {
		t.Fatal("another task's temporary path must not be accepted")
	}
	if isVideoTranscodeTemporaryPath(`D:\videos\movie.mp4.part`, "task-1") {
		t.Fatal("ordinary part file must not be treated as an owned temporary file")
	}
}

func TestValidateVideoTranscodeSourceCompatibilityRejectsBitmapSubtitleInMP4(t *testing.T) {
	info := processorsffmpeg.VideoFormatInfo{
		Streams: []processorsffmpeg.VideoStreamInfo{
			{CodecType: "video", CodecName: "h264"},
			{CodecType: "subtitle", CodecName: "hdmv_pgs_subtitle"},
		},
	}
	config := DefaultVideoTranscodeConfig()
	if err := validateVideoTranscodeSourceCompatibility(info, config); err == nil {
		t.Fatal("MP4 output must reject bitmap subtitles before running FFmpeg")
	}
	config.Container = "mkv"
	if err := validateVideoTranscodeSourceCompatibility(info, config); err != nil {
		t.Fatalf("MKV should preserve bitmap subtitles: %v", err)
	}
}

func TestValidateVideoTranscodeOutputInfoChecksStreamsAndTargets(t *testing.T) {
	source := processorsffmpeg.VideoFormatInfo{
		Streams: []processorsffmpeg.VideoStreamInfo{
			{CodecType: "video", CodecName: "h264", Width: 1920, Height: 1080, AverageFrameRate: "30/1"},
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "subtitle", CodecName: "subrip"},
		},
	}
	source.Format.Duration = "60"
	output := processorsffmpeg.VideoFormatInfo{
		Streams: []processorsffmpeg.VideoStreamInfo{
			{CodecType: "video", CodecName: "h264", Width: 1280, Height: 720, AverageFrameRate: "24/1"},
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "subtitle", CodecName: "mov_text"},
		},
	}
	output.Format.Duration = "60"
	config := DefaultVideoTranscodeConfig()
	config.VideoCodec = "h264"
	config.AudioCodec = "aac"
	config.ResolutionHeight = 720
	config.FrameRate = 24
	if err := validateVideoTranscodeOutputInfo(source, output, 60, config); err != nil {
		t.Fatalf("matching output should pass validation: %v", err)
	}

	missingAudio := output
	missingAudio.Streams = []processorsffmpeg.VideoStreamInfo{
		output.Streams[0],
		output.Streams[2],
	}
	if err := validateVideoTranscodeOutputInfo(source, missingAudio, 60, config); err == nil {
		t.Fatal("missing audio stream must be rejected")
	}

	wrongFrameRate := output
	wrongFrameRate.Streams = append([]processorsffmpeg.VideoStreamInfo(nil), output.Streams...)
	wrongFrameRate.Streams[0].AverageFrameRate = "30/1"
	if err := validateVideoTranscodeOutputInfo(source, wrongFrameRate, 60, config); err == nil {
		t.Fatal("unexpected frame rate must be rejected")
	}
}
