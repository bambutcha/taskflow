package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bambutcha/taskflow/internal/model"
	"github.com/bambutcha/taskflow/internal/repository"
	"github.com/bambutcha/taskflow/internal/service"
)

func setupTestHandler() *TaskHandler {
	repo := repository.NewMemoryRepository()
	manager := service.NewTaskManager(repo, 2)
	return NewTaskHandler(manager)
}

func TestTaskHandler_CreateTask(t *testing.T) {
	handler := setupTestHandler()

	// Подготавливаем запрос
	reqBody := CreateTaskRequest{ID: "test-task-1"}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.CreateTask(w, req)

	// Проверяем ответ
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

	if task.Status != model.StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", task.Status)
	}
}

func TestTaskHandler_CreateTask_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_CreateTask_EmptyID(t *testing.T) {
	handler := setupTestHandler()

	reqBody := CreateTaskRequest{ID: ""}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_CreateTask_Duplicate(t *testing.T) {
	handler := setupTestHandler()

	reqBody := CreateTaskRequest{ID: "duplicate-task"}
	jsonData, _ := json.Marshal(reqBody)

	// Создаем первую задачу
	req1 := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.CreateTask(w1, req1)

	// Пытаемся создать дубликат
	req2 := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.CreateTask(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w2.Code)
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	handler := setupTestHandler()

	// Сначала создаем задачу через менеджер
	_, err := handler.taskManager.CreateTask("get-test-task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Запрашиваем задачу через HTTP
	req := httptest.NewRequest(http.MethodGet, "/tasks/get-test-task", nil)
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

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/tasks/non-existent", nil)
	w := httptest.NewRecorder()

	handler.GetTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	handler := setupTestHandler()

	// Создаем задачу
	_, err := handler.taskManager.CreateTask("delete-test-task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Удаляем задачу
	req := httptest.NewRequest(http.MethodDelete, "/tasks/delete-test-task", nil)
	w := httptest.NewRecorder()

	handler.DeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Проверяем, что задача действительно удалилась
	_, err = handler.taskManager.GetTask("delete-test-task")
	if err == nil {
		t.Error("Expected task to be deleted, but it still exists")
	}
}

func TestTaskHandler_DeleteTask_NotFound(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/tasks/non-existent", nil)
	w := httptest.NewRecorder()

	handler.DeleteTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_ExtractTaskID(t *testing.T) {
	handler := setupTestHandler()

	tests := []struct {
		path     string
		expected string
	}{
		{"/tasks/my-task-1", "my-task-1"},
		{"/tasks/test-123", "test-123"},
		{"/tasks/", ""},
		{"/other/path", ""},
	}

	for _, test := range tests {
		result := handler.extractTaskID(test.path)
		if result != test.expected {
			t.Errorf("extractTaskID(%s) = %s, want %s", test.path, result, test.expected)
		}
	}
}
