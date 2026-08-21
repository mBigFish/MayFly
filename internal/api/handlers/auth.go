// Package handlers 定义 HTTP 请求处理器。
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/dto"
	"github.com/webshell-manager/webshell-manager/internal/auth"
)

// AuthHandler 认证相关处理器。
type AuthHandler struct {
	svc     *auth.Service
	limiter *auth.RateLimiter
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(svc *auth.Service, limiter *auth.RateLimiter) *AuthHandler {
	return &AuthHandler{svc: svc, limiter: limiter}
}

// Login 处理登录请求。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}

	// 登录限流（按 IP + 用户名）。
	key := c.ClientIP() + ":" + req.Username
	if !h.limiter.Allow(key) {
		c.JSON(http.StatusTooManyRequests, dto.Error(429, "登录尝试过于频繁，请稍后再试"))
		return
	}

	token, user, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials, auth.ErrUserDisabled:
			c.JSON(http.StatusUnauthorized, dto.Error(401, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.Error(500, "登录失败"))
		}
		return
	}

	// 登录成功，重置限流计数。
	h.limiter.Reset(key)

	// 组装用户角色。
	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{
		Token: token,
		User: dto.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Roles:    roles,
		},
	}))
}
