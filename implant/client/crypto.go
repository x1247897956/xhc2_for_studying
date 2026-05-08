package client

import (
	"fmt"
	"math/rand/v2"

	"xhc2_for_studying/protocol"
)

// pickEncoder 随机选择一个已注册的编码器，返回其 ID 和实例。
func pickEncoder() (int, protocol.Encoder, error) {
	// 模拟：随机选 Plain(0) 或 Base64(1)
	id := rand.IntN(2)
	enc, ok := protocol.GetEncoderByID(id)
	if !ok {
		return 0, nil, fmt.Errorf("encoder %d not found", id)
	}
	return id, enc, nil
}

// encodeBody 用 encoderID 对应的编码器编码请求体。
func encodeBody(data []byte, encoderID int) ([]byte, error) {
	enc, ok := protocol.GetEncoderByID(encoderID)
	if !ok {
		return nil, fmt.Errorf("encoder %d not found", encoderID)
	}
	return enc.Encode(data)
}

// decodeBody 用 encoderID 对应的编码器解码响应体。
func decodeBody(data []byte, encoderID int) ([]byte, error) {
	enc, ok := protocol.GetEncoderByID(encoderID)
	if !ok {
		return nil, fmt.Errorf("encoder %d not found", encoderID)
	}
	return enc.Decode(data)
}

// RequestContext 封装一次请求的 nonce 和 encoder 信息。
// Implant 用它编码请求、解码响应；Server 端从 nonce 还原出同样的 encoder。
type RequestContext struct {
	Nonce     int
	EncoderID int
}

// NewRequestContext 生成一个新的请求上下文：随机选 encoder，生成对应的 nonce。
func NewRequestContext(modulus int) *RequestContext {
	encoderID, _, err := pickEncoder()
	if err != nil {
		// 兜底：用 Plain
		encoderID = 0
	}
	return &RequestContext{
		Nonce:     protocol.GenerateNonce(encoderID, modulus),
		EncoderID: encoderID,
	}
}
