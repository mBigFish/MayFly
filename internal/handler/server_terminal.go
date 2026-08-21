package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"mayfly/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// ServerTerminalHandler 服务器 SSH 交互式终端
type ServerTerminalHandler struct{}

// NewServerTerminalHandler 创建服务器终端 handler
func NewServerTerminalHandler() *ServerTerminalHandler {
	return &ServerTerminalHandler{}
}

// getServerByID 从存储中获取服务器配置
func getServerByID(id int) (model.Server, bool) {
	ensureServerStore()
	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()
	for _, s := range serverStore.Servers {
		if s.ID == id {
			return s, true
		}
	}
	return model.Server{}, false
}

// sshConfigFromServer 根据服务器配置构造 SSH 客户端配置
func sshConfigFromServer(s model.Server) (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod
	if s.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(s.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("私钥解析失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if s.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.Password))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("必须提供密码或私钥")
	}
	return &ssh.ClientConfig{
		User:            s.Username,
		Auth:            authMethods,
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

// Handle 处理服务器 SSH 终端 WebSocket 连接
func (h *ServerTerminalHandler) Handle(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器 ID"})
		return
	}
	s, ok := getServerByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
		return
	}

	cfg, err := sshConfigFromServer(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	port := s.Port
	if port == 0 {
		port = 22
	}
	address := fmt.Sprintf("%s:%d", s.Host, port)

	conn, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 写锁（goroutine 并发写 WebSocket 需要互斥）
	var writeMu sync.Mutex
	safeWrite := func(msg gin.H) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.WriteJSON(msg)
	}

	// WebSocket 保活：设置读超时 + 定期发 ping，防止空闲被中间层断开
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	pingCtx, pingCancel := context.WithCancel(context.Background())
	defer pingCancel()
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticker.C:
				writeMu.Lock()
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

	// 连接 SSH
	client, err := ssh.Dial("tcp", address, cfg)
	if err != nil {
		safeWrite(gin.H{"type": "error", "data": "SSH 连接失败: " + err.Error()})
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		safeWrite(gin.H{"type": "error", "data": "创建会话失败: " + err.Error()})
		return
	}
	defer session.Close()

	// 请求 PTY
	if err := session.RequestPty("xterm-256color", 24, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		safeWrite(gin.H{"type": "error", "data": "请求 PTY 失败: " + err.Error()})
		return
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	if err := session.Shell(); err != nil {
		safeWrite(gin.H{"type": "error", "data": "启动 Shell 失败: " + err.Error()})
		return
	}

	safeWrite(gin.H{"type": "ready"})

	// stdout/stderr → WebSocket
	outDone := make(chan struct{})
	go func() {
		defer close(outDone)
		buf := make([]byte, 8192)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				safeWrite(gin.H{"type": "output", "data": string(buf[:n])})
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				safeWrite(gin.H{"type": "output", "data": string(buf[:n])})
			}
			if err != nil {
				return
			}
		}
	}()

	// 主循环：读 WebSocket → stdin
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var m struct {
			Type string `json:"type"`
			Data string `json:"data"`
			Cols int    `json:"cols"`
			Rows int    `json:"rows"`
		}
		if err := json.Unmarshal(msg, &m); err != nil {
			continue
		}
		switch m.Type {
		case "input":
			stdin.Write([]byte(m.Data))
		case "resize":
			if m.Cols > 0 && m.Rows > 0 {
				session.WindowChange(m.Rows, m.Cols)
			}
		}
	}
	stdin.Close()
	<-outDone
	safeWrite(gin.H{"type": "exit"})
}
