package protocol

import "math/rand/v2"

// GenerateNonce 生成一个携带 encoderID 信息的 nonce。
//
// 公式: nonce = randInt * modulus + encoderID
// Server 端通过 nonce % modulus 还原出 encoderID。
//
// modulus 必须 > 0 且 > 所有 encoderID，否则取模结果会冲突。
func GenerateNonce(encoderID int, modulus int) int {
	if modulus <= 0 {
		modulus = 256
	}
	// 保证 randInt >= 1，避免 nonce 太小看起来可疑
	randInt := rand.IntN(10000-1) + 1
	return randInt*modulus + encoderID
}

// ExtractEncoderID 从 nonce 中提取编码器 ID。
//
// 公式: encoderID = nonce % modulus
func ExtractEncoderID(nonce int, modulus int) int {
	if modulus <= 0 {
		return 0
	}
	return nonce % modulus
}
