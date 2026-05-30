// Package cryptography provides utilities for managing Age encryption key pairs.
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

// EnsureAgeKeyPair checks whether the server Age key pair exists on disk and
// returns it. If the keys are missing, it generates and persists a new pair
// before returning it.
func EnsureAgeKeyPair() (publicKey string, privateKey string, err error) {
	if _, err := os.Stat(pubFile); err == nil {
		if _, err := os.Stat(priFile); err == nil {
			return LoadKeyPair()
		}
	}
	return GenerateAndSaveKeyPair()
}

// GenerateAndSaveKeyPair creates a new X25519 Age identity, writes both keys
// to disk, and returns the public and private key strings.
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

// LoadKeyPair reads the Age public and private keys from disk and returns them
// as strings.
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
