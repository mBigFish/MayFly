package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"mayfly/config"
	"mayfly/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket 消息类型
const (
	MsgInput  = "input"
	MsgResize = "resize"
	MsgOutput = "output"
	MsgClosed = "closed"
	MsgError  = "error"
)

// WSMessage WebSocket 消息格式
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ResizeData 窗口大小调整数据
type ResizeData struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// TerminalHandler 终端处理器
type TerminalHandler struct {
	sessionService *service.SessionService
	upgrader       websocket.Upgrader
}

// NewTerminalHandler 创建终端处理器
func NewTerminalHandler(ss *service.SessionService) *TerminalHandler {
	return &TerminalHandler{
		sessionService: ss,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境应限制来源
			},
		},
	}
}

// HandleTerminal 处理 WebSocket 终端连接
func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	token := c.Query("token")
	claims, err := ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败"})
		return
	}
	_ = claims

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer conn.Close()

	cfg := config.Get()

	// 创建终端会话
	ptySession, err := service.NewTerminal(cfg.Shell, 80, 24)
	if err != nil {
		h.sendError(conn, "创建终端失败: "+err.Error())
		return
	}

	// 创建会话记录
	session, _ := h.sessionService.CreateSession("")
	h.sessionService.StorePTY(session.ID, ptySession)
	defer func() {
		ptySession.Close()
		h.sessionService.RemoveSession(session.ID)
	}()

	log.Printf("终端会话已创建: %s (用户: %s)", session.ID, claims.Username)

	// goroutine: PTY 输出 -> WebSocket
	var once sync.Once
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptySession.Read(buf)
			if err != nil {
				once.Do(func() {
					h.sendMessage(conn, MsgClosed, "终端已关闭")
				})
				return
			}
			if n > 0 {
				h.sendMessage(conn, MsgOutput, string(buf[:n]))
			}
		}
	}()

	// 主循环: 读取 WebSocket 消息
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case MsgInput:
			var input string
			if err := json.Unmarshal(msg.Data, &input); err == nil {
				ptySession.Write([]byte(input))
			}
		case MsgResize:
			var resize ResizeData
			if err := json.Unmarshal(msg.Data, &resize); err == nil {
				ptySession.Resize(resize.Cols, resize.Rows)
			}
		}
	}
}

// ListSessions 列出所有会话
func (h *TerminalHandler) ListSessions(c *gin.Context) {
	sessions := h.sessionService.ListSessions()
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// sendMessage 发送 WebSocket 消息
func (h *TerminalHandler) sendMessage(conn *websocket.Conn, msgType string, data string) {
	// 用 json.Marshal 正确转义字符串，避免终端输出中的控制字符（ESC、引号、反斜杠等）
	// 破坏 JSON 结构导致前端解析失败
	dataJSON, _ := json.Marshal(data)
	msg := WSMessage{
		Type: msgType,
		Data: json.RawMessage(dataJSON),
	}
	conn.WriteJSON(msg)
}

// sendError 发送错误消息
func (h *TerminalHandler) sendError(conn *websocket.Conn, errMsg string) {
	h.sendMessage(conn, MsgError, errMsg)
}
