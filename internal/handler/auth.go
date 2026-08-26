package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"mayfly/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// Claims JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Login 处理登录请求
func Login(c *gin.Context) {
	ip := c.ClientIP()

	// 防爆破：IP 已被锁定时直接拒绝
	if limiter.isLocked(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "尝试次数过多，请稍后再试"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	if req.Username != cfg.Username || req.Password != cfg.Password {
		// 失败延迟，减缓爆破速度
		time.Sleep(loginFailDelay)
		limiter.recordFailure(ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 登录成功，清除该 IP 的失败记录
	limiter.reset(ip)

	// 生成 JWT
	claims := &Claims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.SessionTimeout) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}

// ValidateToken 验证 JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Get()
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// EncodeJSON 编码工具
func EncodeJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
