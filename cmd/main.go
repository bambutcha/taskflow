package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bambutcha/taskflow/internal/handler"
	"github.com/bambutcha/taskflow/internal/repository"
	"github.com/bambutcha/taskflow/internal/service"
)

func main() {
	fmt.Println("Taskflow API starting...")

	repo := repository.NewMemoryRepository()
	taskManager := service.NewTaskManager(repo, 3)
	taskHandler := handler.NewTaskHandler(taskManager)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tasks" {
			taskHandler.CreateTask(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetTask(w, r)
		case http.MethodDelete:
			taskHandler.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Taskflow API is running!"))
}
