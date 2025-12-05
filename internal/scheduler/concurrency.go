package scheduler

import (
	"context"
	"fmt"
	"sync"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// ConcurrencyController manages concurrent execution of subtasks
type ConcurrencyController struct {
	repository domain.TaskRepository

	// Map of parent task ID to its concurrency state
	mu     sync.RWMutex
	states map[string]*parentTaskState
}

// parentTaskState tracks the concurrency state for a parent task
type parentTaskState struct {
	mu               sync.Mutex
	parentID         string
	concurrencyLimit int
	runningCount     int
	waitingQueue     []*domain.Task
}

// NewConcurrencyController creates a new ConcurrencyController
func NewConcurrencyController(repository domain.TaskRepository) *ConcurrencyController {
	return &ConcurrencyController{
		repository: repository,
		states:     make(map[string]*parentTaskState),
	}
}

// CanExecuteSubtask checks if a subtask can be executed based on concurrency limits
// Returns true if the subtask can execute immediately, false if it should wait
func (c *ConcurrencyController) CanExecuteSubtask(ctx context.Context, task *domain.Task) (bool, error) {
	// If task has no parent, it's not a subtask - allow execution
	if task.ParentID == nil || *task.ParentID == "" {
		return true, nil
	}

	// Get parent task to check concurrency limit
	parent, err := c.repository.Get(ctx, *task.ParentID)
	if err != nil {
		return false, fmt.Errorf("failed to get parent task: %w", err)
	}

	if parent == nil {
		return false, fmt.Errorf("parent task not found: %s", *task.ParentID)
	}

	// If no concurrency limit is set, allow execution
	if parent.ConcurrencyLimit <= 0 {
		return true, nil
	}

	// Get or create state for this parent task
	state := c.getOrCreateState(*task.ParentID, parent.ConcurrencyLimit)

	state.mu.Lock()
	defer state.mu.Unlock()

	// Check if we're under the limit
	if state.runningCount < state.concurrencyLimit {
		state.runningCount++
		logger.Debug("Subtask allowed to execute",
			zap.String("taskID", task.ID),
			zap.String("parentID", *task.ParentID),
			zap.Int("runningCount", state.runningCount),
			zap.Int("limit", state.concurrencyLimit))
		return true, nil
	}

	// Add to waiting queue
	state.waitingQueue = append(state.waitingQueue, task)
	logger.Debug("Subtask added to waiting queue",
		zap.String("taskID", task.ID),
		zap.String("parentID", *task.ParentID),
		zap.Int("queueLength", len(state.waitingQueue)),
		zap.Int("runningCount", state.runningCount),
		zap.Int("limit", state.concurrencyLimit))

	return false, nil
}

// OnSubtaskCompleted should be called when a subtask completes execution
// Returns the next subtask to execute, if any
func (c *ConcurrencyController) OnSubtaskCompleted(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	// If task has no parent, nothing to do
	if task.ParentID == nil || *task.ParentID == "" {
		return nil, nil
	}

	// Get state for this parent task
	c.mu.RLock()
	state, exists := c.states[*task.ParentID]
	c.mu.RUnlock()

	if !exists {
		// No state means no concurrency control was active
		return nil, nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// Decrement running count
	if state.runningCount > 0 {
		state.runningCount--
	}

	logger.Debug("Subtask completed",
		zap.String("taskID", task.ID),
		zap.String("parentID", *task.ParentID),
		zap.Int("runningCount", state.runningCount),
		zap.Int("queueLength", len(state.waitingQueue)))

	// Check if there are waiting tasks
	if len(state.waitingQueue) == 0 {
		return nil, nil
	}

	// Get next task from queue
	nextTask := state.waitingQueue[0]
	state.waitingQueue = state.waitingQueue[1:]
	state.runningCount++

	logger.Info("Next subtask dequeued for execution",
		zap.String("taskID", nextTask.ID),
		zap.String("parentID", *task.ParentID),
		zap.Int("runningCount", state.runningCount),
		zap.Int("remainingInQueue", len(state.waitingQueue)))

	return nextTask, nil
}

// GetWaitingSubtasks returns all subtasks waiting in the queue for a parent task
func (c *ConcurrencyController) GetWaitingSubtasks(parentID string) []*domain.Task {
	c.mu.RLock()
	state, exists := c.states[parentID]
	c.mu.RUnlock()

	if !exists {
		return nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// Return a copy of the waiting queue
	result := make([]*domain.Task, len(state.waitingQueue))
	copy(result, state.waitingQueue)

	return result
}

// GetRunningCount returns the number of currently running subtasks for a parent task
func (c *ConcurrencyController) GetRunningCount(parentID string) int {
	c.mu.RLock()
	state, exists := c.states[parentID]
	c.mu.RUnlock()

	if !exists {
		return 0
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	return state.runningCount
}

// CleanupParentState removes the state for a parent task (call when parent completes)
func (c *ConcurrencyController) CleanupParentState(parentID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.states, parentID)
	logger.Debug("Cleaned up parent task state", zap.String("parentID", parentID))
}

// getOrCreateState gets or creates the state for a parent task
func (c *ConcurrencyController) getOrCreateState(parentID string, limit int) *parentTaskState {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.states[parentID]
	if !exists {
		state = &parentTaskState{
			parentID:         parentID,
			concurrencyLimit: limit,
			runningCount:     0,
			waitingQueue:     make([]*domain.Task, 0),
		}
		c.states[parentID] = state
		logger.Debug("Created new parent task state",
			zap.String("parentID", parentID),
			zap.Int("limit", limit))
	}

	return state
}
