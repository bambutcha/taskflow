package model

import "time"

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID          string     `json:"id"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
}

func NewTask(id string) *Task {
	return &Task{
		ID:        id,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

func (t *Task) IsCompleted() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed
}

func (t *Task) IsRunning() bool {
	return t.Status == StatusRunning
}
