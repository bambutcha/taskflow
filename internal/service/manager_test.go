package service

import (
	"strings"
	"testing"
	"time"

	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
)

func TestTaskManager_CreateTask(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManagerForTesting(repo, 2)

	task, err := manager.CreateTask("test-1")
	if err != nil {
		t.Errorf("CreateTask() error = %v, want nil", err)
	}

	if task.ID != "test-1" {
		t.Errorf("CreateTask() ID = %v, want test-1", task.ID)
	}

	if task.Status != model.StatusPending {
		t.Errorf("CreateTask() Status = %v, want %v", task.Status, model.StatusPending)
	}

	savedTask, err := manager.GetTask("test-1")
	if err != nil {
		t.Errorf("GetTask() error = %v, want nil", err)
	}

	if savedTask.ID != "test-1" {
		t.Errorf("GetTask() ID = %v, want test-1", savedTask.ID)
	}
}

func TestTaskManager_CreateTask_Duplicate(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManagerForTesting(repo, 2)

	_, err := manager.CreateTask("test-1")
	if err != nil {
		t.Errorf("First CreateTask() error = %v, want nil", err)
	}

	_, err = manager.CreateTask("test-1")
	if err == nil {
		t.Error("Second CreateTask() error = nil, want error")
	}
}

func TestTaskManager_DeleteTask(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManagerForTesting(repo, 0)

	_, err := manager.CreateTask("test-1")
	if err != nil {
		t.Errorf("CreateTask() error = %v, want nil", err)
	}

	err = manager.DeleteTask("test-1")
	if err != nil {
		t.Errorf("DeleteTask() error = %v, want nil", err)
	}

	_, err = manager.GetTask("test-1")
	if err == nil {
		t.Error("GetTask() after delete error = nil, want error")
	}
}

func TestTaskManager_DeleteTask_NotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManagerForTesting(repo, 2)

	err := manager.DeleteTask("non-existent")
	if err == nil {
		t.Error("DeleteTask() error = nil, want error")
	}
}

func TestTaskManager_TaskExecution(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManagerForTesting(repo, 1)

	_, err := manager.CreateTask("test-execution")
	if err != nil {
		t.Errorf("CreateTask() error = %v, want nil", err)
	}

	time.Sleep(200 * time.Millisecond)

	updatedTask, err := manager.GetTask("test-execution")
	if err != nil {
		t.Errorf("GetTask() error = %v, want nil", err)
	}

	if updatedTask.Status != model.StatusCompleted {
		t.Errorf("Task status = %v, want %v", updatedTask.Status, model.StatusCompleted)
	}

	if updatedTask.Result == "" {
		t.Error("Task result is empty, expected some result")
	}

	if !strings.Contains(updatedTask.Result, "completed by worker") {
		t.Errorf("Task result doesn't contain expected text: %s", updatedTask.Result)
	}
}
