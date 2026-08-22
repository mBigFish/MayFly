# WebShell Manager

面向授权安全测试、CTF/靶场及内部安全验证环境的 WebShell 管理与分析平台。

> 本项目仅用于授权安全测试、CTF、靶场和内部安全验证环境。测试目标必须使用用户拥有或明确授权的环境。

## 技术栈

- 后端：Go 1.24+（Gin、GORM、SQLite、Zap、JWT、bcrypt）
- 前端：Vue 3 + TypeScript + Vite + Element Plus + Pinia + Vue Router + Axios

## 当前进度

已完成：
- **Phase 1**：项目初始化、SQLite、Config、Logger
- **Phase 2**：User、Login、RBAC、Target CRUD
- **Phase 3**：Transport、Protocol Interface、Request/Response、Operation
- **Phase 4**：Operation 完整实现、Session、Terminal（WebSocket + xterm.js）、File Manager

尚未实现：Plugin、Request Inspector、Audit、Task、Dashboard 等（Phase 5+）。

## 目录结构

```
webshell-manager/
├── cmd/server/               # 服务端入口
├── internal/
│   ├── api/                  # 路由、处理器、DTO
│   ├── auth/                 # 用户/角色/权限、JWT、登录限流
│   ├── target/               # Target 模型、仓储、服务
│   ├── transport/            # 传输抽象（HTTP 实现，SSRF 防护）
│   ├── protocol/             # 协议抽象 + 注册表 + PHP 适配器
│   ├── operation/            # 统一操作抽象
│   ├── crypto/               # 敏感字段加密（AES-GCM）
│   ├── migrations/           # 具体数据库迁移
│   ├── config/               # 配置加载
│   ├── logger/               # 日志模块（zap）
│   └── database/             # 数据库连接与迁移框架
├── frontend/                 # Vue3 前端
├── configs/config.yaml       # 服务端配置
├── migrations/               # 迁移脚本目录（占位）
├── docs/ scripts/ tests/     # 占位目录
├── Makefile
└── go.mod
```

## 运行

### 后端

```bash
# 首次启动需设置初始 admin 密码（默认 admin123）
export ADMIN_PASSWORD=your-password
go run ./cmd/server
# 或指定配置文件
go run ./cmd/server -config configs/config.yaml
```

默认监听 `0.0.0.0:8080`，默认管理员账号 `admin`。

### 前端

```bash
cd frontend
npm install
npm run dev
```

## API

统一前缀 `/api/v1`，返回格式 `{"code":0,"message":"ok","data":...}`。

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | 无 | 登录，返回 JWT |
| GET | `/api/v1/targets` | target:read | 目标列表 |
| POST | `/api/v1/targets` | target:create | 创建目标 |
| GET | `/api/v1/targets/:id` | target:read | 目标详情 |
| PUT | `/api/v1/targets/:id` | target:update | 更新目标 |
| DELETE | `/api/v1/targets/:id` | target:delete | 删除目标 |
| POST | `/api/v1/targets/:id/check` | target:read | 目标探活（Protocol.Check） |
| GET | `/api/v1/targets/:id/files` | file:read | 列出目录 |
| POST | `/api/v1/targets/:id/files/read` | file:read | 读取文件 |
| POST | `/api/v1/targets/:id/files/write` | file:write | 写入文件 |
| POST | `/api/v1/targets/:id/files/rename` | file:write | 重命名 |
| POST | `/api/v1/targets/:id/files/mkdir` | file:write | 创建目录 |
| POST | `/api/v1/targets/:id/files/delete` | file:delete | 删除 |
| POST | `/api/v1/sessions` | - | 创建会话 |
| GET | `/api/v1/sessions` | - | 会话列表 |
| DELETE | `/api/v1/sessions/:id` | - | 关闭会话 |

WebSocket：`/ws/v1/session/:id?token=<jwt>`（终端）。

鉴权方式：请求头 `Authorization: Bearer <token>`。

## 测试

```bash
go test ./...
```

## 构建

```bash
make build          # 编译后端
make frontend-build # 构建前端
```

## 安全说明

- 密码使用 bcrypt 哈希存储。
- Target 的 Cookies、Headers 等敏感字段使用 AES-GCM 加密后落库。
- 登录接口带内存限流（默认每分钟 5 次）。
- Target URL 仅允许 http/https 协议，防止 SSRF。
- 生产环境务必修改 `configs/config.yaml` 中的 `jwt_secret` 与 `encryption_key`。
