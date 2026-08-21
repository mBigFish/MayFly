package model

import "time"

// Node 表示一个目标 WebShell 节点
type Node struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`          // 显示名称
	URL       string    `json:"url"`           // WebShell 脚本完整 URL
	Pass      string    `json:"pass"`          // 连接密码（POST 字段名 / key）
	Type      string    `json:"type"`          // php / jsp / aspx / asp
	Group     string    `json:"group"`        // 分组名称，空则为"默认"
	Remark    string    `json:"remark"`        // 备注
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 最近一次连接测试结果（持久化，刷新/切换页面不丢失）
	LastTestStatus  string     `json:"last_test_status"`  // untested / testing / ok / fail
	LastTestTime    *time.Time `json:"last_test_time"`    // 测试时间
	LastTestMessage string     `json:"last_test_message"` // 结果详情（连接成功或错误信息）
}