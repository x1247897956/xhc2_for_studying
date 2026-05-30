package client

import (
	"xhc2_for_studying/protocol"
)

// encryptAndEncode encrypts the plaintext, picks a random encoder, and returns
// the encoded body together with an encoder negotiation nonce. The server
// derives the encoder ID from the nonce via nonce % EncoderModulus.
func (c *Client) encryptAndEncode(plaintext []byte) (encodedBody []byte, encoderNonce int, err error) {
	packet, _, err := c.cipherCtx.Encrypt(plaintext)
	if err != nil {
		return nil, 0, err
	}

	enc := protocol.RandomEncoder()
	encoded, err := enc.Encode(packet)
	if err != nil {
		return nil, 0, err
	}

	nonce := protocol.GenerateNonce(enc.ID(), c.c2Profile.EncoderModulus)
	return encoded, nonce, nil
}
