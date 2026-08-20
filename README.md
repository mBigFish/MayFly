# Mayfly WebShell

一个基于 **Go + Web** 的开源 WebShell 管理工具（网页版蚁剑/菜刀/冰蝎类工具），用于授权渗透测试与自有资产的管理。

> ⚠️ **合规声明**：本工具仅用于**授权的安全测试、渗透测试**以及**管理自有资产**。请勿用于任何未授权访问，使用者需自行承担相应法律责任。

## 核心架构

```
┌─ 管理端（Web 界面）──────┐          HTTP 请求          ┌─ 目标服务器 ────────┐
│  节点管理 / 命令执行      │  ───────────────────────►   │  一句话 WebShell 脚本 │
│  文件管理 / 数据库 / 终端  │  （base64 + JSON 协议）      │  shell.php / jsp /   │
└─────────────────────────┘  ◄───────────────────────   │  aspx / asp          │
                                                     └─────────────────────┘
```

- **管理端**：单一 Go 二进制 + 网页前端，浏览器访问，跨平台、易部署
- **服务端脚本**：内置 4 种语言的功能型 WebShell（PHP / JSP / ASPX / ASP），通过「脚本生成器」一键生成

## 功能特性

- **节点管理** - 多目标节点增删改查，按连接密码 / 语言类型区分
- **命令执行** - 在目标服务器执行系统命令并回显结果
- **文件管理** - 目录浏览、文件读取/编辑、上传、下载、删除、重命名、新建目录
- **数据库管理** - 连接目标 MySQL 并执行 SQL（PHP 端，结果表格化展示）
- **虚拟终端** - 基于命令执行的伪交互终端（xterm.js，维护工作目录）
- **脚本生成器** - 一键生成/复制/下载 4 种语言的 WebShell 脚本，支持自定义连接密码
- **JWT 认证** - 登录鉴权，节点数据本地持久化（JSON）

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go + Gin + gorilla/websocket |
| 前端 | 原生 JS + xterm.js + FontAwesome |
| 认证 | JWT (golang-jwt) |
| 存储 | 本地 JSON 文件 |

## 快速开始

### 1. 编译

```bash
go mod tidy
go build -o mayfly .
```

Windows 也可直接双击 `start.bat` 启动。

### 2. 运行

```bash
# 默认配置（端口 8080，账号 admin/mayfly123）
./mayfly

# 自定义配置
MAYFLY_PORT=9090 MAYFLY_USER=myuser MAYFLY_PASS=mypassword ./mayfly
```

### 3. 使用流程

1. 打开 `http://localhost:8080`，使用 `admin / mayfly123` 登录
2. 点击侧边栏「脚本生成器」，选择语言生成 WebShell 脚本
3. 将生成的脚本部署到目标 Web 服务器（Web 可访问目录）
4. 回到节点列表，点击「+」添加节点，填写脚本 URL、语言类型、连接密码
5. 选中节点后即可使用「命令执行 / 文件管理 / 数据库 / 虚拟终端」

## 服务端脚本支持

| 语言 | 文件 | 命令执行 | 文件管理 | 数据库 | 说明 |
|------|------|:---:|:---:|:---:|------|
| PHP  | `shell.php`  | ✅ | ✅ | ✅ MySQL | 最通用，PDO/mysqli |
| JSP  | `shell.jsp`  | ✅ | ✅ | — | 依赖 JDK 内置 Nashorn（JDK 8~14） |
| ASPX | `shell.aspx` | ✅ | ✅ | — | .NET 内置序列化，IIS |
| ASP  | `shell.asp`  | ✅ | ✅ | — | VBScript + WScript.Shell，明文协议 |

## 通信协议

管理端与 WebShell 脚本之间的通信协议（非 ASP）：

- **请求**：`POST` 到脚本 URL，表单字段 `<连接密码> = base64(json)`
  - `json = {"action": "cmd|fileList|...", "params": {...}}`
- **响应**：HTTP body = `base64(json)`，`json = {"status":"ok|error","data":"base64(结果)","message":""}`

ASP 使用明文表单协议（VBScript 无 base64/JSON 能力）。

## 配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `MAYFLY_PORT` | 8080 | 监听端口 |
| `MAYFLY_USER` | admin | 登录用户名 |
| `MAYFLY_PASS` | mayfly123 | 登录密码 |
| `MAYFLY_JWT_SECRET` | mayfly-secret-key-change-in-production | JWT 签名密钥 |

## 项目结构

```
Mayfly/
├── main.go                      # 程序入口，路由注册
├── go.mod
├── config/
│   └── config.go                # 配置管理
├── payloads/                    # WebShell 服务端脚本模板
│   ├── shell.php
│   ├── shell.jsp
│   ├── shell.aspx
│   └── shell.asp
├── internal/
│   ├── handler/
│   │   ├── auth.go              # 登录/JWT
│   │   ├── node.go              # 节点 CRUD + 核心操作 API
│   │   ├── terminal.go          # 本地 WebSSH 终端（保留）
│   │   └── terminal_ws.go       # 虚拟终端 WebSocket
│   ├── middleware/
│   │   └── auth.go              # 认证中间件
│   ├── model/
│   │   ├── node.go              # 节点模型
│   │   └── session.go
│   ├── store/
│   │   └── store.go             # 节点 JSON 持久化
│   └── service/
│       ├── shell.go             # WebShell 客户端 + 协议编解码
│       └── ...                  # session/terminal 服务
└── web/
    ├── index.html               # 管理端主界面
    ├── login.html               # 登录页
    └── static/
        ├── css/style.css
        └── js/{app.js, login.js}
```

## 安全注意事项

1. **修改默认密码** - 务必通过环境变量修改登录账号与密码
2. **使用 HTTPS** - 管理端建议配合 Nginx 使用 HTTPS
3. **修改脚本连接密码** - 生成 WebShell 脚本时设置高强度连接密码，而非默认 `mayfly`
4. **最小权限** - 管理端无需以 root 运行
5. **合规使用** - 仅对已授权目标使用本工具

## 后续规划

- [ ] 数据库管理支持更多语言/类型（JSP JDBC、ASPX SQL Server）
- [ ] 真 PTY 虚拟终端（持久进程 + 管道）
- [ ] 通信加密（自定义编解码器，规避明文特征）
- [ ] 批量命令/文件分片下载（大文件）
- [ ] 系统信息面板（CPU/内存/进程列表）
- [ ] 节点分组与搜索

## License

MIT