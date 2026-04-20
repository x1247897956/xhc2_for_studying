package client

import (
	"errors"
	"net/http"
	"strings"
	"time"
	
	"xhc2_for_studying/implant/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
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
	}, nil
}
