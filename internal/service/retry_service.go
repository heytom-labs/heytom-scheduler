package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// RetryService handles retry logic for failed tasks
type RetryService struct {
	repo domain.TaskRepository
}

// NewRetryService creates a new RetryService instance
func NewRetryService(repo domain.TaskRepository) *RetryService {
	return &RetryService{
		repo: repo,
	}
}

// ShouldRetry determines if a task should be retried based on its retry policy
func (s *RetryService) ShouldRetry(task *domain.Task) bool {
	if task == nil {
		return false
	}

	// No retry policy means no retry
	if task.RetryPolicy == nil {
		return false
	}

	// Check if max retries is set to 0 (no retry)
	if task.RetryPolicy.MaxRetries == 0 {
		return false
	}

	// Check if we haven't exceeded max retries
	return task.RetryCount < task.RetryPolicy.MaxRetries
}

// CalculateRetryDelay calculates the delay before the next retry attempt
// Supports exponential backoff based on the BackoffFactor
func (s *RetryService) CalculateRetryDelay(task *domain.Task) time.Duration {
	if task == nil || task.RetryPolicy == nil {
		return 0
	}

	baseInterval := task.RetryPolicy.RetryInterval
	backoffFactor := task.RetryPolicy.BackoffFactor

	// If backoff factor is 0 or 1, use fixed interval
	if backoffFactor <= 1.0 {
		return baseInterval
	}

	// Calculate exponential backoff: baseInterval * (backoffFactor ^ retryCount)
	multiplier := math.Pow(backoffFactor, float64(task.RetryCount))
	delay := time.Duration(float64(baseInterval) * multiplier)

	logger.Debug("Calculated retry delay",
		zap.String("taskID", task.ID),
		zap.Int("retryCount", task.RetryCount),
		zap.Duration("baseInterval", baseInterval),
		zap.Float64("backoffFactor", backoffFactor),
		zap.Duration("calculatedDelay", delay))

	return delay
}

// ScheduleRetry schedules a task for retry by updating its status and next execution time
func (s *RetryService) ScheduleRetry(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if !s.ShouldRetry(task) {
		return fmt.Errorf("task should not be retried")
	}

	// Increment retry count
	task.RetryCount++

	// Calculate retry delay
	retryDelay := s.CalculateRetryDelay(task)

	// Update task status to pending for retry
	task.Status = domain.TaskStatusPending
	task.UpdatedAt = time.Now()

	// For scheduled tasks, update the scheduled time
	if task.ScheduleConfig == nil {
		task.ScheduleConfig = &domain.ScheduleConfig{}
	}

	nextRetryTime := time.Now().Add(retryDelay)
	task.ScheduleConfig.ScheduledTime = &nextRetryTime

	// Persist the updated task
	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to schedule retry: %w", err)
	}

	logger.Info("Task scheduled for retry",
		zap.String("taskID", task.ID),
		zap.Int("retryCount", task.RetryCount),
		zap.Int("maxRetries", task.RetryPolicy.MaxRetries),
		zap.Duration("retryDelay", retryDelay),
		zap.Time("nextRetryTime", nextRetryTime))

	return nil
}

// HandleTaskFailure handles a failed task by determining if it should be retried
// Returns true if the task was scheduled for retry, false otherwise
func (s *RetryService) HandleTaskFailure(ctx context.Context, task *domain.Task, err error) (bool, error) {
	if task == nil {
		return false, fmt.Errorf("task cannot be nil")
	}

	logger.Info("Handling task failure",
		zap.String("taskID", task.ID),
		zap.Error(err),
		zap.Int("retryCount", task.RetryCount))

	// Check if we should retry
	if s.ShouldRetry(task) {
		// Schedule retry
		if retryErr := s.ScheduleRetry(ctx, task); retryErr != nil {
			logger.Error("Failed to schedule retry",
				zap.String("taskID", task.ID),
				zap.Error(retryErr))
			return false, retryErr
		}
		return true, nil
	}

	// No more retries, mark as failed
	task.Status = domain.TaskStatusFailed
	task.UpdatedAt = time.Now()
	now := time.Now()
	if task.CompletedAt == nil {
		task.CompletedAt = &now
	}

	if updateErr := s.repo.Update(ctx, task); updateErr != nil {
		logger.Error("Failed to update task status to failed",
			zap.String("taskID", task.ID),
			zap.Error(updateErr))
		return false, updateErr
	}

	logger.Info("Task marked as failed (no more retries)",
		zap.String("taskID", task.ID),
		zap.Int("retryCount", task.RetryCount))

	return false, nil
}

// GetRetryInfo returns information about the retry status of a task
func (s *RetryService) GetRetryInfo(task *domain.Task) map[string]interface{} {
	if task == nil {
		return nil
	}

	info := map[string]interface{}{
		"retryCount": task.RetryCount,
		"canRetry":   s.ShouldRetry(task),
	}

	if task.RetryPolicy != nil {
		info["maxRetries"] = task.RetryPolicy.MaxRetries
		info["retryInterval"] = task.RetryPolicy.RetryInterval.String()
		info["backoffFactor"] = task.RetryPolicy.BackoffFactor

		if s.ShouldRetry(task) {
			nextDelay := s.CalculateRetryDelay(task)
			info["nextRetryDelay"] = nextDelay.String()
		}
	}

	return info
}
