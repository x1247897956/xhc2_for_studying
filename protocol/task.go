package protocol

import (
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

type TaskResult struct {
	TaskID    string    `json:"task_id"`
	ImplantID string    `json:"implant_id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Completed time.Time `json:"completed"`
	Output    string    `json:"output,omitempty"`
}
