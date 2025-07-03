# Taskflow API

HTTP API сервис на Go для управления долгосрочными задачами (3-5 минут выполнения).

## 📋 Описание

Taskflow API позволяет создавать, удалять и отслеживать статус выполнения асинхронных IO-bound задач. Все данные хранятся в памяти приложения, что делает сервис легким и простым в использовании.

## ✨ Возможности

- ✅ **Асинхронное выполнение задач** - задачи выполняются в фоне через worker pool
- ✅ **REST API** - простой HTTP интерфейс для управления задачами  
- ✅ **In-memory хранение** - без внешних зависимостей
- ✅ **Thread-safe операции** - безопасная работа с несколькими горутинами
- ✅ **Health check** - мониторинг состояния сервиса
- ✅ **CORS поддержка** - готов для веб-приложений
- ✅ **Структурированное логирование** - JSON логи для production
- ✅ **Graceful shutdown** - корректная остановка сервера

## 🚀 Быстрый старт

### Требования

- Go 1.24+

### Установка и запуск

```bash
# Клонируем репозиторий
git clone https://github.com/bambutcha/taskflow.git
cd taskflow

# Устанавливаем зависимости
go mod tidy

# Запускаем сервер
go run cmd/main.go
```

=== ИЛИ c Docker ===

```bash
# Клонируем репозиторий
git clone https://github.com/bambutcha/taskflow.git
cd taskflow

docker-compose up -d --build
```

Сервер запустится на порту 8080.

### Проверка работы

```bash
curl http://localhost:8080
# Ответ: Taskflow API is running!

curl http://localhost:8080/health
# Ответ: JSON с состоянием сервиса
```

## 📚 API Документация

### Базовый URL
```
http://localhost:8080
```

### Эндпоинты

| Метод | Путь | Описание | Статус коды |
|-------|------|----------|-------------|
| `GET` | `/` | Проверка работы API | 200 |
| `GET` | `/health` | Health check сервиса | 200, 503 |
| `POST` | `/tasks` | Создание новой задачи | 201, 400, 409, 500 |
| `GET` | `/tasks/{id}` | Получение статуса задачи | 200, 400, 404, 500 |
| `DELETE` | `/tasks/{id}` | Удаление задачи | 204, 400, 404, 409, 500 |

### Модель данных

**Task Object:**
```json
{
  "id": "my-task-1",
  "status": "pending",
  "created_at": "2025-07-02T19:30:00Z",
  "started_at": "2025-07-02T19:30:05Z",
  "completed_at": "2025-07-02T19:35:15Z", 
  "result": "Task completed by worker 2",
  "error": ""
}
```

**Статусы задач:**
- `pending` - задача создана, ожидает выполнения
- `running` - задача выполняется воркером
- `completed` - задача завершена успешно
- `failed` - задача завершена с ошибкой

### Примеры использования

#### 1. Создание задачи

**Запрос:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"id": "data-processing-job"}'
```

**Ответ (201 Created):**
```json
{
  "id": "data-processing-job",
  "status": "pending",
  "created_at": "2025-07-02T19:30:00Z"
}
```

#### 2. Получение статуса задачи

**Запрос:**
```bash
curl http://localhost:8080/tasks/data-processing-job
```

**Ответ (200 OK) - Выполняется:**
```json
{
  "id": "data-processing-job", 
  "status": "running",
  "created_at": "2025-07-02T19:30:00Z",
  "started_at": "2025-07-02T19:30:05Z"
}
```

**Ответ (200 OK) - Завершена:**
```json
{
  "id": "data-processing-job",
  "status": "completed", 
  "created_at": "2025-07-02T19:30:00Z",
  "started_at": "2025-07-02T19:30:05Z",
  "completed_at": "2025-07-02T19:35:15Z",
  "result": "Task completed by worker 2"
}
```

#### 3. Удаление задачи

**Запрос:**
```bash
curl -X DELETE http://localhost:8080/tasks/data-processing-job
```

**Ответ (204 No Content):** *(пустое тело ответа)*

#### 4. Health Check

**Запрос:**
```bash
curl http://localhost:8080/health
```

**Ответ (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-07-02T19:45:00Z", 
  "uptime": "15m30s",
  "service": "Taskflow API",
  "version": "1.0.0",
  "metrics": {
    "active_workers": 3,
    "total_tasks": 5,
    "pending_tasks": 1,
    "running_tasks": 2, 
    "completed_tasks": 2,
    "failed_tasks": 0
  },
  "checks": {
    "workers": "ok",
    "memory": "ok", 
    "storage": "ok"
  }
}
```

### Обработка ошибок

Все ошибки возвращаются в JSON формате:

```json
{
  "error": "Task ID is required"
}
```

**Коды ошибок:**
- `400 Bad Request` - неверный формат запроса
- `404 Not Found` - задача не найдена
- `409 Conflict` - задача уже существует или выполняется
- `500 Internal Server Error` - внутренняя ошибка сервера
- `503 Service Unavailable` - сервис неработоспособен

## 🏗️ Архитектура

### Компоненты

- **HTTP Server** - обработка REST API с CORS поддержкой
- **Task Manager** - управление жизненным циклом задач
- **Worker Pool** - пул горутин для параллельного выполнения
- **In-Memory Repository** - thread-safe хранилище задач
- **Health Check** - мониторинг состояния сервиса

### Структура проекта

```
taskflow/
├── cmd/
│   └── main.go              # Точка входа
├── internal/
│   ├── handler/             # HTTP обработчики
│   │   ├── handler.go       # API эндпоинты
│   │   ├── handler_test.go  # Тесты API
│   │   ├── health.go        # Health check
│   │   └── health_test.go   # Тесты health check
│   ├── model/
│   │   └── task.go          # Модели данных
│   ├── repository/
│   │   ├── memory.go        # In-memory хранилище
│   │   └── memory_test.go   # Тесты репозитория
│   └── service/
│       ├── manager.go       # Бизнес-логика
│       └── manager_test.go  # Тесты менеджера
├── go.mod                   # Go модуль
├── go.sum                   # Зависимости
├── .env.example             # Пример .env файла
├── Dockerfile               # Dockerfile
├── docker-compose.yml       # Docker Compose файл
└── README.md                # Документация
```

## 🧪 Тестирование

Проект включает тесты с использованием testify:

```bash
# Запуск всех тестов
go test ./...

# Запуск конкретного пакета
go test ./internal/handler/

# Запуск с verbose выводом
go test -v ./...
```

**Типы тестов:**
- Unit тесты для всех компонентов
- HTTP тесты для API эндпоинтов  
- Concurrency тесты для thread safety
- Integration тесты для health check

## 📊 Мониторинг

### Логирование

Сервис использует структурированное JSON логирование:

```json
{"level":"info","msg":"Creating task","task_id":"my-task","time":"2025-07-02T19:30:00Z"}
{"level":"info","msg":"Task completed successfully","task_id":"my-task","worker_id":2,"duration":"4m15s","time":"2025-07-02T19:34:15Z"}
```

### Метрики через Health Check

- Количество активных воркеров
- Общее количество задач
- Распределение по статусам
- Uptime сервиса
- Состояние компонентов

## 🚀 Production

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o taskflow cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/taskflow .
CMD ["./taskflow"]
```

### Environment Variables

```bash
PORT=8080          # Порт сервера (по умолчанию 8080)
WORKERS=3          # Количество воркеров (по умолчанию 3) 
LOG_LEVEL=info     # Уровень логирования
```

### Load Balancer Health Check

```yaml
health_check:
  path: /health
  interval: 30s
  timeout: 5s
  healthy_threshold: 2
  unhealthy_threshold: 3
```

## 📝 Changelog

### v1.0.0
- Базовый REST API для управления задачами
- Асинхронное выполнение через worker pool
- In-memory хранилище с thread safety
- Health check эндпоинт с метриками
- Структурированное логирование
- Graceful shutdown
- Comprehensive test suite

## 📄 Лицензия

(MIT LICENSE)[https://github.com/bambutcha/taskflow/blob/master/LICENSE]

## 🤝 Contributing

1. Fork проект
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request
