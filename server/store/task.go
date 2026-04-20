package store

import (
	"errors"
	"sync"
	"time"
	
	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"
)

var ErrServerTaskNotFound = errors.New("task not found")

type ServerTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*core.ServerTask
}

func (s *ServerTaskStore) AddTask(task *core.ServerTask) {
	if task == nil || task.TaskID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	
	taskCopy := *task
	s.tasks[task.TaskID] = &taskCopy
}

func (s *ServerTaskStore) GetTask(taskID string) (*core.ServerTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, ErrServerTaskNotFound
	}
	taskCopy := *task
	return &taskCopy, nil
}

func NewServerTaskStore() *ServerTaskStore {
	return &ServerTaskStore{
		tasks: make(map[string]*core.ServerTask),
	}
}

func (s *ServerTaskStore) GetPendingTasksByImplantID(implantID string) []*core.ServerTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*core.ServerTask, 0)
	for _, task := range s.tasks {
		if task.ImplantID != implantID {
			continue
		}
		if task.Status != "pending" {
			continue
		}
		
		taskCopy := *task
		result = append(result, &taskCopy)
	}
	return result
}

func (s *ServerTaskStore) UpdateTask(taskID string, taskRes protocol.TaskResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, ok := s.tasks[taskID]
	if !ok {
		return ErrServerTaskNotFound
	}
	
	s.updateTaskResult(task, taskRes)
	s.updateTaskStatus(task, taskRes)
	s.updateTaskCompleted(task, taskRes)
	
	return nil
}

func (s *ServerTaskStore) updateTaskResult(task *core.ServerTask, taskRes protocol.TaskResult) {
	task.Result = taskRes
}

func (s *ServerTaskStore) updateTaskStatus(task *core.ServerTask, taskRes protocol.TaskResult) {
	task.Status = taskRes.Status
	task.Error = taskRes.Error
}

func (s *ServerTaskStore) updateTaskCompleted(task *core.ServerTask, taskRes protocol.TaskResult) {
	if !taskRes.Completed.IsZero() {
		task.CompletedAt = taskRes.Completed
		return
	}
	task.CompletedAt = time.Now()
}
