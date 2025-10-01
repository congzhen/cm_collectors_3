package middleware

import (
	"cm_collectors_server/core"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	// 视频流限流相关
	videoStreamLimiters  = make(map[string]*rate.Limiter)
	videoStreamLimitedAt = make(map[string]time.Time)
	limiterMutex         sync.Mutex
	limitedAtMutex       sync.Mutex
	limiterCleanupOnce   sync.Once
	limiterMessage       = "请求过于频繁，请稍后再试"
)

// VideoStreamRateLimitMiddleware 视频流接口限流中间件
// 根据客户端IP进行限流控制
func VideoStreamRateLimitMiddleware() gin.HandlerFunc {
	// 启动清理协程
	limiterCleanupOnce.Do(func() {
		go cleanupVideoStreamLimiters()
	})

	return func(c *gin.Context) {
		// 检查是否启用限流
		if !core.Config.General.VideoRateLimit.Enabled {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 检查是否在冷却期内
		if isVideoStreamRateLimited(clientIP) {
			fmt.Println("######################### 限流器工作中：", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": limiterMessage,
			})
			c.Abort()
			return
		}

		// 获取该IP对应的限流器
		limiter := getVideoStreamLimiter(clientIP)

		// 应用限流
		if !limiter.Allow() {
			// 记录被限流的时间
			setVideoStreamRateLimited(clientIP)
			fmt.Println("######################### 限流器工作中：", clientIP)
			// 如果没有可用令牌，拒绝请求
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": limiterMessage,
			})
			c.Abort()
			return
		}
		// 令牌充足，继续处理请求
		c.Next()
	}
}

// 检查IP是否在限流冷却期内
func isVideoStreamRateLimited(ip string) bool {
	limitedAtMutex.Lock()
	defer limitedAtMutex.Unlock()

	if limitedTime, exists := videoStreamLimitedAt[ip]; exists {
		// 如果在5秒内，则认为仍在冷却期
		if time.Since(limitedTime) < 5*time.Second {
			return true
		}
	}
	return false
}

// 记录IP被限流的时间
func setVideoStreamRateLimited(ip string) {
	limitedAtMutex.Lock()
	defer limitedAtMutex.Unlock()

	videoStreamLimitedAt[ip] = time.Now()
}

// 获取或创建特定IP的限流器
func getVideoStreamLimiter(ip string) *rate.Limiter {
	limiterMutex.Lock()
	defer limiterMutex.Unlock()

	limiter, exists := videoStreamLimiters[ip]
	if !exists {
		// 为每个IP创建独立的限流器
		limiter = rate.NewLimiter(
			rate.Limit(core.Config.General.VideoRateLimit.RequestsPerSecond),
			core.Config.General.VideoRateLimit.Burst)
		videoStreamLimiters[ip] = limiter
	}

	return limiter
}

// 定期清理不活跃的限流器以防止内存泄漏
func cleanupVideoStreamLimiters() {
	for {
		time.Sleep(time.Minute)
		limiterMutex.Lock()
		for ip, limiter := range videoStreamLimiters {
			// 如果限流器的令牌数接近最大值，说明很久没有使用了
			if limiter.Tokens() >= float64(limiter.Burst())-0.1 {
				delete(videoStreamLimiters, ip)
			}
		}
		limiterMutex.Unlock()

		// 同时清理限流记录
		limitedAtMutex.Lock()
		for ip, limitedTime := range videoStreamLimitedAt {
			// 清理1分钟前的限流记录
			if time.Since(limitedTime) > 1*time.Minute {
				fmt.Println("######################### 清理限流桶：", ip)
				delete(videoStreamLimitedAt, ip)
			}
		}
		limitedAtMutex.Unlock()
	}
}
