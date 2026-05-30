package protocol

// Message is a generic envelope used for bidirectional communication
// between the operator console and the server.
type Message struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Payload        string `json:"payload"`
	UnknownPayload bool   `json:"unknownPayload"`
}
