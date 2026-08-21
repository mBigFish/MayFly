package target

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/webshell-manager/webshell-manager/internal/crypto"
)

// Service 目标服务，负责业务逻辑、敏感字段加密与输入校验。
type Service struct {
	repo    *Repository
	encrypt *crypto.AESGCM
}

// NewService 创建目标服务。
func NewService(repo *Repository, encrypt *crypto.AESGCM) *Service {
	return &Service{repo: repo, encrypt: encrypt}
}

// Create 校验并创建目标（敏感字段加密后入库）。
func (s *Service) Create(ctx context.Context, t *Target) error {
	if err := s.validate(t); err != nil {
		return err
	}
	if err := s.encryptSensitive(t); err != nil {
		return err
	}
	return s.repo.Create(ctx, t)
}

// Get 获取目标并解密敏感字段。
func (s *Service) Get(ctx context.Context, id uint) (*Target, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.decryptSensitive(t); err != nil {
		return nil, err
	}
	return t, nil
}

// List 分页列出目标（解密敏感字段）。
func (s *Service) List(ctx context.Context, offset, limit int) ([]Target, int64, error) {
	targets, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range targets {
		if err := s.decryptSensitive(&targets[i]); err != nil {
			return nil, 0, err
		}
	}
	return targets, total, nil
}

// Update 校验并更新目标。
// 为支持部分更新：先读取现有记录，请求中未提供的敏感字段（空字符串）保留原值。
func (s *Service) Update(ctx context.Context, t *Target) error {
	if t.ID == 0 {
		return errors.New("目标 ID 不能为空")
	}

	// 读取现有记录以保留未提供的敏感字段。
	existing, err := s.repo.GetByID(ctx, t.ID)
	if err != nil {
		return err
	}

	// 敏感字段：请求未提供（空字符串）时保留原值（已加密）。
	if t.Cookies == "" {
		t.Cookies = existing.Cookies
	}
	if t.Headers == "" {
		t.Headers = existing.Headers
	}

	// 基础字段校验与默认值（用合并后的值）。
	if err := s.validate(t); err != nil {
		return err
	}

	// 仅对新提供的敏感字段加密，已有密文不再重复加密。
	if t.Cookies != "" && t.Cookies != existing.Cookies {
		enc, err := s.encrypt.Encrypt(t.Cookies)
		if err != nil {
			return fmt.Errorf("加密 Cookies 失败: %w", err)
		}
		t.Cookies = enc
	}
	if t.Headers != "" && t.Headers != existing.Headers {
		enc, err := s.encrypt.Encrypt(t.Headers)
		if err != nil {
			return fmt.Errorf("加密 Headers 失败: %w", err)
		}
		t.Headers = enc
	}

	return s.repo.Update(ctx, t)
}

// Delete 删除目标。
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// MaskSensitive 对敏感字段做掩码，用于列表等只读场景避免明文泄露。
func MaskSensitive(t *Target) {
	if t.Cookies != "" {
		t.Cookies = "***"
	}
	if t.Headers != "" {
		t.Headers = "***"
	}
}

// encryptSensitive 加密敏感字段。
func (s *Service) encryptSensitive(t *Target) error {
	if t.Cookies != "" {
		enc, err := s.encrypt.Encrypt(t.Cookies)
		if err != nil {
			return fmt.Errorf("加密 Cookies 失败: %w", err)
		}
		t.Cookies = enc
	}
	if t.Headers != "" {
		enc, err := s.encrypt.Encrypt(t.Headers)
		if err != nil {
			return fmt.Errorf("加密 Headers 失败: %w", err)
		}
		t.Headers = enc
	}
	return nil
}

// decryptSensitive 解密敏感字段（解密失败时置空，避免阻断读取）。
func (s *Service) decryptSensitive(t *Target) error {
	if t.Cookies != "" {
		if dec, err := s.encrypt.Decrypt(t.Cookies); err == nil {
			t.Cookies = dec
		} else {
			t.Cookies = ""
		}
	}
	if t.Headers != "" {
		if dec, err := s.encrypt.Decrypt(t.Headers); err == nil {
			t.Headers = dec
		} else {
			t.Headers = ""
		}
	}
	return nil
}

// validate 校验目标字段，尤其是 URL（SSRF 防护，spec 第 23 节）。
func (s *Service) validate(t *Target) error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("目标名称不能为空")
	}
	if strings.TrimSpace(t.URL) == "" {
		return errors.New("目标 URL 不能为空")
	}

	u, err := url.Parse(t.URL)
	if err != nil {
		return fmt.Errorf("URL 格式非法: %w", err)
	}

	// 协议限制：仅允许 http/https。
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("仅允许 http/https 协议")
	}
	if u.Host == "" {
		return errors.New("URL 缺少主机名")
	}

	// 超时默认值。
	if t.Timeout <= 0 {
		t.Timeout = 30
	}
	if t.Method == "" {
		t.Method = "POST"
	}
	if t.Type == "" {
		t.Type = "webshell"
	}
	if t.Protocol == "" {
		t.Protocol = "php"
	}
	if t.Encoding == "" {
		t.Encoding = "utf-8"
	}

	return nil
}
