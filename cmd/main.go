package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/bambutcha/taskflow/internal/handler"
	"github.com/bambutcha/taskflow/internal/repository"
	"github.com/bambutcha/taskflow/internal/service"
)

func main() {
	config := loadConfig()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	if level, err := logrus.ParseLevel(config.LogLevel); err == nil {
		log.SetLevel(level)
	}

	log.Info("Taskflow API starting...")

	repo := repository.NewMemoryRepository()
	taskManager := service.NewTaskManager(repo, config.Workers)
	taskHandler := handler.NewTaskHandler(taskManager)
	healthHandler := handler.NewHealthHandler(taskManager)

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/health", healthHandler.Health).Methods("GET")
	r.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	r.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")

	corsOptions := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type"})

	corsHandler := handlers.CORS(corsOptions, corsMethods, corsHeaders)(r)
	loggedHandler := handlers.LoggingHandler(os.Stdout, corsHandler)

	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      loggedHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.WithField("port", config.Port).Info("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

type Config struct {
	Port     string
	Workers  int
	LogLevel string
}

func loadConfig() Config {
	port := getEnv("PORT", "8080")
	workersStr := getEnv("WORKERS", "3")
	logLevel := getEnv("LOG_LEVEL", "info")

	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		workers = 3
	}

	return Config{
		Port:     port,
		Workers:  workers,
		LogLevel: logLevel,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Taskflow API is running!"))
}
