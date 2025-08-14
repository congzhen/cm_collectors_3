package processorsffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Transcode struct{}

// VideoStreamTranscode 处理视频流式转码并直接写入HTTP响应
func (t Transcode) VideoStreamTranscode(c *gin.Context, src string) error {
	ffmpegPath, err := FFmpeg{}.IsFFmpegAvailable()
	if err != nil {
		return fmt.Errorf("FFmpeg不可用: %v", err)
	}

	// 解析Range请求
	rangeHeader := c.GetHeader("Range")
	start, end, err := t.parseRangeHeader(rangeHeader)
	if err != nil {
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return err
	}

	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 构建ffmpeg命令进行流式转码
	// 使用更适合流媒体播放的参数
	args := []string{
		"-i", src,
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-preset", "superfast",
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-ar", "44100",
		"-b:a", "128k",
		"-f", "mp4",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof",
		"pipe:1",
	}

	if start > 0 {
		args = append([]string{"-ss", strconv.FormatInt(start, 10)}, args...)
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)

	// 设置stderr用于错误处理
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法创建stderr管道: %v", err)
	}

	// 设置stdout用于数据传输
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法创建stdout管道: %v", err)
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("无法启动ffmpeg: %v", err)
	}

	// 设置响应头
	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Connection", "keep-alive")

	// 处理Range请求
	if rangeHeader != "" {
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/*", start, end))
		c.Status(http.StatusPartialContent)
	} else {
		c.Status(http.StatusOK)
	}

	// 启动错误监控goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// 可以记录FFmpeg的输出日志用于调试
			// fmt.Printf("FFmpeg log: %s\n", scanner.Text())
		}
	}()

	// 监听客户端断开连接
	clientDisconnected := c.Request.Context().Done()

	// 实时传输数据
	buffer := make([]byte, 64*1024) // 64KB缓冲区以提高性能
	for {
		select {
		case <-clientDisconnected:
			// 客户端断开连接，取消FFmpeg进程
			cancel()
			cmd.Process.Kill()
			wg.Wait()
			return nil
		default:
			n, err := stdout.Read(buffer)
			if n > 0 {
				// 写入HTTP响应
				_, writeErr := c.Writer.Write(buffer[:n])
				if writeErr != nil {
					// 客户端断开连接，正常情况
					cancel()
					cmd.Process.Kill()
					wg.Wait()
					return nil
				}
				// 确保数据被发送
				if flusher, ok := c.Writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			if err == io.EOF {
				goto finish
			}
			if err != nil {
				cancel()
				cmd.Process.Kill()
				goto finish
			}
		}
	}

finish:
	// 等待命令完成
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		// 超时，强制杀死进程
		cmd.Process.Kill()
	case err := <-done:
		// 正常完成或出错
		if err != nil {
			// 记录错误但不返回给客户端以避免连接中断
		}
	}

	wg.Wait()
	return nil
}

// parseRangeHeader 解析Range请求头
func (t Transcode) parseRangeHeader(rangeHeader string) (start, end int64, err error) {
	if rangeHeader == "" {
		return 0, 0, nil
	}

	// 简单解析bytes=begin-格式
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("无效的Range请求头: %s", rangeHeader)
	}

	rangeValue := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeValue, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("无效的Range格式: %s", rangeHeader)
	}

	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("无效的起始位置: %s", parts[0])
		}
	}

	// 结束位置可以为空，表示到文件末尾
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("无效的结束位置: %s", parts[1])
		}
	}

	return start, end, nil
}
