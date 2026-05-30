package protocol

import "math/rand/v2"

// GenerateNonce produces a nonce that encodes the given encoderID so the
// server can recover it later.
//
// The nonce is computed as:
//
//	nonce = randInt * modulus + encoderID
//
// The server recovers encoderID via nonce % modulus.
//
// modulus must be greater than zero and greater than all encoder IDs,
// otherwise the modulo extraction will produce collisions.
func GenerateNonce(encoderID int, modulus int) int {
	if modulus <= 0 {
		modulus = 256
	}
	// Ensure randInt >= 1 so the nonce is never suspiciously small.
	randInt := rand.IntN(10000-1) + 1
	return randInt*modulus + encoderID
}

// ExtractEncoderID recovers the encoder identifier from a nonce.
//
// It computes:
//
//	encoderID = nonce % modulus
func ExtractEncoderID(nonce int, modulus int) int {
	if modulus <= 0 {
		return 0
	}
	return nonce % modulus
}
