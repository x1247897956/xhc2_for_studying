package store

import (
	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"
)

// BeaconStore persists beacon metadata and check-in state.
type BeaconStore interface {
	Add(beacon *core.Beacon) error
	Get(id string) (*core.Beacon, error)
	UpdateCheckIn(id string, lastCheckIn int64, remoteAddress string) error
	ListIDs() []string
}

// ServerTaskStore persists operator tasks and implant task results.
type ServerTaskStore interface {
	AddTask(task *core.ServerTask) error
	GetTask(taskID string) (*core.ServerTask, error)
	GetPendingTasksByImplantID(implantID string) []*core.ServerTask
	UpdateTask(taskID string, taskRes protocol.TaskResult) error
}

// ImplantStore persists generated implant metadata needed for key exchange.
type ImplantStore interface {
	Set(pubKeyDigest string, record *ImplantRecord) error
	Get(pubKeyDigest string) (*ImplantRecord, bool)
}
