// Command server 是 WebShell Manager 的服务端入口。
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webshell-manager/webshell-manager/internal/api"
	"github.com/webshell-manager/webshell-manager/internal/auth"
	"github.com/webshell-manager/webshell-manager/internal/config"
	"github.com/webshell-manager/webshell-manager/internal/crypto"
	"github.com/webshell-manager/webshell-manager/internal/database"
	"github.com/webshell-manager/webshell-manager/internal/logger"
	"github.com/webshell-manager/webshell-manager/internal/migrations"
	"github.com/webshell-manager/webshell-manager/internal/protocol"
	"github.com/webshell-manager/webshell-manager/internal/protocol/adapters"
	"github.com/webshell-manager/webshell-manager/internal/target"
	"github.com/webshell-manager/webshell-manager/internal/transport"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

// run 承载主流程，返回错误以便统一处理，确保 defer 能正常执行。
func run() error {
	// 解析命令行参数。
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置。
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	// 初始化日志。
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Path); err != nil {
		return err
	}
	defer logger.Sync()

	logger.Logger.Info("WebShell Manager 启动")

	// 连接数据库。
	db, err := database.Open(&cfg.Database)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Logger.Error("关闭数据库失败")
		}
	}()

	// 执行数据库迁移。
	if err := database.Migrate(db, migrations.All()); err != nil {
		return err
	}

	// 初始化加密器（敏感字段加密）。
	encrypt, err := crypto.NewAESGCM([]byte(cfg.Security.EncryptionKey))
	if err != nil {
		return err
	}

	// 初始化认证服务。
	jwtTTL := time.Duration(cfg.Security.SessionTimeout) * time.Second
	authSvc := auth.NewService(cfg.Security.JWTSecret, jwtTTL)
	rateLimiter := auth.NewRateLimiter(5, time.Minute)

	// 初始化目标服务。
	targetRepo := target.NewRepository(db)
	targetSvc := target.NewService(targetRepo, encrypt)

	// 初始化传输层与协议注册表。
	// allowLocal=true 允许访问内网，便于本地授权测试服务验证（生产可改为 false）。
	httpTransport := transport.NewHTTPTransport(30*time.Second, true)
	protocolReg := protocol.NewRegistry()
	protocolReg.Register(adapters.NewPHPAdapter(httpTransport))

	// 确保初始 admin 用户存在（密码从环境变量读取）。
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
		logger.Logger.Warn("未设置 ADMIN_PASSWORD，使用默认密码 admin123，请在生产环境修改")
	}
	if err := auth.EnsureAdminUser(context.Background(), "admin", adminPassword); err != nil {
		return err
	}

	// 构建路由并启动 HTTP 服务。
	router := api.NewRouter(&api.Dependencies{
		AuthSvc:     authSvc,
		TargetSvc:   targetSvc,
		RateLimiter: rateLimiter,
		ProtocolReg: protocolReg,
	})

	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,
	}

	logger.Logger.Info("HTTP 服务启动，监听 " + cfg.Addr())

	// 监听系统信号，实现优雅关闭。
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		logger.Logger.Info("收到退出信号 " + sig.String() + "，正在优雅关闭")
	}

	// 优雅关闭：等待进行中的请求处理完成。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Error("HTTP 服务关闭失败")
		return err
	}

	logger.Logger.Info("HTTP 服务已关闭")
	return nil
}
