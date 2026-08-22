package service

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"mayfly/internal/crypto"
	"mayfly/internal/model"
)

// TerminalSession 终端会话
type TerminalSession struct {
	ID        string
	Type      string // local / ssh
	Cmd       *exec.Cmd
	Pty       io.ReadWriteCloser
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	SSHStdin   io.WriteCloser
	SSHStdout  io.Reader
	WS        *websocket.Conn
	mu        sync.Mutex
	closed    bool
}

var (
	terminalSessions   = make(map[string]*TerminalSession)
	terminalSessionsMu sync.RWMutex
)

// StartLocalTerminal 启动本地终端
func StartLocalTerminal(ws *websocket.Conn, shell string) (*TerminalSession, error) {
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "powershell.exe"
		} else {
			shell = "bash"
		}
	}

	cmd := exec.Command(shell)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("启动终端失败: %w", err)
	}

	session := &TerminalSession{
		ID:   fmt.Sprintf("local-%d", time.Now().UnixNano()),
		Type: "local",
		Cmd:  cmd,
		Pty:  ptmx,
		WS:   ws,
	}

	terminalSessionsMu.Lock()
	terminalSessions[session.ID] = session
	terminalSessionsMu.Unlock()

	// 读取 PTY 输出并发送到 WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				break
			}
			session.mu.Lock()
			if session.closed {
				session.mu.Unlock()
				break
			}
			_ = ws.WriteMessage(websocket.TextMessage, buf[:n])
			session.mu.Unlock()
		}
		session.Close()
	}()

	return session, nil
}

// StartSSHTerminal 启动 SSH 交互式终端
func StartSSHTerminal(db *gorm.DB, ws *websocket.Conn, serverID uint, cols, rows int) (*TerminalSession, error) {
	server, err := GetServerByID(db, serverID)
	if err != nil {
		return nil, fmt.Errorf("服务器不存在: %w", err)
	}

	password, _ := crypto.Decrypt(server.Password)
	privateKey, _ := crypto.Decrypt(server.PrivateKey)

	sshConfig := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	if password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(password))
	}
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err == nil {
			sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
		}
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}

	sshSession, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("创建 SSH 会话失败: %w", err)
	}

	// 请求 PTY
	term := "xterm-256color"
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSession.RequestPty(term, rows, cols, modes); err != nil {
		client.Close()
		return nil, fmt.Errorf("请求 PTY 失败: %w", err)
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("获取 stdin 失败: %w", err)
	}
	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("获取 stdout 失败: %w", err)
	}

	// 启动远程 shell
	if err := sshSession.Shell(); err != nil {
		client.Close()
		return nil, fmt.Errorf("启动 shell 失败: %w", err)
	}

	session := &TerminalSession{
		ID:         fmt.Sprintf("ssh-%d", time.Now().UnixNano()),
		Type:       "ssh",
		SSHClient:  client,
		SSHSession: sshSession,
		SSHStdin:   stdin,
		SSHStdout:  stdout,
		WS:         ws,
	}

	terminalSessionsMu.Lock()
	terminalSessions[session.ID] = session
	terminalSessionsMu.Unlock()

	// 读取 SSH stdout 并发送到 WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			session.mu.Lock()
			if session.closed {
				session.mu.Unlock()
				break
			}
			_ = ws.WriteMessage(websocket.TextMessage, buf[:n])
			session.mu.Unlock()
		}
		// 等待 SSH 会话结束
		_ = sshSession.Wait()
		session.Close()
	}()

	return session, nil
}

// WriteInput 向终端写入输入
func (s *TerminalSession) WriteInput(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("会话已关闭")
	}
	if s.Type == "ssh" {
		_, err := s.SSHStdin.Write(data)
		return err
	}
	_, err := s.Pty.Write(data)
	return err
}

// ResizeWindow 调整终端窗口大小
func (s *TerminalSession) ResizeWindow(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("会话已关闭")
	}
	if s.Type == "ssh" && s.SSHSession != nil {
		return s.SSHSession.WindowChange(rows, cols)
	}
	return nil
}

// Close 关闭终端会话
func (s *TerminalSession) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	if s.Type == "ssh" {
		if s.SSHStdin != nil {
			s.SSHStdin.Close()
		}
		if s.SSHSession != nil {
			s.SSHSession.Close()
		}
		if s.SSHClient != nil {
			s.SSHClient.Close()
		}
	} else {
		if s.Pty != nil {
			s.Pty.Close()
		}
		if s.Cmd != nil && s.Cmd.Process != nil {
			_ = s.Cmd.Process.Kill()
		}
	}
	_ = s.WS.Close()

	terminalSessionsMu.Lock()
	delete(terminalSessions, s.ID)
	terminalSessionsMu.Unlock()
}

// GetTerminalSession 获取终端会话
func GetTerminalSession(id string) *TerminalSession {
	terminalSessionsMu.RLock()
	defer terminalSessionsMu.RUnlock()
	return terminalSessions[id]
}

// 确保 model 包被引用
var _ = model.Server{}
