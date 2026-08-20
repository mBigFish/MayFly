package service

// Terminal 终端会话接口，屏蔽不同操作系统的 PTY 实现差异
type Terminal interface {
	Read(buf []byte) (int, error)
	Write(data []byte) (int, error)
	Resize(cols, rows uint16) error
	Close() error
}