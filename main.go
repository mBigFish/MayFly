package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mayfly/config"
	"mayfly/internal/handler"
	"mayfly/internal/middleware"
	"mayfly/internal/service"
	"mayfly/internal/store"

	"github.com/gin-gonic/gin"
)

//go:embed web
var webFS embed.FS

//go:embed payloads
var payloadsFS embed.FS

// resolvePath 优先使用可执行文件所在目录下的资源路径，找不到时回退到当前工作目录。
// 这样无论进入目录运行，还是从 Finder 双击运行（工作目录为用户主目录），都能正确加载 web/、data/ 等资源。
func resolvePath(sub string) string {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(filepath.Join(filepath.Dir(exe), sub)); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return sub
}

func main() {
	// 初始化配置
	config.Init()

	cfg := config.Get()

	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化节点存储
	nodeStore, err := store.New(resolvePath("data/nodes.json"))
	if err != nil {
		log.Fatalf("初始化节点存储失败: %v", err)
	}
	// 命令执行历史缓存存储
	cmdHistoryStore, err := store.NewCmdHistory(resolvePath("data/cmd_history.json"))
	if err != nil {
		log.Fatalf("初始化命令历史存储失败: %v", err)
	}
	nodeHandler := handler.NewNodeHandler(nodeStore, cmdHistoryStore, payloadsFS)
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

	// 静态文件服务：前端已通过 go:embed 编译进二进制，单文件即可运行，无需外部 web/ 目录
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("加载内置前端资源失败: %v", err)
	}
	staticRoot, err := fs.Sub(webRoot, "static")
	if err != nil {
		log.Fatalf("加载内置静态资源失败: %v", err)
	}
	r.StaticFS("/static", http.FS(staticRoot))
	r.StaticFileFS("/login", "login.html", http.FS(webRoot))
	// 首页：直接返回嵌入的 index.html（不能用 StaticFileFS("/")，否则路径 "/" 会被当成目录触发 301 重定向导致白屏）
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(webRoot, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 审计日志存储（需在路由注册前创建，供登录接口使用）
	auditStore := handler.NewAuditStore(resolvePath("data/audit_log.json"), 2000)
	// 初始化登录审计（登录接口不经过审计中间件，需单独记录）
	handler.InitLoginAudit(auditStore)

	// 公开 API
	api := r.Group("/api")
	{
		api.POST("/login", handler.Login)
	}

	// 运行时设置存储
	settingsStore := handler.NewSettingsStore(resolvePath("data/settings.json"))
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
		authAPI.PUT("/system/theme", systemHandler.UpdateTheme)
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