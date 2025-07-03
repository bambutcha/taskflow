package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/bambutcha/taskflow/internal/service"
)

type HealthHandler struct {
	taskManager *service.TaskManager
	logger      *logrus.Logger
	startTime   time.Time
}

func NewHealthHandler(taskManager *service.TaskManager) *HealthHandler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	return &HealthHandler{
		taskManager: taskManager,
		logger:      logger,
		startTime:   time.Now(),
	}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Metrics   HealthMetrics     `json:"metrics"`
	Checks    map[string]string `json:"checks"`
}

type HealthMetrics struct {
	ActiveWorkers int `json:"active_workers"`
	TotalTasks    int `json:"total_tasks"`
	PendingTasks  int `json:"pending_tasks"`
	RunningTasks  int `json:"running_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks   int `json:"failed_tasks"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Processing health check request")
	
	uptime := time.Since(h.startTime)
	metrics := h.collectMetrics()
	checks := h.performChecks()
	
	status := "healthy"
	for _, checkStatus := range checks {
		if checkStatus != "ok" {
			status = "unhealthy"
			break
		}
	}
	
	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Service:   "Taskflow API",
		Version:   "1.0.0",
		Metrics:   metrics,
		Checks:    checks,
	}
	
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
	
	h.logger.WithFields(logrus.Fields{
		"status":        status,
		"uptime":        uptime.String(),
		"total_tasks":   metrics.TotalTasks,
		"running_tasks": metrics.RunningTasks,
	}).Info("Health check completed")
}

func (h *HealthHandler) collectMetrics() HealthMetrics {
	tasks, err := h.taskManager.GetAllTasks()
	if err != nil {
		h.logger.WithField("error", err.Error()).Warn("Failed to collect task metrics")
		return HealthMetrics{
			ActiveWorkers: h.taskManager.GetWorkerCount(),
		}
	}
	
	var pending, running, completed, failed int
	
	for _, task := range tasks {
		switch task.Status {
		case "pending":
			pending++
		case "running":
			running++
		case "completed":
			completed++
		case "failed":
			failed++
		}
	}
	
	return HealthMetrics{
		ActiveWorkers:  h.taskManager.GetWorkerCount(),
		TotalTasks:     len(tasks),
		PendingTasks:   pending,
		RunningTasks:   running,
		CompletedTasks: completed,
		FailedTasks:    failed,
	}
}

func (h *HealthHandler) performChecks() map[string]string {
	checks := make(map[string]string)
	
	checks["workers"] = "ok"
	if h.taskManager.GetWorkerCount() == 0 {
		checks["workers"] = "no_workers"
	}
	
	checks["memory"] = "ok"
	
	checks["storage"] = "ok"
	_, err := h.taskManager.GetAllTasks()
	if err != nil {
		checks["storage"] = "error"
	}
	
	return checks
}
