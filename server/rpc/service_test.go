package rpc

import (
	"bytes"
	"testing"

	"xhc2_for_studying/protocol"
	serverConfig "xhc2_for_studying/server/config"
	"xhc2_for_studying/server/store"
)

func TestGenerateImplantUsesDefaultsAndReturnsBinary(t *testing.T) {
	implantStore := store.NewImplantStore()
	cfg := &serverConfig.ServerConfig{
		AgePublicKey: "age1server",
		C2Profile:    testC2Profile(),
		GenerateDefaults: serverConfig.GenerateDefaults{
			ServerURL:  "http://127.0.0.1:8024",
			PathPrefix: "/cdn/assets",
			Interval:   5,
			Jitter:     1,
			GOOS:       "linux",
			GOARCH:     "amd64",
		},
	}

	var captured GenerateImplantBuildRequest
	svc := &C2RPC{
		ImplantStore: implantStore,
		Config:       cfg,
		BuildImplant: func(req GenerateImplantBuildRequest) (*GenerateImplantBuildResult, error) {
			captured = req
			return &GenerateImplantBuildResult{
				Digest: "digest-1",
				Binary: []byte("implant-binary"),
			}, nil
		},
	}

	var reply GenerateImplantResponse
	err := svc.GenerateImplant(GenerateImplantRequest{}, &reply)
	if err != nil {
		t.Fatalf("GenerateImplant returned error: %v", err)
	}

	if captured.ServerURL != cfg.GenerateDefaults.ServerURL {
		t.Fatalf("ServerURL = %q, want %q", captured.ServerURL, cfg.GenerateDefaults.ServerURL)
	}
	if captured.PathPrefix != cfg.GenerateDefaults.PathPrefix {
		t.Fatalf("PathPrefix = %q, want %q", captured.PathPrefix, cfg.GenerateDefaults.PathPrefix)
	}
	if captured.Interval != cfg.GenerateDefaults.Interval {
		t.Fatalf("Interval = %d, want %d", captured.Interval, cfg.GenerateDefaults.Interval)
	}
	if captured.Jitter != cfg.GenerateDefaults.Jitter {
		t.Fatalf("Jitter = %d, want %d", captured.Jitter, cfg.GenerateDefaults.Jitter)
	}
	if captured.GOOS != cfg.GenerateDefaults.GOOS || captured.GOARCH != cfg.GenerateDefaults.GOARCH {
		t.Fatalf("target = %s/%s, want %s/%s", captured.GOOS, captured.GOARCH, cfg.GenerateDefaults.GOOS, cfg.GenerateDefaults.GOARCH)
	}
	if captured.ServerPublicKey != cfg.AgePublicKey {
		t.Fatalf("ServerPublicKey = %q, want %q", captured.ServerPublicKey, cfg.AgePublicKey)
	}
	if captured.C2Profile != cfg.C2Profile {
		t.Fatal("C2Profile was not passed through from server config")
	}
	if captured.ImplantStore != implantStore {
		t.Fatal("ImplantStore was not passed through from service state")
	}
	if reply.Digest != "digest-1" {
		t.Fatalf("Digest = %q, want digest-1", reply.Digest)
	}
	if !bytes.Equal(reply.Binary, []byte("implant-binary")) {
		t.Fatalf("Binary = %q, want implant-binary", string(reply.Binary))
	}
	if reply.Filename != "implant-linux-amd64" {
		t.Fatalf("Filename = %q, want implant-linux-amd64", reply.Filename)
	}
}

func TestGenerateImplantRequestOverridesDefaults(t *testing.T) {
	cfg := &serverConfig.ServerConfig{
		AgePublicKey: "age1server",
		C2Profile:    testC2Profile(),
		GenerateDefaults: serverConfig.GenerateDefaults{
			ServerURL:  "http://127.0.0.1:8024",
			PathPrefix: "/cdn/assets",
			Interval:   5,
			Jitter:     1,
			GOOS:       "linux",
			GOARCH:     "amd64",
		},
	}

	var captured GenerateImplantBuildRequest
	svc := &C2RPC{
		ImplantStore: store.NewImplantStore(),
		Config:       cfg,
		BuildImplant: func(req GenerateImplantBuildRequest) (*GenerateImplantBuildResult, error) {
			captured = req
			return &GenerateImplantBuildResult{
				Digest: "digest-2",
				Binary: []byte("windows-binary"),
			}, nil
		},
	}

	req := GenerateImplantRequest{
		ServerURL:  "http://10.0.0.1:8024",
		PathPrefix: "/api",
		Interval:   10,
		Jitter:     int64Ptr(3),
		GOOS:       "windows",
		GOARCH:     "arm64",
	}

	var reply GenerateImplantResponse
	if err := svc.GenerateImplant(req, &reply); err != nil {
		t.Fatalf("GenerateImplant returned error: %v", err)
	}

	if captured.ServerURL != req.ServerURL {
		t.Fatalf("ServerURL = %q, want %q", captured.ServerURL, req.ServerURL)
	}
	if captured.PathPrefix != req.PathPrefix {
		t.Fatalf("PathPrefix = %q, want %q", captured.PathPrefix, req.PathPrefix)
	}
	if captured.Interval != req.Interval {
		t.Fatalf("Interval = %d, want %d", captured.Interval, req.Interval)
	}
	if captured.Jitter != *req.Jitter {
		t.Fatalf("Jitter = %d, want %d", captured.Jitter, *req.Jitter)
	}
	if captured.GOOS != req.GOOS || captured.GOARCH != req.GOARCH {
		t.Fatalf("target = %s/%s, want %s/%s", captured.GOOS, captured.GOARCH, req.GOOS, req.GOARCH)
	}
	if reply.Filename != "implant-windows-arm64.exe" {
		t.Fatalf("Filename = %q, want implant-windows-arm64.exe", reply.Filename)
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}

func testC2Profile() *protocol.C2Profile {
	return &protocol.C2Profile{
		PathSegments: []protocol.HTTPC2PathSegment{
			{Value: "api", IsFile: false},
			{Value: "app", IsFile: true},
		},
		KeyExchangeExtensions: []string{".html"},
		RegisterExtensions:    []string{".php"},
		PollExtensions:        []string{".js"},
		ResultExtensions:      []string{".json"},
		MinPathLength:         1,
		MaxPathLength:         1,
		NonceMode:             protocol.NonceModeURLParam,
		EncoderModulus:        256,
	}
}
