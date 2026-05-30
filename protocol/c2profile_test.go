package protocol

import "testing"

func TestExtensionMapResolvesPollAndResult(t *testing.T) {
	extMap := ExtensionMap{
		KeyExchange: ".html",
		Register:    ".json",
		Poll:        ".js",
		Result:      ".php",
	}

	tests := map[string]string{
		".html": MsgKeyExchange,
		".json": MsgRegister,
		".js":   MsgPoll,
		".php":  MsgResult,
		".txt":  "",
	}

	for ext, want := range tests {
		if got := extMap.ExtToMsgType(ext); got != want {
			t.Fatalf("ExtToMsgType(%q) = %q, want %q", ext, got, want)
		}
	}
}
