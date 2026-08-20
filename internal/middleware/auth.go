package middleware

import (
	"net/http"
	"strings"

	"mayfly/internal/handler"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件，验证 JWT token
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 或 query 参数中获取 token
		token := c.GetHeader("Authorization")
		if token != "" {
			// 去掉 "Bearer " 前缀
			token = strings.TrimPrefix(token, "Bearer ")
		} else {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		claims, err := handler.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("token", token)
		c.Next()
	}
}
