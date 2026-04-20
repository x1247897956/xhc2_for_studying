package identity

import (
	"os"
	"os/user"
	"runtime"
)

type HostInfo struct {
	Hostname string
	Username string
	OS       string
	Arch     string
}

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
