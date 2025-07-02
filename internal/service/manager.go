package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
)

type TaskManager struct {
	repo       repository.TaskRepository
	workerPool chan *model.Task
	workers    int
	testMode   bool
}

func NewTaskManager(repo repository.TaskRepository, workers int) *TaskManager {
	manager := &TaskManager{
		repo:       repo,
		workerPool: make(chan *model.Task, workers*2),
		workers:    workers,
		testMode:   false,
	}

	manager.startWorkers()
	return manager
}

func NewTaskManagerForTesting(repo repository.TaskRepository, workers int) *TaskManager {
	manager := &TaskManager{
		repo:       repo,
		workerPool: make(chan *model.Task, workers*2),
		workers:    workers,
		testMode:   true,
	}

	manager.startWorkers()
	return manager
}

func (tm *TaskManager) CreateTask(id string) (*model.Task, error) {
	task := model.NewTask(id)

	err := tm.repo.Create(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	select {
	case tm.workerPool <- task:
	default:
	}

	return task, nil
}

func (tm *TaskManager) GetTask(id string) (*model.Task, error) {
	return tm.repo.GetByID(id)
}

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

func (tm *TaskManager) GetAllTasks() ([]*model.Task, error) {
	return tm.repo.GetAll()
}

func (tm *TaskManager) startWorkers() {
	for i := 0; i < tm.workers; i++ {
		go tm.worker(i + 1)
	}
}

func (tm *TaskManager) worker(workerID int) {
	for task := range tm.workerPool {
		tm.executeTask(task, workerID)
	}
}

func (tm *TaskManager) executeTask(task *model.Task, workerID int) {
	now := time.Now()
	task.Status = model.StatusRunning
	task.StartedAt = &now

	err := tm.repo.Update(task)
	if err != nil {
		return
	}

	var duration time.Duration
	if tm.testMode {
		duration = 100 * time.Millisecond
	} else {
		duration = time.Duration(3+rand.Intn(3)) * time.Minute
	}

	time.Sleep(duration)

	completedAt := time.Now()
	task.Status = model.StatusCompleted
	task.CompletedAt = &completedAt
	task.Result = fmt.Sprintf("Task completed by worker %d", workerID)

	tm.repo.Update(task)
}