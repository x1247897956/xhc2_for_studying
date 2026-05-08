package protocol

import (
	"encoding/base64"
	"fmt"
)

// Encoder 定义了请求/响应体的编解码行为。
// Implant 和 Server 都通过这个接口来处理 C2 流量。
type Encoder interface {
	// ID 返回编码器的唯一标识。nonce % EncoderModulus 的结果必须等于这个 ID。
	ID() int
	// Name 返回编码器的可读名称。
	Name() string
	// Encode 将原始字节编码为传输格式。
	Encode(data []byte) ([]byte, error)
	// Decode 将传输格式解码回原始字节。
	Decode(data []byte) ([]byte, error)
}

// PlainEncoder — 不做任何编码，原始字节直接传输。
// ID=0，方便演示：当 nonce % Modulus == 0 时选中。
type PlainEncoder struct{}

func (e *PlainEncoder) ID() int                            { return 0 }
func (e *PlainEncoder) Name() string                       { return "plain" }
func (e *PlainEncoder) Encode(data []byte) ([]byte, error) { return data, nil }
func (e *PlainEncoder) Decode(data []byte) ([]byte, error) { return data, nil }

// Base64Encoder — 标准 Base64 编码。ID=1。
type Base64Encoder struct{}

func (e *Base64Encoder) ID() int      { return 1 }
func (e *Base64Encoder) Name() string { return "base64" }
func (e *Base64Encoder) Encode(data []byte) ([]byte, error) {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	return encoded, nil
}
func (e *Base64Encoder) Decode(data []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return decoded[:n], nil
}

// encoderRegistry 持有所有可用的编码器实例，按 ID 索引。
// 在 init() 中注册，确保 Implant 和 Server 都有一致的编码器集合。
var encoderRegistry = map[int]Encoder{}

func RegisterEncoder(e Encoder) {
	encoderRegistry[e.ID()] = e
}

// GetEncoderByID 根据 ID 查找编码器。ID 来自 nonce % EncoderModulus。
func GetEncoderByID(id int) (Encoder, bool) {
	enc, ok := encoderRegistry[id]
	return enc, ok
}

// MustGetEncoder 根据 ID 获取编码器，找不到则 panic（仅用于测试/初始化）。
func MustGetEncoder(id int) Encoder {
	enc, ok := GetEncoderByID(id)
	if !ok {
		panic(fmt.Sprintf("encoder %d not registered", id))
	}
	return enc
}

func init() {
	RegisterEncoder(&PlainEncoder{})
	RegisterEncoder(&Base64Encoder{})
}
