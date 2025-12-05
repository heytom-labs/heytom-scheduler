# HTTP API Documentation

This package provides the HTTP/REST API interface for the Task Scheduler System using the Gin web framework.

## Overview

The HTTP API exposes endpoints for managing tasks through RESTful operations. All endpoints follow REST conventions and return JSON responses.

## Base URL

```
http://localhost:8080/api/v1
```

## Endpoints

### Health Check

Check if the server is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "time": "2024-01-01T12:00:00Z"
}
```

### Create Task

Create a new task.

**Endpoint:** `POST /api/v1/tasks`

**Request Body:**
```json
{
  "name": "My Task",
  "description": "Task description",
  "execution_mode": "immediate",
  "schedule_config": {
    "scheduled_time": "2024-12-31T23:59:59Z",
    "interval": "5m",
    "cron_expr": "0 0 * * *"
  },
  "callback_config": {
    "protocol": "http",
    "url": "http://example.com/callback",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer token"
    },
    "timeout": "30s",
    "is_async": false
  },
  "retry_policy": {
    "max_retries": 3,
    "retry_interval": "5s",
    "backoff_factor": 2.0
  },
  "concurrency_limit": 5,
  "alert_policy": {
    "enable_failure_alert": true,
    "retry_threshold": 2,
    "timeout_threshold": "1h",
    "channels": [
      {
        "type": "email",
        "config": {
          "to": "admin@example.com"
        }
      }
    ]
  },
  "metadata": {
    "key": "value"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Task",
  "description": "Task description",
  "execution_mode": "immediate",
  "status": "pending",
  "retry_count": 0,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Get Task

Retrieve a specific task by ID.

**Endpoint:** `GET /api/v1/tasks/:id`

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Task",
  "description": "Task description",
  "execution_mode": "immediate",
  "status": "running",
  "retry_count": 0,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "started_at": "2024-01-01T12:00:01Z"
}
```

### List Tasks

Retrieve a list of tasks with optional filtering.

**Endpoint:** `GET /api/v1/tasks`

**Query Parameters:**
- `status` (optional): Filter by task status (pending, running, success, failed, cancelled)
- `parent_id` (optional): Filter by parent task ID
- `node_id` (optional): Filter by node ID
- `limit` (optional): Maximum number of tasks to return
- `offset` (optional): Number of tasks to skip

**Example:** `GET /api/v1/tasks?status=pending&limit=10`

**Response:** `200 OK`
```json
{
  "tasks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Task 1",
      "status": "pending",
      "execution_mode": "immediate",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Task 2",
      "status": "pending",
      "execution_mode": "scheduled",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "count": 2
}
```

### Cancel Task

Cancel a task and its subtasks.

**Endpoint:** `DELETE /api/v1/tasks/:id`

**Response:** `200 OK`
```json
{
  "message": "task cancelled successfully",
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Get Task Status

Retrieve the current status of a task.

**Endpoint:** `GET /api/v1/tasks/:id/status`

**Response:** `200 OK`
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "retry_count": 1,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:05Z",
  "started_at": "2024-01-01T12:00:01Z",
  "completed_at": null
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "code": "optional_error_code"
}
```

### Error Codes

- `invalid_request` (400): The request is malformed or missing required fields
- `invalid_input` (400): Input validation failed
- `invalid_execution_mode` (400): Invalid execution mode specified
- `invalid_schedule_config` (400): Invalid schedule configuration
- `not_found` (404): The requested task was not found
- `internal_error` (500): An internal server error occurred

## Execution Modes

- `immediate`: Execute the task immediately after creation
- `scheduled`: Execute the task at a specific time (requires `scheduled_time`)
- `interval`: Execute the task repeatedly at fixed intervals (requires `interval`)
- `cron`: Execute the task based on a cron expression (requires `cron_expr`)

## Task Status

- `pending`: Task is waiting to be executed
- `running`: Task is currently executing
- `success`: Task completed successfully
- `failed`: Task execution failed
- `cancelled`: Task was cancelled

## Duration Format

Duration strings use Go's duration format:
- `5s` - 5 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `1h30m` - 1 hour and 30 minutes

## Example Usage

### Create an Immediate Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Process Data",
    "description": "Process incoming data",
    "execution_mode": "immediate"
  }'
```

### Create a Scheduled Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nightly Backup",
    "description": "Backup database",
    "execution_mode": "scheduled",
    "schedule_config": {
      "scheduled_time": "2024-12-31T23:00:00Z"
    }
  }'
```

### Create a Cron Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Report",
    "description": "Generate daily report",
    "execution_mode": "cron",
    "schedule_config": {
      "cron_expr": "0 9 * * *"
    }
  }'
```

### Get Task Status

```bash
curl http://localhost:8080/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000/status
```

### List Pending Tasks

```bash
curl "http://localhost:8080/api/v1/tasks?status=pending&limit=10"
```

### Cancel a Task

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

## CORS Support

The API includes CORS middleware that allows cross-origin requests from any origin. In production, you should configure this to only allow specific origins.

## Logging

All HTTP requests are logged with the following information:
- HTTP method
- Request path
- Query parameters
- Response status code
- Request latency
- Client IP address

## Server Configuration

The server can be configured with the following parameters:

```go
serverConfig := &http.ServerConfig{
    Host:         "0.0.0.0",
    Port:         8080,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

## Integration Example

See `examples/http_server_example.go` for a complete example of how to set up and run the HTTP server.
