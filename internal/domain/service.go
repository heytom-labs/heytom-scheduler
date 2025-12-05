package domain

import (
	"context"
	"time"
)

// CreateTaskRequest contains parameters for creating a task
type CreateTaskRequest struct {
	Name             string
	Description      string
	ParentID         *string
	ExecutionMode    ExecutionMode
	ScheduleConfig   *ScheduleConfig
	CallbackConfig   *CallbackConfig
	RetryPolicy      *RetryPolicy
	ConcurrencyLimit int
	AlertPolicy      *AlertPolicy
	Metadata         map[string]string
}

// TaskService defines the interface for task business logic
type TaskService interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error)

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// ListTasks retrieves tasks matching the filter
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)

	// CancelTask cancels a task
	CancelTask(ctx context.Context, taskID string) error

	// UpdateTaskStatus updates the status of a task
	UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error
}

// ScheduleService defines the interface for task scheduling logic
type ScheduleService interface {
	// ScheduleTask schedules a task for execution
	ScheduleTask(ctx context.Context, task *Task) error

	// UnscheduleTask removes a task from the schedule
	UnscheduleTask(ctx context.Context, taskID string) error

	// GetNextExecutionTime calculates the next execution time for a task
	GetNextExecutionTime(task *Task) (time.Time, error)
}

// CallbackResponse contains the response from a callback
type CallbackResponse struct {
	Status   TaskStatus
	Output   string
	Error    *string
	Metadata map[string]string
}

// CallbackService defines the interface for callback handling
type CallbackService interface {
	// ExecuteCallback executes a synchronous callback
	ExecuteCallback(ctx context.Context, task *Task, result *ExecutionResult) error

	// ExecuteAsyncCallback executes an asynchronous callback
	ExecuteAsyncCallback(ctx context.Context, task *Task) error

	// HandleCallbackResponse handles the response from an async callback
	HandleCallbackResponse(ctx context.Context, taskID string, response *CallbackResponse) error
}
