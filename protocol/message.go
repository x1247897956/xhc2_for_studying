package protocol

type Message struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Payload        string `json:"payload"`
	UnknownPayload bool   `json:"unknownPayload"`
}
