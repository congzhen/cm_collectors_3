package processors

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Video struct{}

func (v Video) VideoMP4Stream(c *gin.Context, dramaSeriesId string) error {
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
	// 获取Range头部信息
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// 解析Range头部
		re := regexp.MustCompile(`=(\d+)-(\d+)?`)
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

		// 解析结束位置，如果未指定则默认为起始位置+1MB
		var end int64
		if matches[2] != "" {
			end, err = strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid end range"})
				return nil
			}
		} else {
			end = start + 1024*1024*1 // 1MB
		}

		// 确保结束位置不超过文件大小
		if end > fileInfo.Size()-1 {
			end = fileInfo.Size() - 1
		}

		// 设置响应头
		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
		c.Header("Accept-Ranges", "bytes")

		// 设置状态码为206 Partial Content
		c.Status(http.StatusPartialContent)

		// 打开文件并读取指定范围内容
		file, err := os.Open(src)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return nil
		}
		defer file.Close()

		// 设置读取范围
		_, err = file.Seek(start, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek file"})
			return nil
		}

		// 创建缓冲区并读取数据
		buffer := make([]byte, end-start+1)
		_, err = file.Read(buffer)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
			return nil
		}

		// 写入响应
		c.Data(http.StatusPartialContent, "video/mp4", buffer)
	} else {
		// 没有Range头部，返回整个文件
		c.Header("Content-Type", "video/mp4")
		c.Header("Accept-Ranges", "bytes")
		c.File(src)
	}

	return nil
}
