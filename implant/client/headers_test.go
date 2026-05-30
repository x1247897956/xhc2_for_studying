package client

import (
	"net/http"
	"testing"

	"xhc2_for_studying/protocol"
)

func TestApplyRequestHeadersUsesConfiguredUserAgentAndCookie(t *testing.T) {
	c := &Client{
		c2Profile: &protocol.C2Profile{
			UserAgent:         "xhc2-test-agent",
			SessionCookieName: "sid",
		},
		sessionToken: "session-token",
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	c.applyRequestHeaders(req, true)

	if got := req.Header.Get("User-Agent"); got != "xhc2-test-agent" {
		t.Fatalf("User-Agent = %q, want xhc2-test-agent", got)
	}
	if got := req.Header.Get("X-Session-Token"); got != "" {
		t.Fatalf("X-Session-Token = %q, want empty", got)
	}
	cookie, err := req.Cookie("sid")
	if err != nil {
		t.Fatalf("Cookie sid missing: %v", err)
	}
	if cookie.Value != "session-token" {
		t.Fatalf("cookie value = %q, want session-token", cookie.Value)
	}
}
