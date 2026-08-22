package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Router 全局路由引擎
var Router *gin.Engine

// SetupRouter 初始化路由
func SetupRouter(mode string) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		OK(c, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")

	// 认证相关（无需 JWT）
	auth := v1.Group("/auth")
	{
		auth.POST("/login", Login)
	}
	// 认证相关（需要 JWT）
	authAuthed := v1.Group("/auth", JWTAuth())
	{
		authAuthed.POST("/logout", Logout)
		authAuthed.GET("/info", GetUserInfo)
		authAuthed.POST("/change-password", ChangePassword)
	}

	// 目标管理（需要 JWT）
	targets := v1.Group("/targets", JWTAuth())
	{
		targets.GET("", ListTargets)
		targets.GET("/:id", GetTarget)
		targets.POST("", CreateTarget)
		targets.PUT("/:id", UpdateTarget)
		targets.DELETE("/:id", DeleteTarget)
		targets.POST("/:id/check", CheckTarget)
		targets.POST("/:id/execute", ExecuteCommand)
		targets.POST("/:id/files", ListFiles)
		targets.POST("/:id/files/read", ReadFile)
		targets.POST("/:id/files/write", WriteFile)
		targets.POST("/:id/files/delete", DeleteFile)
		targets.POST("/:id/files/rename", RenameFile)
		targets.POST("/:id/files/mkdir", Mkdir)
		targets.POST("/:id/files/download", DownloadFile)
		targets.GET("/:id/info", GetSysInfo)
		targets.GET("/:id/history", ListCmdHistory)
		targets.DELETE("/:id/history", ClearCmdHistory)
		targets.POST("/batch-check", BatchCheck)
		targets.POST("/encrypt", EncryptPassword)
	}

	// SSH 服务器管理（需要 JWT）
	servers := v1.Group("/servers", JWTAuth())
	{
		servers.GET("", ListServers)
		servers.GET("/:id", GetServer)
		servers.POST("", CreateServer)
		servers.PUT("/:id", UpdateServer)
		servers.DELETE("/:id", DeleteServer)
		servers.POST("/:id/test", TestServer)
	}

	// Payload 生成器（需要 JWT）
	payloads := v1.Group("/payloads", JWTAuth())
	{
		payloads.GET("/reverse", GeneratePayloads)
		payloads.GET("/shell", GenerateShellScript)
	}

	// 会话管理（需要 JWT）
	sessions := v1.Group("/sessions", JWTAuth())
	{
		sessions.GET("", ListSessions)
		sessions.GET("/:id", GetSession)
		sessions.POST("/:id/close", CloseSession)
		sessions.DELETE("/:id", DeleteSession)
	}

	// 监听器管理（需要 JWT）
	listeners := v1.Group("/listeners", JWTAuth())
	{
		listeners.GET("", ListListeners)
		listeners.POST("", CreateListener)
		listeners.POST("/:id/start", StartListener)
		listeners.POST("/:id/stop", StopListener)
		listeners.DELETE("/:id", DeleteListener)
	}

	// 任务管理（需要 JWT）
	tasks := v1.Group("/tasks", JWTAuth())
	{
		tasks.GET("", ListTasks)
		tasks.GET("/:id", GetTask)
		tasks.POST("", CreateTask)
		tasks.POST("/:id/cancel", CancelTask)
		tasks.DELETE("/:id", DeleteTask)
	}

	// 审计日志（需要 JWT，仅 admin/operator 可查）
	audit := v1.Group("/audit", JWTAuth())
	{
		audit.GET("", ListAuditLogs)
	}

	// 请求日志 / Request Inspector（需要 JWT）
	reqLogs := v1.Group("/request-logs", JWTAuth())
	{
		reqLogs.GET("", ListRequestLogs)
		reqLogs.GET("/:id", GetRequestLog)
	}

	// 插件管理（需要 JWT）
	plugins := v1.Group("/plugins", JWTAuth())
	{
		plugins.GET("", ListPlugins)
		plugins.POST("/:name/execute", ExecutePlugin)
	}

	// 仪表盘（需要 JWT）
	dashboard := v1.Group("/dashboard", JWTAuth())
	{
		dashboard.GET("", GetDashboard)
	}

	// WebSocket 端点
	r.GET("/ws/terminal", TerminalWS)
	r.GET("/ws/listener", ListenerWS)

	// 前端静态文件服务（SPA）
	distDir := "web/dist"
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		// 尝试可执行文件目录
		if exe, e2 := os.Executable(); e2 == nil {
			distDir = filepath.Join(filepath.Dir(exe), "web", "dist")
		}
	}
	if _, err := os.Stat(distDir); err == nil {
		r.Use(serveSPA(distDir))
	}

	Router = r
	return r
}

// serveSPA 提供 SPA 静态文件服务，支持前端路由
func serveSPA(distDir string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(distDir))
	indexPath := filepath.Join(distDir, "index.html")

	return func(c *gin.Context) {
		// 如果是 API 请求，跳过
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ws") || path == "/health" {
			c.Next()
			return
		}

		// 检查请求的文件是否存在
		fullPath := filepath.Join(distDir, path)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// 文件不存在，返回 index.html（SPA 路由）
		if indexData, err := os.ReadFile(indexPath); err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
			c.Abort()
			return
		}

		c.Next()
	}
}
