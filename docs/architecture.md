# Task Scheduler System - Architecture Overview

## Project Structure

```
task-scheduler-system/
├── cmd/                          # Application entry points
│   └── scheduler/               # Main scheduler service
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   └── domain/                  # Domain models and interfaces
│       ├── task.go              # Task data models
│       ├── service.go           # Service interfaces
│       ├── repository.go        # Repository interface
│       ├── scheduler.go         # Scheduler interfaces
│       ├── discovery.go         # Service discovery interface
│       └── notification.go      # Notification interface
├── pkg/                         # Public libraries
│   └── logger/                  # Logging utilities
│       └── logger.go            # Zap logger wrapper
├── api/                         # API definitions (to be added)
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # Project documentation
```

## Core Domain Models

### Task
The central entity representing a schedulable work unit with:
- Unique identifier
- Execution mode (immediate, scheduled, interval, cron)
- Status tracking (pending, running, success, failed, cancelled)
- Parent-child relationships for subtasks
- Retry policy configuration
- Callback configuration
- Alert policy configuration
- Concurrency limits for subtasks

### ExecutionMode
Defines when and how tasks should be executed:
- **Immediate**: Execute as soon as possible
- **Scheduled**: Execute at a specific time
- **Interval**: Execute repeatedly at fixed intervals
- **Cron**: Execute based on cron expression

### TaskStatus
Tracks the lifecycle state of tasks:
- **Pending**: Waiting to be executed
- **Running**: Currently executing
- **Success**: Completed successfully
- **Failed**: Execution failed
- **Cancelled**: Manually cancelled

## Core Interfaces

### TaskRepository
Data access layer for task persistence:
- Create, Read, Update, Delete operations
- List tasks with filtering
- Retrieve subtasks

### TaskService
Business logic for task management:
- Task creation and validation
- Task retrieval and listing
- Task cancellation
- Status updates

### ScheduleService
Task scheduling logic:
- Schedule tasks for execution
- Calculate next execution time
- Unschedule tasks

### CallbackService
Callback handling:
- Execute synchronous callbacks
- Execute asynchronous callbacks
- Handle callback responses

### Scheduler
Task execution and scheduling:
- Start/stop scheduler
- Submit tasks for execution
- Acquire tasks for execution by nodes
- Release tasks after execution

### DistributedLock
Distributed locking for multi-node deployments:
- Acquire locks with TTL
- Release locks
- Refresh lock TTL

### ServiceDiscovery
Service discovery for callbacks:
- Discover service instances
- Register service instances
- Deregister service instances

### Notifier
Alert notification system:
- Send single alerts
- Send batch alerts

## Technology Stack

- **Language**: Go 1.21+
- **Logging**: Uber Zap (structured logging)
- **Module Management**: Go Modules

## Next Steps

The following components will be implemented in subsequent tasks:
1. Storage layer (MySQL, PostgreSQL, MongoDB adapters)
2. Service layer implementations
3. Scheduling service with cron support
4. Callback service with service discovery
5. Distributed lock implementation
6. HTTP and gRPC API handlers
7. Vue.js admin UI
8. Monitoring and observability
