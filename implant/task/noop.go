package task

import (
	"xhc2_for_studying/protocol"
)

func handleNoop(task protocol.Task, implantID string) *protocol.TaskResult {
	return newSuccessResult(task, implantID, "")
}
