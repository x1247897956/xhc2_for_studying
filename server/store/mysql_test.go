package store

import (
	"testing"
	"time"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGORMStoresPersistBeaconTaskAndImplant(t *testing.T) {
	db := newTestGORMDB(t)
	stores, err := NewGORMStores(db)
	if err != nil {
		t.Fatalf("NewGORMStores returned error: %v", err)
	}

	beacon := &core.Beacon{
		ID:            "beacon-1",
		Hostname:      "host",
		Username:      "user",
		OS:            "linux",
		Arch:          "amd64",
		Interval:      5,
		Jitter:        1,
		LastCheckIn:   1715000000,
		RemoteAddress: "10.0.0.1",
	}
	if err := stores.Beacons.Add(beacon); err != nil {
		t.Fatalf("Add beacon returned error: %v", err)
	}
	gotBeacon, err := stores.Beacons.Get(beacon.ID)
	if err != nil {
		t.Fatalf("Get beacon returned error: %v", err)
	}
	if *gotBeacon != *beacon {
		t.Fatalf("beacon = %+v, want %+v", gotBeacon, beacon)
	}

	task := &core.ServerTask{
		TaskID:    "task-1",
		Type:      protocol.TaskTypeWhoami,
		ImplantID: beacon.ID,
		Status:    protocol.TaskStatusPending,
		CreatedAt: time.Unix(1715000001, 0),
	}
	if err := stores.Tasks.AddTask(task); err != nil {
		t.Fatalf("AddTask returned error: %v", err)
	}
	pending := stores.Tasks.GetPendingTasksByImplantID(beacon.ID)
	if len(pending) != 1 || pending[0].TaskID != task.TaskID {
		t.Fatalf("pending tasks = %+v", pending)
	}

	completed := time.Unix(1715000010, 0).UTC()
	result := protocol.TaskResult{
		TaskID:    task.TaskID,
		ImplantID: beacon.ID,
		Status:    protocol.TaskStatusDone,
		Completed: completed,
		Output:    "root",
	}
	if err := stores.Tasks.UpdateTask(task.TaskID, result); err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	gotTask, err := stores.Tasks.GetTask(task.TaskID)
	if err != nil {
		t.Fatalf("GetTask returned error: %v", err)
	}
	if gotTask.Status != protocol.TaskStatusDone || gotTask.Result.Output != "root" || gotTask.CompletedAt.Unix() != completed.Unix() {
		t.Fatalf("task = %+v", gotTask)
	}

	record := &ImplantRecord{
		ImplantAgePrivateKey: "AGE-SECRET-KEY-1",
		ExtMap: protocol.ExtensionMap{
			KeyExchange: ".woff",
			Register:    ".php",
			Poll:        ".js",
			Result:      ".json",
		},
	}
	if err := stores.Implants.Set("digest-1", record); err != nil {
		t.Fatalf("Set implant returned error: %v", err)
	}
	gotRecord, ok := stores.Implants.Get("digest-1")
	if !ok {
		t.Fatal("implant record was not found")
	}
	if gotRecord.ImplantAgePrivateKey != record.ImplantAgePrivateKey || gotRecord.ExtMap != record.ExtMap {
		t.Fatalf("implant record = %+v, want %+v", gotRecord, record)
	}
}

func TestGORMStoresReturnNotFoundErrors(t *testing.T) {
	db := newTestGORMDB(t)
	stores, err := NewGORMStores(db)
	if err != nil {
		t.Fatalf("NewGORMStores returned error: %v", err)
	}

	if _, err := stores.Beacons.Get("missing"); err != ErrBeaconNotFound {
		t.Fatalf("Beacon Get error = %v, want %v", err, ErrBeaconNotFound)
	}
	if _, err := stores.Tasks.GetTask("missing"); err != ErrServerTaskNotFound {
		t.Fatalf("Task Get error = %v, want %v", err, ErrServerTaskNotFound)
	}
	if err := stores.Beacons.UpdateCheckIn("missing", 1, ""); err != ErrBeaconNotFound {
		t.Fatalf("UpdateCheckIn error = %v, want %v", err, ErrBeaconNotFound)
	}
}

func newTestGORMDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite gorm db: %v", err)
	}
	return db
}
