package service

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	"github.com/robfig/cron/v3"
)

// scheduleService implements the ScheduleService interface
type scheduleService struct {
	cronParser cron.Parser
}

// NewScheduleService creates a new ScheduleService instance
func NewScheduleService() domain.ScheduleService {
	return &scheduleService{
		// Use standard cron parser with seconds support
		cronParser: cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
	}
}

// ScheduleTask schedules a task for execution
func (s *scheduleService) ScheduleTask(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	// Validate execution mode and schedule config
	switch task.ExecutionMode {
	case domain.ExecutionModeImmediate:
		// No validation needed for immediate execution
		return nil

	case domain.ExecutionModeScheduled:
		if task.ScheduleConfig == nil || task.ScheduleConfig.ScheduledTime == nil {
			return fmt.Errorf("scheduled time is required for scheduled execution mode")
		}
		if task.ScheduleConfig.ScheduledTime.Before(time.Now()) {
			return fmt.Errorf("scheduled time cannot be in the past")
		}
		return nil

	case domain.ExecutionModeInterval:
		if task.ScheduleConfig == nil || task.ScheduleConfig.Interval == nil {
			return fmt.Errorf("interval is required for interval execution mode")
		}
		if *task.ScheduleConfig.Interval <= 0 {
			return fmt.Errorf("interval must be positive")
		}
		return nil

	case domain.ExecutionModeCron:
		if task.ScheduleConfig == nil || task.ScheduleConfig.CronExpr == nil {
			return fmt.Errorf("cron expression is required for cron execution mode")
		}
		// Validate cron expression
		_, err := s.cronParser.Parse(*task.ScheduleConfig.CronExpr)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown execution mode: %s", task.ExecutionMode)
	}
}

// UnscheduleTask removes a task from the schedule
func (s *scheduleService) UnscheduleTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	// This is a placeholder - actual implementation would interact with scheduler
	return nil
}

// GetNextExecutionTime calculates the next execution time for a task
func (s *scheduleService) GetNextExecutionTime(task *domain.Task) (time.Time, error) {
	if task == nil {
		return time.Time{}, fmt.Errorf("task cannot be nil")
	}

	if task.ScheduleConfig == nil {
		return time.Time{}, fmt.Errorf("schedule config is required")
	}

	now := time.Now()

	switch task.ExecutionMode {
	case domain.ExecutionModeImmediate:
		// Immediate tasks execute right away
		return now, nil

	case domain.ExecutionModeScheduled:
		if task.ScheduleConfig.ScheduledTime == nil {
			return time.Time{}, fmt.Errorf("scheduled time is required for scheduled execution mode")
		}
		return *task.ScheduleConfig.ScheduledTime, nil

	case domain.ExecutionModeInterval:
		if task.ScheduleConfig.Interval == nil {
			return time.Time{}, fmt.Errorf("interval is required for interval execution mode")
		}
		// For interval tasks, calculate next execution based on last execution
		// If task hasn't started yet, use creation time
		baseTime := task.CreatedAt
		if task.StartedAt != nil {
			baseTime = *task.StartedAt
		}
		if task.CompletedAt != nil {
			baseTime = *task.CompletedAt
		}

		nextTime := baseTime.Add(*task.ScheduleConfig.Interval)
		// If next time is in the past, calculate from now
		if nextTime.Before(now) {
			nextTime = now.Add(*task.ScheduleConfig.Interval)
		}
		return nextTime, nil

	case domain.ExecutionModeCron:
		if task.ScheduleConfig.CronExpr == nil {
			return time.Time{}, fmt.Errorf("cron expression is required for cron execution mode")
		}

		// Parse cron expression
		schedule, err := s.cronParser.Parse(*task.ScheduleConfig.CronExpr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
		}

		// Calculate next execution time from now
		nextTime := schedule.Next(now)
		return nextTime, nil

	default:
		return time.Time{}, fmt.Errorf("unknown execution mode: %s", task.ExecutionMode)
	}
}
