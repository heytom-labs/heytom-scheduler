package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/metrics"

	"github.com/google/uuid"
)

var (
	// ErrTaskNotFound is returned when a task is not found
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
	// ErrInvalidExecutionMode is returned when execution mode is invalid
	ErrInvalidExecutionMode = errors.New("invalid execution mode")
	// ErrInvalidScheduleConfig is returned when schedule config is invalid
	ErrInvalidScheduleConfig = errors.New("invalid schedule config")
)

// taskService implements the TaskService interface
type taskService struct {
	repo domain.TaskRepository
}

// NewTaskService creates a new TaskService instance
func NewTaskService(repo domain.TaskRepository) domain.TaskService {
	return &taskService{
		repo: repo,
	}
}

// CreateTask creates a new task with validation
func (s *taskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	// Validate input
	if err := s.validateCreateTaskRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Generate unique ID
	taskID := uuid.New().String()

	// Create task object
	now := time.Now()
	task := &domain.Task{
		ID:               taskID,
		Name:             req.Name,
		Description:      req.Description,
		ParentID:         req.ParentID,
		ExecutionMode:    req.ExecutionMode,
		ScheduleConfig:   req.ScheduleConfig,
		CallbackConfig:   req.CallbackConfig,
		RetryPolicy:      req.RetryPolicy,
		ConcurrencyLimit: req.ConcurrencyLimit,
		AlertPolicy:      req.AlertPolicy,
		Status:           domain.TaskStatusPending,
		RetryCount:       0,
		CreatedAt:        now,
		UpdatedAt:        now,
		Metadata:         req.Metadata,
	}

	// If parent task is specified, verify it exists
	if req.ParentID != nil && *req.ParentID != "" {
		parent, err := s.repo.Get(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent task not found: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent task not found: %s", *req.ParentID)
		}
	}

	// Save to repository
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Record metrics
	metrics.RecordTaskCreated(string(req.ExecutionMode))

	return task, nil
}

// GetTask retrieves a task by ID
func (s *taskService) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task ID is required", ErrInvalidInput)
	}

	task, err := s.repo.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// ListTasks retrieves tasks matching the filter
func (s *taskService) ListTasks(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	if filter == nil {
		filter = &domain.TaskFilter{}
	}

	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// CancelTask cancels a task and its subtasks
func (s *taskService) CancelTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("%w: task ID is required", ErrInvalidInput)
	}

	// Get the task
	task, err := s.repo.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task == nil {
		return ErrTaskNotFound
	}

	// Check if task can be cancelled
	if task.Status == domain.TaskStatusSuccess || task.Status == domain.TaskStatusCancelled {
		return fmt.Errorf("task cannot be cancelled in status: %s", task.Status)
	}

	// Update task status to cancelled
	task.Status = domain.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	// Cancel all subtasks recursively
	subtasks, err := s.repo.GetSubTasks(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	for _, subtask := range subtasks {
		if err := s.CancelTask(ctx, subtask.ID); err != nil {
			// Log error but continue cancelling other subtasks
			continue
		}
	}

	return nil
}

// UpdateTaskStatus updates the status of a task
func (s *taskService) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error {
	if taskID == "" {
		return fmt.Errorf("%w: task ID is required", ErrInvalidInput)
	}

	// Validate status
	if !isValidTaskStatus(status) {
		return fmt.Errorf("%w: invalid task status: %s", ErrInvalidInput, status)
	}

	// Get the task
	task, err := s.repo.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task == nil {
		return ErrTaskNotFound
	}

	// Validate status transition
	if err := s.validateStatusTransition(task.Status, status); err != nil {
		return err
	}

	// Update status and timestamps
	task.Status = status
	task.UpdatedAt = time.Now()

	switch status {
	case domain.TaskStatusRunning:
		if task.StartedAt == nil {
			now := time.Now()
			task.StartedAt = &now
		}
	case domain.TaskStatusSuccess, domain.TaskStatusFailed, domain.TaskStatusCancelled:
		if task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Update parent task status if this is a subtask
	if task.ParentID != nil && *task.ParentID != "" {
		if err := s.updateParentTaskStatus(ctx, *task.ParentID); err != nil {
			// Log error but don't fail the operation
			return nil
		}
	}

	return nil
}

// validateCreateTaskRequest validates the create task request
func (s *taskService) validateCreateTaskRequest(req *domain.CreateTaskRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}

	if req.Name == "" {
		return errors.New("task name is required")
	}

	// Validate execution mode
	if !isValidExecutionMode(req.ExecutionMode) {
		return fmt.Errorf("%w: %s", ErrInvalidExecutionMode, req.ExecutionMode)
	}

	// Validate schedule config based on execution mode
	if err := s.validateScheduleConfig(req.ExecutionMode, req.ScheduleConfig); err != nil {
		return err
	}

	// Validate retry policy
	if req.RetryPolicy != nil {
		if req.RetryPolicy.MaxRetries < 0 {
			return errors.New("max retries cannot be negative")
		}
		if req.RetryPolicy.RetryInterval < 0 {
			return errors.New("retry interval cannot be negative")
		}
		if req.RetryPolicy.BackoffFactor < 0 {
			return errors.New("backoff factor cannot be negative")
		}
	}

	// Validate concurrency limit
	if req.ConcurrencyLimit < 0 {
		return errors.New("concurrency limit cannot be negative")
	}

	return nil
}

// validateScheduleConfig validates schedule configuration
func (s *taskService) validateScheduleConfig(mode domain.ExecutionMode, config *domain.ScheduleConfig) error {
	switch mode {
	case domain.ExecutionModeImmediate:
		// No schedule config needed
		return nil

	case domain.ExecutionModeScheduled:
		if config == nil || config.ScheduledTime == nil {
			return fmt.Errorf("%w: scheduled time is required for scheduled execution mode", ErrInvalidScheduleConfig)
		}
		if config.ScheduledTime.Before(time.Now()) {
			return fmt.Errorf("%w: scheduled time cannot be in the past", ErrInvalidScheduleConfig)
		}

	case domain.ExecutionModeInterval:
		if config == nil || config.Interval == nil {
			return fmt.Errorf("%w: interval is required for interval execution mode", ErrInvalidScheduleConfig)
		}
		if *config.Interval <= 0 {
			return fmt.Errorf("%w: interval must be positive", ErrInvalidScheduleConfig)
		}

	case domain.ExecutionModeCron:
		if config == nil || config.CronExpr == nil || *config.CronExpr == "" {
			return fmt.Errorf("%w: cron expression is required for cron execution mode", ErrInvalidScheduleConfig)
		}
		// Basic cron expression validation (can be enhanced)
		// For now, just check it's not empty
	}

	return nil
}

// validateStatusTransition validates if a status transition is allowed
func (s *taskService) validateStatusTransition(from, to domain.TaskStatus) error {
	// Define valid transitions
	validTransitions := map[domain.TaskStatus][]domain.TaskStatus{
		domain.TaskStatusPending: {
			domain.TaskStatusRunning,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusRunning: {
			domain.TaskStatusSuccess,
			domain.TaskStatusFailed,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusFailed: {
			domain.TaskStatusRunning, // For retry
		},
		domain.TaskStatusSuccess:   {},
		domain.TaskStatusCancelled: {},
	}

	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("invalid current status: %s", from)
	}

	for _, allowed := range allowedStatuses {
		if allowed == to {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", from, to)
}

// updateParentTaskStatus updates parent task status based on subtasks
func (s *taskService) updateParentTaskStatus(ctx context.Context, parentID string) error {
	// Get all subtasks
	subtasks, err := s.repo.GetSubTasks(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		return nil
	}

	// Aggregate subtask statuses
	allSuccess := true
	anyFailed := false
	anyRunning := false
	anyCancelled := false

	for _, subtask := range subtasks {
		switch subtask.Status {
		case domain.TaskStatusSuccess:
			// Continue checking
		case domain.TaskStatusFailed:
			anyFailed = true
			allSuccess = false
		case domain.TaskStatusRunning:
			anyRunning = true
			allSuccess = false
		case domain.TaskStatusCancelled:
			anyCancelled = true
			allSuccess = false
		case domain.TaskStatusPending:
			allSuccess = false
		}
	}

	// Determine parent status
	var newStatus domain.TaskStatus
	if allSuccess {
		newStatus = domain.TaskStatusSuccess
	} else if anyFailed {
		newStatus = domain.TaskStatusFailed
	} else if anyRunning {
		newStatus = domain.TaskStatusRunning
	} else if anyCancelled {
		newStatus = domain.TaskStatusCancelled
	} else {
		newStatus = domain.TaskStatusPending
	}

	// Get parent task
	parent, err := s.repo.Get(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	if parent == nil {
		return fmt.Errorf("parent task not found: %s", parentID)
	}

	// Update parent status if changed
	if parent.Status != newStatus {
		parent.Status = newStatus
		parent.UpdatedAt = time.Now()

		if err := s.repo.Update(ctx, parent); err != nil {
			return fmt.Errorf("failed to update parent task status: %w", err)
		}
	}

	return nil
}

// isValidExecutionMode checks if the execution mode is valid
func isValidExecutionMode(mode domain.ExecutionMode) bool {
	switch mode {
	case domain.ExecutionModeImmediate,
		domain.ExecutionModeScheduled,
		domain.ExecutionModeInterval,
		domain.ExecutionModeCron:
		return true
	default:
		return false
	}
}

// isValidTaskStatus checks if the task status is valid
func isValidTaskStatus(status domain.TaskStatus) bool {
	switch status {
	case domain.TaskStatusPending,
		domain.TaskStatusRunning,
		domain.TaskStatusSuccess,
		domain.TaskStatusFailed,
		domain.TaskStatusCancelled:
		return true
	default:
		return false
	}
}
