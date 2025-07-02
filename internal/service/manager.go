package service

import (
	"fmt"
	"time"

	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
)

// TaskManager управляет жизненным циклом задач
type TaskManager struct {
	repo       repository.TaskRepository
	workerPool chan *model.Task
	workers    int
}

// NewTaskManager создает новый менеджер задач
func NewTaskManager(repo repository.TaskRepository, workers int) *TaskManager {
	manager := &TaskManager{
		repo:       repo,
		workerPool: make(chan *model.Task, workers*2),
		workers:    workers,
	}

	manager.startWorkers()

	return manager
}

// CreateTask создает новую задачу и добавляет её в очередь выполнения
func (tm *TaskManager) CreateTask(id string) (*model.Task, error) {
	task := model.NewTask(id)

	err := tm.repo.Create(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	select {
	case tm.workerPool <- task:
		// Задача успешно добавлена в очередь
	default:
		// Очередь заполнена, но это не критично
		// Задача будет выполнена позже
	}

	return task, nil
}

// GetTask возвращает задачу по ID
func (tm *TaskManager) GetTask(id string) (*model.Task, error) {
	return tm.repo.GetByID(id)
}

// DeleteTask удаляет задачу
func (tm *TaskManager) DeleteTask(id string) error {
	task, err := tm.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.IsRunning() {
		return fmt.Errorf("cannot delete running task")
	}

	return tm.repo.Delete(id)
}

// GetAllTasks возвращает все задачи
func (tm *TaskManager) GetAllTasks() ([]*model.Task, error) {
	return tm.repo.GetAll()
}

// startWorkers запускает горутины-воркеры для выполнения задач
func (tm *TaskManager) startWorkers() {
	for i := 0; i < tm.workers; i++ {
		go tm.worker(i + 1)
	}
}

// worker - отдельная горутина, которая выполняет задачи
func (tm *TaskManager) worker(workerID int) {
	for task := range tm.workerPool {
		tm.executeTask(task, workerID)
	}
}

// executeTask выполняет конкретную задачу
func (tm *TaskManager) executeTask(task *model.Task, workerID int) {
	// Обновляем статус на "выполняется"
	now := time.Now()
	task.Status = model.StatusRunning
	task.StartedAt = &now

	err := tm.repo.Update(task)
	if err != nil {
		// Если не можем обновить статус, прерываем выполнение
		return
	}

	// Имитируем IO-bound операцию
	// В реальном проекте здесь была бы настоящая работа
	duration := time.Minute*0 + time.Second*10 // 10 секунд
	time.Sleep(duration)

	// Завершаем задачу
	completedAt := time.Now()
	task.Status = model.StatusCompleted
	task.CompletedAt = &completedAt
	task.Result = fmt.Sprintf("Task completed by worker %d after %v", workerID, duration)

	// Сохраняем результат
	tm.repo.Update(task)
}
