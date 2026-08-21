package crypto

import "testing"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	e, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM 失败: %v", err)
	}

	plaintext := "session=abc123; token=secret"
	enc, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}
	if enc == plaintext {
		t.Error("密文不应与明文相同")
	}

	dec, err := e.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt 失败: %v", err)
	}
	if dec != plaintext {
		t.Errorf("解密结果不符，期望 %q，得到 %q", plaintext, dec)
	}
}

func TestInvalidKeyLength(t *testing.T) {
	if _, err := NewAESGCM([]byte("short")); err == nil {
		t.Error("非 32 字节密钥应报错")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	e, _ := NewAESGCM(key)

	if _, err := e.Decrypt("!!!not-base64!!!"); err == nil {
		t.Error("非法 base64 输入应报错")
	}
}
