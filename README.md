# Mayfly

渗透测试管理平台 — 前后端分离架构，Go + Vue 3。

## 功能概览

| 模块 | 说明 |
|------|------|
| 目标管理 | WebShell 目标 CRUD、连接测试、命令执行 |
| 文件管理 | 列目录、读/写/编辑（Monaco Editor）、重命名、删除、下载、上传、新建目录 |
| 终端 | 本地终端 + SSH 交互式终端（WebSocket + xterm.js） |
| SSH 服务器 | 服务器管理、连接测试 |
| 脚本生成 | WebShell 脚本（PHP/JSP/ASP/ASPX，从 payloads/ 目录读取） |
| 任务中心 | 批量连接测试、批量命令执行、Worker Pool 并发 |
| 请求检查器 | 记录每次请求的 Request/Response/Duration/Status |
| 插件系统 | 插件接口 + 注册中心 + 内置插件（SystemInfo/ProcessViewer/NetworkInfo） |
| 审计日志 | 用户操作审计 |
| 用户认证 | JWT + RBAC（admin/operator/viewer） |

## 技术栈

**后端**: Go 1.23 + Gin + GORM + SQLite + Zap + JWT
**前端**: Vue 3 + TypeScript + Vite + Element Plus + Pinia + xterm.js + Monaco Editor

## 快速开始

### 本地开发

```bash
# 1. 构建前端
cd frontend
npm install
npm run build

# 2. 编译后端
cd ..
go build -o mayfly ./cmd/server/

# 3. 启动
./mayfly --config configs/config.yaml
```

访问 `http://localhost:8080`，默认账户 `admin` / `mayfly123`。

### 前端热更新开发

```bash
# 终端 1：启动后端
go build -o mayfly ./cmd/server/ && ./mayfly --config configs/config.yaml

# 终端 2：启动前端 dev server
cd frontend && npm run dev
```

前端运行在 `http://localhost:3000`，API 自动代理到后端 8080。

### Docker 部署

```bash
# 构建并启动
docker compose up -d

# 查看日志
docker compose logs -f

# 停止
docker compose down
```

## 项目结构

```
Mayfly/
├── cmd/server/main.go          # 主入口
├── configs/config.yaml         # 配置文件
├── internal/
│   ├── api/                    # API 路由 + 处理器
│   ├── config/                 # 配置加载
│   ├── crypto/                 # AES 加解密
│   ├── database/               # SQLite + GORM
│   ├── logger/                 # Zap 日志
│   ├── model/                  # 数据模型
│   ├── plugin/                 # 插件系统
│   ├── protocol/               # 协议适配器（PHP/JSP/ASP/ASPX）
│   ├── service/                # 业务逻辑 + Task Worker Pool
│   └── transport/              # HTTP 传输层
├── frontend/                   # Vue 3 前端
│   └── src/
│       ├── api/                # API 调用
│       ├── router/             # Vue Router
│       └── views/              # 页面组件
├── payloads/                   # WebShell 脚本文件
├── web/dist/                   # 前端构建产物（SPA）
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## WebShell 协议

脚本文件位于 `payloads/` 目录，支持自定义连接密码：

| 类型 | 文件 | 协议 |
|------|------|------|
| PHP | shell.php | base64+JSON |
| JSP | shell.jsp | base64+JSON |
| ASPX | shell.aspx | base64+JSON |
| ASP | shell.asp | 明文 |

修改脚本只需编辑文件，无需重新编译。

## 配置说明

`configs/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug  # debug / release

database:
  path: data/mayfly.db

auth:
  jwt_secret: your-secret-key
  jwt_expire: 24h

terminal:
  shell: powershell.exe  # Windows / bash (Linux)

log:
  level: info
  dir: logs
```

## License

MIT
