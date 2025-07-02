package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
)

type TaskManager struct {
	repo       repository.TaskRepository
	workerPool chan *model.Task
	workers    int
	testMode   bool
	logger     *logrus.Logger
}

func NewTaskManager(repo repository.TaskRepository, workers int) *TaskManager {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	manager := &TaskManager{
		repo:       repo,
		workerPool: make(chan *model.Task, workers*2),
		workers:    workers,
		testMode:   false,
		logger:     logger,
	}

	manager.startWorkers()
	manager.logger.WithField("workers", workers).Info("TaskManager initialized")
	return manager
}

func NewTaskManagerForTesting(repo repository.TaskRepository, workers int) *TaskManager {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	manager := &TaskManager{
		repo:       repo,
		workerPool: make(chan *model.Task, workers*2),
		workers:    workers,
		testMode:   true,
		logger:     logger,
	}

	manager.startWorkers()
	return manager
}

func (tm *TaskManager) CreateTask(id string) (*model.Task, error) {
	tm.logger.WithField("task_id", id).Info("Creating task")
	
	task := model.NewTask(id)

	err := tm.repo.Create(task)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"task_id": id,
			"error":   err.Error(),
		}).Error("Failed to create task in repository")
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	select {
	case tm.workerPool <- task:
		tm.logger.WithField("task_id", id).Info("Task queued for execution")
	default:
		tm.logger.WithField("task_id", id).Warn("Worker pool full, task will be processed later")
	}

	tm.logger.WithField("task_id", id).Info("Task created successfully")
	return task, nil
}

func (tm *TaskManager) GetTask(id string) (*model.Task, error) {
	tm.logger.WithField("task_id", id).Debug("Getting task")
	
	task, err := tm.repo.GetByID(id)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"task_id": id,
			"error":   err.Error(),
		}).Warn("Task not found")
		return nil, err
	}
	
	return task, nil
}

func (tm *TaskManager) DeleteTask(id string) error {
	tm.logger.WithField("task_id", id).Info("Deleting task")
	
	task, err := tm.repo.GetByID(id)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"task_id": id,
			"error":   err.Error(),
		}).Warn("Cannot delete task: not found")
		return fmt.Errorf("task not found: %w", err)
	}

	if task.IsRunning() {
		tm.logger.WithFields(logrus.Fields{
			"task_id": id,
			"status":  task.Status,
		}).Warn("Cannot delete running task")
		return fmt.Errorf("cannot delete running task")
	}

	err = tm.repo.Delete(id)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"task_id": id,
			"error":   err.Error(),
		}).Error("Failed to delete task from repository")
		return err
	}

	tm.logger.WithField("task_id", id).Info("Task deleted successfully")
	return nil
}

func (tm *TaskManager) GetAllTasks() ([]*model.Task, error) {
	tm.logger.Debug("Getting all tasks")
	
	tasks, err := tm.repo.GetAll()
	if err != nil {
		tm.logger.WithField("error", err.Error()).Error("Failed to get all tasks")
		return nil, err
	}
	
	tm.logger.WithField("count", len(tasks)).Debug("Retrieved all tasks")
	return tasks, nil
}

func (tm *TaskManager) startWorkers() {
	for i := 0; i < tm.workers; i++ {
		go tm.worker(i + 1)
	}
}

func (tm *TaskManager) worker(workerID int) {
	tm.logger.WithField("worker_id", workerID).Info("Worker started")
	
	for task := range tm.workerPool {
		tm.executeTask(task, workerID)
	}
	
	tm.logger.WithField("worker_id", workerID).Info("Worker stopped")
}

func (tm *TaskManager) executeTask(task *model.Task, workerID int) {
	logger := tm.logger.WithFields(logrus.Fields{
		"task_id":   task.ID,
		"worker_id": workerID,
	})
	
	logger.Info("Starting task execution")
	
	now := time.Now()
	task.Status = model.StatusRunning
	task.StartedAt = &now

	err := tm.repo.Update(task)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to update task status to running")
		return
	}

	logger.Info("Task status updated to running")

	var duration time.Duration
	if tm.testMode {
		duration = 100 * time.Millisecond
	} else {
		duration = time.Duration(3+rand.Intn(3)) * time.Minute
	}

	logger.WithField("duration", duration.String()).Info("Executing IO-bound operation")
	time.Sleep(duration)

	completedAt := time.Now()
	task.Status = model.StatusCompleted
	task.CompletedAt = &completedAt
	task.Result = fmt.Sprintf("Task completed by worker %d", workerID)

	err = tm.repo.Update(task)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to update completed task")
		return
	}

	logger.WithField("duration", duration.String()).Info("Task completed successfully")
}
