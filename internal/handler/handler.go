package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bambutcha/taskflow/internal/service"
)

type TaskHandler struct {
	taskManager *service.TaskManager
}

func NewTaskHandler(taskManager *service.TaskManager) *TaskHandler {
	return &TaskHandler{
		taskManager: taskManager,
	}
}

type CreateTaskRequest struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
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
	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := h.extractTaskID(r.URL.Path)
	if taskID == "" {
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
	if r.Method != http.MethodDelete {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := h.extractTaskID(r.URL.Path)
	if taskID == "" {
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

func (h *TaskHandler) extractTaskID(path string) string {
	if strings.HasPrefix(path, "/tasks/") {
		return strings.TrimPrefix(path, "/tasks/")
	}
	return ""
}

func (h *TaskHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{Error: message}
	json.NewEncoder(w).Encode(errorResp)
}
