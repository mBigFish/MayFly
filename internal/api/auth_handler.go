package api

import (
	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/service"
)

// Login 登录
// POST /api/v1/auth/login
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	db := database.Get()
	user, err := service.GetUserByUsername(db, req.Username)
	if err != nil {
		Fail(c, 401, "用户名或密码错误")
		return
	}

	if user.Status != 1 {
		Fail(c, 403, "账户已被禁用")
		return
	}

	if err := service.VerifyPassword(user.Password, req.Password); err != nil {
		Fail(c, 401, "用户名或密码错误")
		return
	}

	token, err := service.GenerateToken(user)
	if err != nil {
		Fail(c, 500, "生成 token 失败")
		return
	}

	_ = service.UpdateLastLogin(db, user.ID)

	OK(c, gin.H{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
	})
}

// Logout 登出
// POST /api/v1/auth/logout
func Logout(c *gin.Context) {
	// JWT 是无状态的，前端清除 token 即可
	OK(c, nil)
}

// ChangePassword 修改密码
// POST /api/v1/auth/change-password
func ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	userID := c.GetUint("user_id")
	if userID == 0 {
		Fail(c, 401, "未登录")
		return
	}

	if err := service.ChangePassword(database.Get(), userID, req.OldPassword, req.NewPassword); err != nil {
		Fail(c, 400, err.Error())
		return
	}

	OKMsg(c, "密码修改成功", nil)
}

// GetUserInfo 获取当前用户信息
// GET /api/v1/auth/info
func GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		Fail(c, 401, "未登录")
		return
	}

	user, err := service.GetUserByID(database.Get(), userID)
	if err != nil {
		Fail(c, 404, "用户不存在")
		return
	}

	OK(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"email":    user.Email,
		"role":     user.Role,
	})
}
