package protocol

import (
	"bytes"
	"crypto/rand"
	"io"

	"filippo.io/age"
)

// AgeEncryptToServer 用服务端 Age 公钥加密明文。
// serverPublicKey 是服务端 Age 私钥对应的 recipient 字符串。
func AgeEncryptToServer(plaintext []byte, serverPublicKey string) ([]byte, error) {
	recipient, err := age.ParseX25519Recipient(serverPublicKey)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(w, bytes.NewReader(plaintext)); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AgeDecryptFromImplant 用服务端 Age 私钥解密密文。
func AgeDecryptFromImplant(ciphertext []byte, serverPrivateKey string) ([]byte, error) {
	identity, err := age.ParseX25519Identity(serverPrivateKey)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

// GenerateSymmetricKey 生成 32 字节随机对称密钥。
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
