//go:build !windows
// +build !windows

package service

import (
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// unixTerminal Linux/Mac 平台使用 creack/pty 实现的终端会话
type unixTerminal struct {
	ptmx *os.File
	cmd  *exec.Cmd
	mu   sync.Mutex
}

// NewTerminal 创建一个基于 PTY 的终端会话
func NewTerminal(shell string, cols, rows uint16) (Terminal, error) {
	cmd := exec.Command(shell)
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
	if err != nil {
		return nil, err
	}

	return &unixTerminal{
		ptmx: ptmx,
		cmd:  cmd,
	}, nil
}

func (t *unixTerminal) Read(buf []byte) (int, error) {
	return t.ptmx.Read(buf)
}

func (t *unixTerminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ptmx.Write(data)
}

func (t *unixTerminal) Resize(cols, rows uint16) error {
	return pty.Setsize(t.ptmx, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
}

func (t *unixTerminal) Close() error {
	if t.ptmx != nil {
		t.ptmx.Close()
	}
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}
	return nil
}