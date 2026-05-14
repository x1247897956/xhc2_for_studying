package client

import (
	"encoding/base64"
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
