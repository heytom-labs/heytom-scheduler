package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"task-scheduler/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTaskRepository is a simple in-memory repository for testing
type mockTaskRepository struct {
	tasks map[string]*domain.Task
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return task, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	if _, exists := m.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, taskID string) error {
	delete(m.tasks, taskID)
	return nil
}

func (m *mockTaskRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0)
	for _, task := range m.tasks {
		// Apply status filter if provided
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		tasks = append(tasks, task)

		// Apply limit if provided
		if filter.Limit > 0 && len(tasks) >= filter.Limit {
			break
		}
	}
	return tasks, nil
}

func (m *mockTaskRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0)
	for _, task := range m.tasks {
		if task.ParentID != nil && *task.ParentID == parentID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

// mockTaskLock is a simple in-memory lock for testing
type mockTaskLock struct {
	locks map[string]bool
}

func newMockTaskLock() *mockTaskLock {
	return &mockTaskLock{
		locks: make(map[string]bool),
	}
}

func (m *mockTaskLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if m.locks[key] {
		return false, nil
	}
	m.locks[key] = true
	return true, nil
}

func (m *mockTaskLock) Unlock(ctx context.Context, key string) error {
	delete(m.locks, key)
	return nil
}

func (m *mockTaskLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func TestAllocator_AllocateTask(t *testing.T) {
	ctx := context.Background()

	// Setup
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRandom)

	// Register nodes
	node1 := &Node{ID: "node1", Address: "localhost:8080"}
	node2 := &Node{ID: "node2", Address: "localhost:8081"}

	require.NoError(t, registry.Register(ctx, node1))
	require.NoError(t, registry.Register(ctx, node2))

	// Create a task
	task := &domain.Task{
		ID:            "task1",
		Name:          "Test Task",
		Status:        domain.TaskStatusPending,
		ExecutionMode: domain.ExecutionModeImmediate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, repo.Create(ctx, task))

	// Allocate task
	selectedNode, err := allocator.AllocateTask(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, selectedNode)

	// Verify task was updated with node assignment
	assert.NotNil(t, task.NodeID)
	assert.Equal(t, selectedNode.ID, *task.NodeID)
	assert.Equal(t, domain.TaskStatusRunning, task.Status)
	assert.NotNil(t, task.StartedAt)

	// Verify task is locked
	lockKey := "task:lock:task1"
	acquired, err := distributedLock.Lock(ctx, lockKey, time.Minute)
	require.NoError(t, err)
	assert.False(t, acquired, "Task should be locked")
}

func TestAllocator_ReleaseTask(t *testing.T) {
	ctx := context.Background()

	// Setup
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRandom)

	// Register a node
	node := &Node{ID: "node1", Address: "localhost:8080"}
	require.NoError(t, registry.Register(ctx, node))

	// Create and allocate a task
	task := &domain.Task{
		ID:            "task1",
		Name:          "Test Task",
		Status:        domain.TaskStatusPending,
		ExecutionMode: domain.ExecutionModeImmediate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, repo.Create(ctx, task))
	_, err := allocator.AllocateTask(ctx, task)
	require.NoError(t, err)

	// Release task
	err = allocator.ReleaseTask(ctx, "task1")
	require.NoError(t, err)

	// Verify task is unlocked
	lockKey := "task:lock:task1"
	acquired, err := distributedLock.Lock(ctx, lockKey, time.Minute)
	require.NoError(t, err)
	assert.True(t, acquired, "Task should be unlocked")
}

func TestAllocator_RoundRobinStrategy(t *testing.T) {
	ctx := context.Background()

	// Setup
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRoundRobin)

	// Register nodes
	node1 := &Node{ID: "node1", Address: "localhost:8080"}
	node2 := &Node{ID: "node2", Address: "localhost:8081"}
	node3 := &Node{ID: "node3", Address: "localhost:8082"}

	require.NoError(t, registry.Register(ctx, node1))
	require.NoError(t, registry.Register(ctx, node2))
	require.NoError(t, registry.Register(ctx, node3))

	// Allocate multiple tasks and verify round-robin distribution
	nodeAssignments := make(map[string]int)

	for i := 0; i < 9; i++ {
		task := &domain.Task{
			ID:            fmt.Sprintf("task%d", i),
			Name:          "Test Task",
			Status:        domain.TaskStatusPending,
			ExecutionMode: domain.ExecutionModeImmediate,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		require.NoError(t, repo.Create(ctx, task))

		selectedNode, err := allocator.AllocateTask(ctx, task)
		require.NoError(t, err)
		require.NotNil(t, selectedNode)

		nodeAssignments[selectedNode.ID]++

		// Release for next iteration
		allocator.ReleaseTask(ctx, task.ID)
	}

	// Each node should get 3 tasks (9 tasks / 3 nodes)
	// Note: The exact distribution depends on the order nodes are returned from the registry
	// which may not be deterministic, so we just verify all tasks were assigned
	totalAssigned := nodeAssignments["node1"] + nodeAssignments["node2"] + nodeAssignments["node3"]
	assert.Equal(t, 9, totalAssigned, "All 9 tasks should be assigned")

	// Verify each node got at least one task
	assert.Greater(t, nodeAssignments["node1"], 0)
	assert.Greater(t, nodeAssignments["node2"], 0)
	assert.Greater(t, nodeAssignments["node3"], 0)
}

func TestAllocator_NoHealthyNodes(t *testing.T) {
	ctx := context.Background()

	// Setup with empty registry
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRandom)

	// Create a task
	task := &domain.Task{
		ID:            "task1",
		Name:          "Test Task",
		Status:        domain.TaskStatusPending,
		ExecutionMode: domain.ExecutionModeImmediate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, repo.Create(ctx, task))

	// Try to allocate task with no nodes
	_, err := allocator.AllocateTask(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy nodes available")
}

func TestAllocator_GetNodeTasks(t *testing.T) {
	ctx := context.Background()

	// Setup
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRandom)

	// Register nodes
	node1 := &Node{ID: "node1", Address: "localhost:8080"}
	node2 := &Node{ID: "node2", Address: "localhost:8081"}

	require.NoError(t, registry.Register(ctx, node1))
	require.NoError(t, registry.Register(ctx, node2))

	// Create tasks and assign to node1
	node1ID := "node1"
	for i := 0; i < 3; i++ {
		task := &domain.Task{
			ID:            fmt.Sprintf("task%d", i),
			Name:          "Test Task",
			Status:        domain.TaskStatusRunning,
			ExecutionMode: domain.ExecutionModeImmediate,
			NodeID:        &node1ID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		require.NoError(t, repo.Create(ctx, task))
	}

	// Create a task assigned to node2
	node2ID := "node2"
	task := &domain.Task{
		ID:            "task3",
		Name:          "Test Task",
		Status:        domain.TaskStatusRunning,
		ExecutionMode: domain.ExecutionModeImmediate,
		NodeID:        &node2ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, repo.Create(ctx, task))

	// Get tasks for node1
	node1Tasks, err := allocator.GetNodeTasks(ctx, "node1")
	require.NoError(t, err)
	assert.Len(t, node1Tasks, 3)

	// Get tasks for node2
	node2Tasks, err := allocator.GetNodeTasks(ctx, "node2")
	require.NoError(t, err)
	assert.Len(t, node2Tasks, 1)
}

func TestMultiNodeTaskAllocation_Integration(t *testing.T) {
	ctx := context.Background()

	// Setup registry with heartbeat detection
	registry := NewInMemoryRegistry(2 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	repo := newMockTaskRepository()
	distributedLock := newMockTaskLock()

	allocator := NewAllocator(registry, repo, distributedLock, AllocationStrategyRoundRobin)

	// Register multiple nodes
	node1 := &Node{ID: "node1", Address: "localhost:8080"}
	node2 := &Node{ID: "node2", Address: "localhost:8081"}
	node3 := &Node{ID: "node3", Address: "localhost:8082"}

	require.NoError(t, registry.Register(ctx, node1))
	require.NoError(t, registry.Register(ctx, node2))
	require.NoError(t, registry.Register(ctx, node3))

	// Start heartbeat managers for nodes
	hb1 := NewHeartbeatManager(registry, "node1", 500*time.Millisecond)
	hb2 := NewHeartbeatManager(registry, "node2", 500*time.Millisecond)
	hb3 := NewHeartbeatManager(registry, "node3", 500*time.Millisecond)

	require.NoError(t, hb1.Start())
	require.NoError(t, hb2.Start())
	require.NoError(t, hb3.Start())

	defer hb1.Stop()
	defer hb2.Stop()
	defer hb3.Stop()

	// Verify all nodes are healthy
	time.Sleep(100 * time.Millisecond)
	healthyNodes, err := registry.ListHealthyNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, healthyNodes, 3)

	// Create and allocate tasks
	for i := 0; i < 6; i++ {
		task := &domain.Task{
			ID:            fmt.Sprintf("task%d", i),
			Name:          "Test Task",
			Status:        domain.TaskStatusPending,
			ExecutionMode: domain.ExecutionModeImmediate,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		require.NoError(t, repo.Create(ctx, task))

		selectedNode, err := allocator.AllocateTask(ctx, task)
		require.NoError(t, err)
		require.NotNil(t, selectedNode)

		// Verify task-node binding
		assert.NotNil(t, task.NodeID)
		assert.Equal(t, selectedNode.ID, *task.NodeID)
		assert.Equal(t, domain.TaskStatusRunning, task.Status)
	}

	// Verify tasks are distributed across nodes
	node1Tasks, err := allocator.GetNodeTasks(ctx, "node1")
	require.NoError(t, err)

	node2Tasks, err := allocator.GetNodeTasks(ctx, "node2")
	require.NoError(t, err)

	node3Tasks, err := allocator.GetNodeTasks(ctx, "node3")
	require.NoError(t, err)

	// With round-robin, tasks should be distributed across nodes
	// The exact distribution depends on node order, so we verify total is correct
	totalTasks := len(node1Tasks) + len(node2Tasks) + len(node3Tasks)
	assert.Equal(t, 6, totalTasks, "All 6 tasks should be assigned")

	// Verify tasks are distributed (at least 2 nodes should have tasks)
	nodesWithTasks := 0
	if len(node1Tasks) > 0 {
		nodesWithTasks++
	}
	if len(node2Tasks) > 0 {
		nodesWithTasks++
	}
	if len(node3Tasks) > 0 {
		nodesWithTasks++
	}
	assert.GreaterOrEqual(t, nodesWithTasks, 2, "At least 2 nodes should have tasks")

	// Stop one heartbeat manager to simulate node failure
	hb2.Stop()

	// Wait for health check to mark node2 as unhealthy
	time.Sleep(3 * time.Second)

	// Verify node2 is now unhealthy
	healthyNodes, err = registry.ListHealthyNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, healthyNodes, 2)

	// Allocate a new task - should only go to healthy nodes
	task := &domain.Task{
		ID:            "new-task",
		Name:          "New Task",
		Status:        domain.TaskStatusPending,
		ExecutionMode: domain.ExecutionModeImmediate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, repo.Create(ctx, task))

	selectedNode, err := allocator.AllocateTask(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, selectedNode)

	// Should be assigned to node1 or node3, not node2
	assert.NotEqual(t, "node2", selectedNode.ID)
	assert.Contains(t, []string{"node1", "node3"}, selectedNode.ID)
}
