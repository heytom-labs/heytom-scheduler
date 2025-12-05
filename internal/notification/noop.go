package notification

import (
	"context"

	"task-scheduler/internal/domain"
)

// NoOpNotifier is a notifier that does nothing
type NoOpNotifier struct{}

// NewNoOpNotifier creates a new no-op notifier
func NewNoOpNotifier() *NoOpNotifier {
	return &NoOpNotifier{}
}

// Send does nothing
func (n *NoOpNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	return nil
}

// SendBatch does nothing
func (n *NoOpNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
	return nil
}
