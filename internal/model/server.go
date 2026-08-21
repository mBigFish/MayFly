package model

import "time"

// Server 服务器资源
type Server struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`          // 服务器名称
	Group       string    `json:"group"`         // 分组名称
	Host        string    `json:"host"`          // IP地址
	Port        int       `json:"port"`          // SSH端口
	Username    string    `json:"username"`      // SSH用户名
	Password    string    `json:"password"`      // SSH密码（存储加密）
	PrivateKey  string    `json:"private_key"`   // SSH私钥，可选（存储加密）
	Description string    `json:"description"`   // 描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 最近一次连接测试结果（持久化，刷新/切换页面不丢失）
	LastTestStatus  string     `json:"last_test_status"`  // untested / testing / ok / fail
	LastTestTime    *time.Time `json:"last_test_time"`    // 测试时间
	LastTestMessage string     `json:"last_test_message"` // 结果详情（主机名或错误信息）
}

// ServerGroup 服务器分组
type ServerGroup struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ServerStore 服务器数据存储
type ServerStore struct {
	Servers     []Server     `json:"servers"`
	NextID      int          `json:"next_id"`
	Groups      []string     `json:"groups"`
}

// TestSSHRequest SSH测试连接请求
type TestSSHRequest struct {
	ServerID   int    `json:"server_id"`  // 服务器ID（可选，传入则持久化测试结果）
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}
