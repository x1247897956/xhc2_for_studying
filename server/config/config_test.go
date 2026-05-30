package config

import "testing"

func TestDecodeStaticConfigLoadsGenerateDefaultsAndC2Profile(t *testing.T) {
	raw := []byte(`{
	  "listen_addr": ":8024",
	  "generate_defaults": {
	    "server_url": "http://127.0.0.1:8024",
	    "path_prefix": "/api/v1",
	    "interval": 5,
	    "jitter": 0,
	    "goos": "linux",
	    "goarch": "amd64"
	  },
	  "c2_profile": {
	    "path_segments": [
	      {"value": "api", "is_file": false},
	      {"value": "sync", "is_file": true}
	    ],
	    "keyexchange_extensions": [".html"],
	    "register_extensions": [".json"],
	    "poll_extensions": [".js"],
	    "result_extensions": [".php"],
	    "user_agent": "xhc2-test-agent",
	    "session_cookie_name": "sid",
	    "min_path_length": 1,
	    "max_path_length": 1,
	    "nonce_mode": "urlparam",
	    "encoder_modulus": 256
	  }
	}`)

	cfg, err := decodeStaticConfig(raw)
	if err != nil {
		t.Fatalf("decodeStaticConfig returned error: %v", err)
	}
	if cfg.GenerateDefaults.PathPrefix != "/api/v1" {
		t.Fatalf("PathPrefix = %q, want /api/v1", cfg.GenerateDefaults.PathPrefix)
	}
	if len(cfg.C2Profile.PollExtensions) != 1 || cfg.C2Profile.PollExtensions[0] != ".js" {
		t.Fatalf("PollExtensions = %#v, want [.js]", cfg.C2Profile.PollExtensions)
	}
	if len(cfg.C2Profile.ResultExtensions) != 1 || cfg.C2Profile.ResultExtensions[0] != ".php" {
		t.Fatalf("ResultExtensions = %#v, want [.php]", cfg.C2Profile.ResultExtensions)
	}
	if cfg.C2Profile.UserAgent != "xhc2-test-agent" {
		t.Fatalf("UserAgent = %q, want xhc2-test-agent", cfg.C2Profile.UserAgent)
	}
	if cfg.C2Profile.SessionCookieName != "sid" {
		t.Fatalf("SessionCookieName = %q, want sid", cfg.C2Profile.SessionCookieName)
	}
	if cfg.Database.Driver != "memory" {
		t.Fatalf("Database.Driver = %q, want memory", cfg.Database.Driver)
	}
}

func TestDecodeStaticConfigUsesMySQLDSNFromEnvironment(t *testing.T) {
	t.Setenv("C2_MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/xhc2?parseTime=true")

	cfg, err := decodeStaticConfig([]byte(`{
	  "listen_addr": ":8024",
	  "generate_defaults": {
	    "server_url": "http://127.0.0.1:8024",
	    "interval": 5,
	    "jitter": 0,
	    "goos": "linux",
	    "goarch": "amd64"
	  },
	  "c2_profile": {
	    "path_segments": [{"value": "api", "is_file": false}],
	    "keyexchange_extensions": [".html"],
	    "register_extensions": [".json"],
	    "poll_extensions": [".js"],
	    "result_extensions": [".php"],
	    "min_path_length": 1,
	    "max_path_length": 1
	  }
	}`))
	if err != nil {
		t.Fatalf("decodeStaticConfig returned error: %v", err)
	}
	if cfg.Database.Driver != "mysql" {
		t.Fatalf("Database.Driver = %q, want mysql", cfg.Database.Driver)
	}
	if cfg.Database.DSN != "user:pass@tcp(127.0.0.1:3306)/xhc2?parseTime=true" {
		t.Fatalf("Database.DSN = %q", cfg.Database.DSN)
	}
}
