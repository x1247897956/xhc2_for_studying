package core

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
