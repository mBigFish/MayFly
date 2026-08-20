//go:build windows
// +build windows

package service

import (
	"sync"

	"github.com/UserExistsError/conpty"
)

// windowsTerminal Windows 平台使用 ConPTY 实现的终端会话
type windowsTerminal struct {
	cpty *conpty.ConPty
	mu   sync.Mutex
}

// NewTerminal 创建一个基于 Windows ConPTY 的终端会话
func NewTerminal(shell string, cols, rows uint16) (Terminal, error) {
	cpty, err := conpty.Start(shell, conpty.ConPtyDimensions(int(cols), int(rows)))
	if err != nil {
		return nil, err
	}
	return &windowsTerminal{cpty: cpty}, nil
}

func (t *windowsTerminal) Read(buf []byte) (int, error) {
	return t.cpty.Read(buf)
}

func (t *windowsTerminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cpty.Write(data)
}

func (t *windowsTerminal) Resize(cols, rows uint16) error {
	return t.cpty.Resize(int(cols), int(rows))
}

func (t *windowsTerminal) Close() error {
	if t.cpty == nil {
		return nil
	}
	return t.cpty.Close()
}
