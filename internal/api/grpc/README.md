# gRPC API

This package implements the gRPC protocol layer for the Task Scheduler system.

## Overview

The gRPC API provides high-performance RPC interfaces for task management operations. It implements the same business logic as the HTTP API but uses Protocol Buffers for efficient serialization.

## Service Definition

The service is defined in `api/task_scheduler.proto` and includes the following methods:

- `CreateTask` - Create a new task
- `GetTask` - Retrieve a task by ID
- `ListTasks` - List tasks with filtering
- `CancelTask` - Cancel a task
- `GetTaskStatus` - Get task status information

## Usage

### Starting the gRPC Server

```go
import (
    grpcapi "task-scheduler/internal/api/grpc"
    "task-scheduler/internal/service"
)

// Initialize task service
taskService := service.NewTaskService(repo)

// Create gRPC server on port 9090
grpcServer := grpcapi.NewServer(taskService, 9090)

// Start server
if err := grpcServer.Start(); err != nil {
    log.Fatal(err)
}

// Graceful shutdown
defer grpcServer.Stop()
```

### Testing with grpcurl

List available services:
```bash
grpcurl -plaintext localhost:9090 list
```

Describe a service:
```bash
grpcurl -plaintext localhost:9090 describe taskscheduler.TaskSchedulerService
```

Create a task:
```bash
grpcurl -plaintext -d '{
  "name": "test-task",
  "description": "Test task",
  "execution_mode": "EXECUTION_MODE_IMMEDIATE"
}' localhost:9090 taskscheduler.TaskSchedulerService/CreateTask
```

Get a task:
```bash
grpcurl -plaintext -d '{"task_id": "task-123"}' \
  localhost:9090 taskscheduler.TaskSchedulerService/GetTask
```

List tasks:
```bash
grpcurl -plaintext -d '{"limit": 10}' \
  localhost:9090 taskscheduler.TaskSchedulerService/ListTasks
```

Cancel a task:
```bash
grpcurl -plaintext -d '{"task_id": "task-123"}' \
  localhost:9090 taskscheduler.TaskSchedulerService/CancelTask
```

Get task status:
```bash
grpcurl -plaintext -d '{"task_id": "task-123"}' \
  localhost:9090 taskscheduler.TaskSchedulerService/GetTaskStatus
```

## Client Example

### Go Client

```go
import (
    "context"
    "task-scheduler/api/pb"
    "google.golang.org/grpc"
)

// Connect to server
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Create client
client := pb.NewTaskSchedulerServiceClient(conn)

// Create a task
resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
    Name:          "my-task",
    Description:   "My task description",
    ExecutionMode: pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created task: %s\n", resp.Task.Id)
```

## Error Handling

The gRPC handler converts domain errors to appropriate gRPC status codes:

- `codes.InvalidArgument` - Invalid input or validation errors
- `codes.NotFound` - Task not found
- `codes.Internal` - Internal server errors

## Protocol Buffers

The Protocol Buffer definitions are in `api/task_scheduler.proto`. To regenerate the Go code:

```bash
protoc --go_out=. --go_opt=module=task-scheduler \
       --go-grpc_out=. --go-grpc_opt=module=task-scheduler \
       api/task_scheduler.proto
```

## Features

- **Type Safety**: Strong typing through Protocol Buffers
- **Performance**: Efficient binary serialization
- **Streaming**: Support for bidirectional streaming (future enhancement)
- **Reflection**: gRPC reflection enabled for debugging
- **Error Handling**: Consistent error codes and messages
- **Validation**: Request validation at the protocol layer

## Architecture

```
Client
  ↓
gRPC Handler (internal/api/grpc/handler.go)
  ↓
Task Service (internal/service/task_service.go)
  ↓
Repository (internal/storage/*)
```

The handler is responsible for:
1. Request validation
2. Protocol conversion (proto ↔ domain)
3. Error handling and status code mapping
4. Calling the appropriate service methods

## Comparison with HTTP API

Both HTTP and gRPC APIs provide the same functionality but differ in:

| Feature | HTTP | gRPC |
|---------|------|------|
| Protocol | REST/JSON | Protocol Buffers |
| Performance | Good | Excellent |
| Browser Support | Native | Requires proxy |
| Debugging | Easy (curl) | Requires tools (grpcurl) |
| Type Safety | Runtime | Compile-time |
| Streaming | Limited | Full support |

Choose HTTP for:
- Browser-based clients
- Simple debugging
- Wide compatibility

Choose gRPC for:
- High-performance requirements
- Strong typing needs
- Service-to-service communication
- Streaming requirements
