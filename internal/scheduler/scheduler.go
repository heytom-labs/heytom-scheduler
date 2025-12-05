package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/internal/service"
	"task-scheduler/pkg/logger"
	"task-scheduler/pkg/metrics"

	"go.uber.org/zap"
)

// Config contains scheduler configuration
type Config struct {
	WorkerCount       int           // Number of worker goroutines
	QueueSize         int           // Size of task queue
	PollInterval      time.Duration // Interval for polling scheduled tasks
	LockTTL           time.Duration // TTL for distributed locks
	HeartbeatInterval time.Duration // Interval for node heartbeat
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() *Config {
	return &Config{
		WorkerCount:       10,
		QueueSize:         1000,
		PollInterval:      time.Second,
		LockTTL:           time.Minute,
		HeartbeatInterval: 10 * time.Second,
	}
}

// NotificationService defines the interface for sending alerts
type NotificationService interface {
	EvaluateAndSendAlert(ctx context.Context, task *domain.Task, reason string, errorMsg *string)
}

// scheduler implements the Scheduler interface
type scheduler struct {
	config                *Config
	repository            domain.TaskRepository
	scheduleService       domain.ScheduleService
	callbackService       domain.CallbackService
	distributedLock       domain.DistributedLock
	retryService          *service.RetryService
	concurrencyController *ConcurrencyController
	notificationService   NotificationService
	nodeID                string

	taskQueue chan *domain.Task
	workers   []*worker
	ticker    *time.Ticker

	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// worker represents a task execution worker
type worker struct {
	id        int
	scheduler *scheduler
	ctx       context.Context
}

// NewScheduler creates a new Scheduler instance
func NewScheduler(
	config *Config,
	repository domain.TaskRepository,
	scheduleService domain.ScheduleService,
	callbackService domain.CallbackService,
	distributedLock domain.DistributedLock,
	nodeID string,
) domain.Scheduler {
	if config == nil {
		config = DefaultConfig()
	}

	return &scheduler{
		config:                config,
		repository:            repository,
		scheduleService:       scheduleService,
		callbackService:       callbackService,
		distributedLock:       distributedLock,
		retryService:          service.NewRetryService(repository),
		concurrencyController: NewConcurrencyController(repository),
		notificationService:   nil, // Will be set via SetNotificationService
		nodeID:                nodeID,
		taskQueue:             make(chan *domain.Task, config.QueueSize),
		workers:               make([]*worker, 0, config.WorkerCount),
	}
}

// SetNotificationService sets the notification service for the scheduler
func (s *scheduler) SetNotificationService(ns NotificationService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notificationService = ns
}

// Start starts the scheduler
func (s *scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	// Start worker pool
	for i := 0; i < s.config.WorkerCount; i++ {
		w := &worker{
			id:        i,
			scheduler: s,
			ctx:       s.ctx,
		}
		s.workers = append(s.workers, w)
		s.wg.Add(1)
		go w.run()
	}

	// Start task polling goroutine
	s.ticker = time.NewTicker(s.config.PollInterval)
	s.wg.Add(1)
	go s.pollTasks()

	logger.Info("Scheduler started",
		zap.String("nodeID", s.nodeID),
		zap.Int("workerCount", s.config.WorkerCount))

	// Record metrics
	metrics.SetActiveWorkersCount(s.config.WorkerCount)

	return nil
}

// Stop stops the scheduler
func (s *scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.running = false
	s.mu.Unlock()

	// Stop ticker
	if s.ticker != nil {
		s.ticker.Stop()
	}

	// Cancel context to stop all workers
	if s.cancel != nil {
		s.cancel()
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Close task queue
	close(s.taskQueue)

	logger.Info("Scheduler stopped", zap.String("nodeID", s.nodeID))

	return nil
}

// SubmitTask submits a task for execution
func (s *scheduler) SubmitTask(task *domain.Task) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	// Validate task can be scheduled
	if err := s.scheduleService.ScheduleTask(s.ctx, task); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	// For immediate execution, add to queue directly
	if task.ExecutionMode == domain.ExecutionModeImmediate {
		select {
		case s.taskQueue <- task:
			logger.Debug("Task submitted to queue",
				zap.String("taskID", task.ID),
				zap.String("mode", string(task.ExecutionMode)))
			// Update queue size metric
			metrics.SetTaskQueueSize(len(s.taskQueue))
			return nil
		case <-s.ctx.Done():
			return fmt.Errorf("scheduler is shutting down")
		default:
			return fmt.Errorf("task queue is full")
		}
	}

	// For scheduled tasks, they will be picked up by pollTasks
	logger.Debug("Task scheduled",
		zap.String("taskID", task.ID),
		zap.String("mode", string(task.ExecutionMode)))

	return nil
}

// AcquireTask acquires a task for execution by a node
func (s *scheduler) AcquireTask(nodeID string) (*domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil, fmt.Errorf("scheduler is not running")
	}

	// Get pending tasks
	filter := &domain.TaskFilter{
		Status: func() *domain.TaskStatus {
			status := domain.TaskStatusPending
			return &status
		}(),
		Limit: 10,
	}

	tasks, err := s.repository.List(s.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending tasks: %w", err)
	}

	// Try to acquire a task with distributed lock
	for _, task := range tasks {
		lockKey := fmt.Sprintf("task:lock:%s", task.ID)
		acquired, err := s.distributedLock.Lock(s.ctx, lockKey, s.config.LockTTL)
		if err != nil {
			logger.Warn("Failed to acquire lock",
				zap.String("taskID", task.ID),
				zap.Error(err))
			continue
		}

		if acquired {
			// Update task with node assignment
			task.NodeID = &nodeID
			task.Status = domain.TaskStatusRunning
			task.UpdatedAt = time.Now()
			now := time.Now()
			task.StartedAt = &now

			if err := s.repository.Update(s.ctx, task); err != nil {
				// Release lock if update fails
				s.distributedLock.Unlock(s.ctx, lockKey)
				logger.Error("Failed to update task",
					zap.String("taskID", task.ID),
					zap.Error(err))
				continue
			}

			logger.Info("Task acquired",
				zap.String("taskID", task.ID),
				zap.String("nodeID", nodeID))

			return task, nil
		}
	}

	return nil, nil // No task available
}

// ReleaseTask releases a task after execution
func (s *scheduler) ReleaseTask(taskID string) error {
	lockKey := fmt.Sprintf("task:lock:%s", taskID)
	if err := s.distributedLock.Unlock(s.ctx, lockKey); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	logger.Debug("Task released", zap.String("taskID", taskID))
	return nil
}

// pollTasks periodically checks for scheduled tasks that are ready to execute
func (s *scheduler) pollTasks() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.checkScheduledTasks()
		}
	}
}

// checkScheduledTasks checks for tasks that are ready to execute
func (s *scheduler) checkScheduledTasks() {
	// Get pending tasks
	filter := &domain.TaskFilter{
		Status: func() *domain.TaskStatus {
			status := domain.TaskStatusPending
			return &status
		}(),
		Limit: 100,
	}

	tasks, err := s.repository.List(s.ctx, filter)
	if err != nil {
		logger.Error("Failed to list pending tasks", zap.Error(err))
		return
	}

	now := time.Now()

	for _, task := range tasks {
		// Skip immediate tasks (already handled in SubmitTask)
		if task.ExecutionMode == domain.ExecutionModeImmediate {
			continue
		}

		// Check if task is ready to execute
		nextTime, err := s.scheduleService.GetNextExecutionTime(task)
		if err != nil {
			logger.Error("Failed to get next execution time",
				zap.String("taskID", task.ID),
				zap.Error(err))
			continue
		}

		// If next execution time has passed, add to queue
		if nextTime.Before(now) || nextTime.Equal(now) {
			select {
			case s.taskQueue <- task:
				logger.Debug("Scheduled task added to queue",
					zap.String("taskID", task.ID),
					zap.Time("scheduledTime", nextTime))
			case <-s.ctx.Done():
				return
			default:
				logger.Warn("Task queue is full, skipping task",
					zap.String("taskID", task.ID))
			}
		}
	}
}

// worker.run executes tasks from the queue
func (w *worker) run() {
	defer w.scheduler.wg.Done()

	logger.Debug("Worker started", zap.Int("workerID", w.id))

	for {
		select {
		case <-w.ctx.Done():
			logger.Debug("Worker stopped", zap.Int("workerID", w.id))
			return
		case task, ok := <-w.scheduler.taskQueue:
			if !ok {
				logger.Debug("Worker stopped (queue closed)", zap.Int("workerID", w.id))
				return
			}
			w.executeTask(task)
		}
	}
}

// executeTask executes a single task
func (w *worker) executeTask(task *domain.Task) {
	startTime := time.Now()

	logger.Info("Executing task",
		zap.Int("workerID", w.id),
		zap.String("taskID", task.ID),
		zap.String("name", task.Name))

	// Record task execution start
	metrics.RecordTaskExecuted("started")

	// Check concurrency limits for subtasks
	canExecute, err := w.scheduler.concurrencyController.CanExecuteSubtask(w.ctx, task)
	if err != nil {
		logger.Error("Failed to check concurrency limits",
			zap.String("taskID", task.ID),
			zap.Error(err))
		return
	}

	if !canExecute {
		// Task was added to waiting queue, don't execute now
		logger.Info("Task added to waiting queue due to concurrency limit",
			zap.String("taskID", task.ID))
		return
	}

	// Transition to running state
	if err := w.transitionTaskStatus(task, domain.TaskStatusRunning); err != nil {
		logger.Error("Failed to transition task to running",
			zap.String("taskID", task.ID),
			zap.Error(err))
		return
	}

	// Execute callback if configured
	var executionErr error
	if task.CallbackConfig != nil {
		result := &domain.ExecutionResult{
			TaskID:    task.ID,
			Status:    domain.TaskStatusRunning,
			StartTime: time.Now(),
		}

		if task.CallbackConfig.IsAsync {
			executionErr = w.scheduler.callbackService.ExecuteAsyncCallback(w.ctx, task)
		} else {
			executionErr = w.scheduler.callbackService.ExecuteCallback(w.ctx, task, result)
		}

		result.EndTime = time.Now()

		if executionErr != nil {
			errMsg := executionErr.Error()
			result.Error = &errMsg
			result.Status = domain.TaskStatusFailed
		} else {
			result.Status = domain.TaskStatusSuccess
		}
	}

	// Determine final status
	var finalStatus domain.TaskStatus
	if executionErr != nil {
		errMsg := executionErr.Error()

		// Check for timeout condition
		if task.AlertPolicy != nil && task.AlertPolicy.TimeoutThreshold > 0 && task.StartedAt != nil {
			elapsed := time.Since(*task.StartedAt)
			if elapsed >= task.AlertPolicy.TimeoutThreshold {
				// Send timeout alert
				if w.scheduler.notificationService != nil {
					w.scheduler.notificationService.EvaluateAndSendAlert(w.ctx, task, "timeout", &errMsg)
				}
			}
		}

		// Use retry service to handle failure
		retried, retryErr := w.scheduler.retryService.HandleTaskFailure(w.ctx, task, executionErr)
		if retryErr != nil {
			logger.Error("Failed to handle task failure",
				zap.String("taskID", task.ID),
				zap.Error(retryErr))
			// Still try to mark as failed
			finalStatus = domain.TaskStatusFailed
			if err := w.transitionTaskStatus(task, finalStatus); err != nil {
				logger.Error("Failed to transition task to failed status",
					zap.String("taskID", task.ID),
					zap.Error(err))
			}
			// Send failure alert
			if w.scheduler.notificationService != nil {
				w.scheduler.notificationService.EvaluateAndSendAlert(w.ctx, task, "failure", &errMsg)
			}
			return
		}

		if retried {
			// Task was scheduled for retry, no need to transition status
			// The retry service already updated the task status to pending
			logger.Info("Task scheduled for retry",
				zap.String("taskID", task.ID),
				zap.Int("retryCount", task.RetryCount))

			// Record retry metrics
			metrics.RecordTaskRetry(string(task.ExecutionMode))

			// Check if retry threshold is exceeded and send alert
			if w.scheduler.notificationService != nil {
				w.scheduler.notificationService.EvaluateAndSendAlert(w.ctx, task, "retry_threshold", &errMsg)
			}

			// Release the task lock so it can be picked up again
			if err := w.scheduler.ReleaseTask(task.ID); err != nil {
				logger.Error("Failed to release task",
					zap.String("taskID", task.ID),
					zap.Error(err))
			}

			// Check if there's a waiting subtask to execute (for failed subtasks)
			nextTask, err := w.scheduler.concurrencyController.OnSubtaskCompleted(w.ctx, task)
			if err != nil {
				logger.Error("Failed to get next waiting subtask",
					zap.String("taskID", task.ID),
					zap.Error(err))
			} else if nextTask != nil {
				// Submit the next waiting task to the queue
				select {
				case w.scheduler.taskQueue <- nextTask:
					logger.Info("Next waiting subtask submitted to queue after retry",
						zap.String("nextTaskID", nextTask.ID),
						zap.String("retriedTaskID", task.ID))
				case <-w.ctx.Done():
					logger.Warn("Cannot submit next task, scheduler is shutting down",
						zap.String("nextTaskID", nextTask.ID))
				default:
					logger.Warn("Task queue is full, cannot submit next waiting task",
						zap.String("nextTaskID", nextTask.ID))
				}
			}
			return
		}

		// No retry, task is already marked as failed by retry service
		logger.Info("Task execution failed (no more retries)",
			zap.String("taskID", task.ID))

		// Record failure metrics
		metrics.RecordTaskFailure(string(task.ExecutionMode))
		duration := time.Since(startTime).Seconds()
		metrics.RecordTaskExecutionDuration(string(task.ExecutionMode), string(domain.TaskStatusFailed), duration)

		// Send failure alert (task permanently failed)
		if w.scheduler.notificationService != nil {
			w.scheduler.notificationService.EvaluateAndSendAlert(w.ctx, task, "failure", &errMsg)
		}

		// Release the task lock
		if err := w.scheduler.ReleaseTask(task.ID); err != nil {
			logger.Error("Failed to release task",
				zap.String("taskID", task.ID),
				zap.Error(err))
		}

		// Check if there's a waiting subtask to execute (for permanently failed subtasks)
		nextTask, err := w.scheduler.concurrencyController.OnSubtaskCompleted(w.ctx, task)
		if err != nil {
			logger.Error("Failed to get next waiting subtask",
				zap.String("taskID", task.ID),
				zap.Error(err))
		} else if nextTask != nil {
			// Submit the next waiting task to the queue
			select {
			case w.scheduler.taskQueue <- nextTask:
				logger.Info("Next waiting subtask submitted to queue after failure",
					zap.String("nextTaskID", nextTask.ID),
					zap.String("failedTaskID", task.ID))
			case <-w.ctx.Done():
				logger.Warn("Cannot submit next task, scheduler is shutting down",
					zap.String("nextTaskID", nextTask.ID))
			default:
				logger.Warn("Task queue is full, cannot submit next waiting task",
					zap.String("nextTaskID", nextTask.ID))
			}
		}
		return
	} else {
		finalStatus = domain.TaskStatusSuccess
	}

	// Transition to final state (success case)
	if err := w.transitionTaskStatus(task, finalStatus); err != nil {
		logger.Error("Failed to transition task to final status",
			zap.String("taskID", task.ID),
			zap.String("status", string(finalStatus)),
			zap.Error(err))
		return
	}

	// Release the task lock
	if err := w.scheduler.ReleaseTask(task.ID); err != nil {
		logger.Error("Failed to release task",
			zap.String("taskID", task.ID),
			zap.Error(err))
	}

	// Check if there's a waiting subtask to execute
	nextTask, err := w.scheduler.concurrencyController.OnSubtaskCompleted(w.ctx, task)
	if err != nil {
		logger.Error("Failed to get next waiting subtask",
			zap.String("taskID", task.ID),
			zap.Error(err))
	} else if nextTask != nil {
		// Submit the next waiting task to the queue
		select {
		case w.scheduler.taskQueue <- nextTask:
			logger.Info("Next waiting subtask submitted to queue",
				zap.String("nextTaskID", nextTask.ID),
				zap.String("completedTaskID", task.ID))
		case <-w.ctx.Done():
			logger.Warn("Cannot submit next task, scheduler is shutting down",
				zap.String("nextTaskID", nextTask.ID))
		default:
			logger.Warn("Task queue is full, cannot submit next waiting task",
				zap.String("nextTaskID", nextTask.ID))
		}
	}

	logger.Info("Task execution completed",
		zap.String("taskID", task.ID),
		zap.String("status", string(finalStatus)))

	// Record metrics
	duration := time.Since(startTime).Seconds()
	metrics.RecordTaskCompleted(string(finalStatus))
	metrics.RecordTaskExecutionDuration(string(task.ExecutionMode), string(finalStatus), duration)
}

// transitionTaskStatus transitions a task to a new status
func (w *worker) transitionTaskStatus(task *domain.Task, newStatus domain.TaskStatus) error {
	// Validate state transition
	if !isValidTransition(task.Status, newStatus) {
		return fmt.Errorf("invalid state transition from %s to %s", task.Status, newStatus)
	}

	// Update task status
	task.Status = newStatus
	task.UpdatedAt = time.Now()

	// Set timestamps based on status
	now := time.Now()
	switch newStatus {
	case domain.TaskStatusRunning:
		if task.StartedAt == nil {
			task.StartedAt = &now
		}
	case domain.TaskStatusSuccess, domain.TaskStatusFailed, domain.TaskStatusCancelled:
		if task.CompletedAt == nil {
			task.CompletedAt = &now
		}
	}

	// Persist to repository
	if err := w.scheduler.repository.Update(w.ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	logger.Debug("Task status transitioned",
		zap.String("taskID", task.ID),
		zap.String("status", string(newStatus)))

	return nil
}

// isValidTransition checks if a state transition is valid
func isValidTransition(from, to domain.TaskStatus) bool {
	// Valid transitions:
	// pending -> running
	// pending -> cancelled
	// running -> success
	// running -> failed
	// running -> cancelled
	// failed -> pending (retry)

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
			domain.TaskStatusPending, // For retry
		},
	}

	allowedTransitions, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}

	return false
}
