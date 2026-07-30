package processorsffmpeg

import (
	"context"
	"sync"
	"time"
)

type GPUEncoderCapability struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	VideoCodecs []string `json:"videoCodecs"`
}

var (
	gpuEncoderCapabilitiesOnce sync.Once
	gpuEncoderCapabilities     []GPUEncoderCapability
)

// AvailableGPUEncoders 不只检查 FFmpeg 是否编译了编码器，还实际编码一帧，
// 避免向前端展示当前机器无法初始化的 GPU。
func (FFmpeg) AvailableGPUEncoders() []GPUEncoderCapability {
	gpuEncoderCapabilitiesOnce.Do(func() {
		type candidate struct {
			id       string
			label    string
			encoders map[string]string
		}
		candidates := []candidate{
			{id: "nvenc", label: "NVIDIA NVENC", encoders: map[string]string{"h264": "h264_nvenc", "h265": "hevc_nvenc"}},
			{id: "qsv", label: "Intel Quick Sync", encoders: map[string]string{"h264": "h264_qsv", "h265": "hevc_qsv"}},
			{id: "amf", label: "AMD AMF", encoders: map[string]string{"h264": "h264_amf", "h265": "hevc_amf"}},
		}
		ffmpegPath, err := (FFmpeg{}).IsFFmpegAvailable()
		if err != nil {
			return
		}
		for _, item := range candidates {
			capability := GPUEncoderCapability{ID: item.id, Label: item.label}
			for _, codec := range []string{"h264", "h265"} {
				if testGPUEncoder(ffmpegPath, item.encoders[codec]) {
					capability.VideoCodecs = append(capability.VideoCodecs, codec)
				}
			}
			if len(capability.VideoCodecs) > 0 {
				gpuEncoderCapabilities = append(gpuEncoderCapabilities, capability)
			}
		}
	})
	return append([]GPUEncoderCapability(nil), gpuEncoderCapabilities...)
}

func testGPUEncoder(ffmpegPath, encoder string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := createCommandContext(
		ctx,
		ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "color=size=320x180:rate=1",
		"-frames:v", "1", "-an",
		"-c:v", encoder,
		"-f", "null", "-",
	)
	return cmd.Run() == nil
}
