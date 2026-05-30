// Package config loads and validates the beacon configuration embedded in the
// implant binary.
package config

import (
	_ "embed"
	"encoding/json"
	"errors"

	"xhc2_for_studying/protocol"
)

//go:embed implant.json
var embeddedConfig []byte

// BeaconConfig holds all configuration parameters for the beacon implant.
// It is deserialized from the embedded implant.json.
type BeaconConfig struct {
	ServerURL            string                `json:"server_url"`
	PathPrefix           string                `json:"path_prefix"`
	ImplantAgePublicKey  string                `json:"implant_age_public_key"`
	ImplantAgePrivateKey string                `json:"implant_age_private_key"`
	Interval             int64                 `json:"interval"`
	Jitter               int64                 `json:"jitter"`
	ServerPublicKey      string                `json:"server_public_key"`
	C2Profile            protocol.C2Profile    `json:"c2_profile"`
	ExtMap               protocol.ExtensionMap `json:"ext_map"`
}

// Load reads and validates the embedded beacon configuration.
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

// Validate checks that all required fields in BeaconConfig are set to
// valid values. It returns an error describing the first invalid field.
func (c *BeaconConfig) Validate() error {
	if c == nil {
		return errors.New("beacon config is nil")
	}
	if c.ServerURL == "" {
		return errors.New("server_url is required")
	}
	if c.ImplantAgePublicKey == "" {
		return errors.New("implant_age_public_key is required")
	}
	if c.ImplantAgePrivateKey == "" {
		return errors.New("implant_age_private_key is required")
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
	if len(c.C2Profile.KeyExchangeExtensions) == 0 {
		return errors.New("c2_profile.keyexchange_extensions is required")
	}
	if len(c.C2Profile.RegisterExtensions) == 0 {
		return errors.New("c2_profile.register_extensions is required")
	}
	if len(c.C2Profile.PollExtensions) == 0 {
		return errors.New("c2_profile.poll_extensions is required")
	}
	if len(c.C2Profile.ResultExtensions) == 0 {
		return errors.New("c2_profile.result_extensions is required")
	}
	if c.C2Profile.EncoderModulus <= 0 {
		return errors.New("c2_profile.encoder_modulus must be > 0")
	}
	return nil
}
