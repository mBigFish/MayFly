package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"mayfly/internal/config"
	"mayfly/internal/database"
	"mayfly/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TerminalWS WebSocket 终端端点
// GET /ws/terminal?type=local          本地终端
// GET /ws/terminal?type=ssh&server_id=1  SSH 终端
func TerminalWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	terminalType := c.Query("type")
	if terminalType == "" {
		terminalType = "local"
	}

	var session *service.TerminalSession

	if terminalType == "ssh" {
		// SSH 终端
		serverIDStr := c.Query("server_id")
		if serverIDStr == "" {
			_ = ws.WriteMessage(websocket.TextMessage, []byte("\x1b[31m错误: 缺少 server_id 参数\x1b[0m"))
			return
		}
		serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
		if err != nil {
			_ = ws.WriteMessage(websocket.TextMessage, []byte("\x1b[31m错误: server_id 参数无效\x1b[0m"))
			return
		}

		cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
		rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))

		session, err = service.StartSSHTerminal(database.Get(), ws, uint(serverID), cols, rows)
		if err != nil {
			_ = ws.WriteMessage(websocket.TextMessage, []byte("\x1b[31mSSH 连接失败: "+err.Error()+"\x1b[0m"))
			return
		}
	} else {
		// 本地终端
		cfg := config.Get()
		shell := cfg.Terminal.Shell
		session, err = service.StartLocalTerminal(ws, shell)
		if err != nil {
			_ = ws.WriteMessage(websocket.TextMessage, []byte("\x1b[31m"+err.Error()+"\x1b[0m"))
			return
		}
	}

	defer session.Close()

	// 读取 WebSocket 消息并写入终端
	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			break
		}

		// 支持窗口大小调整消息（JSON 格式）
		if msgType == websocket.TextMessage && len(data) > 0 && data[0] == '{' {
			// 尝试解析 resize 消息
			var resizeMsg struct {
				Type string `json:"type"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			if json.Unmarshal(data, &resizeMsg) == nil && resizeMsg.Type == "resize" {
				_ = session.ResizeWindow(resizeMsg.Cols, resizeMsg.Rows)
				continue
			}
		}

		if err := session.WriteInput(data); err != nil {
			break
		}
	}
}

// ListenerWS WebSocket 监听器终端端点
// GET /ws/listener?id=xxx
func ListenerWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			break
		}
		_ = ws.WriteMessage(websocket.TextMessage, data)
	}
}

// 保持引用
var _ = time.Now
