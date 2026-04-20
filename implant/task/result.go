package task

import (
	"time"
	
	"xhc2_for_studying/protocol"
)

func newSuccessResult(task protocol.Task, implantID, output string) *protocol.TaskResult {
	return &protocol.TaskResult{
		TaskID:    task.ID,
		ImplantID: implantID,
		Status:    protocol.TaskStatusDone,
		Completed: time.Now(),
		Output:    output,
	}
}

func newFailedResult(task protocol.Task, implantID, errMsg string) *protocol.TaskResult {
	return &protocol.TaskResult{
		TaskID:    task.ID,
		ImplantID: implantID,
		Status:    protocol.TaskStatusFailed,
		Error:     errMsg,
		Completed: time.Now(),
	}
}
