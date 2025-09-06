package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Crawler相关常量，以CR开头表示Crawler
const (
	CR_DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	CR_DefaultTimeout   = 30 * time.Second
	CR_RefererHeader    = "Referer"
)

// CR_Crawler 爬虫结构体，以CR开头表示Crawler
type CR_Crawler struct {
	UserAgent string
	Timeout   time.Duration
	// 添加重试次数
	MaxRetries int
}

// CR_NewCrawler 创建新的爬虫实例，以CR开头表示Crawler
func CR_NewCrawler() *CR_Crawler {
	return &CR_Crawler{
		UserAgent:  CR_DefaultUserAgent,
		Timeout:    CR_DefaultTimeout,
		MaxRetries: 3, // 默认重试3次
	}
}

// CR_GetData 根据给定URL获取数据的[]byte，以CR开头表示Crawler
func CR_GetData(url string, referer string) ([]byte, error) {
	crawler := CR_NewCrawler()
	// 如果没有提供referer，则尝试从URL推断
	if referer == "" {
		// 简单地将URL作为页面URL（去掉文件名部分）
		// 这是一种常见的防盗链策略，将所在目录作为来源
		referer = CR_inferRefererFromURL(url)
	}
	return crawler.FetchData(url, referer)
}

// CR_inferRefererFromURL 从URL推断可能的referer
func CR_inferRefererFromURL(url string) string {
	// 这里可以实现更复杂的逻辑来推断referer
	// 例如：去掉URL中的文件名部分，保留目录结构

	// 简单实现：去掉URL中最后一个'/'后的内容（假设是文件名）
	lastSlashIndex := strings.LastIndex(url, "/")
	if lastSlashIndex > 0 {
		return url[:lastSlashIndex+1] // 保留到最后一个'/'
	}

	return url
}

// FetchData 根据给定URL获取数据的[]byte
func (c *CR_Crawler) FetchData(url string, referer string) ([]byte, error) {
	var lastErr error

	// 实现重试机制
	for i := 0; i <= c.MaxRetries; i++ {
		// 在重试之间添加随机延迟，避免过于频繁的请求
		if i > 0 {
			delay := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(delay)
		}

		data, err := c.fetchDataOnce(url, referer)
		if err == nil {
			return data, nil
		}

		lastErr = err
		// 如果是连接被强制关闭的错误，则进行重试
		if c.shouldRetry(err) {
			continue
		}

		// 其他错误直接返回
		return nil, err
	}

	return nil, fmt.Errorf("经过 %d 次尝试后仍然失败: %v", c.MaxRetries+1, lastErr)
}

// fetchDataOnce 执行单次数据获取
func (c *CR_Crawler) fetchDataOnce(url string, referer string) ([]byte, error) {
	// 创建HTTP客户端，添加Transport配置
	client := &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置更完整的请求头，模拟真实浏览器
	c.setRealisticHeaders(req, referer)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	return data, nil
}

// setRealisticHeaders 设置更真实的请求头
func (c *CR_Crawler) setRealisticHeaders(req *http.Request, referer string) {
	// 设置User-Agent
	req.Header.Set("User-Agent", c.UserAgent)

	// 设置浏览器相关头信息
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")

	// 添加更多浏览器头信息
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	// 保持连接
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Accept-Charset", "utf-8")

	// 如果提供了referer，则设置referer头
	if referer != "" {
		req.Header.Set(CR_RefererHeader, referer)
	} else {
		// 如果没有提供referer，使用请求的URL作为默认referer
		req.Header.Set(CR_RefererHeader, req.URL.String())
	}
}

// ContainsAny 检查字符串是否包含任意一个子串
func (c *CR_Crawler) ContainsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// shouldRetry 判断是否应该重试
func (c *CR_Crawler) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 针对连接被强制关闭的错误进行重试
	if c.ContainsAny(errStr, []string{
		"connection was forcibly closed",
		"connection reset by peer",
		"broken pipe",
		"timeout",
	}) {
		return true
	}

	return false
}

// CR_SetUserAgent 设置用户代理，以CR开头表示Crawler
func (c *CR_Crawler) CR_SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
}

// CR_SetTimeout 设置超时时间，以CR开头表示Crawler
func (c *CR_Crawler) CR_SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// CR_SetMaxRetries 设置最大重试次数
func (c *CR_Crawler) CR_SetMaxRetries(maxRetries int) {
	c.MaxRetries = maxRetries
}
