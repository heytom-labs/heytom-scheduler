package domain

import (
	"context"
	"time"
)

// Scheduler defines the interface for task execution and scheduling
type Scheduler interface {
	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop stops the scheduler
	Stop(ctx context.Context) error

	// SubmitTask submits a task for execution
	SubmitTask(task *Task) error

	// AcquireTask acquires a task for execution by a node
	AcquireTask(nodeID string) (*Task, error)

	// ReleaseTask releases a task after execution
	ReleaseTask(taskID string) error
}

// DistributedLock defines the interface for distributed locking
type DistributedLock interface {
	// Lock acquires a distributed lock
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Unlock releases a distributed lock
	Unlock(ctx context.Context, key string) error

	// Refresh refreshes the TTL of a lock
	Refresh(ctx context.Context, key string, ttl time.Duration) error
}
