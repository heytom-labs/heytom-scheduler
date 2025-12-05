package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/internal/service"

	"github.com/stretchr/testify/assert"
)

// mockCallbackServiceWithFailures is a callback service that fails N times before succeeding
type mockCallbackServiceWithFailures struct {
	failCount    int
	currentCount int
	mu           sync.Mutex
}

func (m *mockCallbackServiceWithFailures) ExecuteCallback(ctx context.Context, task *domain.Task, result *domain.ExecutionResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentCount++

	if m.currentCount <= m.failCount {
		return errors.New("simulated failure")
	}
	return nil
}

func (m *mockCallbackServiceWithFailures) ExecuteAsyncCallback(ctx context.Context, task *domain.Task) error {
	return m.ExecuteCallback(ctx, task, nil)
}

func (m *mockCallbackServiceWithFailures) HandleCallbackResponse(ctx context.Context, taskID string, response *domain.CallbackResponse) error {
	return nil
}

// TestRetryWithExponentialBackoff tests that retry delays increase exponentially
func TestRetryWithExponentialBackoff(t *testing.T) {
	_ = context.Background()

	mockRepo := newMockRepository()

	task := &domain.Task{
		ID:         "backoff-test-task",
		Status:     domain.TaskStatusFailed,
		RetryCount: 0,
		RetryPolicy: &domain.RetryPolicy{
			MaxRetries:    3,
			RetryInterval: 1 * time.Second,
			BackoffFactor: 2.0,
		},
	}

	retryService := service.NewRetryService(mockRepo)

	// Test delay calculations
	delay1 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 1*time.Second, delay1, "First retry should have 1s delay (2^0)")

	task.RetryCount = 1
	delay2 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 2*time.Second, delay2, "Second retry should have 2s delay (2^1)")

	task.RetryCount = 2
	delay3 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 4*time.Second, delay3, "Third retry should have 4s delay (2^2)")
}

// TestRetryMaxRetriesExceeded tests that tasks fail after max retries
func TestRetryMaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()

	mockRepo := newMockRepository()

	task := &domain.Task{
		ID:         "max-retry-test-task",
		Status:     domain.TaskStatusRunning,
		RetryCount: 3,
		RetryPolicy: &domain.RetryPolicy{
			MaxRetries:    3,
			RetryInterval: 1 * time.Second,
			BackoffFactor: 1.0,
		},
	}

	retryService := service.NewRetryService(mockRepo)

	// Try to handle failure when max retries exceeded
	retried, err := retryService.HandleTaskFailure(ctx, task, errors.New("test error"))
	assert.NoError(t, err)
	assert.False(t, retried, "Task should not be retried when max retries exceeded")
	assert.Equal(t, domain.TaskStatusFailed, task.Status)
	assert.NotNil(t, task.CompletedAt)
}

// TestRetryWithNoRetryPolicy tests that tasks without retry policy fail immediately
func TestRetryWithNoRetryPolicy(t *testing.T) {
	ctx := context.Background()

	mockRepo := newMockRepository()

	task := &domain.Task{
		ID:          "no-retry-policy-task",
		Status:      domain.TaskStatusRunning,
		RetryPolicy: nil,
	}

	retryService := service.NewRetryService(mockRepo)

	// Try to handle failure with no retry policy
	retried, err := retryService.HandleTaskFailure(ctx, task, errors.New("test error"))
	assert.NoError(t, err)
	assert.False(t, retried, "Task should not be retried without retry policy")
	assert.Equal(t, domain.TaskStatusFailed, task.Status)
}

// TestRetryFixedInterval tests retry with fixed interval (no backoff)
func TestRetryFixedInterval(t *testing.T) {
	_ = context.Background()

	mockRepo := newMockRepository()

	task := &domain.Task{
		ID:         "fixed-interval-task",
		Status:     domain.TaskStatusFailed,
		RetryCount: 0,
		RetryPolicy: &domain.RetryPolicy{
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
			BackoffFactor: 1.0, // Fixed interval
		},
	}

	retryService := service.NewRetryService(mockRepo)

	// Test delay calculations with fixed interval
	delay1 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 5*time.Second, delay1, "All retries should have 5s delay with backoff factor 1.0")

	task.RetryCount = 1
	delay2 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 5*time.Second, delay2, "All retries should have 5s delay with backoff factor 1.0")

	task.RetryCount = 2
	delay3 := retryService.CalculateRetryDelay(task)
	assert.Equal(t, 5*time.Second, delay3, "All retries should have 5s delay with backoff factor 1.0")
}
