package cryptography

import (
	"os"

	"filippo.io/age"
)

const (
	keyDir  = "server/cryptography"
	pubFile = keyDir + "/age_public.key"
	priFile = keyDir + "/age_private.key"
)

// EnsureAgeKeyPair 确保服务端 Age 密钥对存在，不存在则生成。
// 返回公钥和私钥字符串。
func EnsureAgeKeyPair() (publicKey string, privateKey string, err error) {
	if _, err := os.Stat(pubFile); err == nil {
		if _, err := os.Stat(priFile); err == nil {
			return LoadKeyPair()
		}
	}
	return GenerateAndSaveKeyPair()
}

// GenerateAndSaveKeyPair 生成新的 X25519 密钥对并写盘。
func GenerateAndSaveKeyPair() (publicKey string, privateKey string, err error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}

	pubStr := identity.Recipient().String()
	priStr := identity.String()

	if err := os.MkdirAll(keyDir, 0750); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(priFile, []byte(priStr), 0600); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(pubFile, []byte(pubStr), 0644); err != nil {
		return "", "", err
	}

	return pubStr, priStr, nil
}

// LoadKeyPair 从文件加载密钥对。
func LoadKeyPair() (publicKey string, privateKey string, err error) {
	pubBytes, err := os.ReadFile(pubFile)
	if err != nil {
		return "", "", err
	}
	priBytes, err := os.ReadFile(priFile)
	if err != nil {
		return "", "", err
	}
	return string(pubBytes), string(priBytes), nil
}
