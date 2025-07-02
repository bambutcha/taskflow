package service

import (
	"testing"
	"time"

	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
)

func TestTaskManager_CreateTask(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManager(repo, 2)

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

	// Проверяем, что задача сохранилась в репозитории
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
	manager := NewTaskManager(repo, 2)

	// Создаем первую задачу
	_, err := manager.CreateTask("test-1")
	if err != nil {
		t.Errorf("First CreateTask() error = %v, want nil", err)
	}

	// Пытаемся создать дубликат
	_, err = manager.CreateTask("test-1")
	if err == nil {
		t.Error("Second CreateTask() error = nil, want error")
	}
}

func TestTaskManager_DeleteTask(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManager(repo, 2)

	// Создаем задачу
	_, err := manager.CreateTask("test-1")
	if err != nil {
		t.Errorf("CreateTask() error = %v, want nil", err)
	}

	// Удаляем задачу (пока она в статусе pending)
	err = manager.DeleteTask("test-1")
	if err != nil {
		t.Errorf("DeleteTask() error = %v, want nil", err)
	}

	// Проверяем, что задача удалилась
	_, err = manager.GetTask("test-1")
	if err == nil {
		t.Error("GetTask() after delete error = nil, want error")
	}
}

func TestTaskManager_DeleteTask_NotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := NewTaskManager(repo, 2)

	err := manager.DeleteTask("non-existent")
	if err == nil {
		t.Error("DeleteTask() error = nil, want error")
	}
}

func TestTaskManager_TaskExecution(t *testing.T) {
	// Этот тест проверяет, что задачи действительно выполняются
	// но делаем время выполнения короче для быстрого тестирования
	repo := repository.NewMemoryRepository()
	manager := NewTaskManager(repo, 1)

	_, err := manager.CreateTask("test-execution")
	if err != nil {
		t.Errorf("CreateTask() error = %v, want nil", err)
	}

	// Ждем немного, чтобы воркер успел подхватить задачу
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что статус изменился с pending
	updatedTask, err := manager.GetTask("test-execution")
	if err != nil {
		t.Errorf("GetTask() error = %v, want nil", err)
	}

	// Задача должна быть либо в статусе running, либо уже completed
	// (в зависимости от того, насколько быстро выполнится)
	if updatedTask.Status == model.StatusPending {
		t.Errorf("Task status is still pending, expected it to be picked up by worker")
	}
}
