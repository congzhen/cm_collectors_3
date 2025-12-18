package middleware

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"net/url"

	"github.com/gin-gonic/gin"
)

func getSourceDomain(c *gin.Context) string {
	// 尝试从 Origin 头部获取
	if origin := c.GetHeader("Origin"); origin != "" {
		if u, err := url.Parse(origin); err == nil {
			return u.Hostname()
		}
	}

	// 尝试从 Referer 头部获取
	if referer := c.GetHeader("Referer"); referer != "" {
		if u, err := url.Parse(referer); err == nil {
			return u.Hostname()
		}
	}

	return ""
}

// AdminLoginApiMiddleware 管理员登录验证中间件
// 该中间件用于验证访问接口的用户是否为管理员
// 如果验证通过，将adminToken信息存入上下文并继续执行后续操作
// 如果验证失败，将中断请求并返回权限错误
func AdminLoginApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未启用管理员登录验证，则直接跳过验证
		if core.Config.General.IsAdminLogin == false {
			c.Next()
			return
		}
		// 从cookie中获取adminToken
		adminTokenString, err := c.Cookie("adminToken")
		if err != nil || adminTokenString == "" {
			if getSourceDomain(c) == "127.0.0.1" {
				// 如果是本地访问，获取请求头中的adminToken
				tokenStr := c.GetHeader("adminToken")
				if tokenStr != "" {
					adminTokenString = tokenStr
				} else {
					response.FailPermissions(c)
					c.Abort()
					return
				}
			} else {
				response.FailPermissions(c)
				c.Abort()
				return
			}
		}
		// 解析token并验证用户类型是否为管理员
		adminToken, err := processors.Login{}.JWTParseToken(adminTokenString)
		if err != nil || adminToken.UserType != datatype.ENUM_UserType_Admin {
			response.FailPermissions(c)
			c.Abort()
			return
		}
		// 将解析后的token信息存入上下文，供后续处理使用
		c.Set("adminToken", adminToken)
		c.Next()
	}

}
