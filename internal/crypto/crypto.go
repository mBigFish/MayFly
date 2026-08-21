package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// prefix 加密字段前缀，用于区分密文与历史明文数据（向后兼容）
const prefix = "enc:"

// secretKey AES-256 密钥（32 字节），由固定盐派生。
// 该方案为“混淆级加密”：防止数据文件被直接打开即可读到明文凭据，
// 适用于本地单用户工具。如需更强保护可替换为用户主密码派生。
var secretKey = sha256Key("mayfly-webshell-manager-secret-salt-v1")

func sha256Key(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// isEncrypted 判断字符串是否已是密文
func isEncrypted(s string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Encrypt 使用 AES-256-GCM 加密明文，返回带前缀的 base64 密文。
// 空字符串原样返回；已加密值原样返回（幂等）。
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" || isEncrypted(plaintext) {
		return plaintext, nil
	}

	block, err := aes.NewCipher(secretKey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密带前缀的密文；空字符串或历史明文原样返回（向后兼容）。
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" || !isEncrypted(ciphertext) {
		return ciphertext, nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext[len(prefix):])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("密文数据不完整")
	}

	nonce, sealed := data[:ns], data[ns:]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}