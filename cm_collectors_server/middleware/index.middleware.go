package middleware

import (
	"cm_collectors_server/core"
	"cm_collectors_server/response"
	"cm_collectors_server/utils"

	"github.com/gin-gonic/gin"
)

func AdminLoginApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if core.Config.General.IsAdminLogin == false {
			c.Next()
			return
		}
		tokenString, err := c.Cookie("token")
		var token *utils.UserTokenCustomClaims
		if err != nil || tokenString == "" {
			response.FailPermissions(c)
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}

}
