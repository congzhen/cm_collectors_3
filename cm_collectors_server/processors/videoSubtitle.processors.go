package processors

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html/charset"
)

type VideoSubtitle struct{}

// GetVideoSubtitle 获取指定语言的字幕文件
func (t VideoSubtitle) GetVideoSubtitle(c *gin.Context, dramaSeriesId string) error {
	// 获取视频文件路径
	videoSrc, err := ResourcesDramaSeries{}.GetSrc(dramaSeriesId)
	if err != nil {
		return err
	}

	// 获取查询参数
	lang := c.Query("lang")     // 如: zh, en, jp 等
	format := c.Query("format") // 如: srt, vtt, ass 等

	// 查找字幕文件
	subtitleSrc, _, err := t.findLocalizedSubtitleFile(videoSrc, lang, format)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Subtitle file not found"})
		return nil
	}

	// 检查字幕文件是否存在
	_, err = os.Stat(subtitleSrc)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Subtitle file not found"})
		return nil
	}
	// 如果请求转换为 WebVTT 或文件本身就是 WebVTT
	if filepath.Ext(subtitleSrc) == ".vtt" {
		c.Header("Content-Type", "text/vtt; charset=utf-8")
		c.File(subtitleSrc)
		return nil
	} else {
		err = t.convertToVTT(c, subtitleSrc)
		if err != nil {
			// 如果转换失败，直接返回原文件
			contentType := "text/plain; charset=utf-8"
			c.Header("Content-Type", contentType)
			c.File(subtitleSrc)
			return nil
		}
		return nil
	}
}

func (t VideoSubtitle) convertToVTT(c *gin.Context, subtitlePath string) error {
	// 自动检测并读取字幕文件
	sub, err := astisub.OpenFile(subtitlePath)
	if err != nil {
		// 如果直接读取失败，尝试处理编码问题
		sub, err = t.openSubtitleWithEncodingFix(subtitlePath)
		if err != nil {
			return fmt.Errorf("failed to open subtitle file: %v", err)
		}
	}

	// 检查是否成功读取到字幕项
	if sub == nil || len(sub.Items) == 0 {
		return fmt.Errorf("no subtitles to write")
	}

	// 设置 Content-Type
	c.Header("Content-Type", "text/vtt; charset=utf-8")
	if sub.Metadata == nil {
		sub.Metadata = &astisub.Metadata{}
	}

	// 写入 WebVTT 格式
	return sub.WriteToWebVTT(c.Writer)
}

// openSubtitleWithEncodingFix 尝试修复编码问题后打开字幕文件
func (t VideoSubtitle) openSubtitleWithEncodingFix(subtitlePath string) (*astisub.Subtitles, error) {
	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(subtitlePath))

	// 首先尝试检测文件编码
	file, err := os.Open(subtitlePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件前几个字节用于编码检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// 检测编码
	contentType := http.DetectContentType(buffer[:n])
	_, detectedCharset, _ := charset.DetermineEncoding(buffer[:n], contentType)

	// 重新打开文件进行完整读取
	file, err = os.Open(subtitlePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 如果检测到非UTF-8编码，进行转换
	if detectedCharset != "" && strings.ToLower(detectedCharset) != "utf-8" {
		// 使用 charset 包正确创建解码器
		reader, err := charset.NewReaderLabel(detectedCharset, file)
		if err != nil {
			// 如果转换失败，尝试直接打开
			return astisub.OpenFile(subtitlePath)
		}
		// 根据文件扩展名选择正确的读取方法
		switch ext {
		case ".srt":
			return astisub.ReadFromSRT(reader)
		case ".ass", ".ssa":
			return astisub.ReadFromSSA(reader)
		case ".vtt":
			return astisub.ReadFromWebVTT(reader)
		default:
			// 尝试自动检测
			return astisub.ReadFromSRT(reader)
		}
	}

	// 如果是UTF-8编码，直接打开
	return astisub.OpenFile(subtitlePath)
}

// findLocalizedSubtitleFile 查找指定语言的字幕文件
func (VideoSubtitle) findLocalizedSubtitleFile(videoSrc string, lang string, format string) (string, string, error) {
	dir := filepath.Dir(videoSrc)
	baseName := strings.TrimSuffix(filepath.Base(videoSrc), filepath.Ext(videoSrc))

	// 构造可能的字幕文件名
	var candidates []string

	if lang != "" {
		// 带语言后缀的文件名优先
		if format != "" {
			candidates = append(candidates, baseName+"."+lang+"."+format)
		} else {
			// 尝试各种格式
			candidates = append(candidates, baseName+"."+lang+".srt")
			candidates = append(candidates, baseName+"."+lang+".ass")
			candidates = append(candidates, baseName+"."+lang+".ssa")
			candidates = append(candidates, baseName+"."+lang+".vtt")
		}
	}

	// 不带语言后缀的文件名
	if format != "" {
		candidates = append(candidates, baseName+"."+format)
	} else {
		candidates = append(candidates, baseName+".srt")
		candidates = append(candidates, baseName+".ass")
		candidates = append(candidates, baseName+".ssa")
		candidates = append(candidates, baseName+".vtt")
	}

	// 支持的Content-Type映射
	contentTypes := map[string]string{
		".srt": "text/plain; charset=utf-8",
		".ass": "text/plain; charset=utf-8",
		".ssa": "text/plain; charset=utf-8",
		".vtt": "text/vtt; charset=utf-8",
		".sub": "text/plain; charset=utf-8",
		".idx": "text/plain; charset=utf-8",
	}

	// 按顺序查找存在的文件
	for _, candidate := range candidates {
		fullPath := filepath.Join(dir, candidate)
		if _, err := os.Stat(fullPath); err == nil {
			ext := filepath.Ext(fullPath)
			contentType := contentTypes[ext]
			if contentType == "" {
				contentType = "text/plain; charset=utf-8"
			}
			return fullPath, contentType, nil
		}
	}

	return "", "", fmt.Errorf("no subtitle file found")
}
