package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"mayfly/internal/crypto"
	"mayfly/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

var (
	serverStore   *model.ServerStore
	serverStoreMu sync.Mutex
	serverFile    string
)

// ensureServerStore 确保服务器存储已初始化（线程安全）
func ensureServerStore() {
	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()
	if serverStore == nil {
		initServerStore()
	}
}

// initServerStore 初始化服务器存储（不加锁，调用方负责加锁）
func initServerStore() {
	dataDir := filepath.Join(".", "data")
	os.MkdirAll(dataDir, 0755)
	serverFile = filepath.Join(dataDir, "servers.json")

	data, err := os.ReadFile(serverFile)
	if err != nil {
		// 文件不存在，创建空存储
		serverStore = &model.ServerStore{
			Servers: []model.Server{},
			NextID:  1,
			Groups:  []string{},
		}
		saveServerStore()
		return
	}

	serverStore = &model.ServerStore{}
	if err := json.Unmarshal(data, serverStore); err != nil {
		serverStore = &model.ServerStore{
			Servers: []model.Server{},
			NextID:  1,
			Groups:  []string{},
		}
		saveServerStore()
		return
	}

	// 解密敏感字段（兼容历史明文数据，原样返回）
	for i := range serverStore.Servers {
		serverStore.Servers[i].Password, _ = crypto.Decrypt(serverStore.Servers[i].Password)
		serverStore.Servers[i].PrivateKey, _ = crypto.Decrypt(serverStore.Servers[i].PrivateKey)
	}
}

// saveServerStore 保存服务器存储到文件（敏感字段加密后落盘）
func saveServerStore() {
	// 深拷贝后加密，避免污染内存中的明文数据
	snap := *serverStore
	snap.Servers = make([]model.Server, len(serverStore.Servers))
	for i, s := range serverStore.Servers {
		snap.Servers[i] = s
		snap.Servers[i].Password, _ = crypto.Encrypt(s.Password)
		snap.Servers[i].PrivateKey, _ = crypto.Encrypt(s.PrivateKey)
	}
	data, _ := json.MarshalIndent(&snap, "", "  ")
	os.WriteFile(serverFile, data, 0644)
}

// saveServerResult 更新指定服务器的最近测试结果并落盘（调用方需持锁）
func saveServerResult(id int, status, message string) {
	for i := range serverStore.Servers {
		if serverStore.Servers[i].ID == id {
			now := time.Now()
			serverStore.Servers[i].LastTestStatus = status
			serverStore.Servers[i].LastTestTime = &now
			serverStore.Servers[i].LastTestMessage = message
			saveServerStore()
			return
		}
	}
}

// GetServers 获取服务器列表
func GetServers(c *gin.Context) {
	ensureServerStore()

	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()

	// 构建分组列表
	groupMap := make(map[string]int)
	for _, s := range serverStore.Servers {
		group := s.Group
		if group == "" {
			group = "默认"
		}
		groupMap[group]++
	}

	groups := []model.ServerGroup{}
	for name, count := range groupMap {
		groups = append(groups, model.ServerGroup{Name: name, Count: count})
	}

	c.JSON(http.StatusOK, gin.H{
		"servers": serverStore.Servers,
		"groups":  groups,
	})
}

// CreateServer 创建服务器
func CreateServer(c *gin.Context) {
	ensureServerStore()
	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()

	var s model.Server
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if s.Host == "" || s.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP地址和用户名不能为空"})
		return
	}

	if s.Port == 0 {
		s.Port = 22
	}

	s.ID = serverStore.NextID
	serverStore.NextID++
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	serverStore.Servers = append(serverStore.Servers, s)
	saveServerStore()

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "server": s})
}

// UpdateServer 更新服务器
func UpdateServer(c *gin.Context) {
	ensureServerStore()
	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()

	var s model.Server
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	for i := range serverStore.Servers {
		if serverStore.Servers[i].ID == s.ID {
			s.CreatedAt = serverStore.Servers[i].CreatedAt
			s.UpdatedAt = time.Now()
			serverStore.Servers[i] = s
			saveServerStore()
			c.JSON(http.StatusOK, gin.H{"message": "更新成功", "server": s})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
}

// DeleteServer 删除服务器
func DeleteServer(c *gin.Context) {
	ensureServerStore()
	serverStoreMu.Lock()
	defer serverStoreMu.Unlock()

	var req struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	for i, s := range serverStore.Servers {
		if s.ID == req.ID {
			serverStore.Servers = append(serverStore.Servers[:i], serverStore.Servers[i+1:]...)
			saveServerStore()
			c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
}

// TestSSHConnection 测试SSH连接
func TestSSHConnection(c *gin.Context) {
	var req model.TestSSHRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if req.Host == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP地址和用户名不能为空"})
		return
	}

	port := req.Port
	if port == 0 {
		port = 22
	}

	// 持久化测试结果（仅当指定了服务器ID）
	record := func(status, msg string) {
		if req.ServerID <= 0 {
			return
		}
		ensureServerStore()
		serverStoreMu.Lock()
		saveServerResult(req.ServerID, status, msg)
		serverStoreMu.Unlock()
	}

	var authMethods []ssh.AuthMethod

	// 优先使用私钥
	if req.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(req.PrivateKey))
		if err != nil {
			msg := "私钥解析失败: " + err.Error()
			record("fail", msg)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
			return
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 密码认证
	if req.Password != "" {
		authMethods = append(authMethods, ssh.Password(req.Password))
	}

	if len(authMethods) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供密码或私钥"})
		return
	}

	config := &ssh.ClientConfig{
		User:            req.Username,
		Auth:            authMethods,
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", req.Host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		msg := "连接失败: " + err.Error()
		record("fail", msg)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}
	defer client.Close()

	// 尝试执行命令验证
	session, err := client.NewSession()
	if err != nil {
		msg := "创建会话失败: " + err.Error()
		record("fail", msg)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput("hostname")
	if err != nil {
		msg := "执行命令失败: " + err.Error()
		record("fail", msg)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}

	hostname := string(output)
	// 去除末尾换行
	if len(hostname) > 0 && hostname[len(hostname)-1] == '\n' {
		hostname = hostname[:len(hostname)-1]
	}

	record("ok", hostname)
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "连接成功",
		"hostname": hostname,
	})
}
