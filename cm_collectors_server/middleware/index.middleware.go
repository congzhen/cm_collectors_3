package middleware

import (
	"cm_collectors_server/core"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

func AdminLoginApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if core.Config.General.IsAdminLogin == false {
			c.Next()
			return
		}
		adminTokenString, err := c.Cookie("adminToken")
		if err != nil || adminTokenString == "" {
			response.FailPermissions(c)
			c.Abort()
			return
		}
		adminToken, err := processors.Login{}.JWTParseToken(adminTokenString)
		if err != nil {
			response.FailPermissions(c)
			c.Abort()
			return
		}
		c.Set("adminToken", adminToken)
		c.Next()
	}

}
