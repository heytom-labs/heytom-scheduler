package node

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// AllocationStrategy defines how tasks are allocated to nodes
type AllocationStrategy string

const (
	AllocationStrategyRoundRobin  AllocationStrategy = "round_robin"
	AllocationStrategyRandom      AllocationStrategy = "random"
	AllocationStrategyLeastLoaded AllocationStrategy = "least_loaded"
)

// Allocator handles task allocation to nodes
type Allocator struct {
	registry   Registry
	repository domain.TaskRepository
	lock       domain.DistributedLock
	strategy   AllocationStrategy

	// For round-robin strategy
	lastNodeIndex int
}

// NewAllocator creates a new task allocator
func NewAllocator(
	registry Registry,
	repository domain.TaskRepository,
	lock domain.DistributedLock,
	strategy AllocationStrategy,
) *Allocator {
	return &Allocator{
		registry:      registry,
		repository:    repository,
		lock:          lock,
		strategy:      strategy,
		lastNodeIndex: 0,
	}
}

// AllocateTask allocates a task to a node
func (a *Allocator) AllocateTask(ctx context.Context, task *domain.Task) (*Node, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	// Get healthy nodes
	nodes, err := a.registry.ListHealthyNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list healthy nodes: %w", err)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	// Select a node based on strategy
	var selectedNode *Node
	switch a.strategy {
	case AllocationStrategyRoundRobin:
		selectedNode = a.selectRoundRobin(nodes)
	case AllocationStrategyRandom:
		selectedNode = a.selectRandom(nodes)
	case AllocationStrategyLeastLoaded:
		selectedNode, err = a.selectLeastLoaded(ctx, nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to select least loaded node: %w", err)
		}
	default:
		selectedNode = a.selectRandom(nodes)
	}

	// Try to acquire lock for the task
	lockKey := fmt.Sprintf("task:lock:%s", task.ID)
	acquired, err := a.lock.Lock(ctx, lockKey, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return nil, fmt.Errorf("task is already locked")
	}

	// Update task with node assignment
	task.NodeID = &selectedNode.ID
	task.Status = domain.TaskStatusRunning
	task.UpdatedAt = time.Now()
	now := time.Now()
	task.StartedAt = &now

	if err := a.repository.Update(ctx, task); err != nil {
		// Release lock if update fails
		a.lock.Unlock(ctx, lockKey)
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	logger.Info("Task allocated to node",
		zap.String("taskID", task.ID),
		zap.String("nodeID", selectedNode.ID))

	return selectedNode, nil
}

// ReleaseTask releases a task from a node
func (a *Allocator) ReleaseTask(ctx context.Context, taskID string) error {
	lockKey := fmt.Sprintf("task:lock:%s", taskID)

	if err := a.lock.Unlock(ctx, lockKey); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	logger.Debug("Task released", zap.String("taskID", taskID))

	return nil
}

// selectRoundRobin selects a node using round-robin strategy
func (a *Allocator) selectRoundRobin(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	a.lastNodeIndex = (a.lastNodeIndex + 1) % len(nodes)
	return nodes[a.lastNodeIndex]
}

// selectRandom selects a node randomly
func (a *Allocator) selectRandom(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	return nodes[rand.Intn(len(nodes))]
}

// selectLeastLoaded selects the node with the least number of running tasks
func (a *Allocator) selectLeastLoaded(ctx context.Context, nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Count running tasks per node
	nodeTasks := make(map[string]int)
	for _, node := range nodes {
		nodeTasks[node.ID] = 0
	}

	// Get all running tasks
	filter := &domain.TaskFilter{
		Status: func() *domain.TaskStatus {
			status := domain.TaskStatusRunning
			return &status
		}(),
	}

	tasks, err := a.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list running tasks: %w", err)
	}

	// Count tasks per node
	for _, task := range tasks {
		if task.NodeID != nil {
			if _, exists := nodeTasks[*task.NodeID]; exists {
				nodeTasks[*task.NodeID]++
			}
		}
	}

	// Find node with least tasks
	var leastLoadedNode *Node
	minTasks := int(^uint(0) >> 1) // Max int

	for _, node := range nodes {
		taskCount := nodeTasks[node.ID]
		if taskCount < minTasks {
			minTasks = taskCount
			leastLoadedNode = node
		}
	}

	return leastLoadedNode, nil
}

// GetNodeTasks returns all tasks assigned to a specific node
func (a *Allocator) GetNodeTasks(ctx context.Context, nodeID string) ([]*domain.Task, error) {
	filter := &domain.TaskFilter{
		Status: func() *domain.TaskStatus {
			status := domain.TaskStatusRunning
			return &status
		}(),
	}

	allTasks, err := a.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Filter tasks for this node
	nodeTasks := make([]*domain.Task, 0)
	for _, task := range allTasks {
		if task.NodeID != nil && *task.NodeID == nodeID {
			nodeTasks = append(nodeTasks, task)
		}
	}

	return nodeTasks, nil
}
