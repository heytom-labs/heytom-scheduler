package node

import (
	"context"
	"testing"
	"time"

	"task-scheduler/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockRegistry struct {
	mock.Mock
}

func (m *mockRegistry) Register(ctx context.Context, node *Node) error {
	args := m.Called(ctx, node)
	return args.Error(0)
}

func (m *mockRegistry) Deregister(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *mockRegistry) UpdateHeartbeat(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *mockRegistry) GetNode(ctx context.Context, nodeID string) (*Node, error) {
	args := m.Called(ctx, nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Node), args.Error(1)
}

func (m *mockRegistry) ListNodes(ctx context.Context) ([]*Node, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Node), args.Error(1)
}

func (m *mockRegistry) ListHealthyNodes(ctx context.Context) ([]*Node, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Node), args.Error(1)
}

func (m *mockRegistry) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *mockRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *mockRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

type mockLock struct {
	mock.Mock
}

func (m *mockLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *mockLock) Unlock(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func TestFailoverManager_TransferTask(t *testing.T) {
	ctx := context.Background()

	mockReg := new(mockRegistry)
	mockRepo := new(mockRepository)
	mockLck := new(mockLock)

	allocator := NewAllocator(mockReg, mockRepo, mockLck, AllocationStrategyRandom)
	failover := NewFailoverManager(mockReg, mockRepo, allocator, 10*time.Second)

	// Setup test data
	nodeID := "node1"
	task := &domain.Task{
		ID:     "task1",
		Name:   "Test Task",
		Status: domain.TaskStatusRunning,
		NodeID: &nodeID,
	}

	healthyNode := &Node{
		ID:     "node2",
		Status: NodeStatusHealthy,
	}

	// Setup mocks
	mockLck.On("Unlock", mock.Anything, "task:lock:task1").Return(nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil).Times(2)
	mockReg.On("ListHealthyNodes", mock.Anything).Return([]*Node{healthyNode}, nil)
	mockLck.On("Lock", mock.Anything, "task:lock:task1", mock.AnythingOfType("time.Duration")).Return(true, nil)

	// Execute transfer
	err := failover.transferTask(ctx, task)
	require.NoError(t, err)

	// Verify task was updated - AllocateTask sets it to running
	assert.NotNil(t, task.NodeID)

	mockLck.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

func TestFailoverManager_CheckAndTransferTasks(t *testing.T) {
	mockReg := new(mockRegistry)
	mockRepo := new(mockRepository)
	mockLck := new(mockLock)

	allocator := NewAllocator(mockReg, mockRepo, mockLck, AllocationStrategyRandom)
	failover := NewFailoverManager(mockReg, mockRepo, allocator, 10*time.Second)

	// Setup test data
	unhealthyNode := &Node{
		ID:     "node1",
		Status: NodeStatusUnhealthy,
	}

	healthyNode := &Node{
		ID:     "node2",
		Status: NodeStatusHealthy,
	}

	nodeID := "node1"
	task := &domain.Task{
		ID:     "task1",
		Name:   "Test Task",
		Status: domain.TaskStatusRunning,
		NodeID: &nodeID,
	}

	// Setup mocks - use mock.Anything for context to avoid type mismatch
	mockReg.On("ListNodes", mock.Anything).Return([]*Node{unhealthyNode, healthyNode}, nil)

	// For GetNodeTasks
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*domain.TaskFilter")).Return([]*domain.Task{task}, nil)

	// For transfer
	mockLck.On("Unlock", mock.Anything, "task:lock:task1").Return(nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil).Times(2)
	mockReg.On("ListHealthyNodes", mock.Anything).Return([]*Node{healthyNode}, nil)
	mockLck.On("Lock", mock.Anything, "task:lock:task1", mock.AnythingOfType("time.Duration")).Return(true, nil)

	// Execute
	err := failover.checkAndTransferTasks()
	require.NoError(t, err)

	mockReg.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLck.AssertExpectations(t)
}

func TestFailoverManager_ForceTransferTask(t *testing.T) {
	ctx := context.Background()

	mockReg := new(mockRegistry)
	mockRepo := new(mockRepository)
	mockLck := new(mockLock)

	allocator := NewAllocator(mockReg, mockRepo, mockLck, AllocationStrategyRandom)
	failover := NewFailoverManager(mockReg, mockRepo, allocator, 10*time.Second)

	// Setup test data
	nodeID := "node1"
	task := &domain.Task{
		ID:     "task1",
		Name:   "Test Task",
		Status: domain.TaskStatusRunning,
		NodeID: &nodeID,
	}

	targetNode := &Node{
		ID:     "node2",
		Status: NodeStatusHealthy,
	}

	// Setup mocks
	mockRepo.On("Get", mock.Anything, "task1").Return(task, nil)
	mockReg.On("GetNode", mock.Anything, "node2").Return(targetNode, nil)
	mockLck.On("Unlock", mock.Anything, "task:lock:task1").Return(nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

	// Execute
	err := failover.ForceTransferTask(ctx, "task1", "node2")
	require.NoError(t, err)

	// Verify task was updated with new node
	assert.Equal(t, "node2", *task.NodeID)
	assert.Equal(t, domain.TaskStatusRunning, task.Status)

	mockRepo.AssertExpectations(t)
	mockReg.AssertExpectations(t)
	mockLck.AssertExpectations(t)
}
