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

// ErrReplayAttack is returned when a previously seen ciphertext packet
// is replayed, indicating a potential replay attack.
var ErrReplayAttack = errors.New("replay attack detected")

// CipherContext wraps a ChaCha20Poly1305 AEAD cipher and provides replay
// protection by caching SHA-256 digests of every received ciphertext.
type CipherContext struct {
	aead   cipher.AEAD
	replay sync.Map // map[[32]byte]bool, caches SHA-256 digests of seen packets.
}

// NewCipherContext creates a new encryption context from a 32-byte key.
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

// Encrypt seals plaintext and returns a packet containing nonce followed
// by ciphertext and authentication tag. It also returns the base64-encoded
// nonce for URL transport.
func (c *CipherContext) Encrypt(plaintext []byte) (packet []byte, nonceB64 string, err error) {
	nonce := make([]byte, c.aead.NonceSize()) // 12 bytes for XChaCha20.
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	packet = make([]byte, 0, len(nonce)+len(sealed))
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet, base64.StdEncoding.EncodeToString(nonce), nil
}

// Decrypt extracts the nonce from the packet, checks for replay via the
// SHA-256 digest cache, and then decrypts and verifies the authentication
// tag.
func (c *CipherContext) Decrypt(packet []byte) ([]byte, error) {
	// Check for replay: compute the SHA-256 digest of the ciphertext.
	digest := sha256.Sum256(packet)

	// If the digest already exists, this is a replayed packet.
	if _, exists := c.replay.LoadOrStore(digest, true); exists {
		return nil, ErrReplayAttack
	}

	// Proceed with normal decryption.
	nonceSize := c.aead.NonceSize()
	if len(packet) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce := packet[:nonceSize]
	sealed := packet[nonceSize:]
	return c.aead.Open(nil, nonce, sealed, nil)
}

// NonceSize is the size of the ChaCha20Poly1305 nonce in bytes (12).
const NonceSize = chacha20poly1305.NonceSize
