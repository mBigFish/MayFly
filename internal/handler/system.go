package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"mayfly/config"

	"github.com/gin-gonic/gin"
)

// ===== 审计日志 =====

// AuditRecord 审计日志记录
type AuditRecord struct {
	Time    time.Time `json:"time"`
	User    string    `json:"user"`
	Action  string    `json:"action"`
	Target  string    `json:"target"`
	Detail  string    `json:"detail"`
	IP      string    `json:"ip"`
}

// AuditStore 审计日志存储
type AuditStore struct {
	mu      sync.RWMutex
	records []AuditRecord
	file    string
	maxSize int
}

// NewAuditStore 创建审计日志存储
func NewAuditStore(file string, maxSize int) *AuditStore {
	s := &AuditStore{file: file, maxSize: maxSize}
	s.load()
	return s
}

func (s *AuditStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		s.records = []AuditRecord{}
		return
	}
	_ = json.Unmarshal(data, &s.records)
}

func (s *AuditStore) save() {
	data, _ := json.MarshalIndent(s.records, "", "  ")
	_ = os.WriteFile(s.file, data, 0600)
}

// Log 记录一条审计日志
func (s *AuditStore) Log(user, action, target, detail, ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, AuditRecord{
		Time:   time.Now(),
		User:   user,
		Action: action,
		Target: target,
		Detail: detail,
		IP:     ip,
	})
	// 超出上限时截断
	if len(s.records) > s.maxSize {
		s.records = s.records[len(s.records)-s.maxSize:]
	}
	s.save()
}

// List 返回审计日志（最新的在前）
func (s *AuditStore) List(limit int) []AuditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := len(s.records)
	if limit <= 0 || limit > n {
		limit = n
	}
	result := make([]AuditRecord, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.records[n-1-i]
	}
	return result
}

// Clear 清空审计日志
func (s *AuditStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = []AuditRecord{}
	s.save()
}

// ===== 运行时设置 =====

// RuntimeSettings 运行时可修改的设置
type RuntimeSettings struct {
	SessionTimeout int    `json:"session_timeout"`
	Shell          string `json:"shell"`
}

// SettingsStore 设置存储
type SettingsStore struct {
	mu       sync.RWMutex
	settings RuntimeSettings
	file     string
}

// NewSettingsStore 创建设置存储
func NewSettingsStore(file string) *SettingsStore {
	s := &SettingsStore{file: file}
	// 默认值从 config 读取
	cfg := config.Get()
	s.settings = RuntimeSettings{
		SessionTimeout: cfg.SessionTimeout,
		Shell:          cfg.Shell,
	}
	s.load()
	return s
}

func (s *SettingsStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.settings)
}

func (s *SettingsStore) save() {
	data, _ := json.MarshalIndent(s.settings, "", "  ")
	_ = os.WriteFile(s.file, data, 0600)
}

// Get 获取当前设置
func (s *SettingsStore) Get() RuntimeSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// Update 更新设置
func (s *SettingsStore) Update(rs RuntimeSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rs.SessionTimeout > 0 {
		s.settings.SessionTimeout = rs.SessionTimeout
	}
	if rs.Shell != "" {
		s.settings.Shell = rs.Shell
	}
	s.save()
	// 同步到运行时 config
	cfg := config.Get()
	cfg.SessionTimeout = s.settings.SessionTimeout
	cfg.Shell = s.settings.Shell
}

// ===== SystemHandler =====

// SystemHandler 系统管理
type SystemHandler struct {
	audit    *AuditStore
	settings *SettingsStore
	startAt  time.Time
}

// NewSystemHandler 创建系统管理 Handler
func NewSystemHandler(audit *AuditStore, settings *SettingsStore) *SystemHandler {
	return &SystemHandler{audit: audit, settings: settings, startAt: time.Now()}
}

// SysInfo 返回系统信息
// GET /api/system/info
func (h *SystemHandler) SysInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startAt)
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	mins := int(uptime.Minutes()) % 60
	secs := int(uptime.Seconds()) % 60

	uptimeStr := ""
	if days > 0 {
		uptimeStr = fmt.Sprintf("%d天 %d小时 %d分 %d秒", days, hours, mins, secs)
	} else if hours > 0 {
		uptimeStr = fmt.Sprintf("%d小时 %d分 %d秒", hours, mins, secs)
	} else {
		uptimeStr = fmt.Sprintf("%d分 %d秒", mins, secs)
	}

	// 数据文件大小
	dataFiles := map[string]int64{}
	_ = filepath.Walk("data", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			dataFiles[filepath.Base(path)] = info.Size()
		}
		return nil
	})

	// 审计日志条数
	auditCount := len(h.audit.List(0))

	c.JSON(http.StatusOK, gin.H{
		"version":     "1.0.0",
		"go_version":  runtime.Version(),
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"cpu_num":     runtime.NumCPU(),
		"goroutines":  runtime.NumGoroutine(),
		"uptime":      uptimeStr,
		"uptime_secs": int(uptime.Seconds()),
		"mem_alloc":   m.Alloc / 1024 / 1024,
		"mem_total":   m.TotalAlloc / 1024 / 1024,
		"mem_sys":     m.Sys / 1024 / 1024,
		"gc_count":    m.NumGC,
		"data_files":  dataFiles,
		"audit_count": auditCount,
		"start_time":  h.startAt.Format("2006-01-02 15:04:05"),
	})
}

// GetSettings 获取运行时设置
// GET /api/system/settings
func (h *SystemHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, h.settings.Get())
}

// UpdateSettings 更新运行时设置
// PUT /api/system/settings
func (h *SystemHandler) UpdateSettings(c *gin.Context) {
	var rs RuntimeSettings
	if err := c.ShouldBindJSON(&rs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if rs.SessionTimeout < 1 || rs.SessionTimeout > 1440 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话超时时间需在 1~1440 分钟之间"})
		return
	}
	h.settings.Update(rs)
	user := c.GetString("username")
	h.audit.Log(user, "修改设置", "系统设置",
		fmt.Sprintf("session_timeout=%d, shell=%s", rs.SessionTimeout, rs.Shell),
		c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "设置已更新"})
}

// ChangePassword 修改密码
// POST /api/system/password
func (h *SystemHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	cfg := config.Get()
	if req.OldPassword != cfg.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度不能少于 6 位"})
		return
	}
	// 运行时修改密码（不持久化到环境变量，重启后恢复）
	cfg.Password = req.NewPassword
	user := c.GetString("username")
	h.audit.Log(user, "修改密码", "认证配置", "密码已更新", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "密码已修改（重启后恢复为环境变量的值）"})
}

// GetAuditLogs 获取审计日志
// GET /api/system/audit-logs?limit=100
func (h *SystemHandler) GetAuditLogs(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
		if limit <= 0 {
			limit = 100
		}
	}
	logs := h.audit.List(limit)
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": len(h.audit.List(0))})
}

// ClearAuditLogs 清空审计日志
// DELETE /api/system/audit-logs
func (h *SystemHandler) ClearAuditLogs(c *gin.Context) {
	user := c.GetString("username")
	h.audit.Log(user, "清空审计日志", "审计日志", "", c.ClientIP())
	h.audit.Clear()
	c.JSON(http.StatusOK, gin.H{"message": "审计日志已清空"})
}

// ExportData 导出所有数据
// GET /api/system/export
func (h *SystemHandler) ExportData(c *gin.Context) {
	result := map[string]json.RawMessage{}
	files, _ := filepath.Glob("data/*.json")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err == nil {
			result[filepath.Base(f)] = data
		}
	}
	user := c.GetString("username")
	h.audit.Log(user, "导出数据", "系统数据", fmt.Sprintf("导出 %d 个文件", len(result)), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"data": result, "exported_at": time.Now().Format("2006-01-02 15:04:05")})
}

// ImportData 导入数据
// POST /api/system/import
func (h *SystemHandler) ImportData(c *gin.Context) {
	var req struct {
		Data map[string]json.RawMessage `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	imported := 0
	for name, raw := range req.Data {
		// 只允许 .json 文件
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		// 防止路径穿越
		clean := filepath.Base(name)
		path := filepath.Join("data", clean)
		if err := os.WriteFile(path, raw, 0600); err == nil {
			imported++
		}
	}
	user := c.GetString("username")
	h.audit.Log(user, "导入数据", "系统数据", fmt.Sprintf("导入 %d 个文件", imported), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功导入 %d 个数据文件，需重启服务生效", imported)})
}

// ===== 审计中间件 =====

// AuditMiddleware 创建审计中间件，记录关键操作
func AuditMiddleware(audit *AuditStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续处理
		c.Next()

		// 只记录成功的写操作（2xx）
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			return
		}

		user := c.GetString("username")
		if user == "" {
			user = "anonymous"
		}
		path := c.Request.URL.Path
		action := method + " " + path

		// 简化记录
		audit.Log(user, action, path, "", c.ClientIP())
	}
}
