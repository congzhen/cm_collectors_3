package middleware

import (
	"cm_collectors_server/internal/maintenance"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UpdateMaintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if !strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/api/update/") || p == "/api/health" {
			c.Next()
			return
		}
		if !maintenance.Begin() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"status": false, "msg": "正在准备安装更新，请稍后重试"})
			return
		}
		defer maintenance.End()
		c.Next()
	}
}
