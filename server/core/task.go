package core

import (
	"time"

	"xhc2_for_studying/protocol"
)

// ServerTask represents a task issued by the operator, including its lifecycle
// status, result payload, and timestamps.
type ServerTask struct {
	TaskID      string
	Type        string
	ImplantID   string
	Payload     string
	Status      string
	Result      protocol.TaskResult
	CreatedAt   time.Time
	CompletedAt time.Time
	Error       string
}
