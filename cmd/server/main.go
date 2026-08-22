package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"mayfly/internal/api"
	"mayfly/internal/config"
	"mayfly/internal/database"
	"mayfly/internal/logger"
	"mayfly/internal/model"
	_ "mayfly/internal/protocol/adapters" // 注册协议适配器
	"mayfly/internal/service"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		// 尝试可执行文件目录
		if exe, e2 := os.Executable(); e2 == nil {
			p2 := filepath.Join(filepath.Dir(exe), "configs", "config.yaml")
			cfg, err = config.Load(p2)
		}
		if err != nil {
			fmt.Printf("无法加载配置文件，使用默认配置: %v\n", err)
			cfg = &config.Config{}
			cfg.Server.Host = "0.0.0.0"
			cfg.Server.Port = "8080"
			cfg.Server.Mode = "debug"
		}
	}
	config.Set(cfg)

	// 2. 初始化日志
	logger.Init(
		cfg.Log.Level,
		cfg.Log.Encoding,
		cfg.Log.Output,
		cfg.Log.MaxSize,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAge,
	)
	defer logger.Sync()
	logger.S().Info("Mayfly 启动中...")

	// 3. 初始化数据库
	if err := database.Init(cfg.Database.Driver, cfg.Database.DSN); err != nil {
		logger.S().Fatalw("数据库初始化失败", "error", err)
	}
	logger.S().Info("数据库初始化成功")

	// 3.1 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.User{}, &model.Target{}, &model.Session{},
		&model.AuditLog{}, &model.Task{}, &model.Listener{},
		&model.Server{}, &model.CmdHistory{}, &model.RequestLog{},
	); err != nil {
		logger.S().Fatalw("数据库迁移失败", "error", err)
	}
	logger.S().Info("数据库迁移完成")

	// 3.2 初始化管理员账户
	if err := service.InitAdminIfEmpty(database.Get(), cfg.Auth.InitUsername, cfg.Auth.InitPassword); err != nil {
		logger.S().Errorw("初始化管理员失败", "error", err)
	} else {
		logger.S().Info("管理员账户检查完成")
	}

	// 4. 初始化路由
	r := api.SetupRouter(cfg.Server.Mode)

	// 5. 启动服务
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.S().Infof("Mayfly 服务启动，监听地址: %s，模式: %s", addr, cfg.Server.Mode)
	if err := r.Run(addr); err != nil {
		logger.S().Fatalw("服务启动失败", "error", err)
	}
}
