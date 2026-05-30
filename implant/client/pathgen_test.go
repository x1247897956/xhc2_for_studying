package client

import (
	"strings"
	"testing"

	"xhc2_for_studying/protocol"
)

func TestBuildRandomURLPrependsPathPrefix(t *testing.T) {
	profile := &protocol.C2Profile{
		PathSegments: []protocol.HTTPC2PathSegment{
			{Value: "api", IsFile: false},
			{Value: "chunk", IsFile: true},
		},
		MinPathLength: 1,
		MaxPathLength: 1,
	}

	u, err := buildRandomURL("http://127.0.0.1:8024", "/cdn/assets", profile, ".js")
	if err != nil {
		t.Fatalf("buildRandomURL returned error: %v", err)
	}

	if !strings.HasPrefix(u.Path, "/cdn/assets/") {
		t.Fatalf("path %q does not include fixed prefix", u.Path)
	}
	if !strings.HasSuffix(u.Path, "/chunk.js") {
		t.Fatalf("path %q does not include randomized file segment", u.Path)
	}
}

func TestRandomPathDoesNotRepeatDirectorySegments(t *testing.T) {
	profile := &protocol.C2Profile{
		PathSegments: []protocol.HTTPC2PathSegment{
			{Value: "client", IsFile: false},
			{Value: "session", IsFile: false},
			{Value: "sync", IsFile: false},
			{Value: "collect", IsFile: true},
		},
		MinPathLength: 3,
		MaxPathLength: 3,
	}

	for range 100 {
		segments := randomPath(profile, ".php")
		seen := map[string]bool{}
		for _, segment := range segments[:len(segments)-1] {
			if seen[segment] {
				t.Fatalf("directory segment %q repeated in path %#v", segment, segments)
			}
			seen[segment] = true
		}
	}
}
