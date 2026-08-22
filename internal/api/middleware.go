package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mayfly/internal/service"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 放行登录接口
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/auth/login") {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "未登录或 token 缺失",
			})
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "token 格式错误",
			})
			c.Abort()
			return
		}

		// 解析 JWT
		claims, err := service.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "token 无效或已过期",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "权限不足",
		})
		c.Abort()
	}
}

// Recovery 全局错误恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "服务器内部错误",
		})
		c.Abort()
	})
}
