package protocol

import (
	"time"
)

// Task represents a single command or action that the server wants an
// implant to execute.
type Task struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

// TaskResult carries the outcome of a completed (or failed) task back to
// the server.
type TaskResult struct {
	TaskID    string    `json:"task_id"`
	ImplantID string    `json:"implant_id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Completed time.Time `json:"completed"`
	Output    string    `json:"output,omitempty"`
}
