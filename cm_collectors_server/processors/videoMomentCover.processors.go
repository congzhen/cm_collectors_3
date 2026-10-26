package processors

import (
	"cm_collectors_server/models"
	ff "cm_collectors_server/processorsFFmpeg"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const momentCoverRoot = "./db/videoMomentCovers"

var momentCoverMutex sync.Mutex
var videoMomentCapture = captureVideoMomentCover

func momentCoverDir(id string) string {
	return filepath.Join(momentCoverRoot, fmt.Sprintf("%x", sha256.Sum256([]byte(id))))
}
func momentCoverPath(v models.VideoMoment) string {
	return filepath.Join(momentCoverDir(v.ID), fmt.Sprintf("%x.jpg", sha256.Sum256([]byte(strconv.FormatFloat(v.Start, 'g', -1, 64)))))
}
func previewFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

// Covers are persistent bookmark assets, independent of the disposable timeline cache.
// A source replacement never changes an already saved cover at this bookmark position.
func persistMomentCover(v models.VideoMoment) (string, error) {
	path := momentCoverPath(v)
	if previewFileExists(path) {
		return path, nil
	}
	momentCoverMutex.Lock()
	defer momentCoverMutex.Unlock()
	if previewFileExists(path) {
		return path, nil
	}
	var current models.VideoMoment
	if err := videoMomentsDB().First(&current, "id = ?", v.ID).Error; err != nil {
		return "", err
	}
	if current.Start != v.Start {
		return "", errors.New("片段位置已更新，请刷新")
	}
	d, _, _, err := momentSource(v.DramaSeriesID)
	if err != nil {
		return "", err
	}
	setting, err := (VideoMoments{}).Setting()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := videoMomentCapture(d.Src, v.Start, setting.Width, path); err != nil {
		return "", err
	}
	return path, nil
}
func (VideoMoments) Cover(id string) (string, error) {
	var v models.VideoMoment
	if err := videoMomentsDB().First(&v, "id = ?", id).Error; err != nil {
		return "", err
	}
	return persistMomentCover(v)
}
func captureVideoMomentCover(src string, start float64, width int, path string) error {
	tool, err := (ff.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	temporary := path + ".tmp"
	defer os.Remove(temporary)
	height := width * 9 / 16 / 2 * 2
	cmd := ff.CreateCommandContext(ctx, tool, "-nostdin", "-v", "error", "-ss", strconv.FormatFloat(start, 'f', 3, 64), "-i", src, "-map", "0:v:0", "-frames:v", "1", "-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", width, height, width, height), "-threads", "1", "-q:v", "5", "-f", "image2", "-y", temporary)
	if err := cmd.Run(); err != nil {
		return errors.New("片段首图生成失败或超时")
	}
	if !previewFileExists(temporary) {
		return errors.New("未能读取片段首图")
	}
	return os.Rename(temporary, path)
}
