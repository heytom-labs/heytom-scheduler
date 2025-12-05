package notification

import (
	"context"
	"fmt"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// MultiNotifier sends notifications to multiple notifiers
type MultiNotifier struct {
	notifiers []domain.Notifier
}

// NewMultiNotifier creates a new multi-notifier
func NewMultiNotifier(notifiers ...domain.Notifier) *MultiNotifier {
	return &MultiNotifier{
		notifiers: notifiers,
	}
}

// Send sends an alert to all notifiers
func (m *MultiNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	var lastErr error
	successCount := 0

	for i, notifier := range m.notifiers {
		if err := notifier.Send(ctx, alert); err != nil {
			logger.Error("Failed to send notification",
				zap.Int("notifier_index", i),
				zap.Error(err),
			)
			lastErr = err
		} else {
			successCount++
		}
	}

	// Return error only if all notifiers failed
	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("all notifiers failed, last error: %w", lastErr)
	}

	return nil
}

// SendBatch sends multiple alerts to all notifiers
func (m *MultiNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
	var lastErr error
	successCount := 0

	for i, notifier := range m.notifiers {
		if err := notifier.SendBatch(ctx, alerts); err != nil {
			logger.Error("Failed to send batch notification",
				zap.Int("notifier_index", i),
				zap.Error(err),
			)
			lastErr = err
		} else {
			successCount++
		}
	}

	// Return error only if all notifiers failed
	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("all notifiers failed, last error: %w", lastErr)
	}

	return nil
}
