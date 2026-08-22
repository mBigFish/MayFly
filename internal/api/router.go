// Package api 定义 HTTP 路由与处理器装配。
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/handlers"
	"github.com/webshell-manager/webshell-manager/internal/auth"
	"github.com/webshell-manager/webshell-manager/internal/file"
	"github.com/webshell-manager/webshell-manager/internal/protocol"
	"github.com/webshell-manager/webshell-manager/internal/session"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// Dependencies 是路由装配所需的依赖集合。
type Dependencies struct {
	AuthSvc     *auth.Service
	TargetSvc   *target.Service
	RateLimiter *auth.RateLimiter
	ProtocolReg *protocol.Registry
	SessionMgr  *session.Manager
}

// NewRouter 构建并返回 Gin 引擎，注册所有路由。
func NewRouter(deps *Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	authHandler := handlers.NewAuthHandler(deps.AuthSvc, deps.RateLimiter)
	targetHandler := handlers.NewTargetHandler(deps.TargetSvc)
	targetCheckHandler := handlers.NewTargetCheckHandler(deps.TargetSvc, deps.ProtocolReg)

	fileSvc := file.NewService(deps.ProtocolReg)
	fileHandler := handlers.NewFileHandler(deps.TargetSvc, fileSvc)
	sessionHandler := handlers.NewSessionHandler(deps.TargetSvc, deps.SessionMgr)
	terminalHandler := handlers.NewTerminalHandler(deps.SessionMgr, deps.ProtocolReg, deps.AuthSvc)

	api := r.Group("/api/v1")
	{
		// 认证接口（无需鉴权）。
		api.POST("/auth/login", authHandler.Login)

		// 目标接口（需鉴权 + 权限）。
		targets := api.Group("/targets")
		targets.Use(auth.JWTMiddleware(deps.AuthSvc))
		{
			targets.GET("", auth.RequirePermission(auth.PermTargetRead), targetHandler.List)
			targets.POST("", auth.RequirePermission(auth.PermTargetCreate), targetHandler.Create)
			targets.GET("/:id", auth.RequirePermission(auth.PermTargetRead), targetHandler.Get)
			targets.PUT("/:id", auth.RequirePermission(auth.PermTargetUpdate), targetHandler.Update)
			targets.DELETE("/:id", auth.RequirePermission(auth.PermTargetDelete), targetHandler.Delete)
			targets.POST("/:id/check", auth.RequirePermission(auth.PermTargetRead), targetCheckHandler.Check)

			// 文件管理（需 file:read / file:write / file:delete 权限）。
			targets.GET("/:id/files", auth.RequirePermission(auth.PermFileRead), fileHandler.List)
			targets.POST("/:id/files/read", auth.RequirePermission(auth.PermFileRead), fileHandler.Read)
			targets.POST("/:id/files/write", auth.RequirePermission(auth.PermFileWrite), fileHandler.Write)
			targets.POST("/:id/files/rename", auth.RequirePermission(auth.PermFileWrite), fileHandler.Rename)
			targets.POST("/:id/files/mkdir", auth.RequirePermission(auth.PermFileWrite), fileHandler.Mkdir)
			targets.POST("/:id/files/delete", auth.RequirePermission(auth.PermFileDelete), fileHandler.Delete)
		}

		// 会话接口（需鉴权）。
		sessions := api.Group("/sessions")
		sessions.Use(auth.JWTMiddleware(deps.AuthSvc))
		{
			sessions.POST("", sessionHandler.Create)
			sessions.GET("", sessionHandler.List)
			sessions.DELETE("/:id", sessionHandler.Close)
		}
	}

	// WebSocket 终端（需通过 query 参数携带 token，见前端）。
	ws := r.Group("/ws/v1")
	{
		ws.GET("/session/:id", terminalHandler.ServeWS)
	}

	return r
}
