package processorsffmpeg

import "testing"

func TestGetVideoBasicInfoSelectsPrimaryStreams(t *testing.T) {
	info := VideoFormatInfo{}
	info.Format.Duration = "120.5"
	info.Format.Size = "1024"
	info.Format.BitRate = "9000000"
	info.Format.FormatName = "mov,mp4,m4a,3gp,3g2,mj2"

	cover := VideoStreamInfo{CodecType: "video", CodecName: "mjpeg", Width: 1200, Height: 1200}
	cover.Disposition.AttachedPic = 1
	secondary := VideoStreamInfo{CodecType: "video", CodecName: "h264", Width: 1920, Height: 1080}
	primary := VideoStreamInfo{
		CodecType: "video", CodecName: "hevc", Profile: "Main 10",
		PixelFormat: "yuv420p10le", Width: 3840, Height: 2160,
		AverageFrameRate: "24000/1001", BitRate: "8000000",
	}
	primary.Disposition.Default = 1
	audio := VideoStreamInfo{CodecType: "audio", CodecName: "aac", Channels: 6, SampleRate: "48000"}
	audio.Disposition.Default = 1
	info.Streams = []VideoStreamInfo{cover, secondary, primary, audio}

	basic := (VideoInfo{}).GetVideoBasicInfoByVideoFormatInfo(info)
	if basic.Width != 3840 || basic.Height != 2160 {
		t.Fatalf("unexpected dimensions: %dx%d", basic.Width, basic.Height)
	}
	if basic.VideoCodec != "hevc" || basic.VideoProfile != "Main 10" {
		t.Fatalf("unexpected video codec: %#v", basic)
	}
	if basic.FrameRate != 23.976 || basic.FrameRateRaw != "24000/1001" {
		t.Fatalf("unexpected frame rate: %v (%s)", basic.FrameRate, basic.FrameRateRaw)
	}
	if basic.BitDepth != 10 {
		t.Fatalf("unexpected bit depth: %d", basic.BitDepth)
	}
	if basic.AudioCodec != "aac" || basic.AudioChannels != 6 || basic.AudioSampleRate != 48000 {
		t.Fatalf("unexpected audio metadata: %#v", basic)
	}
	if basic.Duration != "120.5" || basic.ContainerFormat == "" {
		t.Fatalf("format fallback was not applied: %#v", basic)
	}
}

func TestPrimaryVideoStreamFallsBackToLargestValidStream(t *testing.T) {
	info := VideoFormatInfo{
		Streams: []VideoStreamInfo{
			{CodecType: "video", CodecName: "h264", Width: 1280, Height: 720},
			{CodecType: "video", CodecName: "vp9", Width: 1920, Height: 1080},
		},
	}
	stream := (VideoInfo{}).PrimaryVideoStream(info)
	if stream == nil || stream.CodecName != "vp9" {
		t.Fatalf("expected largest stream, got %#v", stream)
	}
}
