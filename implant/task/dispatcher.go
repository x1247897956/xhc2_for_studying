// Package task dispatches incoming tasks from the C2 server to the
// appropriate handler and returns formatted results.
package task

import (
	"fmt"
	"time"

	"xhc2_for_studying/protocol"
)

// Dispatch routes a task to the handler matching its type and returns the
// result. Unknown task types are reported as failed.
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
