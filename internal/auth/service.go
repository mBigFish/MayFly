package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/webshell-manager/webshell-manager/internal/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 认证相关错误。
var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserDisabled       = errors.New("用户已被禁用")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidToken       = errors.New("无效的令牌")
)

// Service 认证服务。
type Service struct {
	jwtSecret []byte
	jwtTTL    time.Duration
}

// NewService 创建认证服务。
func NewService(jwtSecret string, jwtTTL time.Duration) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
		jwtTTL:    jwtTTL,
	}
}

// Claims 是 JWT 的自定义载荷。
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// HashPassword 使用 bcrypt 生成密码哈希。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 校验明文密码与哈希是否匹配。
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Login 校验用户名密码并签发 JWT。
func (s *Service) Login(ctx context.Context, username, password string) (string, *User, error) {
	var user User
	err := database.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !user.Enabled {
		return "", nil, ErrUserDisabled
	}

	if !VerifyPassword(user.PasswordHash, password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(&user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// issueToken 签发 JWT。
func (s *Service) issueToken(user *User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "webshell-manager",
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ParseToken 解析并校验 JWT，返回 Claims。
func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GetUser 根据用户 ID 加载用户及其角色权限。
func GetUser(ctx context.Context, id uint) (*User, error) {
	var user User
	err := database.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// HasPermission 判断用户是否拥有指定权限。
func (u *User) HasPermission(code string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Code == code {
				return true
			}
		}
	}
	return false
}
