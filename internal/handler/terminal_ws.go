package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mayfly/internal/model"
	"mayfly/internal/service"
	"mayfly/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// NodeParam 中间件：解析 :id 参数并加载节点到上下文
func NodeParam(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		node, ok := s.Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
			c.Abort()
			return
		}
		c.Set("node", node)
		c.Next()
	}
}

// TerminalWSHandler 虚拟终端（基于命令执行的伪终端）
type TerminalWSHandler struct {
	store *store.Store
}

// NewTerminalWSHandler 创建终端 handler
func NewTerminalWSHandler(s *store.Store) *TerminalWSHandler {
	return &TerminalWSHandler{store: s}
}

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle 处理虚拟终端 WebSocket 连接
func (h *TerminalWSHandler) Handle(c *gin.Context) {
	node, ok := c.Get("node")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	cl := service.NewClient(node.(*model.Node))
	windows := strings.EqualFold(node.(*model.Node).Type, "aspx") || strings.EqualFold(node.(*model.Node).Type, "asp")

	conn, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	cwd := ""
	if info, err := cl.SysInfo(); err == nil {
		for _, line := range strings.Split(info, "\n") {
			if strings.HasPrefix(line, "CWD:") {
				cwd = strings.TrimSpace(strings.TrimPrefix(line, "CWD:"))
			}
		}
	}

	send := func(msgType, data string) error {
		return conn.WriteJSON(gin.H{"type": msgType, "data": data})
	}

	prompt := func() string {
		if windows {
			return cwd + "> "
		}
		return "\x1b[32m" + cwd + "\x1b[0m$ "
	}

	_ = send("output", "\x1b[36mMayfly 虚拟终端（命令模式，每次输入执行一条命令）\x1b[0m\r\n")
	_ = send("output", prompt())

	for {
		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}
		switch msg.Type {
		case "input":
			var line string
			_ = json.Unmarshal(msg.Data, &line)
			line = strings.TrimSpace(strings.TrimRight(line, "\r\n"))
			if line == "" {
				_ = send("output", prompt())
				continue
			}

			// cd 命令：通过 shell 解析真实路径并更新 cwd
			if line == "cd" || strings.HasPrefix(line, "cd ") || line == "cd .." || strings.HasPrefix(line, "cd\\") || strings.HasPrefix(line, "cd/") {
				target := strings.TrimSpace(strings.TrimPrefix(line, "cd"))
				var pwdCmd string
				if windows {
					pwdCmd = "cd"
					if cwd != "" {
						pwdCmd = `cd /d "` + cwd + `" 2>nul && cd /d "` + target + `" && cd`
					} else {
						pwdCmd = `cd /d "` + target + `" && cd`
					}
				} else {
					if cwd != "" {
						pwdCmd = `cd "` + cwd + `" && cd "` + target + `" 2>/dev/null && pwd`
					} else {
						pwdCmd = `cd "` + target + `" 2>/dev/null && pwd`
					}
				}
				out, e := cl.Exec(pwdCmd)
				if e != nil || strings.TrimSpace(out) == "" {
					_ = send("output", "cd: 目录不存在\r\n")
				} else {
					cwd = strings.TrimSpace(strings.Split(out, "\n")[0])
				}
				_ = send("output", prompt())
				continue
			}

			// 普通命令：在维护的 cwd 下执行
			var fullCmd string
			if windows {
				if cwd != "" {
					fullCmd = `cd /d "` + cwd + `" && ` + line
				} else {
					fullCmd = line
				}
			} else {
				if cwd != "" {
					fullCmd = `cd "` + cwd + `" && ` + line
				} else {
					fullCmd = line
				}
			}
			out, e := cl.Exec(fullCmd)
			if e != nil {
				_ = send("output", "\x1b[31m"+e.Error()+"\x1b[0m\r\n")
			} else {
				if out != "" {
					_ = send("output", out)
					if !strings.HasSuffix(out, "\n") {
						_ = send("output", "\r\n")
					}
				}
			}
			_ = send("output", prompt())

		case "close":
			return
		}
	}
}