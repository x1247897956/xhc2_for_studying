package handlers

import (
	"testing"

	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/store"
)

func TestHandlePollReturnsPendingTasks(t *testing.T) {
	beaconStore := store.NewBeaconStore()
	taskStore := store.NewServerTaskStore()
	beaconStore.Add(&core.Beacon{ID: "beacon-1"})
	taskStore.AddTask(&core.ServerTask{
		TaskID:    "task-1",
		ImplantID: "beacon-1",
		Type:      protocol.TaskTypeWhoami,
		Status:    protocol.TaskStatusPending,
	})

	resp, err := HandlePoll(beaconStore, taskStore, &beaconProtocol.PollRequest{BeaconID: "beacon-1"}, "127.0.0.1")
	if err != nil {
		t.Fatalf("HandlePoll returned error: %v", err)
	}
	if len(resp.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want 1", len(resp.Tasks))
	}
	if resp.Tasks[0].ID != "task-1" {
		t.Fatalf("task ID = %q, want task-1", resp.Tasks[0].ID)
	}
}

func TestHandleResultUpdatesTask(t *testing.T) {
	beaconStore := store.NewBeaconStore()
	taskStore := store.NewServerTaskStore()
	beaconStore.Add(&core.Beacon{ID: "beacon-1"})
	taskStore.AddTask(&core.ServerTask{
		TaskID:    "task-1",
		ImplantID: "beacon-1",
		Type:      protocol.TaskTypeWhoami,
		Status:    protocol.TaskStatusPending,
	})

	req := &beaconProtocol.ResultRequest{
		BeaconID: "beacon-1",
		TaskResult: protocol.TaskResult{
			TaskID:    "task-1",
			ImplantID: "beacon-1",
			Status:    protocol.TaskStatusDone,
			Output:    "root",
		},
	}
	if _, err := HandleResult(beaconStore, taskStore, req, "127.0.0.1"); err != nil {
		t.Fatalf("HandleResult returned error: %v", err)
	}

	task, err := taskStore.GetTask("task-1")
	if err != nil {
		t.Fatalf("GetTask returned error: %v", err)
	}
	if task.Status != protocol.TaskStatusDone {
		t.Fatalf("task status = %q, want %q", task.Status, protocol.TaskStatusDone)
	}
	if task.Result.Output != "root" {
		t.Fatalf("task output = %q, want root", task.Result.Output)
	}
}
