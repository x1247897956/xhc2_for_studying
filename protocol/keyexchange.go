package protocol

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"filippo.io/age"
)

// pubKeyDigestLen is the fixed output size of SHA-256 in bytes (32).
const pubKeyDigestLen = 32

// GenerateSymmetricKey produces a 32-byte random symmetric key for use
// with ChaCha20Poly1305.
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateAgeKeyPair generates a new X25519 Age identity and returns
// the PEM-encoded private key and the Bech32-encoded public key string.
func GenerateAgeKeyPair() (privateKey, publicKey string, err error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}
	return identity.String(), identity.Recipient().String(), nil
}

// PubKeyDigestRaw returns the raw 32-byte SHA-256 digest of the given
// Age public key.
func PubKeyDigestRaw(agePublicKey string) []byte {
	digest := sha256.Sum256([]byte(agePublicKey))
	return digest[:]
}

// PubKeyDigest returns the hex-encoded SHA-256 digest of the given Age
// public key. This digest is used as a stable implant identifier.
func PubKeyDigest(agePublicKey string) string {
	return hex.EncodeToString(PubKeyDigestRaw(agePublicKey))
}

// computeHMAC computes an HMAC-SHA256 of the symmetric key using the
// implant's Age private key as the HMAC secret.
// The HMAC key is SHA256(implantAgePrivateKey).
func computeHMAC(implantAgePrivateKey string, symmetricKey []byte) []byte {
	hk := sha256.Sum256([]byte(implantAgePrivateKey))
	mac := hmac.New(sha256.New, hk[:])
	mac.Write(symmetricKey)
	return mac.Sum(nil)
}

// BuildKeyExchangePacket constructs the key exchange packet sent from
// the implant to the server.
//
// Packet layout:
//
//	[32-byte SHA256(implant public key)] + [Age-encrypted(HMAC || symmetric key)]
func BuildKeyExchangePacket(symmetricKey []byte, implantAgePublicKey, implantAgePrivateKey, serverAgePublicKey string) ([]byte, error) {
	// Compute HMAC-SHA256 where key = SHA256(implant private key), message = symmetric key.
	macValue := computeHMAC(implantAgePrivateKey, symmetricKey)

	// Plaintext: HMAC || symmetric key.
	plaintext := make([]byte, 0, sha256.Size+len(symmetricKey))
	plaintext = append(plaintext, macValue...)
	plaintext = append(plaintext, symmetricKey...)

	// Encrypt the plaintext with the server's Age public key.
	recipient, err := age.ParseX25519Recipient(serverAgePublicKey)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	ciphertext := buf.Bytes()

	// Prefix the ciphertext with SHA256(implant public key).
	digest := PubKeyDigestRaw(implantAgePublicKey)

	packet := make([]byte, 0, pubKeyDigestLen+len(ciphertext))
	packet = append(packet, digest...)
	packet = append(packet, ciphertext...)
	return packet, nil
}

// VerifyAndDecryptKeyExchange decrypts a key exchange packet on the
// server side, verifies the HMAC, and returns the symmetric key. An
// error is returned if the HMAC does not match or if decryption fails.
func VerifyAndDecryptKeyExchange(packet []byte, serverAgePrivateKey, implantAgePrivateKey string) ([]byte, error) {
	if len(packet) < pubKeyDigestLen {
		return nil, errors.New("packet too short")
	}

	ciphertext := packet[pubKeyDigestLen:]

	// Decrypt using the server's Age private key.
	identity, err := age.ParseX25519Identity(serverAgePrivateKey)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, err
	}
	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(plaintext) < sha256.Size {
		return nil, errors.New("plaintext too short")
	}

	// Split the plaintext into HMAC and symmetric key.
	receivedMAC := plaintext[:sha256.Size]
	symmetricKey := plaintext[sha256.Size:]

	// Verify the HMAC.
	expectedMAC := computeHMAC(implantAgePrivateKey, symmetricKey)
	if !hmac.Equal(receivedMAC, expectedMAC) {
		return nil, errors.New("HMAC verification failed")
	}

	return symmetricKey, nil
}

// ExtractPubKeyDigest extracts the 32-byte public key digest from the
// beginning of a packet and returns it as a hex-encoded string. The
// boolean indicates whether the packet was long enough to contain a
// digest.
func ExtractPubKeyDigest(packet []byte) (string, bool) {
	if len(packet) < pubKeyDigestLen {
		return "", false
	}
	return hex.EncodeToString(packet[:pubKeyDigestLen]), true
}
