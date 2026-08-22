package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/webshell-manager/webshell-manager/internal/auth"
	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/protocol"
	"github.com/webshell-manager/webshell-manager/internal/session"
)

// upgrader 将 HTTP 升级为 WebSocket。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域（后续可按需收紧为同源校验）。
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TerminalHandler 终端 WebSocket 处理器。
type TerminalHandler struct {
	manager  *session.Manager
	registry *protocol.Registry
	authSvc  *auth.Service
}

// NewTerminalHandler 创建终端处理器。
func NewTerminalHandler(manager *session.Manager, registry *protocol.Registry, authSvc *auth.Service) *TerminalHandler {
	return &TerminalHandler{manager: manager, registry: registry, authSvc: authSvc}
}

// ServeWS 处理 /ws/v1/session/:id 的 WebSocket 连接。
// 终端与 Session 绑定，通过 Session 获取目标，执行命令。
// token 通过 query 参数传递（浏览器 WebSocket 无法自定义 header）。
func (h *TerminalHandler) ServeWS(c *gin.Context) {
	// JWT 鉴权。
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
		return
	}
	if _, err := h.authSvc.ParseToken(token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
		return
	}

	id := c.Param("id")
	s, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 设置读写超时。
	conn.SetReadLimit(64 * 1024)

	// 读取命令循环。
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType != websocket.TextMessage {
			continue
		}

		cmd := string(data)

		// 通过协议执行命令。
		p, err := h.registry.Get(s.Target().Protocol)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("错误: "+err.Error()))
			continue
		}

		res, err := p.Execute(c.Request.Context(), s.Target(), &operation.Operation{
			Type:   operation.OperationCommand,
			Params: map[string]any{"cmd": cmd},
		})
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("错误: "+err.Error()))
			continue
		}

		output := res.Output
		if res.Error != "" {
			output = res.Error
		}
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
			break
		}

		// 刷新会话活跃时间。
		h.manager.Touch(id)
	}
}
