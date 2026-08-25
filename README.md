# Mayfly WebShell

<p align="center">
  <strong>一款基于 Go + Web 的开源 WebShell 管理工具</strong><br>
  <sub>网页版蚁剑 / 菜刀 / 冰蝎类工具，用于授权渗透测试与自有资产管控</sub>
</p>

<p align="center">
  <img alt="Go Version" src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-blue.svg">
  <img alt="Platform" src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey">
  <img alt="Docker" src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker">
</p>

---

> ⚠️ **合规声明**：本工具仅用于**授权的安全测试、渗透测试**以及**管理自有资产**。请勿用于任何未授权访问，使用者需自行承担相应法律责任。

## 项目背景

在渗透测试与资产管理中，传统 WebShell 管理工具（如中国菜刀、蚁剑、冰蝎）均为桌面图形化客户端，必须运行在带有 GUI 环境的本地主机上。这带来一个实际问题：

1. **本地直连容易溯源** — 从本地机器直接连接目标 WebShell，流量易被追踪定位到真实身份与位置，安全性和隐蔽性不足。
2. **服务器无图形化** — 为规避溯源，通常会使用 VPS / 跳板服务器作为中转，但这些服务器大多是无图形界面的 Linux 环境，无法运行桌面版工具。
3. **工具兼容性差** — 菜刀 / 蚁剑 / 冰蝎等工具依赖 .NET Framework、Java 或 Electron 等运行时，在无 GUI 的服务器上难以部署。

**Mayfly 由此而生** — 一款面向服务器端部署的 WebShell 管理工具，核心设计目标：

- **无图形化依赖** — 单一 Go 二进制 + Web 前端，部署在 VPS 上通过浏览器远程访问，无需桌面环境
- **跨平台运行** — 支持 Linux / macOS / Windows，编译即用，零外部依赖
- **流量中转** — 管理端部署在服务器上，由服务器与目标通信，避免本地直连暴露
- **功能完备** — 节点管理、文件操作、命令执行、虚拟终端、数据库、SSH 服务器管理、反弹 Shell 等一站覆盖
- **容器化部署** — 内置 Docker 支持，一条命令即可在服务器上拉起服务

```
┌─────────────┐     浏览器访问      ┌──────── VPS / 服务器 ────────┐    HTTP 请求    ┌── 目标 ──┐
│  测试者本地   │ ──────────────► │  Mayfly 管理端（无 GUI）        │ ────────────► │  WebShell │
│  （仅需浏览器）│ ◄────────────── │  节点管理 / 文件 / 终端 / ...   │ ◄──────────── │           │
└─────────────┘     Web 界面       └──────────────────────────────┘                └───────────┘
```

> 管理端运行在 VPS 上，本地通过浏览器访问，所有操作由 VPS 中转执行，避免本地 IP 暴露。

## 目录

- [项目背景](#项目背景)
- [核心架构](#核心架构)
- [功能总览](#功能总览)
- [截图预览](#截图预览)
- [快速开始](#快速开始)
- [Docker 部署](#docker-部署)
- [使用流程](#使用流程)
- [服务端脚本支持](#服务端脚本支持)
- [通信协议](#通信协议)
- [配置项](#配置项)
- [项目结构](#项目结构)
- [技术栈](#技术栈)
- [安全注意事项](#安全注意事项)
- [后续规划](#后续规划)
- [License](#license)

## 核心架构

```
┌──────── 管理端（Go 二进制 + Web 界面）──────────┐                ┌── 目标服务器 ────┐
│                                                 │   HTTP 请求    │                  │
│  仪表盘 / 连接列表 / 节点管理                      │ ────────────► │  WebShell 脚本    │
│  文件管理 / 命令执行 / 虚拟终端 / 数据库           │  (base64+JSON) │  php/jsp/aspx/asp│
│  反弹Shell / SSH服务器管理 / 系统管理              │ ◄──────────── │                  │
│                                                 │                └──────────────────┘
│  本地 WebSSH 终端 / SSH 交互终端                   │
│  JWT 认证 / 审计日志 / 数据导入导出                │
│  AES-256-GCM 敏感字段加密                          │
└─────────────────────────────────────────────────┘
```

- **管理端**：单一 Go 二进制 + 网页前端，浏览器访问，跨平台、零依赖部署
- **服务端脚本**：内置 4 种语言的功能型 WebShell（PHP / JSP / ASPX / ASP），通过「脚本生成器」一键生成

## 功能总览

### 节点管理（WebShell 目标）

| 功能 | 说明 |
|------|------|
| 节点 CRUD | 多目标节点增删改查，支持分组、备注、连接密码 |
| 连接列表 | 分组展示、状态筛选（已连通/失败/未测试）、搜索过滤 |
| 批量测试 | 一键测试全部分组或单个连接，结果持久化 |
| 命令执行 | 在目标服务器执行系统命令并回显，支持常用命令快捷面板 |
| 命令历史 | 按节点持久化保存命令执行记录，支持清空 |
| 文件管理 | 目录浏览、文件读取/编辑/上传/下载/删除/重命名/新建目录 |
| 数据库管理 | 连接目标 MySQL 并执行 SQL（PHP 端），结果表格化展示 |
| 虚拟终端 | 基于 xterm.js 的伪交互终端（WebSocket 实时通信） |

### 资源管理（SSH 服务器）

| 功能 | 说明 |
|------|------|
| 服务器 CRUD | 管理 SSH 服务器，支持密码/私钥认证、分组 |
| 连接测试 | SSH 连接测试并持久化结果（主机名、时间、状态） |
| 批量测试 | 一键测试所有服务器连通性 |
| SSH 终端 | 基于 WebSocket 的交互式 SSH 终端，支持多开、心跳保活、断线重连 |

### 反弹 Shell

| 功能 | 说明 |
|------|------|
| 监听管理 | 启动/停止 TCP 监听，实时查看连接输出 |
| Payload 生成 | 一键生成 Bash / Netcat / Python / PHP / Perl / PowerShell 等反弹命令 |
| 分类筛选 | 按语言类型过滤 payload |

### 仪表盘

| 功能 | 说明 |
|------|------|
| 统计卡片 | 节点/服务器/监听器/分组总数及在线状态 |
| 节点类型分布 | PHP / JSP / ASPX / ASP 占比可视化 |
| 分组概览 | 节点与服务器分组合并展示 |
| 最近活动 | 跨节点最近 20 条命令执行记录 |
| 连接告警 | 连接失败节点自动汇总告警 |

### 系统管理

| 功能 | 说明 |
|------|------|
| 系统信息 | 版本、Go 版本、OS、CPU、内存、GC、Goroutine、运行时长 |
| 运行状态 | 内存分配、系统内存、GC 次数、审计日志条数、数据文件列表 |
| 系统设置 | 运行时修改会话超时、默认 Shell（持久化） |
| 修改密码 | 运行时修改登录密码（审计记录） |
| 数据管理 | 导出全部数据为 JSON / 导入数据文件 |
| 审计日志 | 记录所有关键操作（用户、操作、目标、IP），支持查看/清空 |

### 其他特性

- **多主题切换** — 毛玻璃 / 玻璃 / 黑夜三种主题，本地持久化
- **本地 WebSSH 终端** — 管理端本身可提供本地 Shell 终端（跨平台 PTY）
- **敏感字段加密** — SSH 服务器密码/私钥使用 AES-256-GCM 加密存储
- **JWT 认证** — 登录鉴权，Token 过期自动跳转
- **界面状态持久化** — 记住最后选中的节点、标签页、视图，刷新后自动恢复
- **终端自适应** — 切换标签页/窗口聚焦时自动重绘，避免空白

## 截图预览

> 可在项目根目录查看 `node-panel-new-style.png` 了解界面风格

## 快速开始

### 1. 源码编译

```bash
git clone https://github.com/yourname/Mayfly.git
cd Mayfly
go mod tidy
go build -o mayfly .
```

### 2. 运行

```bash
# 默认配置（端口 8080，账号 admin / mayfly123）
./mayfly

# 自定义配置（环境变量优先级最高）
MAYFLY_PORT=9090 MAYFLY_USER=myuser MAYFLY_PASS=mypassword ./mayfly
```

> Windows 也可直接双击 `start.bat` 启动。

### 3. 访问

打开 `http://localhost:8080`，使用配置的账号密码登录。

## Docker 部署

项目自带 `Dockerfile` 和 `docker-compose.yaml`，支持一键容器化部署：

```bash
# 构建并启动（后台运行）
docker compose up -d --build

# 查看日志
docker compose logs -f

# 停止
docker compose down
```

**配置说明**：

- `data/` 目录通过 volume 挂载持久化
- `config/config.yaml` 以只读方式挂载，可通过修改文件后重启生效
- 支持环境变量覆盖（`MAYFLY_PORT` / `MAYFLY_USER` / `MAYFLY_PASS` / `MAYFLY_JWT_SECRET`）

```yaml
# docker-compose.yaml
services:
  mayfly:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./config/config.yaml:/app/config/config.yaml:ro
    environment:
      - GIN_MODE=release
```

## 使用流程

### WebShell 管理

1. 打开管理端，使用账号密码登录
2. 在**连接列表**中点击「添加节点」，填写脚本 URL、语言类型、连接密码
3. 点击「批量测试」或单个节点测试按钮，验证连通性
4. 连通后选中节点，即可使用：
   - **文件管理** — 浏览/编辑/上传/下载文件
   - **命令执行** — 执行系统命令，支持快捷命令面板
   - **虚拟终端** — 基于 xterm.js 的交互终端
   - **数据库** — 连接目标 MySQL 执行 SQL

### 脚本生成

1. 在侧边栏点击「连接列表」→ 添加节点时选择类型
2. 在节点操作中点击「脚本」获取对应语言的 WebShell 脚本
3. 修改脚本中的连接密码，部署到目标 Web 服务器

### SSH 服务器管理

1. 在**资源管理**中点击「添加服务器」，填写 IP、端口、用户名、密码/私钥
2. 点击「测试」验证 SSH 连通性，或「一键测试」批量验证
3. 点击「终端」打开 SSH 交互终端（支持多开）

### 反弹 Shell

1. 在**反弹 Shell**页面输入监听端口，点击「开启监听」
2. 输入监听 IP 和端口，点击「生成」获取各语言 payload
3. 在目标机器执行 payload，管理端实时接收 Shell 连接

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

配置优先级：**环境变量 > config.yaml > 默认值**

| 环境变量 | config.yaml | 默认值 | 说明 |
|----------|-------------|--------|------|
| `MAYFLY_PORT` | `server_port` | `8080` | 监听端口 |
| `MAYFLY_USER` | `username` | `admin` | 登录用户名 |
| `MAYFLY_PASS` | `password` | `mayfly123` | 登录密码 |
| `MAYFLY_JWT_SECRET` | `jwt_secret` | `mayfly-secret-key-change-in-production` | JWT 签名密钥 |
| `MAYFLY_SHELL` | `shell` | 系统默认 | 默认终端 shell |
| `MAYFLY_SESSION_TIMEOUT` | `session_timeout` | `30` | 会话超时（分钟） |

> ⚠️ 生产环境务必修改 `password` 和 `jwt_secret`。

## 项目结构

```
Mayfly/
├── main.go                          # 程序入口，路由注册
├── go.mod / go.sum
├── config/
│   ├── config.go                    # 配置管理（环境变量 + YAML）
│   └── config.yaml                  # 配置文件
├── payloads/                        # WebShell 服务端脚本模板
│   ├── shell.php
│   ├── shell.jsp
│   ├── shell.aspx
│   └── shell.asp
├── internal/
│   ├── crypto/
│   │   └── crypto.go                # AES-256-GCM 加密/解密
│   ├── handler/
│   │   ├── auth.go                  # 登录 / JWT 签发与验证
│   │   ├── node.go                  # 节点 CRUD + 命令执行 + 文件管理 + 数据库
│   │   ├── dashboard.go             # 仪表盘聚合统计
│   │   ├── listener.go              # 反弹 Shell 监听管理
│   │   ├── server.go                # SSH 服务器资源管理
│   │   ├── server_terminal.go       # SSH 交互终端 WebSocket
│   │   ├── terminal.go              # 本地 WebSSH 终端
│   │   ├── terminal_ws.go           # 虚拟终端 WebSocket
│   │   └── system.go                # 系统信息 / 设置 / 审计日志 / 数据导入导出
│   ├── middleware/
│   │   └── auth.go                  # JWT 认证中间件
│   ├── model/
│   │   ├── node.go                  # 节点模型
│   │   ├── server.go                # SSH 服务器模型
│   │   └── session.go              # 会话模型
│   ├── service/
│   │   ├── shell.go                 # WebShell 客户端 + 协议编解码
│   │   ├── session.go               # 会话服务
│   │   ├── terminal.go              # 终端服务
│   │   ├── terminal_unix.go         # Unix PTY 实现
│   │   ├── terminal_windows.go      # Windows PTY 实现
│   │   └── listener.go              # 反弹 Shell 监听服务
│   └── store/
│       ├── store.go                 # 节点数据持久化
│       └── cmd_history.go           # 命令历史持久化
├── web/
│   ├── index.html                   # 管理端主界面
│   ├── login.html                   # 登录页
│   └── static/
│       ├── css/style.css            # 全局样式 + 多主题
│       └── js/
│           ├── app.js               # 前端主逻辑
│           └── login.js             # 登录页逻辑
├── Dockerfile                       # 多阶段构建
├── docker-compose.yaml              # 容器编排
├── .dockerignore
└── start.bat                        # Windows 快速启动
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go + [Gin](https://github.com/gin-gonic/gin) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| SSH | [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) |
| 认证 | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) |
| 加密 | AES-256-GCM |
| 前端终端 | [xterm.js](https://github.com/xtermjs/xterm.js) |
| 前端 UI | 原生 JS + [FontAwesome](https://fontawesome.com/) |
| 存储 | 本地 JSON 文件（零数据库依赖） |

## 安全注意事项

1. **修改默认密码** — 务必通过环境变量或 `config.yaml` 修改登录账号、密码和 JWT 密钥
2. **使用 HTTPS** — 管理端建议配合 Nginx / Caddy 反向代理使用 HTTPS
3. **修改脚本密码** — 生成 WebShell 脚本时设置高强度连接密码，而非默认 `mayfly`
4. **最小权限** — 管理端无需以 root 运行；Docker 部署使用非特权用户
5. **数据加密** — SSH 凭据使用 AES-256-GCM 加密存储，但仍建议保护好 `data/` 目录
6. **合规使用** — 仅对已授权目标使用本工具

## 后续规划

- [ ] 数据库管理支持更多语言/类型（JSP JDBC、ASPX SQL Server）
- [ ] 真 PTY 虚拟终端（持久进程 + 管道）
- [ ] 通信加密（自定义编解码器，规避明文特征）
- [ ] 批量命令/文件分片下载（大文件）
- [ ] 更多反弹 Shell payload 类型
- [ ] 节点标签与多维度筛选

## License

[MIT](LICENSE)

---

<p align="center">
  <sub>如果本项目对您有帮助，欢迎 Star ⭐ 支持</sub>
</p>
