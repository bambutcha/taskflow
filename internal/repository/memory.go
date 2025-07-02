package repository

import (
	"fmt"
	"sync"

	"github.com/bambutcha/taskflow/internal/model"
)

// TaskRepository определяет интерфейс для работы с задачами
type TaskRepository interface {
	Create(task *model.Task) error
	GetByID(id string) (*model.Task, error)
	Update(task *model.Task) error
	Delete(id string) error
	GetAll() ([]*model.Task, error)
}

// MemoryRepository реализует TaskRepository с хранением в памяти
type MemoryRepository struct {
	tasks map[string]*model.Task
	mutex sync.RWMutex
}

// NewMemoryRepository создает новый экземпляр репозитория
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks: make(map[string]*model.Task),
	}
}

// Create добавляет новую задачу в хранилище
func (r *MemoryRepository) Create(task *model.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

// GetByID возвращает задачу по ID
func (r *MemoryRepository) GetByID(id string) (*model.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	return task, nil
}

// Update обновляет существующую задачу
func (r *MemoryRepository) Update(task *model.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("task with ID %s not found", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

// Delete удаляет задачу из хранилища
func (r *MemoryRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("task with ID %s not found", id)
	}

	delete(r.tasks, id)
	return nil
}

// GetAll возвращает все задачи
func (r *MemoryRepository) GetAll() ([]*model.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := make([]*model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}
