// Package identity collects host information used for beacon registration.
package identity

import (
	"os"
	"os/user"
	"runtime"
)

// HostInfo contains basic host identification data sent to the C2 server
// during beacon registration.
type HostInfo struct {
	Hostname string
	Username string
	OS       string
	Arch     string
}

// CollectHostInfo gathers the hostname, username, OS, and architecture
// of the current machine.
func CollectHostInfo() (*HostInfo, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	userName := currentUser.Username
	osName := runtime.GOOS
	archName := runtime.GOARCH

	return &HostInfo{
		Hostname: hostName,
		Username: userName,
		OS:       osName,
		Arch:     archName,
	}, nil
}
