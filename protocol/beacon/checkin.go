// Package beacon defines the request and response structures for beacon
// check-in operations.
package beacon

import "xhc2_for_studying/protocol"

// PollRequest is sent by an implant to retrieve pending tasks.
type PollRequest struct {
	BeaconID string `json:"beacon_id"`
}

// PollResponse carries the list of pending tasks assigned to the implant.
type PollResponse struct {
	Tasks []protocol.Task `json:"tasks"`
}

// ResultRequest is sent by an implant to report the outcome of a task.
type ResultRequest struct {
	BeaconID   string              `json:"beacon_id"`
	TaskResult protocol.TaskResult `json:"task_result"`
}

// ResultResponse acknowledges that a task result was processed.
type ResultResponse struct {
	OK bool `json:"ok"`
}
