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

const prefix = "enc:"

var secretKey = sha256Key("mayfly-webshell-manager-secret-salt-v1")

func sha256Key(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

func isEncrypted(s string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Encrypt 使用 AES-256-GCM 加密明文
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

// Decrypt 解密带前缀的密文
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
