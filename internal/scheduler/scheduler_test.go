package scheduler

import (
	"context"
	"testing"
	"time"

	"task-scheduler/internal/domain"
)

// mockRepository is a mock implementation of TaskRepository for testing
type mockRepository struct {
	tasks map[string]*domain.Task
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockRepository) Create(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	task, ok := m.tasks[taskID]
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (m *mockRepository) Update(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, taskID string) error {
	delete(m.tasks, taskID)
	return nil
}

func (m *mockRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, task := range m.tasks {
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		result = append(result, task)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

func (m *mockRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, task := range m.tasks {
		if task.ParentID != nil && *task.ParentID == parentID {
			result = append(result, task)
		}
	}
	return result, nil
}

// mockScheduleService is a mock implementation of ScheduleService
type mockScheduleService struct{}

func (m *mockScheduleService) ScheduleTask(ctx context.Context, task *domain.Task) error {
	return nil
}

func (m *mockScheduleService) UnscheduleTask(ctx context.Context, taskID string) error {
	return nil
}

func (m *mockScheduleService) GetNextExecutionTime(task *domain.Task) (time.Time, error) {
	return time.Now(), nil
}

// mockCallbackService is a mock implementation of CallbackService
type mockCallbackService struct{}

func (m *mockCallbackService) ExecuteCallback(ctx context.Context, task *domain.Task, result *domain.ExecutionResult) error {
	return nil
}

func (m *mockCallbackService) ExecuteAsyncCallback(ctx context.Context, task *domain.Task) error {
	return nil
}

func (m *mockCallbackService) HandleCallbackResponse(ctx context.Context, taskID string, response *domain.CallbackResponse) error {
	return nil
}

// mockDistributedLock is a mock implementation of DistributedLock
type mockDistributedLock struct {
	locks map[string]bool
}

func newMockDistributedLock() *mockDistributedLock {
	return &mockDistributedLock{
		locks: make(map[string]bool),
	}
}

func (m *mockDistributedLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if m.locks[key] {
		return false, nil
	}
	m.locks[key] = true
	return true, nil
}

func (m *mockDistributedLock) Unlock(ctx context.Context, key string) error {
	delete(m.locks, key)
	return nil
}

func (m *mockDistributedLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

// TestSchedulerStartStop tests starting and stopping the scheduler
func TestSchedulerStartStop(t *testing.T) {
	repo := newMockRepository()
	schedSvc := &mockScheduleService{}
	callbackSvc := &mockCallbackService{}
	lock := newMockDistributedLock()

	config := &Config{
		WorkerCount:  2,
		QueueSize:    10,
		PollInterval: 100 * time.Millisecond,
		LockTTL:      time.Minute,
	}

	scheduler := NewScheduler(config, repo, schedSvc, callbackSvc, lock, "test-node")

	ctx := context.Background()

	// Test Start
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Test Stop
	err = scheduler.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

// TestSchedulerSubmitTask tests submitting a task
func TestSchedulerSubmitTask(t *testing.T) {
	repo := newMockRepository()
	schedSvc := &mockScheduleService{}
	callbackSvc := &mockCallbackService{}
	lock := newMockDistributedLock()

	config := &Config{
		WorkerCount:  2,
		QueueSize:    10,
		PollInterval: 100 * time.Millisecond,
		LockTTL:      time.Minute,
	}

	scheduler := NewScheduler(config, repo, schedSvc, callbackSvc, lock, "test-node")

	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop(ctx)

	// Create a test task
	task := &domain.Task{
		ID:            "test-task-1",
		Name:          "Test Task",
		ExecutionMode: domain.ExecutionModeImmediate,
		Status:        domain.TaskStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Submit task
	err = scheduler.SubmitTask(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
}

// TestStateTransitions tests valid and invalid state transitions
func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.TaskStatus
		to       domain.TaskStatus
		expected bool
	}{
		{"pending to running", domain.TaskStatusPending, domain.TaskStatusRunning, true},
		{"pending to cancelled", domain.TaskStatusPending, domain.TaskStatusCancelled, true},
		{"running to success", domain.TaskStatusRunning, domain.TaskStatusSuccess, true},
		{"running to failed", domain.TaskStatusRunning, domain.TaskStatusFailed, true},
		{"running to cancelled", domain.TaskStatusRunning, domain.TaskStatusCancelled, true},
		{"failed to pending (retry)", domain.TaskStatusFailed, domain.TaskStatusPending, true},
		{"success to running (invalid)", domain.TaskStatusSuccess, domain.TaskStatusRunning, false},
		{"pending to success (invalid)", domain.TaskStatusPending, domain.TaskStatusSuccess, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("isValidTransition(%s, %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}
