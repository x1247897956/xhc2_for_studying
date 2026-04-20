package core

import (
	"time"
	
	"xhc2_for_studying/protocol"
)

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
