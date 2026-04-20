package beacon

// RegisterRequest 是 Beacon 首次上线接入 server 时发送的注册信息。
type RegisterRequest struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Interval int64  `json:"interval"`
	Jitter   int64  `json:"jitter"`
}

// RegisterResponse 返回 server 为该 Beacon 分配的逻辑标识。
type RegisterResponse struct {
	BeaconID string `json:"beacon_id"`
}
