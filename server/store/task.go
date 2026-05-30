package store

import (
	"errors"
	"sync"
	"time"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"
)

// ErrServerTaskNotFound is returned when a task lookup by ID fails.
var ErrServerTaskNotFound = errors.New("task not found")

// MemoryServerTaskStore holds the in-memory registry of all server-side tasks,
// protected by a read-write mutex.
type MemoryServerTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*core.ServerTask
}

// NewServerTaskStore creates and returns an initialized ServerTaskStore.
func NewServerTaskStore() ServerTaskStore {
	return &MemoryServerTaskStore{
		tasks: make(map[string]*core.ServerTask),
	}
}

// AddTask inserts or overwrites the task record keyed by its ID. Nil values
// and empty IDs are silently ignored.
func (s *MemoryServerTaskStore) AddTask(task *core.ServerTask) error {
	if task == nil || task.TaskID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	taskCopy := *task
	s.tasks[task.TaskID] = &taskCopy
	return nil
}

// GetTask returns a copy of the task identified by taskID, or
// ErrServerTaskNotFound.
func (s *MemoryServerTaskStore) GetTask(taskID string) (*core.ServerTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, ErrServerTaskNotFound
	}
	taskCopy := *task
	return &taskCopy, nil
}

// GetPendingTasksByImplantID returns copies of all tasks with status "pending"
// for the given implant ID.
func (s *MemoryServerTaskStore) GetPendingTasksByImplantID(implantID string) []*core.ServerTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*core.ServerTask, 0)
	for _, task := range s.tasks {
		if task.ImplantID != implantID || task.Status != "pending" {
			continue
		}

		taskCopy := *task
		result = append(result, &taskCopy)
	}
	return result
}

// UpdateTask applies the result from a completed task to the stored task
// record.
func (s *MemoryServerTaskStore) UpdateTask(taskID string, taskRes protocol.TaskResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return ErrServerTaskNotFound
	}

	task.Result = taskRes
	task.Status = taskRes.Status
	task.Error = taskRes.Error
	if !taskRes.Completed.IsZero() {
		task.CompletedAt = taskRes.Completed
	} else {
		task.CompletedAt = time.Now()
	}

	return nil
}
