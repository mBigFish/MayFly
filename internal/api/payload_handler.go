package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/service"
)

// GeneratePayloads 生成反向 Shell Payload
// GET /api/v1/payloads/reverse?host=xxx&port=4444
func GeneratePayloads(c *gin.Context) {
	host := c.Query("host")
	port := 0
	if p := c.Query("port"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	if port == 0 {
		port = 4444
	}
	payloads := service.GeneratePayloads(host, port)
	OK(c, payloads)
}

// GenerateShellScript 生成 WebShell 脚本
// GET /api/v1/payloads/shell?type=php&password=xxx
func GenerateShellScript(c *gin.Context) {
	scriptType := c.Query("type")
	if scriptType == "" {
		scriptType = "php"
	}
	password := c.Query("password")
	script, err := service.GenerateWebShellScriptWithPassword(scriptType, password)
	if err != nil {
		Fail(c, 400, err.Error())
		return
	}
	OK(c, gin.H{
		"type":   scriptType,
		"script": script,
	})
}
