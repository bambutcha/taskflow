package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/bambutcha/taskflow/internal/service"
)

type TaskHandler struct {
	taskManager *service.TaskManager
	logger      *logrus.Logger
}

func NewTaskHandler(taskManager *service.TaskManager) *TaskHandler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	return &TaskHandler{
		taskManager: taskManager,
		logger:      logger,
	}
}

type CreateTaskRequest struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid JSON in create task request")
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		h.logger.Warn("Empty task ID in create request")
		h.writeError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, err := h.taskManager.CreateTask(req.ID)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.writeError(w, err.Error(), http.StatusConflict)
		} else {
			h.writeError(w, "Failed to create task", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if taskID == "" {
		h.logger.Warn("Empty task ID in get request")
		h.writeError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
		} else {
			h.writeError(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if taskID == "" {
		h.logger.Warn("Empty task ID in delete request")
		h.writeError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	err := h.taskManager.DeleteTask(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "cannot delete running task") {
			h.writeError(w, "Cannot delete running task", http.StatusConflict)
		} else {
			h.writeError(w, "Failed to delete task", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.logger.WithFields(logrus.Fields{
		"status_code": statusCode,
		"message":     message,
	}).Warn("Sending error response")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{Error: message}
	json.NewEncoder(w).Encode(errorResp)
}
