package domain

import "context"

// TaskFilter defines criteria for filtering tasks
type TaskFilter struct {
	Status   *TaskStatus
	ParentID *string
	NodeID   *string
	Limit    int
	Offset   int
}

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	// Create creates a new task
	Create(ctx context.Context, task *Task) error

	// Get retrieves a task by ID
	Get(ctx context.Context, taskID string) (*Task, error)

	// Update updates an existing task
	Update(ctx context.Context, task *Task) error

	// Delete deletes a task by ID
	Delete(ctx context.Context, taskID string) error

	// List retrieves tasks matching the filter
	List(ctx context.Context, filter *TaskFilter) ([]*Task, error)

	// GetSubTasks retrieves all subtasks of a parent task
	GetSubTasks(ctx context.Context, parentID string) ([]*Task, error)
}
