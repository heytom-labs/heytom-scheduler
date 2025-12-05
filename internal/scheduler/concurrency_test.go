package scheduler

import (
	"context"
	"testing"

	"task-scheduler/internal/domain"
)

// TestConcurrencyController_NoParent tests that tasks without parents are always allowed
func TestConcurrencyController_NoParent(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	task := &domain.Task{
		ID:       "task-1",
		Name:     "Test Task",
		ParentID: nil,
		Status:   domain.TaskStatusPending,
	}

	canExecute, err := controller.CanExecuteSubtask(ctx, task)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !canExecute {
		t.Error("Task without parent should always be allowed to execute")
	}
}

// TestConcurrencyController_NoConcurrencyLimit tests that tasks with no limit are always allowed
func TestConcurrencyController_NoConcurrencyLimit(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task with no concurrency limit
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 0, // No limit
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Create subtask
	subtask := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}

	canExecute, err := controller.CanExecuteSubtask(ctx, subtask)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !canExecute {
		t.Error("Subtask with no concurrency limit should always be allowed to execute")
	}
}

// TestConcurrencyController_WithinLimit tests execution within concurrency limit
func TestConcurrencyController_WithinLimit(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task with concurrency limit of 2
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 2,
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Create first subtask
	subtask1 := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}

	canExecute, err := controller.CanExecuteSubtask(ctx, subtask1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !canExecute {
		t.Error("First subtask should be allowed to execute")
	}

	// Create second subtask
	subtask2 := &domain.Task{
		ID:       "subtask-2",
		Name:     "Subtask 2",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}

	canExecute, err = controller.CanExecuteSubtask(ctx, subtask2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !canExecute {
		t.Error("Second subtask should be allowed to execute (within limit of 2)")
	}

	// Verify running count
	runningCount := controller.GetRunningCount(parentID)
	if runningCount != 2 {
		t.Errorf("Expected running count 2, got %d", runningCount)
	}
}

// TestConcurrencyController_ExceedsLimit tests that tasks are queued when limit is exceeded
func TestConcurrencyController_ExceedsLimit(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task with concurrency limit of 2
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 2,
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Execute first two subtasks (should succeed)
	subtask1 := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask1)

	subtask2 := &domain.Task{
		ID:       "subtask-2",
		Name:     "Subtask 2",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask2)

	// Try to execute third subtask (should be queued)
	subtask3 := &domain.Task{
		ID:       "subtask-3",
		Name:     "Subtask 3",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}

	canExecute, err := controller.CanExecuteSubtask(ctx, subtask3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if canExecute {
		t.Error("Third subtask should not be allowed to execute (exceeds limit)")
	}

	// Verify it's in the waiting queue
	waitingTasks := controller.GetWaitingSubtasks(parentID)
	if len(waitingTasks) != 1 {
		t.Errorf("Expected 1 task in waiting queue, got %d", len(waitingTasks))
	}

	if len(waitingTasks) > 0 && waitingTasks[0].ID != "subtask-3" {
		t.Errorf("Expected subtask-3 in queue, got %s", waitingTasks[0].ID)
	}
}

// TestConcurrencyController_SubtaskCompletion tests that completing a subtask allows next to execute
func TestConcurrencyController_SubtaskCompletion(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task with concurrency limit of 1
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 1,
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Execute first subtask
	subtask1 := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	canExecute, _ := controller.CanExecuteSubtask(ctx, subtask1)
	if !canExecute {
		t.Fatal("First subtask should be allowed")
	}

	// Try to execute second subtask (should be queued)
	subtask2 := &domain.Task{
		ID:       "subtask-2",
		Name:     "Subtask 2",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	canExecute, _ = controller.CanExecuteSubtask(ctx, subtask2)
	if canExecute {
		t.Error("Second subtask should be queued")
	}

	// Complete first subtask
	nextTask, err := controller.OnSubtaskCompleted(ctx, subtask1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if nextTask == nil {
		t.Fatal("Expected next task to be returned")
	}

	if nextTask.ID != "subtask-2" {
		t.Errorf("Expected subtask-2 to be next, got %s", nextTask.ID)
	}

	// Verify running count is still 1 (subtask2 is now running)
	runningCount := controller.GetRunningCount(parentID)
	if runningCount != 1 {
		t.Errorf("Expected running count 1, got %d", runningCount)
	}

	// Verify queue is empty
	waitingTasks := controller.GetWaitingSubtasks(parentID)
	if len(waitingTasks) != 0 {
		t.Errorf("Expected empty waiting queue, got %d tasks", len(waitingTasks))
	}
}

// TestConcurrencyController_MultipleWaitingTasks tests FIFO order of waiting queue
func TestConcurrencyController_MultipleWaitingTasks(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task with concurrency limit of 1
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 1,
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Execute first subtask
	subtask1 := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask1)

	// Queue three more subtasks
	subtask2 := &domain.Task{
		ID:       "subtask-2",
		Name:     "Subtask 2",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask2)

	subtask3 := &domain.Task{
		ID:       "subtask-3",
		Name:     "Subtask 3",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask3)

	subtask4 := &domain.Task{
		ID:       "subtask-4",
		Name:     "Subtask 4",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask4)

	// Verify queue has 3 tasks
	waitingTasks := controller.GetWaitingSubtasks(parentID)
	if len(waitingTasks) != 3 {
		t.Errorf("Expected 3 tasks in waiting queue, got %d", len(waitingTasks))
	}

	// Complete subtask1, should get subtask2
	nextTask, _ := controller.OnSubtaskCompleted(ctx, subtask1)
	if nextTask == nil || nextTask.ID != "subtask-2" {
		t.Errorf("Expected subtask-2, got %v", nextTask)
	}

	// Complete subtask2, should get subtask3
	nextTask, _ = controller.OnSubtaskCompleted(ctx, subtask2)
	if nextTask == nil || nextTask.ID != "subtask-3" {
		t.Errorf("Expected subtask-3, got %v", nextTask)
	}

	// Complete subtask3, should get subtask4
	nextTask, _ = controller.OnSubtaskCompleted(ctx, subtask3)
	if nextTask == nil || nextTask.ID != "subtask-4" {
		t.Errorf("Expected subtask-4, got %v", nextTask)
	}

	// Complete subtask4, should get nil (queue empty)
	nextTask, _ = controller.OnSubtaskCompleted(ctx, subtask4)
	if nextTask != nil {
		t.Errorf("Expected nil (empty queue), got %v", nextTask)
	}
}

// TestConcurrencyController_CleanupParentState tests cleanup of parent state
func TestConcurrencyController_CleanupParentState(t *testing.T) {
	repo := newMockRepository()
	controller := NewConcurrencyController(repo)

	ctx := context.Background()

	// Create parent task
	parentID := "parent-1"
	parent := &domain.Task{
		ID:               parentID,
		Name:             "Parent Task",
		ConcurrencyLimit: 2,
		Status:           domain.TaskStatusPending,
	}
	repo.Create(ctx, parent)

	// Execute a subtask to create state
	subtask := &domain.Task{
		ID:       "subtask-1",
		Name:     "Subtask 1",
		ParentID: &parentID,
		Status:   domain.TaskStatusPending,
	}
	controller.CanExecuteSubtask(ctx, subtask)

	// Verify state exists
	runningCount := controller.GetRunningCount(parentID)
	if runningCount != 1 {
		t.Errorf("Expected running count 1, got %d", runningCount)
	}

	// Cleanup state
	controller.CleanupParentState(parentID)

	// Verify state is gone
	runningCount = controller.GetRunningCount(parentID)
	if runningCount != 0 {
		t.Errorf("Expected running count 0 after cleanup, got %d", runningCount)
	}
}

// TestConcurrencyController_ConcurrentLimit tests various concurrency limits
func TestConcurrencyController_ConcurrentLimit(t *testing.T) {
	tests := []struct {
		name             string
		concurrencyLimit int
		numSubtasks      int
		expectedRunning  int
		expectedWaiting  int
	}{
		{"limit 1, 3 tasks", 1, 3, 1, 2},
		{"limit 2, 5 tasks", 2, 5, 2, 3},
		{"limit 3, 3 tasks", 3, 3, 3, 0},
		{"limit 5, 3 tasks", 5, 3, 3, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			controller := NewConcurrencyController(repo)
			ctx := context.Background()

			// Create parent task
			parentID := "parent-1"
			parent := &domain.Task{
				ID:               parentID,
				Name:             "Parent Task",
				ConcurrencyLimit: tt.concurrencyLimit,
				Status:           domain.TaskStatusPending,
			}
			repo.Create(ctx, parent)

			// Try to execute all subtasks
			for i := 0; i < tt.numSubtasks; i++ {
				subtask := &domain.Task{
					ID:       "subtask-" + string(rune('1'+i)),
					Name:     "Subtask",
					ParentID: &parentID,
					Status:   domain.TaskStatusPending,
				}
				controller.CanExecuteSubtask(ctx, subtask)
			}

			// Verify counts
			runningCount := controller.GetRunningCount(parentID)
			if runningCount != tt.expectedRunning {
				t.Errorf("Expected running count %d, got %d", tt.expectedRunning, runningCount)
			}

			waitingTasks := controller.GetWaitingSubtasks(parentID)
			if len(waitingTasks) != tt.expectedWaiting {
				t.Errorf("Expected %d waiting tasks, got %d", tt.expectedWaiting, len(waitingTasks))
			}
		})
	}
}
