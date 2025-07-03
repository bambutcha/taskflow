package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/bambutcha/taskflow/internal/repository"
	"github.com/bambutcha/taskflow/internal/service"
)

func setupTestHealthHandler() *HealthHandler {
	repo := repository.NewMemoryRepository()
	manager := service.NewTaskManagerForTesting(repo, 2)
	handler := NewHealthHandler(manager)
	handler.logger.SetLevel(logrus.PanicLevel)
	return handler
}

func TestHealthHandler_Health(t *testing.T) {
	handler := setupTestHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Service != "Taskflow API" {
		t.Errorf("Expected service 'Taskflow API', got '%s'", response.Service)
	}

	if response.Metrics.ActiveWorkers != 2 {
		t.Errorf("Expected 2 active workers, got %d", response.Metrics.ActiveWorkers)
	}

	if response.Checks["workers"] != "ok" {
		t.Errorf("Expected workers check 'ok', got '%s'", response.Checks["workers"])
	}

	if response.Checks["storage"] != "ok" {
		t.Errorf("Expected storage check 'ok', got '%s'", response.Checks["storage"])
	}
}

func TestHealthHandler_Health_WithTasks(t *testing.T) {
	handler := setupTestHealthHandler()

	handler.taskManager.CreateTask("test-task-1")
	handler.taskManager.CreateTask("test-task-2")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Metrics.TotalTasks < 2 {
		t.Errorf("Expected at least 2 total tasks, got %d", response.Metrics.TotalTasks)
	}
}

func TestHealthHandler_Health_NoWorkers(t *testing.T) {
	repo := repository.NewMemoryRepository()
	manager := service.NewTaskManagerForTesting(repo, 0)
	handler := NewHealthHandler(manager)
	handler.logger.SetLevel(logrus.PanicLevel)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", response.Status)
	}

	if response.Checks["workers"] != "no_workers" {
		t.Errorf("Expected workers check 'no_workers', got '%s'", response.Checks["workers"])
	}
}