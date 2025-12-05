package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// FailoverManager handles node failure detection and task transfer
type FailoverManager struct {
	registry   Registry
	repository domain.TaskRepository
	allocator  *Allocator

	checkInterval time.Duration

	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(
	registry Registry,
	repository domain.TaskRepository,
	allocator *Allocator,
	checkInterval time.Duration,
) *FailoverManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &FailoverManager{
		registry:      registry,
		repository:    repository,
		allocator:     allocator,
		checkInterval: checkInterval,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the failover manager
func (f *FailoverManager) Start() error {
	f.ticker = time.NewTicker(f.checkInterval)

	f.wg.Add(1)
	go f.failoverLoop()

	logger.Info("Failover manager started",
		zap.Duration("checkInterval", f.checkInterval))

	return nil
}

// Stop stops the failover manager
func (f *FailoverManager) Stop() error {
	f.cancel()

	if f.ticker != nil {
		f.ticker.Stop()
	}

	f.wg.Wait()

	logger.Info("Failover manager stopped")

	return nil
}

// failoverLoop periodically checks for failed nodes and transfers tasks
func (f *FailoverManager) failoverLoop() {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-f.ticker.C:
			if err := f.checkAndTransferTasks(); err != nil {
				logger.Error("Failed to check and transfer tasks", zap.Error(err))
			}
		}
	}
}

// checkAndTransferTasks checks for unhealthy nodes and transfers their tasks
func (f *FailoverManager) checkAndTransferTasks() error {
	ctx := f.ctx

	// Get all nodes
	nodes, err := f.registry.ListNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// Find unhealthy nodes
	unhealthyNodes := make([]*Node, 0)
	for _, node := range nodes {
		if node.Status == NodeStatusUnhealthy {
			unhealthyNodes = append(unhealthyNodes, node)
		}
	}

	if len(unhealthyNodes) == 0 {
		return nil
	}

	logger.Info("Found unhealthy nodes",
		zap.Int("count", len(unhealthyNodes)))

	// Transfer tasks from unhealthy nodes
	for _, node := range unhealthyNodes {
		if err := f.transferNodeTasks(ctx, node.ID); err != nil {
			logger.Error("Failed to transfer tasks from node",
				zap.String("nodeID", node.ID),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// transferNodeTasks transfers all tasks from a failed node to healthy nodes
func (f *FailoverManager) transferNodeTasks(ctx context.Context, nodeID string) error {
	// Get all tasks assigned to this node
	tasks, err := f.allocator.GetNodeTasks(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node tasks: %w", err)
	}

	if len(tasks) == 0 {
		logger.Debug("No tasks to transfer", zap.String("nodeID", nodeID))
		return nil
	}

	logger.Info("Transferring tasks from failed node",
		zap.String("nodeID", nodeID),
		zap.Int("taskCount", len(tasks)))

	// Transfer each task
	transferred := 0
	failed := 0

	for _, task := range tasks {
		if err := f.transferTask(ctx, task); err != nil {
			logger.Error("Failed to transfer task",
				zap.String("taskID", task.ID),
				zap.String("fromNode", nodeID),
				zap.Error(err))
			failed++
			continue
		}
		transferred++
	}

	logger.Info("Task transfer completed",
		zap.String("nodeID", nodeID),
		zap.Int("transferred", transferred),
		zap.Int("failed", failed))

	return nil
}

// transferTask transfers a single task to a healthy node
func (f *FailoverManager) transferTask(ctx context.Context, task *domain.Task) error {
	// Release the old lock first
	if err := f.allocator.ReleaseTask(ctx, task.ID); err != nil {
		logger.Warn("Failed to release old lock",
			zap.String("taskID", task.ID),
			zap.Error(err))
		// Continue anyway
	}

	// Reset task to pending state for reallocation
	task.Status = domain.TaskStatusPending
	task.NodeID = nil
	task.UpdatedAt = time.Now()

	if err := f.repository.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Allocate to a new node
	newNode, err := f.allocator.AllocateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to allocate task to new node: %w", err)
	}

	logger.Info("Task transferred",
		zap.String("taskID", task.ID),
		zap.String("newNodeID", newNode.ID))

	return nil
}

// ForceTransferTask manually transfers a task from one node to another
func (f *FailoverManager) ForceTransferTask(ctx context.Context, taskID string, targetNodeID string) error {
	// Get the task
	task, err := f.repository.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Verify target node is healthy
	targetNode, err := f.registry.GetNode(ctx, targetNodeID)
	if err != nil {
		return fmt.Errorf("failed to get target node: %w", err)
	}

	if targetNode.Status != NodeStatusHealthy {
		return fmt.Errorf("target node is not healthy")
	}

	// Release old lock
	if err := f.allocator.ReleaseTask(ctx, taskID); err != nil {
		logger.Warn("Failed to release old lock",
			zap.String("taskID", taskID),
			zap.Error(err))
	}

	// Update task with new node
	task.NodeID = &targetNodeID
	task.Status = domain.TaskStatusRunning
	task.UpdatedAt = time.Now()

	if err := f.repository.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	logger.Info("Task forcefully transferred",
		zap.String("taskID", taskID),
		zap.String("targetNodeID", targetNodeID))

	return nil
}

// GetFailedNodeTasks returns all tasks from failed nodes
func (f *FailoverManager) GetFailedNodeTasks(ctx context.Context) ([]*domain.Task, error) {
	// Get all unhealthy nodes
	nodes, err := f.registry.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	failedTasks := make([]*domain.Task, 0)

	for _, node := range nodes {
		if node.Status == NodeStatusUnhealthy {
			tasks, err := f.allocator.GetNodeTasks(ctx, node.ID)
			if err != nil {
				logger.Error("Failed to get tasks for failed node",
					zap.String("nodeID", node.ID),
					zap.Error(err))
				continue
			}
			failedTasks = append(failedTasks, tasks...)
		}
	}

	return failedTasks, nil
}
