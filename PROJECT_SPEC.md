# WebShell Manager 项目技术设计与开发规范

**项目名称：** MayFly（WebShell Manager）
 **项目定位：** 面向授权安全测试、CTF/靶场及内部安全验证环境的 WebShell 管理与分析平台
 **后端：** Go
 **前端：** Vue 3 + TypeScript
 **数据库：** SQLite，预留 PostgreSQL
 **部署环境：** Linux Server / Ubuntu Server / windows / Mac
 **运行模式：** Server + Browser
 **主要目标：** 在无图形化 Linux Server 上运行，通过浏览器提供类似传统 WebShell 客户端的统一管理能力。

------

# 1. 项目背景

传统 WebShell 客户端通常依赖桌面 GUI。

本项目希望解决：

```
无 GUI Linux Server
        ↓
运行 WebShell Manager
        ↓
浏览器访问
        ↓
管理授权测试目标
```

项目不是简单复制传统 WebShell 客户端，而是构建一个现代化、模块化、可扩展的 WebShell 管理框架。

核心特点：

- Web 化
- Server 部署
- 多目标管理
- 多协议支持
- 插件化
- 文件管理
- 命令交互
- 请求调试
- 审计日志
- 任务管理
- RBAC
- API 化

------

# 2. 产品目标

第一阶段目标：

```
Target 管理
      ↓
HTTP Transport
      ↓
Protocol Adapter
      ↓
授权测试 WebShell
      ↓
Command / File
```

最终目标：

```
                    WebShell Manager
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
    Targets             Sessions            Tasks
       │                   │                   │
       ▼                   ▼                   ▼
   Connection           Terminal            Worker
       │
       ▼
 Protocol Engine
       │
 ┌─────┼─────┐
 │     │     │
PHP   JSP   ASPX
       │
       ▼
 Plugin Engine
       │
 ┌─────┼─────────────┐
 │     │       │     │
Files Info Network Logs
```

------

# 4. 技术栈

## Backend

```
Go 1.24+
```

推荐：

```
Gin / Echo
gorilla/websocket
GORM
SQLite
Zap / Zerolog
```

如果没有特殊原因，HTTP API 可以使用 Gin。

------

# 5. Frontend

```
Vue 3
TypeScript
Vite
Element Plus
Pinia
Vue Router
Axios
xterm.js
Monaco Editor
```

前端负责：

```
UI
Target
Terminal
File Manager
Request Inspector
Task Center
Dashboard
Settings
```

------

# 6. 项目目录

AI 必须按照以下结构组织项目：

```
webshell-manager/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   │
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware/
│   │   ├── handlers/
│   │   └── dto/
│   │
│   ├── auth/
│   │   ├── service.go
│   │   ├── middleware.go
│   │   └── model.go
│   │
│   ├── target/
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   ├── session/
│   │   ├── session.go
│   │   └── manager.go
│   │
│   ├── transport/
│   │   ├── transport.go
│   │   ├── http.go
│   │   └── websocket.go
│   │
│   ├── protocol/
│   │   ├── protocol.go
│   │   ├── registry.go
│   │   └── adapters/
│   │
│   ├── operation/
│   │   ├── command.go
│   │   ├── file.go
│   │   └── info.go
│   │
│   ├── plugin/
│   │   ├── plugin.go
│   │   ├── registry.go
│   │   └── manager.go
│   │
│   ├── task/
│   │   ├── task.go
│   │   ├── queue.go
│   │   └── worker.go
│   │
│   ├── audit/
│   │   ├── model.go
│   │   └── service.go
│   │
│   ├── database/
│   │   ├── database.go
│   │   └── migration.go
│   │
│   └── config/
│       └── config.go
│
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   ├── layouts/
│   │   ├── views/
│   │   ├── stores/
│   │   ├── router/
│   │   └── types/
│   │
│   └── package.json
│
├── configs/
│   └── config.yaml
│
├── migrations/
│
├── docs/
│
├── scripts/
│
├── tests/
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

------

# 7. 核心架构原则

最重要的原则：

> UI 不允许直接操作 HTTP 或目标。

必须经过：

```
UI
 ↓
API
 ↓
Service
 ↓
Operation
 ↓
Protocol
 ↓
Transport
 ↓
Target
```

例如：

```
FileManager.vue
       ↓
file API
       ↓
FileService
       ↓
Protocol.ReadFile()
       ↓
Transport.Request()
       ↓
Authorized Target
```

------

# 8. Target 模型

Target 是整个系统的核心实体。

建议：

```
type Target struct {
    ID          uint
    Name        string
    URL         string
    Type        string
    Protocol    string
    Method      string
    Headers     string
    Cookies     string
    Timeout     int
    Proxy       string
    Encoding    string
    Remark      string
    GroupID     uint
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

注意：

敏感字段必须加密保存，例如：

```
Cookie
Authorization
Credential
Secret
```

------

# 9. Transport 抽象

定义统一接口：

```
type Transport interface {
    Request(ctx context.Context, req *Request) (*Response, error)
}
```

Request：

```
type Request struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    []byte
}
```

Response：

```
type Response struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
    Duration   time.Duration
}
```

第一阶段实现：

```
HTTP Transport
```

以后可以扩展：

```
Proxy Transport
Custom Transport
```

------

# 10. Protocol 抽象

协议层必须与 HTTP Transport 解耦。

接口：

```
type Protocol interface {


    Name() string


    Check(
        ctx context.Context,
        target *Target,
    ) error


    Execute(
        ctx context.Context,
        target *Target,
        operation *Operation,
    ) (*Result, error)
}
```

不要让：

```
PHP Protocol
```

直接操作：

```
http.Client
```

必须通过：

```
Transport
```

进行通信。

------

# 11. Operation

操作统一抽象：

```
type OperationType string


const (
    OperationCommand OperationType = "command"
    OperationReadFile OperationType = "read_file"
    OperationListDir OperationType = "list_dir"
    OperationWriteFile OperationType = "write_file"
    OperationSystemInfo OperationType = "system_info"
)
```

统一：

```
type Operation struct {
    Type   OperationType
    Params map[string]any
}
```

这样：

```
Terminal
FileManager
SystemInfo
```

都可以使用统一 Operation Engine。

------

# 12. 文件管理

第一阶段支持：

```
List
Read
Write
Upload
Download
Rename
Create Directory
Delete
Search
```

前端：

```
FileTree
    +
FileTable
    +
Monaco Editor
```

文件编辑必须：

```
读取
 ↓
编辑
 ↓
Diff
 ↓
用户确认
 ↓
保存
```

避免 UI 无确认直接覆盖重要文件。

------

# 13. Terminal

Terminal 前端使用：

```
xterm.js
```

功能：

```
命令输入
历史
复制
粘贴
清屏
字体调整
主题
输出搜索
```

Terminal 必须与后端 Session 绑定。

------

# 14. Session Manager

```
type Session struct {
    ID        string
    TargetID  uint
    UserID    uint
    CreatedAt time.Time
    LastSeen  time.Time
}
```

Session Manager：

```
Create()
Get()
Close()
List()
Touch()
```

必须限制：

```
单用户 Session 数量
Session 超时时间
空闲 Session 自动释放
```

------

# 15. Plugin System

插件是项目的重要扩展机制。

统一接口：

```
type Plugin interface {


    ID() string


    Name() string


    Version() string


    Description() string


    Execute(
        ctx context.Context,
        input *PluginInput,
    ) (*PluginResult, error)
}
```

插件示例：

```
FileManager
SystemInfo
ProcessViewer
NetworkInfo
LogViewer
DockerInfo
```

插件不能绕过：

```
Auth
Permission
Audit
Protocol
```

------

# 16. Request Inspector

每一次请求可以生成：

```
Request
Response
Duration
Status
Size
```

前端提供：

```
Raw
Headers
Body
Timing
```

用于授权测试环境中的协议调试。

------

# 17. Audit

所有重要操作必须记录：

```
type AuditLog struct {
    ID        uint
    UserID    uint
    TargetID  uint
    Action    string
    Status    string
    IP        string
    Duration  int64
    CreatedAt time.Time
}
```

例如：

```
LOGIN
TARGET_CONNECT
COMMAND
FILE_READ
FILE_WRITE
FILE_UPLOAD
FILE_DOWNLOAD
FILE_DELETE
PLUGIN_EXECUTE
```

敏感数据不要直接写入日志。

------

# 18. Task Center

任务：

```
Pending
Running
Success
Failed
Cancelled
```

任务模型：

```
type Task struct {
    ID        string
    Type      string
    Status    string
    TargetID  uint
    Progress  int
    Result    string
    Error     string
}
```

以后可以扩展：

```
Worker Pool
Redis
Celery 类似的任务架构
```

但第一版不要引入 Redis。

------

# 19. API

统一：

```
/api/v1/
```

例如：

```
POST   /api/v1/auth/login


GET    /api/v1/targets
POST   /api/v1/targets
PUT    /api/v1/targets/:id
DELETE /api/v1/targets/:id


POST   /api/v1/targets/:id/check


GET    /api/v1/targets/:id/files


POST   /api/v1/targets/:id/files/read


POST   /api/v1/targets/:id/files/write


POST   /api/v1/targets/:id/operations


GET    /api/v1/tasks


GET    /api/v1/audit
```

WebSocket：

```
/ws/v1/session/:id
```

------

# 20. 前端页面

第一阶段：

```
Login
Dashboard
Targets
Target Detail
Terminal
File Manager
Request Inspector
```

第二阶段：

```
Plugins
Tasks
Audit
Users
Settings
```

------

# 21. Dashboard

首页展示：

```
Targets
Online Targets
Sessions
Tasks
Recent Operations
```

不要在 Dashboard 中执行任何目标操作。

------

# 22. 权限模型

至少：

```
admin
operator
auditor
```

权限：

```
target:read
target:create
target:update
target:delete


terminal:execute


file:read
file:write
file:delete


plugin:execute


audit:read


user:manage
```

使用 RBAC。

------

# 23. 安全要求

必须实现：

### Authentication

```
Session/JWT
Password Hash
Login Rate Limit
Session Expiration
```

### Authorization

每一个 API 都检查：

```
User
 ↓
Role
 ↓
Permission
 ↓
Target
```

### CSRF

对于 Cookie Session API 开启 CSRF 防护。

### SSRF

Target URL 属于高风险输入。

必须考虑：

```
URL 校验
协议限制
Redirect 控制
内网访问策略
```

### 文件路径

所有文件路径操作必须进行规范化和权限检查。

------

# 24. 数据库

第一版：

```
SQLite
```

表：

```
users
roles
permissions
user_roles


targets
target_groups


sessions


tasks


plugins


audit_logs
```

不要把密码明文保存。

------

# 25. 配置

```
server:
  host: "0.0.0.0"
  port: 8080


database:
  driver: sqlite
  path: "./data/app.db"


security:
  session_timeout: 3600


logging:
  level: info
  path: "./logs/app.log"
```

------

# 26. Docker

必须支持：

```
docker compose up -d
```

最终：

```
Browser
   ↓
Nginx
   ↓
Go Backend
   ↓
SQLite
```

------

# 27. 单二进制部署

最终目标：

```
webshell-manager
```

包含：

```
Go Backend
Vue Static
Configuration
```

执行：

```
./webshell-manager
```

即可启动。

------

# 28. 开发阶段

AI 必须严格按照以下阶段开发。

## Phase 1

```
Go 项目初始化
Vue3 初始化
SQLite
Config
Logger
```

要求：

**此阶段不要实现任何目标操作。**

------

## Phase 2

实现：

```
User
Login
RBAC
Target CRUD
```

------

## Phase 3

实现：

```
Transport
Protocol Interface
Request / Response
```

使用**本地授权测试服务**验证。

------

## Phase 4

实现：

```
Operation
Session
Terminal UI
File Manager
```

------

## Phase 5

实现：

```
Plugin System
Request Inspector
Audit
```

------

## Phase 6

实现：

```
Task Center
Worker Pool
Dashboard
```

------

## Phase 7

实现：

```
Docker
Build
Single Binary
Deployment
Documentation
```

------

# 29. AI Coding Agent 开发规则

把下面这段作为 AI 的最高优先级开发规则：

```
你是一名资深 Go 后端工程师、Vue3 前端工程师和安全工具架构师。


请严格按照 PROJECT_SPEC.md 开发 WebShell Manager。


开发原则：


1. 不随意改变项目整体架构。
2. 不跨越开发阶段。
3. 每次只实现当前任务。
4. 修改代码前先分析现有代码。
5. 不重复实现已经存在的功能。
6. 优先复用已有 Service、Repository、Interface。
7. Controller 不允许直接访问数据库。
8. UI 不允许直接访问 Transport。
9. Protocol 不允许直接依赖具体 HTTP Client。
10. 所有重要操作必须经过权限检查。
11. 所有重要操作必须产生 Audit Log。
12. 所有外部输入必须进行校验。
13. 所有敏感信息不得明文记录到日志。
14. 不使用全局可变状态保存用户 Session。
15. 所有 goroutine 必须有生命周期管理。
16. 所有资源必须正确关闭。
17. 所有网络请求必须设置 timeout。
18. 所有数据库操作必须支持 context。
19. 所有 API 返回统一 JSON 格式。
20. 所有错误必须进行明确处理。


安全边界：


本项目只用于授权安全测试、CTF、靶场和内部安全验证环境。



测试目标必须使用用户拥有或明确授权的环境。


每完成一个功能：


1. 编译项目
2. 运行单元测试
3. 运行 lint
4. 检查错误处理
5. 检查权限控制
6. 检查日志
7. 更新相关文档


不要一次生成整个项目。


必须按照 Phase 顺序逐步开发。
```

