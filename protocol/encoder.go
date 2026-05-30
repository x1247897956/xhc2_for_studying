package protocol

import (
	"encoding/base64"
	"fmt"
	"math/rand/v2"
)

// Encoder defines the interface for encoding and decoding request/response bodies.
// Both the implant and server use this interface to process C2 traffic.
type Encoder interface {
	// ID returns the encoder's unique identifier.
	// The result of nonce % EncoderModulus must equal this ID.
	ID() int
	// Name returns a human-readable name for the encoder.
	Name() string
	// Encode encodes raw bytes into the transport format.
	Encode(data []byte) ([]byte, error)
	// Decode decodes bytes from the transport format back to raw bytes.
	Decode(data []byte) ([]byte, error)
}

// PlainEncoder passes data through without any encoding.
// ID=0, selected when nonce % EncoderModulus == 0.
type PlainEncoder struct{}

func (e *PlainEncoder) ID() int                            { return 0 }
func (e *PlainEncoder) Name() string                       { return "plain" }
func (e *PlainEncoder) Encode(data []byte) ([]byte, error) { return data, nil }
func (e *PlainEncoder) Decode(data []byte) ([]byte, error) { return data, nil }

// Base64Encoder applies standard Base64 encoding. ID=1.
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

// encoderRegistry holds all registered encoder instances, indexed by ID.
// It is populated in init() to ensure both implant and server share the same encoder set.
var encoderRegistry = map[int]Encoder{}

// RegisterEncoder adds an encoder to the global registry.
func RegisterEncoder(e Encoder) {
	encoderRegistry[e.ID()] = e
}

// GetEncoderByID looks up an encoder by its ID. The ID is derived from nonce % EncoderModulus.
func GetEncoderByID(id int) (Encoder, bool) {
	enc, ok := encoderRegistry[id]
	return enc, ok
}

// RandomEncoder returns a randomly selected encoder from the registry.
func RandomEncoder() Encoder {
	ids := make([]int, 0, len(encoderRegistry))
	for id := range encoderRegistry {
		ids = append(ids, id)
	}
	return encoderRegistry[ids[rand.IntN(len(ids))]]
}

// MustGetEncoder returns the encoder for the given ID, panicking if none is registered.
// Intended for tests and initialization only.
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
