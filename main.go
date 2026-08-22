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
	"mayfly/internal/store"

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

	// 初始化节点存储
	nodeStore, err := store.New("data/nodes.json")
	if err != nil {
		log.Fatalf("初始化节点存储失败: %v", err)
	}
	// 命令执行历史缓存存储
	cmdHistoryStore, err := store.NewCmdHistory("data/cmd_history.json")
	if err != nil {
		log.Fatalf("初始化命令历史存储失败: %v", err)
	}
	nodeHandler := handler.NewNodeHandler(nodeStore, cmdHistoryStore)
	serverTermHandler := handler.NewServerTerminalHandler()
	termWSHandler := handler.NewTerminalWSHandler(nodeStore)
	listenerHandler := handler.NewListenerHandler()

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

	// 公开 API
	api := r.Group("/api")
	{
		api.POST("/login", handler.Login)
	}

	// 审计日志存储
	auditStore := handler.NewAuditStore("data/audit_log.json", 2000)
	// 运行时设置存储
	settingsStore := handler.NewSettingsStore("data/settings.json")
	// 系统管理 Handler
	systemHandler := handler.NewSystemHandler(auditStore, settingsStore)

	// 需要认证的 API
	authAPI := r.Group("/api")
	authAPI.Use(middleware.Auth())
	authAPI.Use(handler.AuditMiddleware(auditStore))
	{
		// 本地 WebSSH 终端（保留原有能力）
		termHandler := handler.NewTerminalHandler(service.NewSessionService())
		authAPI.GET("/terminal", termHandler.HandleTerminal)
		authAPI.GET("/sessions", termHandler.ListSessions)

		// 节点管理
		authAPI.GET("/nodes", nodeHandler.ListNodes)
		authAPI.POST("/nodes", nodeHandler.CreateNode)
		authAPI.POST("/nodes/batch-test", nodeHandler.BatchTest)
		authAPI.PUT("/nodes/:id", nodeHandler.UpdateNode)
		authAPI.DELETE("/nodes/:id", nodeHandler.DeleteNode)

		// 脚本生成器
		authAPI.GET("/scripts/:lang", nodeHandler.GetScript)

		// 节点相关操作（自动解析 :id 并加载节点）
		nodeGroup := authAPI.Group("/nodes/:id")
		nodeGroup.Use(handler.NodeParam(nodeStore))
		{
			nodeGroup.POST("/test", nodeHandler.TestNode)
			nodeGroup.POST("/cmd", nodeHandler.ExecCmd)
			nodeGroup.GET("/cmd/history", nodeHandler.GetCmdHistory)
			nodeGroup.DELETE("/cmd/history", nodeHandler.ClearCmdHistory)
			nodeGroup.POST("/file/list", nodeHandler.ListDir)
			nodeGroup.POST("/file/read", nodeHandler.ReadFile)
			nodeGroup.POST("/file/write", nodeHandler.WriteFile)
			nodeGroup.POST("/file/delete", nodeHandler.DeletePath)
			nodeGroup.POST("/file/rename", nodeHandler.RenamePath)
			nodeGroup.POST("/file/mkdir", nodeHandler.Mkdir)
			nodeGroup.POST("/db", nodeHandler.DBQuery)
			nodeGroup.GET("/terminal", termWSHandler.Handle)
		}

		// 反弹Shell监听管理
		authAPI.POST("/listeners", listenerHandler.StartListener)
		authAPI.GET("/listeners", listenerHandler.ListListeners)
		authAPI.GET("/listeners/:id/output", listenerHandler.GetListenerOutput)
		authAPI.POST("/listeners/:id/stop", listenerHandler.StopListener)
		authAPI.DELETE("/listeners/:id", listenerHandler.DeleteListener)
		authAPI.GET("/reverse-shells", listenerHandler.GeneratePayload)

		// 仪表盘
		dashHandler := handler.NewDashboardHandler(nodeStore, cmdHistoryStore)
		authAPI.GET("/dashboard", dashHandler.Dashboard)

		// 资源管理 - 服务器
		authAPI.GET("/servers", handler.GetServers)
		authAPI.POST("/servers", handler.CreateServer)
		authAPI.PUT("/servers", handler.UpdateServer)
		authAPI.DELETE("/servers", handler.DeleteServer)
		authAPI.POST("/servers/test", handler.TestSSHConnection)
		authAPI.GET("/servers/:id/terminal", serverTermHandler.Handle)

		// 系统管理
		authAPI.GET("/system/info", systemHandler.SysInfo)
		authAPI.GET("/system/settings", systemHandler.GetSettings)
		authAPI.PUT("/system/settings", systemHandler.UpdateSettings)
		authAPI.POST("/system/password", systemHandler.ChangePassword)
		authAPI.GET("/system/audit-logs", systemHandler.GetAuditLogs)
		authAPI.DELETE("/system/audit-logs", systemHandler.ClearAuditLogs)
		authAPI.GET("/system/export", systemHandler.ExportData)
		authAPI.POST("/system/import", systemHandler.ImportData)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Mayfly WebShell 管理器启动于 http://0.0.0.0:%s", cfg.ServerPort)
	log.Printf("默认账号: %s（可通过 MAYFLY_USER / MAYFLY_PASS 修改）", cfg.Username)
	log.Printf("节点数据存储: data/nodes.json")

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}