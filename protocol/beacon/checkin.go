package beacon

import "xhc2_for_studying/protocol"

type CheckinRequest struct {
	BeaconID   string              `json:"beacon_id"`
	TaskResult protocol.TaskResult `json:"task_result"`
}

type CheckinResponse struct {
	Tasks []protocol.Task `json:"tasks"`
}
