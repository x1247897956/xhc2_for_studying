// Package core defines the core domain types shared across the server.
package core

// Beacon represents a registered implant with its host metadata and check-in state.
type Beacon struct {
	ID            string
	Hostname      string
	Username      string
	OS            string
	Arch          string
	Interval      int64
	Jitter        int64
	LastCheckIn   int64
	RemoteAddress string
}
