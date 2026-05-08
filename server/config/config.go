package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	
	"xhc2_for_studying/protocol"
)

//go:embed c2profile.json
var c2ProfileBytes []byte

// ServerConfig 是 Server 端的运行时配置。
type ServerConfig struct {
	ListenAddr string              `json:"listen_addr"`
	C2Profile  *protocol.C2Profile `json:"c2_profile"`
}

// Load 加载 Server 配置。
// C2Profile 从嵌入的 c2profile.json 读取，其他字段使用默认值。
func Load() (*ServerConfig, error) {
	var c2Profile protocol.C2Profile
	if err := json.Unmarshal(c2ProfileBytes, &c2Profile); err != nil {
		return nil, err
	}
	
	cfg := &ServerConfig{
		ListenAddr: ":8024",
		C2Profile:  &c2Profile,
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ServerConfig) Validate() error {
	if c.C2Profile == nil {
		return errors.New("c2_profile is required")
	}
	if len(c.C2Profile.Extensions) == 0 {
		return errors.New("c2_profile.extensions is required")
	}
	if c.C2Profile.MinPathLength < 1 {
		return errors.New("c2_profile.min_path_length must be >= 1")
	}
	if c.C2Profile.MaxPathLength < c.C2Profile.MinPathLength {
		return errors.New("c2_profile.max_path_length must be >= min_path_length")
	}
	if c.C2Profile.EncoderModulus <= 0 {
		return errors.New("c2_profile.encoder_modulus must be > 0")
	}
	if c.C2Profile.NonceMode != protocol.NonceModeURLParam &&
		c.C2Profile.NonceMode != protocol.NonceModeURL {
		return errors.New("c2_profile.nonce_mode must be 'urlparam' or 'url'")
	}
	return nil
}
