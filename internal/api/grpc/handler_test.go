package grpc

import (
	"context"
	"testing"
	"time"

	"task-scheduler/api/pb"
	"task-scheduler/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskService) CancelTask(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

func TestCreateTask(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)

	now := time.Now()
	expectedTask := &domain.Task{
		ID:            "task-123",
		Name:          "test-task",
		Description:   "test description",
		ExecutionMode: domain.ExecutionModeImmediate,
		Status:        domain.TaskStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	mockService.On("CreateTask", mock.Anything, mock.AnythingOfType("*domain.CreateTaskRequest")).
		Return(expectedTask, nil)

	req := &pb.CreateTaskRequest{
		Name:          "test-task",
		Description:   "test description",
		ExecutionMode: pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE,
	}

	resp, err := handler.CreateTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Task)
	assert.Equal(t, "task-123", resp.Task.Id)
	assert.Equal(t, "test-task", resp.Task.Name)
	assert.Equal(t, pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE, resp.Task.ExecutionMode)
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_PENDING, resp.Task.Status)

	mockService.AssertExpectations(t)
}

func TestGetTask(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)

	now := time.Now()
	expectedTask := &domain.Task{
		ID:            "task-123",
		Name:          "test-task",
		Description:   "test description",
		ExecutionMode: domain.ExecutionModeImmediate,
		Status:        domain.TaskStatusRunning,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	mockService.On("GetTask", mock.Anything, "task-123").Return(expectedTask, nil)

	req := &pb.GetTaskRequest{
		TaskId: "task-123",
	}

	resp, err := handler.GetTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Task)
	assert.Equal(t, "task-123", resp.Task.Id)
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_RUNNING, resp.Task.Status)

	mockService.AssertExpectations(t)
}

func TestListTasks(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)

	now := time.Now()
	expectedTasks := []*domain.Task{
		{
			ID:            "task-1",
			Name:          "task-1",
			ExecutionMode: domain.ExecutionModeImmediate,
			Status:        domain.TaskStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "task-2",
			Name:          "task-2",
			ExecutionMode: domain.ExecutionModeScheduled,
			Status:        domain.TaskStatusRunning,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}

	mockService.On("ListTasks", mock.Anything, mock.AnythingOfType("*domain.TaskFilter")).
		Return(expectedTasks, nil)

	req := &pb.ListTasksRequest{
		Limit: 10,
	}

	resp, err := handler.ListTasks(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Count)
	assert.Len(t, resp.Tasks, 2)
	assert.Equal(t, "task-1", resp.Tasks[0].Id)
	assert.Equal(t, "task-2", resp.Tasks[1].Id)

	mockService.AssertExpectations(t)
}

func TestCancelTask(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)

	mockService.On("CancelTask", mock.Anything, "task-123").Return(nil)

	req := &pb.CancelTaskRequest{
		TaskId: "task-123",
	}

	resp, err := handler.CancelTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "task-123", resp.TaskId)
	assert.Contains(t, resp.Message, "cancelled successfully")

	mockService.AssertExpectations(t)
}

func TestGetTaskStatus(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)

	now := time.Now()
	startedAt := now.Add(-5 * time.Minute)
	expectedTask := &domain.Task{
		ID:         "task-123",
		Name:       "test-task",
		Status:     domain.TaskStatusSuccess,
		RetryCount: 2,
		CreatedAt:  now.Add(-10 * time.Minute),
		UpdatedAt:  now,
		StartedAt:  &startedAt,
	}

	mockService.On("GetTask", mock.Anything, "task-123").Return(expectedTask, nil)

	req := &pb.GetTaskStatusRequest{
		TaskId: "task-123",
	}

	resp, err := handler.GetTaskStatus(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "task-123", resp.TaskId)
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_SUCCESS, resp.Status)
	assert.Equal(t, int32(2), resp.RetryCount)
	assert.Greater(t, resp.StartedAt, int64(0))

	mockService.AssertExpectations(t)
}

func TestEnumConversions(t *testing.T) {
	handler := &Handler{}

	// Test ExecutionMode conversions
	assert.Equal(t, domain.ExecutionModeImmediate, handler.convertProtoExecutionModeToDomain(pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE))
	assert.Equal(t, domain.ExecutionModeScheduled, handler.convertProtoExecutionModeToDomain(pb.ExecutionMode_EXECUTION_MODE_SCHEDULED))
	assert.Equal(t, domain.ExecutionModeInterval, handler.convertProtoExecutionModeToDomain(pb.ExecutionMode_EXECUTION_MODE_INTERVAL))
	assert.Equal(t, domain.ExecutionModeCron, handler.convertProtoExecutionModeToDomain(pb.ExecutionMode_EXECUTION_MODE_CRON))

	assert.Equal(t, pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE, handler.convertDomainExecutionModeToProto(domain.ExecutionModeImmediate))
	assert.Equal(t, pb.ExecutionMode_EXECUTION_MODE_SCHEDULED, handler.convertDomainExecutionModeToProto(domain.ExecutionModeScheduled))
	assert.Equal(t, pb.ExecutionMode_EXECUTION_MODE_INTERVAL, handler.convertDomainExecutionModeToProto(domain.ExecutionModeInterval))
	assert.Equal(t, pb.ExecutionMode_EXECUTION_MODE_CRON, handler.convertDomainExecutionModeToProto(domain.ExecutionModeCron))

	// Test TaskStatus conversions
	assert.Equal(t, domain.TaskStatusPending, handler.convertProtoStatusToDomain(pb.TaskStatus_TASK_STATUS_PENDING))
	assert.Equal(t, domain.TaskStatusRunning, handler.convertProtoStatusToDomain(pb.TaskStatus_TASK_STATUS_RUNNING))
	assert.Equal(t, domain.TaskStatusSuccess, handler.convertProtoStatusToDomain(pb.TaskStatus_TASK_STATUS_SUCCESS))
	assert.Equal(t, domain.TaskStatusFailed, handler.convertProtoStatusToDomain(pb.TaskStatus_TASK_STATUS_FAILED))
	assert.Equal(t, domain.TaskStatusCancelled, handler.convertProtoStatusToDomain(pb.TaskStatus_TASK_STATUS_CANCELLED))

	assert.Equal(t, pb.TaskStatus_TASK_STATUS_PENDING, handler.convertDomainStatusToProto(domain.TaskStatusPending))
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_RUNNING, handler.convertDomainStatusToProto(domain.TaskStatusRunning))
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_SUCCESS, handler.convertDomainStatusToProto(domain.TaskStatusSuccess))
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_FAILED, handler.convertDomainStatusToProto(domain.TaskStatusFailed))
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_CANCELLED, handler.convertDomainStatusToProto(domain.TaskStatusCancelled))

	// Test CallbackProtocol conversions
	assert.Equal(t, domain.CallbackProtocolHTTP, handler.convertProtoCallbackProtocolToDomain(pb.CallbackProtocol_CALLBACK_PROTOCOL_HTTP))
	assert.Equal(t, domain.CallbackProtocolGRPC, handler.convertProtoCallbackProtocolToDomain(pb.CallbackProtocol_CALLBACK_PROTOCOL_GRPC))

	assert.Equal(t, pb.CallbackProtocol_CALLBACK_PROTOCOL_HTTP, handler.convertDomainCallbackProtocolToProto(domain.CallbackProtocolHTTP))
	assert.Equal(t, pb.CallbackProtocol_CALLBACK_PROTOCOL_GRPC, handler.convertDomainCallbackProtocolToProto(domain.CallbackProtocolGRPC))
}
