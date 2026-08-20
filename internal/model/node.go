package model

import "time"

// Node 表示一个目标 WebShell 节点
type Node struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`          // 显示名称
	URL       string    `json:"url"`           // WebShell 脚本完整 URL
	Pass      string    `json:"pass"`          // 连接密码（POST 字段名 / key）
	Type      string    `json:"type"`          // php / jsp / aspx / asp
	Remark    string    `json:"remark"`        // 备注
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}