package repository

import (
	"testing"

	"github.com/bambutcha/taskflow/internal/model"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := NewMemoryRepository()
	task := model.NewTask("test-1")

	err := repo.Create(task)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Проверяем, что задача действительно создалась
	savedTask, err := repo.GetByID("test-1")
	if err != nil {
		t.Errorf("GetByID() error = %v, want nil", err)
	}

	if savedTask.ID != "test-1" {
		t.Errorf("GetByID() ID = %v, want test-1", savedTask.ID)
	}

	if savedTask.Status != model.StatusPending {
		t.Errorf("GetByID() Status = %v, want %v", savedTask.Status, model.StatusPending)
	}
}

func TestMemoryRepository_Create_Duplicate(t *testing.T) {
	repo := NewMemoryRepository()
	task := model.NewTask("test-1")

	// Создаем первый раз
	err := repo.Create(task)
	if err != nil {
		t.Errorf("First Create() error = %v, want nil", err)
	}

	// Пытаемся создать еще раз
	err = repo.Create(task)
	if err == nil {
		t.Error("Second Create() error = nil, want error")
	}
}

func TestMemoryRepository_GetByID_NotFound(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.GetByID("non-existent")
	if err == nil {
		t.Error("GetByID() error = nil, want error")
	}
}

func TestMemoryRepository_Update(t *testing.T) {
	repo := NewMemoryRepository()
	task := model.NewTask("test-1")

	// Создаем задачу
	err := repo.Create(task)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Обновляем статус
	task.Status = model.StatusRunning
	err = repo.Update(task)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Проверяем, что статус обновился
	updatedTask, err := repo.GetByID("test-1")
	if err != nil {
		t.Errorf("GetByID() error = %v, want nil", err)
	}

	if updatedTask.Status != model.StatusRunning {
		t.Errorf("Status = %v, want %v", updatedTask.Status, model.StatusRunning)
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	task := model.NewTask("test-1")

	// Создаем задачу
	err := repo.Create(task)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Удаляем
	err = repo.Delete("test-1")
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Проверяем, что задача удалилась
	_, err = repo.GetByID("test-1")
	if err == nil {
		t.Error("GetByID() after delete error = nil, want error")
	}
}

func TestMemoryRepository_GetAll(t *testing.T) {
	repo := NewMemoryRepository()

	// Создаем несколько задач
	task1 := model.NewTask("test-1")
	task2 := model.NewTask("test-2")

	repo.Create(task1)
	repo.Create(task2)

	// Получаем все задачи
	tasks, err := repo.GetAll()
	if err != nil {
		t.Errorf("GetAll() error = %v, want nil", err)
	}

	if len(tasks) != 2 {
		t.Errorf("GetAll() length = %v, want 2", len(tasks))
	}
}
