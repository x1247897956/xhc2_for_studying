package c2

import (
	"encoding/json"
	"net/http"
	"testing"

	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/store"
)

func TestRegisterBindsBeaconIDToSession(t *testing.T) {
	srv := &HTTPServer{
		beaconStore:  store.NewBeaconStore(),
		sessionStore: store.NewSessionStore(),
	}
	session := &store.Session{}
	body, err := json.Marshal(beaconProtocol.RegisterRequest{
		Hostname: "host",
		Username: "user",
		OS:       "linux",
		Arch:     "amd64",
		Interval: 5,
	})
	if err != nil {
		t.Fatalf("marshal register request: %v", err)
	}

	respBody := srv.handleRegisterEncrypted(body, "127.0.0.1", session)
	var resp beaconProtocol.RegisterResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal register response: %v", err)
	}

	if resp.BeaconID == "" {
		t.Fatal("BeaconID is empty")
	}
	if session.BeaconID != resp.BeaconID {
		t.Fatalf("session BeaconID = %q, want %q", session.BeaconID, resp.BeaconID)
	}
}

func TestSessionTokenFromRequestUsesConfiguredCookie(t *testing.T) {
	srv := &HTTPServer{
		c2Profile: &protocol.C2Profile{SessionCookieName: "sid"},
	}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "sid", Value: "session-token"})
	req.Header.Set("X-Session-Token", "legacy-token")

	if got := srv.sessionTokenFromRequest(req); got != "session-token" {
		t.Fatalf("session token = %q, want session-token", got)
	}
}

func TestPollUsesBeaconIDFromSession(t *testing.T) {
	taskStore := store.NewServerTaskStore()
	srv := &HTTPServer{
		beaconStore: store.NewBeaconStore(),
		taskStore:   taskStore,
	}
	srv.beaconStore.Add(&core.Beacon{ID: "beacon-1"})
	taskStore.AddTask(&core.ServerTask{
		TaskID:    "task-1",
		ImplantID: "beacon-1",
		Type:      protocol.TaskTypeWhoami,
		Status:    protocol.TaskStatusPending,
	})

	respBody := srv.handlePollEncrypted("127.0.0.1", &store.Session{BeaconID: "beacon-1"})
	var resp beaconProtocol.PollResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal poll response: %v", err)
	}
	if len(resp.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want 1", len(resp.Tasks))
	}
}
