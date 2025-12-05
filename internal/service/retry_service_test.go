package service

import (
	"context"
	"testing"
	"time"

	"task-scheduler/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func TestShouldRetry(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	retryService := NewRetryService(mockRepo)

	tests := []struct {
		name     string
		task     *domain.Task
		expected bool
	}{
		{
			name:     "nil task should not retry",
			task:     nil,
			expected: false,
		},
		{
			name: "task with no retry policy should not retry",
			task: &domain.Task{
				ID:          "task1",
				RetryPolicy: nil,
			},
			expected: false,
		},
		{
			name: "task with max retries 0 should not retry",
			task: &domain.Task{
				ID: "task2",
				RetryPolicy: &domain.RetryPolicy{
					MaxRetries: 0,
				},
			},
			expected: false,
		},
		{
			name: "task with retries remaining should retry",
			task: &domain.Task{
				ID:         "task3",
				RetryCount: 1,
				RetryPolicy: &domain.RetryPolicy{
					MaxRetries: 3,
				},
			},
			expected: true,
		},
		{
			name: "task at max retries should not retry",
			task: &domain.Task{
				ID:         "task4",
				RetryCount: 3,
				RetryPolicy: &domain.RetryPolicy{
					MaxRetries: 3,
				},
			},
			expected: false,
		},
		{
			name: "task exceeding max retries should not retry",
			task: &domain.Task{
				ID:         "task5",
				RetryCount: 5,
				RetryPolicy: &domain.RetryPolicy{
					MaxRetries: 3,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retryService.ShouldRetry(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateRetryDelay(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	retryService := NewRetryService(mockRepo)

	tests := []struct {
		name     string
		task     *domain.Task
		expected time.Duration
	}{
		{
			name:     "nil task should return 0",
			task:     nil,
			expected: 0,
		},
		{
			name: "task with no retry policy should return 0",
			task: &domain.Task{
				ID:          "task1",
				RetryPolicy: nil,
			},
			expected: 0,
		},
		{
			name: "fixed interval (backoff factor 0)",
			task: &domain.Task{
				ID:         "task2",
				RetryCount: 2,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 5 * time.Second,
					BackoffFactor: 0,
				},
			},
			expected: 5 * time.Second,
		},
		{
			name: "fixed interval (backoff factor 1)",
			task: &domain.Task{
				ID:         "task3",
				RetryCount: 3,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 10 * time.Second,
					BackoffFactor: 1.0,
				},
			},
			expected: 10 * time.Second,
		},
		{
			name: "exponential backoff - first retry",
			task: &domain.Task{
				ID:         "task4",
				RetryCount: 0,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 1 * time.Second,
					BackoffFactor: 2.0,
				},
			},
			expected: 1 * time.Second, // 1 * 2^0 = 1
		},
		{
			name: "exponential backoff - second retry",
			task: &domain.Task{
				ID:         "task5",
				RetryCount: 1,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 1 * time.Second,
					BackoffFactor: 2.0,
				},
			},
			expected: 2 * time.Second, // 1 * 2^1 = 2
		},
		{
			name: "exponential backoff - third retry",
			task: &domain.Task{
				ID:         "task6",
				RetryCount: 2,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 1 * time.Second,
					BackoffFactor: 2.0,
				},
			},
			expected: 4 * time.Second, // 1 * 2^2 = 4
		},
		{
			name: "exponential backoff with 1.5 factor",
			task: &domain.Task{
				ID:         "task7",
				RetryCount: 2,
				RetryPolicy: &domain.RetryPolicy{
					RetryInterval: 10 * time.Second,
					BackoffFactor: 1.5,
				},
			},
			expected: 22500 * time.Millisecond, // 10 * 1.5^2 = 22.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retryService.CalculateRetryDelay(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScheduleRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("should schedule retry successfully", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		task := &domain.Task{
			ID:         "task1",
			Status:     domain.TaskStatusFailed,
			RetryCount: 0,
			RetryPolicy: &domain.RetryPolicy{
				MaxRetries:    3,
				RetryInterval: 5 * time.Second,
				BackoffFactor: 1.0,
			},
		}

		mockRepo.On("Update", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.ID == "task1" &&
				t.Status == domain.TaskStatusPending &&
				t.RetryCount == 1 &&
				t.ScheduleConfig != nil &&
				t.ScheduleConfig.ScheduledTime != nil
		})).Return(nil)

		err := retryService.ScheduleRetry(ctx, task)
		assert.NoError(t, err)
		assert.Equal(t, 1, task.RetryCount)
		assert.Equal(t, domain.TaskStatusPending, task.Status)
		assert.NotNil(t, task.ScheduleConfig)
		assert.NotNil(t, task.ScheduleConfig.ScheduledTime)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for nil task", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		err := retryService.ScheduleRetry(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("should return error if task should not retry", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		task := &domain.Task{
			ID:         "task2",
			RetryCount: 3,
			RetryPolicy: &domain.RetryPolicy{
				MaxRetries: 3,
			},
		}

		err := retryService.ScheduleRetry(ctx, task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "should not be retried")
	})
}

func TestHandleTaskFailure(t *testing.T) {
	ctx := context.Background()

	t.Run("should schedule retry when retries available", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		task := &domain.Task{
			ID:         "task1",
			Status:     domain.TaskStatusRunning,
			RetryCount: 0,
			RetryPolicy: &domain.RetryPolicy{
				MaxRetries:    3,
				RetryInterval: 5 * time.Second,
				BackoffFactor: 2.0,
			},
		}

		mockRepo.On("Update", ctx, mock.Anything).Return(nil)

		retried, err := retryService.HandleTaskFailure(ctx, task, assert.AnError)
		assert.NoError(t, err)
		assert.True(t, retried)
		assert.Equal(t, 1, task.RetryCount)
		assert.Equal(t, domain.TaskStatusPending, task.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should mark as failed when no retries left", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		task := &domain.Task{
			ID:         "task2",
			Status:     domain.TaskStatusRunning,
			RetryCount: 3,
			RetryPolicy: &domain.RetryPolicy{
				MaxRetries:    3,
				RetryInterval: 5 * time.Second,
			},
		}

		mockRepo.On("Update", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.ID == "task2" &&
				t.Status == domain.TaskStatusFailed &&
				t.CompletedAt != nil
		})).Return(nil)

		retried, err := retryService.HandleTaskFailure(ctx, task, assert.AnError)
		assert.NoError(t, err)
		assert.False(t, retried)
		assert.Equal(t, domain.TaskStatusFailed, task.Status)
		assert.NotNil(t, task.CompletedAt)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should mark as failed when no retry policy", func(t *testing.T) {
		mockRepo := new(MockTaskRepository)
		retryService := NewRetryService(mockRepo)

		task := &domain.Task{
			ID:          "task3",
			Status:      domain.TaskStatusRunning,
			RetryPolicy: nil,
		}

		mockRepo.On("Update", ctx, mock.MatchedBy(func(t *domain.Task) bool {
			return t.ID == "task3" && t.Status == domain.TaskStatusFailed
		})).Return(nil)

		retried, err := retryService.HandleTaskFailure(ctx, task, assert.AnError)
		assert.NoError(t, err)
		assert.False(t, retried)
		assert.Equal(t, domain.TaskStatusFailed, task.Status)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetRetryInfo(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	retryService := NewRetryService(mockRepo)

	t.Run("should return nil for nil task", func(t *testing.T) {
		info := retryService.GetRetryInfo(nil)
		assert.Nil(t, info)
	})

	t.Run("should return basic info for task without retry policy", func(t *testing.T) {
		task := &domain.Task{
			ID:          "task1",
			RetryCount:  0,
			RetryPolicy: nil,
		}

		info := retryService.GetRetryInfo(task)
		assert.NotNil(t, info)
		assert.Equal(t, 0, info["retryCount"])
		assert.Equal(t, false, info["canRetry"])
	})

	t.Run("should return full info for task with retry policy", func(t *testing.T) {
		task := &domain.Task{
			ID:         "task2",
			RetryCount: 1,
			RetryPolicy: &domain.RetryPolicy{
				MaxRetries:    3,
				RetryInterval: 5 * time.Second,
				BackoffFactor: 2.0,
			},
		}

		info := retryService.GetRetryInfo(task)
		assert.NotNil(t, info)
		assert.Equal(t, 1, info["retryCount"])
		assert.Equal(t, true, info["canRetry"])
		assert.Equal(t, 3, info["maxRetries"])
		assert.Equal(t, "5s", info["retryInterval"])
		assert.Equal(t, 2.0, info["backoffFactor"])
		assert.NotNil(t, info["nextRetryDelay"])
	})
}
