// Package config loads and validates the server's runtime configuration,
// including the C2 profile, implant generation defaults, and Age key pair.
package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/cryptography"
)

//go:embed server.json
var serverConfigBytes []byte

// GenerateDefaults holds the base implant generation options that RPC callers
// may override per request.
type GenerateDefaults struct {
	ServerURL  string `json:"server_url"`
	PathPrefix string `json:"path_prefix"`
	Interval   int64  `json:"interval"`
	Jitter     int64  `json:"jitter"`
	GOOS       string `json:"goos"`
	GOARCH     string `json:"goarch"`
}

// DatabaseConfig selects the server persistence backend.
type DatabaseConfig struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// ServerConfig holds the server's runtime configuration, including the
// listen address, Age key pair, C2 profile, and generation defaults.
type ServerConfig struct {
	ListenAddr       string              `json:"listen_addr"`
	AgePublicKey     string              // Age public key, populated at startup.
	AgePrivateKey    string              // Age private key, populated at startup.
	C2Profile        *protocol.C2Profile `json:"c2_profile"`
	GenerateDefaults GenerateDefaults    `json:"generate_defaults"`
	Database         DatabaseConfig      `json:"database"`
}

// Load reads the embedded C2 profile and generation defaults, ensures an Age
// key pair exists, and returns a validated ServerConfig.
func Load() (*ServerConfig, error) {
	cfg, err := decodeStaticConfig(serverConfigBytes)
	if err != nil {
		return nil, err
	}

	// Ensure an Age key pair exists (generates one if needed).
	pubKey, priKey, err := cryptography.EnsureAgeKeyPair()
	if err != nil {
		return nil, err
	}

	cfg.AgePublicKey = pubKey
	cfg.AgePrivateKey = priKey

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func decodeStaticConfig(data []byte) (*ServerConfig, error) {
	var raw struct {
		ListenAddr       string              `json:"listen_addr"`
		C2Profile        *protocol.C2Profile `json:"c2_profile"`
		GenerateDefaults GenerateDefaults    `json:"generate_defaults"`
		Database         DatabaseConfig      `json:"database"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	cfg := &ServerConfig{
		ListenAddr:       raw.ListenAddr,
		C2Profile:        raw.C2Profile,
		GenerateDefaults: raw.GenerateDefaults,
		Database:         raw.Database,
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8024"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "memory"
	}
	if dsn := os.Getenv("C2_MYSQL_DSN"); dsn != "" {
		cfg.Database.Driver = "mysql"
		cfg.Database.DSN = dsn
	}
	return cfg, nil
}

// Validate checks that the C2 profile has all required extension lists
// and valid path length constraints.
func (c *ServerConfig) Validate() error {
	if c.C2Profile == nil {
		return errors.New("c2_profile is required")
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
	if c.C2Profile.MinPathLength < 1 {
		return errors.New("c2_profile.min_path_length must be >= 1")
	}
	if c.C2Profile.MaxPathLength < c.C2Profile.MinPathLength {
		return errors.New("c2_profile.max_path_length must be >= min_path_length")
	}
	if c.GenerateDefaults.ServerURL == "" {
		return errors.New("generate_defaults.server_url is required")
	}
	if c.GenerateDefaults.Interval <= 0 {
		return errors.New("generate_defaults.interval must be greater than zero")
	}
	if c.GenerateDefaults.Jitter < 0 {
		return errors.New("generate_defaults.jitter must be greater than or equal to zero")
	}
	if c.GenerateDefaults.GOOS == "" {
		return errors.New("generate_defaults.goos is required")
	}
	if c.GenerateDefaults.GOARCH == "" {
		return errors.New("generate_defaults.goarch is required")
	}
	switch c.Database.Driver {
	case "memory":
	case "mysql":
		if c.Database.DSN == "" {
			return errors.New("database.dsn is required for mysql")
		}
	default:
		return errors.New("database.driver must be memory or mysql")
	}
	return nil
}
