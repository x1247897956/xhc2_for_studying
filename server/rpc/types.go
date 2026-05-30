package rpc

// Empty is used as a placeholder value for RPC methods that take no arguments.
type Empty struct{}

// CreateTaskRequest is sent by the client to create a new task for a specific
// implant.
type CreateTaskRequest struct {
	TaskType  string
	ImplantID string
	Payload   string
}

// CreateTaskResponse contains the server-assigned task ID after creation.
type CreateTaskResponse struct {
	TaskID string
}

// GenerateImplantRequest carries per-build options that override the server's
// implant generation defaults when set.
type GenerateImplantRequest struct {
	ServerURL  string
	PathPrefix string
	Interval   int64
	Jitter     *int64
	GOOS       string
	GOARCH     string
}

// GenerateImplantResponse returns the compiled implant and metadata needed by
// the RPC client to save it locally.
type GenerateImplantResponse struct {
	Digest   string
	Filename string
	Binary   []byte
}

// TaskResultRequest queries a single task by its ID to retrieve the current
// result.
type TaskResultRequest struct {
	TaskID string
}

// TaskResultResponse holds the full result of a task, including output,
// status, and completion time.
type TaskResultResponse struct {
	TaskID    string
	Type      string
	ImplantID string
	Status    string
	Error     string
	Output    string
	Completed string
}

// BeaconInfo is returned by ListBeacons with a snapshot of a registered
// beacon's metadata.
type BeaconInfo struct {
	ID          string
	Hostname    string
	Username    string
	OS          string
	Arch        string
	Interval    int64
	Jitter      int64
	LastCheckIn int64
}

// TaskInfo is a summary row returned by ListTasks for each pending task.
type TaskInfo struct {
	TaskID    string
	ImplantID string
	Type      string
	Status    string
}
