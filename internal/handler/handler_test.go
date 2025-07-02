package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
	"github.com/bambutcha/taskflow/internal/service"
)

func setupTestHandler() *TaskHandler {
	repo := repository.NewMemoryRepository()
	manager := service.NewTaskManagerForTesting(repo, 2)
	handler := NewTaskHandler(manager)
	handler.logger.SetLevel(logrus.PanicLevel)
	return handler
}

func setupTestHandlerNoWorkers() *TaskHandler {
	repo := repository.NewMemoryRepository()
	manager := service.NewTaskManagerForTesting(repo, 0)
	handler := NewTaskHandler(manager)
	handler.logger.SetLevel(logrus.PanicLevel)
	return handler
}

func TestTaskHandler_CreateTask(t *testing.T) {
	handler := setupTestHandler()

	reqBody := CreateTaskRequest{ID: "test-task-1"}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var task model.Task
	err := json.NewDecoder(w.Body).Decode(&task)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if task.ID != "test-task-1" {
		t.Errorf("Expected task ID 'test-task-1', got '%s'", task.ID)
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	handler := setupTestHandler()

	_, err := handler.taskManager.CreateTask("get-test-task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks/get-test-task", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "get-test-task"})
	w := httptest.NewRecorder()

	handler.GetTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var task model.Task
	err = json.NewDecoder(w.Body).Decode(&task)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if task.ID != "get-test-task" {
		t.Errorf("Expected task ID 'get-test-task', got '%s'", task.ID)
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	handler := setupTestHandlerNoWorkers()

	_, err := handler.taskManager.CreateTask("delete-test-task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/tasks/delete-test-task", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "delete-test-task"})
	w := httptest.NewRecorder()

	handler.DeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	_, err = handler.taskManager.GetTask("delete-test-task")
	if err == nil {
		t.Error("Expected task to be deleted, but it still exists")
	}
}
