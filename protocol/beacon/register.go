// Package beacon defines the request and response structures for beacon
// registration and check-in.
package beacon

// RegisterRequest carries the host metadata an implant sends during its
// initial registration with the C2 server.
type RegisterRequest struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Interval int64  `json:"interval"`
	Jitter   int64  `json:"jitter"`
}

// RegisterResponse returns the logical beacon identifier assigned by the
// server after a successful registration.
type RegisterResponse struct {
	BeaconID string `json:"beacon_id"`
}
