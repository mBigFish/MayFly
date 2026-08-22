package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/service"
)

// ListServers 查询 SSH 服务器列表
// GET /api/v1/servers?keyword=xxx&group=xxx
func ListServers(c *gin.Context) {
	keyword := c.Query("keyword")
	group := c.Query("group")
	servers, err := service.ListServers(database.Get(), keyword, group)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OK(c, servers)
}

// GetServer 查询单个服务器
// GET /api/v1/servers/:id
func GetServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	server, err := service.GetServerByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "服务器不存在")
		return
	}
	OK(c, server)
}

// CreateServer 创建服务器
// POST /api/v1/servers
func CreateServer(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Host       string `json:"host" binding:"required"`
		Port       int    `json:"port" binding:"required"`
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
		Group      string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	server := &model.Server{
		Name:       req.Name,
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Group:      req.Group,
	}
	if server.Port == 0 {
		server.Port = 22
	}

	if err := service.CreateServer(database.Get(), server); err != nil {
		Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	OKMsg(c, "创建成功", server)
}

// UpdateServer 更新服务器
// PUT /api/v1/servers/:id
func UpdateServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	server, err := service.GetServerByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "服务器不存在")
		return
	}

	var req struct {
		Name       string `json:"name"`
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
		Group      string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Host != "" {
		server.Host = req.Host
	}
	if req.Port > 0 {
		server.Port = req.Port
	}
	if req.Username != "" {
		server.Username = req.Username
	}
	if req.Password != "" {
		server.Password = req.Password
	}
	if req.PrivateKey != "" {
		server.PrivateKey = req.PrivateKey
	}
	if req.Group != "" {
		server.Group = req.Group
	}

	if err := service.UpdateServer(database.Get(), server); err != nil {
		Fail(c, 500, "更新失败: "+err.Error())
		return
	}
	OKMsg(c, "更新成功", server)
}

// DeleteServer 删除服务器
// DELETE /api/v1/servers/:id
func DeleteServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteServer(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "删除失败: "+err.Error())
		return
	}
	OKMsg(c, "已删除", nil)
}

// TestServer 测试 SSH 连接
// POST /api/v1/servers/:id/test
func TestServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.TestSSHConnection(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "连接失败: "+err.Error())
		return
	}
	OKMsg(c, "连接成功", nil)
}
