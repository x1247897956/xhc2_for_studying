package protocol

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
)

// ErrReplayAttack 表示检测到重放攻击。
var ErrReplayAttack = errors.New("replay attack detected")

// CipherContext 封装 ChaCha20Poly1305 AEAD 对称加解密，并内置重放保护。
type CipherContext struct {
	aead   cipher.AEAD
	replay sync.Map // map[[32]byte]bool，SHA256 摘要缓存
}

// NewCipherContext 用 32 字节密钥创建加密上下文。
func NewCipherContext(key []byte) (*CipherContext, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size, need 32 bytes")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	return &CipherContext{aead: aead}, nil
}

// Encrypt 返回 packet(nonce+密文+标签) 和 nonce 的 base64 字符串（用于 URL）。
func (c *CipherContext) Encrypt(plaintext []byte) (packet []byte, nonceB64 string, err error) {
	nonce := make([]byte, c.aead.NonceSize()) // 12 字节
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	packet = make([]byte, 0, len(nonce)+len(sealed))
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet, base64.StdEncoding.EncodeToString(nonce), nil
}

// Decrypt 从 packet 中拆出 nonce，先验重放，再解密并验证认证标签。
func (c *CipherContext) Decrypt(packet []byte) ([]byte, error) {
	// 1. 防重放检查：计算密文 SHA256 摘要
	digest := sha256.Sum256(packet)

	// 2. 如果摘要已存在，则是重放攻击
	if _, exists := c.replay.LoadOrStore(digest, true); exists {
		return nil, ErrReplayAttack
	}

	// 3. 正常解密
	nonceSize := c.aead.NonceSize()
	if len(packet) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce := packet[:nonceSize]
	sealed := packet[nonceSize:]
	return c.aead.Open(nil, nonce, sealed, nil)
}

// NonceSize = 12
const NonceSize = chacha20poly1305.NonceSize
