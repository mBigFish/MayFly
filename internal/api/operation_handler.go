package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mayfly/internal/crypto"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/service"
)

// CheckTarget 测试目标连接
// POST /api/v1/targets/:id/check
func CheckTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	if err := protocol.CheckTarget(ctx, target); err != nil {
		_ = service.UpdateTargetStatus(database.Get(), uint(id), "offline")
		Fail(c, 500, "连接失败: "+err.Error())
		return
	}

	_ = service.UpdateTargetStatus(database.Get(), uint(id), "online")
	OKMsg(c, "连接成功", gin.H{"status": "online"})
}

// ExecuteCommand 执行命令
// POST /api/v1/targets/:id/execute
func ExecuteCommand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpCommand,
		Params: map[string]string{"cmd": req.Command},
	})
	if err != nil {
		Fail(c, 500, "执行失败: "+err.Error())
		return
	}

	// 记录审计日志
	service.LogAction(database.Get(), c.GetUint("user_id"), c.GetString("username"),
		"execute", "target", uint(id), "command: "+req.Command, c.ClientIP(), "success")

	// 保存命令历史
	_ = service.CreateCmdHistory(database.Get(), &model.CmdHistory{
		TargetID: uint(id),
		UserID:   c.GetUint("user_id"),
		Username: c.GetString("username"),
		Command:  req.Command,
		Output:   string(result.Data),
	})

	OK(c, gin.H{
		"output":   string(result.Data),
		"duration": result.Message,
	})
}

// ListFiles 列目录
// POST /api/v1/targets/:id/files
func ListFiles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path string `json:"path"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Path == "" {
		req.Path = "/"
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpListDir,
		Params: map[string]string{"path": req.Path},
	})
	if err != nil {
		Fail(c, 500, "列目录失败: "+err.Error())
		return
	}

	// 尝试解析 JSON 格式的文件列表
	var entries []*protocol.FileEntry
	if err := json.Unmarshal(result.Data, &entries); err == nil {
		OK(c, &protocol.FileList{
			Path:    req.Path,
			Entries: entries,
		})
		return
	}

	// 纯文本格式，直接返回
	OK(c, gin.H{
		"path":    req.Path,
		"raw":     string(result.Data),
	})
}

// ReadFile 读文件
// POST /api/v1/targets/:id/files/read
func ReadFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpReadFile,
		Params: map[string]string{"path": req.Path},
	})
	if err != nil {
		Fail(c, 500, "读取失败: "+err.Error())
		return
	}

	OK(c, gin.H{
		"path":    req.Path,
		"content": string(result.Data),
	})
}

// WriteFile 写文件
// POST /api/v1/targets/:id/files/write
func WriteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err = executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpWriteFile,
		Params: map[string]string{"path": req.Path, "content": req.Content},
	})
	if err != nil {
		Fail(c, 500, "写入失败: "+err.Error())
		return
	}

	OKMsg(c, "写入成功", nil)
}

// DeleteFile 删除文件
// POST /api/v1/targets/:id/files/delete
func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err = executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpDeleteFile,
		Params: map[string]string{"path": req.Path},
	})
	if err != nil {
		Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	OKMsg(c, "删除成功", nil)
}

// RenameFile 重命名文件/目录
// POST /api/v1/targets/:id/files/rename
func RenameFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		OldPath string `json:"old_path" binding:"required"`
		NewPath string `json:"new_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err = executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpRenameFile,
		Params: map[string]string{"oldPath": req.OldPath, "newPath": req.NewPath},
	})
	if err != nil {
		Fail(c, 500, "重命名失败: "+err.Error())
		return
	}

	OKMsg(c, "重命名成功", nil)
}

// Mkdir 创建目录
// POST /api/v1/targets/:id/files/mkdir
func Mkdir(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err = executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpMkdir,
		Params: map[string]string{"path": req.Path},
	})
	if err != nil {
		Fail(c, 500, "创建目录失败: "+err.Error())
		return
	}

	OKMsg(c, "创建成功", nil)
}

// DownloadFile 下载文件
// POST /api/v1/targets/:id/files/download
func DownloadFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpReadFile,
		Params: map[string]string{"path": req.Path},
	})
	if err != nil {
		Fail(c, 500, "下载失败: "+err.Error())
		return
	}

	// 提取文件名
	filename := req.Path
	if idx := strings.LastIndex(req.Path, "/"); idx >= 0 {
		filename = req.Path[idx+1:]
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(200, "application/octet-stream", result.Data)
}

// GetSysInfo 获取系统信息
// GET /api/v1/targets/:id/info
func GetSysInfo(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	result, err := executeAndLog(ctx, c, target, &protocol.Operation{
		Type:   protocol.OpSysInfo,
		Params: map[string]string{},
	})
	if err != nil {
		Fail(c, 500, "获取信息失败: "+err.Error())
		return
	}

	OK(c, gin.H{
		"info": string(result.Data),
	})
}

// EncryptPassword 加密目标密码（工具接口）
// POST /api/v1/targets/encrypt
func EncryptPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	enc, err := crypto.Encrypt(req.Password)
	if err != nil {
		Fail(c, 500, "加密失败")
		return
	}
	OK(c, gin.H{"encrypted": enc})
}

// BatchCheck 批量测试连接
// POST /api/v1/targets/batch-check
func BatchCheck(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	results := make([]gin.H, 0, len(req.IDs))
	for _, id := range req.IDs {
		target, err := service.GetTargetByID(database.Get(), id)
		if err != nil {
			results = append(results, gin.H{"id": id, "status": "offline", "error": "目标不存在"})
			continue
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		err = protocol.CheckTarget(ctx, target)
		cancel()
		if err != nil {
			_ = service.UpdateTargetStatus(database.Get(), id, "offline")
			results = append(results, gin.H{"id": id, "name": target.Name, "status": "offline", "error": err.Error()})
		} else {
			_ = service.UpdateTargetStatus(database.Get(), id, "online")
			results = append(results, gin.H{"id": id, "name": target.Name, "status": "online"})
		}
	}
	OK(c, results)
}

// 确保 model 包被引用
var _ = model.Target{}
