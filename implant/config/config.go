package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
)

const localConfigPath = "implant/config/implant.json"

//go:embed implant.example.json
var embeddedExampleConfig []byte

type BeaconConfig struct {
	Interval  int64  `json:"interval"`
	Jitter    int64  `json:"jitter"`
	ServerURL string `json:"server_url"`
}

func Load() (*BeaconConfig, error) {
	configBytes := embeddedExampleConfig
	if localConfig, err := os.ReadFile(localConfigPath); err == nil {
		configBytes = localConfig
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var cfg BeaconConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *BeaconConfig) Validate() error {
	if c == nil {
		return errors.New("beacon config is nil")
	}
	if c.ServerURL == "" {
		return errors.New("server_url is required")
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than zero")
	}
	if c.Jitter < 0 {
		return errors.New("jitter must be greater than or equal to zero")
	}
	return nil
}
