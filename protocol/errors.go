package protocol

// ErrorResponse is a generic JSON error envelope returned by the server
// when a request cannot be processed.
type ErrorResponse struct {
	Error string `json:"error"`
}
