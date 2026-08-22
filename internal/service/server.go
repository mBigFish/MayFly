package service

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"mayfly/internal/crypto"
	"mayfly/internal/model"
)

// ListServers 查询 SSH 服务器列表
func ListServers(db *gorm.DB, keyword, group string) ([]model.Server, error) {
	var servers []model.Server
	q := db.Model(&model.Server{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if group != "" {
		q = q.Where("group = ?", group)
	}
	err := q.Order("id DESC").Find(&servers).Error
	return servers, err
}

// GetServerByID 按 ID 查询服务器
func GetServerByID(db *gorm.DB, id uint) (*model.Server, error) {
	var server model.Server
	if err := db.First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// CreateServer 创建服务器
func CreateServer(db *gorm.DB, s *model.Server) error {
	// 加密密码和私钥
	if s.Password != "" {
		enc, err := crypto.Encrypt(s.Password)
		if err != nil {
			return err
		}
		s.Password = enc
	}
	if s.PrivateKey != "" {
		enc, err := crypto.Encrypt(s.PrivateKey)
		if err != nil {
			return err
		}
		s.PrivateKey = enc
	}
	return db.Create(s).Error
}

// UpdateServer 更新服务器
func UpdateServer(db *gorm.DB, s *model.Server) error {
	// 如果密码不为空，加密
	if s.Password != "" {
		enc, err := crypto.Encrypt(s.Password)
		if err != nil {
			return err
		}
		s.Password = enc
	}
	if s.PrivateKey != "" {
		enc, err := crypto.Encrypt(s.PrivateKey)
		if err != nil {
			return err
		}
		s.PrivateKey = enc
	}
	return db.Save(s).Error
}

// DeleteServer 删除服务器
func DeleteServer(db *gorm.DB, id uint) error {
	return db.Delete(&model.Server{}, id).Error
}

// TestSSHConnection 测试 SSH 连接
func TestSSHConnection(db *gorm.DB, id uint) error {
	server, err := GetServerByID(db, id)
	if err != nil {
		return err
	}

	password, _ := crypto.Decrypt(server.Password)
	privateKey, _ := crypto.Decrypt(server.PrivateKey)

	config := &ssh.ClientConfig{
		User: server.Username,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err == nil {
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		now := time.Now()
		db.Exec("UPDATE servers SET last_test_status = ?, last_test_message = ?, last_test_time = ? WHERE id = ?",
			"fail", err.Error(), now, id)
		return err
	}
	client.Close()

	now := time.Now()
	db.Exec("UPDATE servers SET last_test_status = ?, last_test_message = ?, last_test_time = ? WHERE id = ?",
		"ok", "连接成功", now, id)
	return nil
}

// SSHSession SSH 会话
type SSHSession struct {
	Client  *ssh.Client
	Session *ssh.Session
}

// CreateSSHSession 创建 SSH 会话
func CreateSSHSession(db *gorm.DB, id uint) (*SSHSession, error) {
	server, err := GetServerByID(db, id)
	if err != nil {
		return nil, err
	}

	password, _ := crypto.Decrypt(server.Password)
	privateKey, _ := crypto.Decrypt(server.PrivateKey)

	config := &ssh.ClientConfig{
		User: server.Username,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err == nil {
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, err
	}

	return &SSHSession{Client: client, Session: session}, nil
}

// Close 关闭 SSH 会话
func (s *SSHSession) Close() {
	if s.Session != nil {
		s.Session.Close()
	}
	if s.Client != nil {
		s.Client.Close()
	}
}
