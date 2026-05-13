package client

import (
	"encoding/base64"

	"xhc2_for_studying/protocol"
)

// encryptAndEncode 加密明文并 Base64 编码。
// 返回: encodedBody, nonceBase64（用于放 URL）, error。
func (c *Client) encryptAndEncode(plaintext []byte) (encodedBody []byte, nonceB64 string, err error) {
	packet, nonceB64, err := c.cipherCtx.Encrypt(plaintext)
	if err != nil {
		return nil, "", err
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(packet)))
	base64.StdEncoding.Encode(encoded, packet)
	return encoded, nonceB64, nil
}

// decodeAndDecrypt 解码 Base64 响应并解密。
func (c *Client) decodeAndDecrypt(data []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return nil, err
	}
	return c.cipherCtx.Decrypt(decoded[:n])
}

// encryptAndEncodeBody 等价于 encryptAndEncode，供包内使用。
func (c *Client) encryptAndEncodeBody(plaintext []byte) (body []byte, nonceB64 string, err error) {
	return c.encryptAndEncode(plaintext)
}

// ensure 编译期检查 Encoder 注册。
var _ = protocol.NonceSize // 确保 protocol 包被正确引用
