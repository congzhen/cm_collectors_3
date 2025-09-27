package processors

import (
	"cm_collectors_server/core"
	processorscache "cm_collectors_server/processorsCache"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Video struct{}

func (v Video) VideoMP4Stream(c *gin.Context, dramaSeriesId string, needEncoding bool) error {
	src, err := ResourcesDramaSeries{}.GetSrc(dramaSeriesId)
	if err != nil {
		return err
	}
	// 检查文件是否存在
	fileInfo, err := os.Stat(src)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return nil
	}

	if !needEncoding {
		fmt.Println("######################### 原始流(不检测编码)")
		return v.VideoMP4Stream_Play(c, fileInfo, src)
	}

	// 获取视频格式信息并检查是否需要转码
	needTranscode := false

	//formatInfo, err := processorsffmpeg.VideoInfo{}.GetVideoFormatInfo(src)
	formatInfo, err := processorscache.CacheVideoInfoLastUse{}.GetVideoInfoHandle(src)
	if err != nil {
		// 如果无法获取格式信息，假设需要转码以确保兼容性
		fmt.Printf("警告: 无法获取视频格式信息，将进行转码以确保兼容性: %v\n", err)
		needTranscode = true
	} else {
		// 检查视频是否与Web兼容
		pf_videoInfo := processorsffmpeg.VideoInfo{}
		// 设置支持的编解码器
		pf_videoInfo.SetSupportedVideoCodecs(core.Config.Play.PlayVideoFormats)
		pf_videoInfo.SetSupportedAudioCodecs(core.Config.Play.PlayAudioFormats)
		needTranscode = !pf_videoInfo.IsWebCompatible(formatInfo)
	}
	if needTranscode {
		fmt.Println("######################### 转码流")
		// 如果需要转码，使用流式转码
		return v.VideoMP4Stream_TranscodePlay(c, src)
	} else {
		fmt.Println("######################### 原始流")
		return v.VideoMP4Stream_Play(c, fileInfo, src)
	}

}

func (v Video) VideoMP4Stream_Play(c *gin.Context, fileInfo fs.FileInfo, src string) error {
	// 获取Range头部信息
	rangeHeader := c.GetHeader("Range")

	// 为TVBox和投屏设备优化的响应头
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Connection", "keep-alive")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Content-Type", "video/mp4")

	// 缓冲区倍率
	var bufferRatio int64 = 2

	if rangeHeader != "" {
		// 解析Range头部
		re := regexp.MustCompile(`bytes=(\d+)-(\d*)`)
		matches := re.FindStringSubmatch(rangeHeader)
		if len(matches) < 2 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid range format"})
			return nil
		}

		// 解析起始位置
		start, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid start range"})
			return nil
		}

		// 解析结束位置
		var end int64
		if matches[2] != "" {
			end, err = strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid end range"})
				return nil
			}
		} else {
			// 对于TVBox和4K高码率视频，使用更大的默认块大小以减少请求次数
			// 根据文件大小动态调整块大小
			var chunkSize int64
			fileSize := fileInfo.Size()
			switch {
			case fileSize > 10*1024*1024*1024: // 大于10GB (超大4K文件)
				chunkSize = 16 * 1024 * 1024 // 16MB
			case fileSize > 4*1024*1024*1024: // 大于4GB
				chunkSize = 8 * 1024 * 1024 // 8MB
			case fileSize > 1*1024*1024*1024: // 大于1GB
				chunkSize = 4 * 1024 * 1024 // 4MB
			default:
				chunkSize = 2 * 1024 * 1024 // 2MB
			}

			// 增加倍率
			chunkSize = chunkSize * bufferRatio

			end = start + chunkSize
			if end > fileSize-1 {
				end = fileSize - 1
			}
		}

		// 确保结束位置不超过文件大小
		if end > fileInfo.Size()-1 {
			end = fileInfo.Size() - 1
		}

		// 设置响应头
		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))

		// 针对TVBox和投屏设备的优化
		c.Header("Content-Disposition", "inline")

		// 设置状态码为206 Partial Content
		c.Status(http.StatusPartialContent)

		// 使用文件句柄缓存优化并发读取
		file, err := processorscache.CacheFileLastUse{}.GetFileHandle(src)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return nil
		}

		/*
			// 打开文件并读取指定范围内容
			file, err := os.Open(src)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
				return nil
			}
			defer func() {
				// 确保文件被正确关闭
				if cerr := file.Close(); cerr != nil {
					fmt.Printf("Warning: failed to close file: %v\n", cerr)
				}
			}()
		*/

		// 设置读取范围
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek file"})
			return nil
		}

		// 根据TVBox的特性使用优化的缓冲区大小
		// 对于4K高码率视频使用更大的缓冲区以减少系统调用次数
		bufferSize := 64 * 1024 // 64KB缓冲区，适合PotPlayer
		fileSize := fileInfo.Size()
		switch {
		case fileSize > 10*1024*1024*1024: // 超大文件 (10GB+)
			bufferSize = 256 * 1024 // 256KB缓冲区
		case fileSize > 4*1024*1024*1024: // 大文件 (4GB+)
			bufferSize = 128 * 1024 // 128KB缓冲区
		case fileSize > 1*1024*1024*1024: // 中等文件 (1GB+)
			bufferSize = 96 * 1024 // 96KB缓冲区
		}

		buffer := make([]byte, bufferSize)
		bytesToRead := end - start + 1
		bytesRead := int64(0)

		// 分块读取并传输数据
		for bytesRead < bytesToRead {
			// 检查客户端是否断开连接
			select {
			case <-c.Request.Context().Done():
				// 客户端断开连接，这是正常情况
				return nil
			default:
			}

			// 计算本次读取的字节数
			chunkSize := int64(len(buffer))
			if bytesRead+chunkSize > bytesToRead {
				chunkSize = bytesToRead - bytesRead
			}

			// 读取数据
			n, err := file.Read(buffer[:chunkSize])
			if err != nil && err != io.EOF {
				// 发生读取错误
				fmt.Printf("Error reading file: %v\n", err)
				return nil
			}

			if n > 0 {
				// 写入响应
				_, writeErr := c.Writer.Write(buffer[:n])
				if writeErr != nil {
					// 客户端断开连接是正常情况
					return nil
				}

				// 确保数据被发送
				if flusher, ok := c.Writer.(http.Flusher); ok {
					// 每次读取后立即flush以确保数据及时传输给播放器
					flusher.Flush()
				}
				bytesRead += int64(n)
			}

			if err == io.EOF {
				break
			}
		}
	} else {
		// 没有Range头部，返回整个文件
		c.Header("Content-Type", "video/mp4")
		c.Header("Accept-Ranges", "bytes")
		c.Header("Connection", "keep-alive")
		c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

		// 设置状态码
		c.Status(http.StatusOK)

		// 使用http.ServeFile处理整个文件传输
		http.ServeFile(c.Writer, c.Request, src)
	}
	return nil
}

// VideoMP4Stream_TranscodePlay 实现流式转码，边转码边传输
func (v Video) VideoMP4Stream_TranscodePlay(c *gin.Context, src string) error {
	// 直接使用转码器处理流式传输
	err := processorsffmpeg.Transcode{}.VideoStreamTranscode(c, src)
	if err != nil {
		fmt.Printf("视频转码失败: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "视频转码失败: " + err.Error(),
		})
		return err
	}
	return nil
}

func (v Video) GetVideoM3u8(dramaSeriesId string) ([]byte, error) {
	dramaSeries, err := ResourcesDramaSeries{}.Info(dramaSeriesId)
	if err != nil {
		return nil, err
	}

	m3u8FileSrc := dramaSeries.Src + ".m3u8"

	// 检查文件是否存在
	if _, err := os.Stat(m3u8FileSrc); err == nil {
		// 文件存在，读取文件内容
		content, err := os.ReadFile(m3u8FileSrc)
		if err != nil {
			return nil, fmt.Errorf("读取m3u8文件失败: %v", err)
		}
		return content, nil
	} else if os.IsNotExist(err) {
		// 文件不存在，创建m3u8文件
		content, err := processorsffmpeg.M3U8{}.CreateM3u8File(dramaSeriesId, dramaSeries.Src, m3u8FileSrc)
		if err != nil {
			return nil, fmt.Errorf("创建m3u8文件失败: %v", err)
		}
		return content, nil
	} else {
		// 其他错误
		return nil, fmt.Errorf("检查文件状态时出错: %v", err)
	}
}
func (v Video) PlayVideoM3u8(dramaSeriesId string, start, duration float64) (*exec.Cmd, error) {
	dramaSeries, err := ResourcesDramaSeries{}.Info(dramaSeriesId)
	if err != nil {
		return nil, err
	}
	return processorsffmpeg.M3U8{}.PlayVideoM3u8(dramaSeries.Src, start, duration)
}
