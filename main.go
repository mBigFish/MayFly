package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"mayfly/config"
	"mayfly/internal/handler"
	"mayfly/internal/middleware"
	"mayfly/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.Init()

	cfg := config.Get()

	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 禁用静态资源和页面的浏览器缓存，确保前端更新即时生效
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" || path == "/login" || strings.HasPrefix(path, "/static") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// 静态文件服务
	r.Static("/static", "./web/static")
	r.StaticFile("/login", "./web/login.html")
	r.StaticFile("/", "./web/index.html")

	// API 路由
	api := r.Group("/api")
	{
		api.POST("/login", handler.Login)
	}

	// 需要认证的 API
	authAPI := r.Group("/api")
	authAPI.Use(middleware.Auth())
	{
		// 终端 WebSocket
		termHandler := handler.NewTerminalHandler(service.NewSessionService())
		authAPI.GET("/terminal", termHandler.HandleTerminal)
		authAPI.GET("/sessions", termHandler.ListSessions)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Mayfly WebShell 启动于 http://0.0.0.0:%s", cfg.ServerPort)
	log.Printf("默认用户: %s (请通过环境变量 MAYFLY_USER/MAYFLY_PASS 修改)", cfg.Username)
	log.Printf("WebSocket 终端: ws://0.0.0.0:%s/api/terminal?token=<JWT>", cfg.ServerPort)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
