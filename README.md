# Mayfly WebShell

一个基于 Go + WebSocket + xterm.js 的 Web 终端管理工具，专为无图形界面的 Linux 服务器设计。

## 功能特性

- **Web 终端** - 通过浏览器直接访问服务器终端，支持完整的交互式 Shell
- **多会话管理** - 同时打开多个终端标签页，快速切换
- **JWT 认证** - 安全的登录认证机制
- **自适应窗口** - 终端大小随浏览器窗口自动调整
- **多主题支持** - 深色 / 浅色 / Solarized 主题切换
- **快捷键** - `Ctrl+Shift+T` 快速新建终端
- **全屏模式** - 一键全屏终端
- **现代 UI** - 暗色主题，简洁美观

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go + Gin + gorilla/websocket + creack/pty |
| 前端 | 原生 JS + xterm.js + FontAwesome |
| 认证 | JWT (golang-jwt) |

## 快速开始

### 1. 编译

```bash
# 在项目根目录执行
go mod tidy
go build -o mayfly
```

### 2. 运行

```bash
# 使用默认配置运行 (端口 8080, 用户 admin/mayfly123)
./mayfly

# 自定义配置
MAYFLY_PORT=9090 \
MAYFLY_USER=myuser \
MAYFLY_PASS=mypassword \
MAYFLY_JWT_SECRET=my-secret \
./mayfly
```

### 3. 访问

浏览器打开 `http://<服务器IP>:8080`，输入用户名密码登录即可使用终端。

## 配置项

所有配置通过环境变量设置：

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `MAYFLY_PORT` | 8080 | 监听端口 |
| `MAYFLY_USER` | admin | 登录用户名 |
| `MAYFLY_PASS` | mayfly123 | 登录密码 |
| `MAYFLY_JWT_SECRET` | mayfly-secret-key-change-in-production | JWT 签名密钥 |
| `MAYFLY_SHELL` | bash | 默认 Shell |
| `MAYFLY_SESSION_TIMEOUT` | 30 | 会话超时（分钟） |

## 项目结构

```
Mayfly/
├── main.go                     # 程序入口
├── go.mod
├── config/
│   └── config.go               # 配置管理
├── internal/
│   ├── handler/
│   │   ├── auth.go             # 登录/JWT 处理
│   │   └── terminal.go         # WebSocket 终端处理
│   ├── middleware/
│   │   └── auth.go             # 认证中间件
│   ├── model/
│   │   └── session.go          # 会话模型
│   └── service/
│       ├── pty.go              # PTY 管理
│       └── session.go          # 会话服务
└── web/
    ├── index.html              # 终端主页面
    ├── login.html              # 登录页面
    └── static/
        ├── css/
        │   └── style.css       # 全局样式
        └── js/
            ├── app.js          # 终端页面逻辑
            └── login.js        # 登录页面逻辑
```

## 部署建议

### 使用 systemd 服务

创建 `/etc/systemd/system/mayfly.service`：

```ini
[Unit]
Description=Mayfly WebShell
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/mayfly
ExecStart=/opt/mayfly/mayfly
Environment=MAYFLY_PORT=8080
Environment=MAYFLY_USER=admin
Environment=MAYFLY_PASS=your-strong-password
Environment=MAYFLY_JWT_SECRET=your-random-secret
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable mayfly
sudo systemctl start mayfly
```

### 使用 Nginx 反向代理 + HTTPS

```nginx
server {
    listen 443 ssl;
    server_name shell.example.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 安全注意事项

1. **修改默认密码** - 务必通过环境变量修改默认用户名和密码
2. **使用 HTTPS** - 生产环境建议配合 Nginx 使用 HTTPS，避免凭据明文传输
3. **修改 JWT Secret** - 使用随机字符串作为 JWT 签名密钥
4. **防火墙限制** - 仅开放必要端口，限制访问来源
5. **最小权限** - 建议以非 root 用户运行（但需确保该用户有 shell 访问权限）

## 后续规划

- [ ] 文件管理器（上传/下载/浏览）
- [ ] 系统监控面板（CPU/内存/磁盘/网络）
- [ ] SSH 远程连接管理
- [ ] 终端录制与回放
- [ ] 多用户支持与权限管理
- [ ] 命令历史搜索
- [ ] 自定义快捷键
- [ ] Docker 部署支持

## License

MIT
