package notification

import (
	"context"

	"task-scheduler/internal/domain"
)

// SchedulerNotificationAdapter adapts NotificationService for use in the scheduler
type SchedulerNotificationAdapter struct {
	service *NotificationService
}

// NewSchedulerNotificationAdapter creates a new adapter
func NewSchedulerNotificationAdapter(service *NotificationService) *SchedulerNotificationAdapter {
	return &SchedulerNotificationAdapter{
		service: service,
	}
}

// EvaluateAndSendAlert evaluates and sends an alert based on the reason
func (sna *SchedulerNotificationAdapter) EvaluateAndSendAlert(ctx context.Context, task *domain.Task, reason string, errorMsg *string) {
	var alertReason AlertReason

	switch reason {
	case "failure":
		alertReason = AlertReasonFailure
	case "retry_threshold":
		alertReason = AlertReasonRetryThreshold
	case "timeout":
		alertReason = AlertReasonTimeout
	default:
		alertReason = AlertReason(reason)
	}

	sna.service.EvaluateAndSendAlert(ctx, task, alertReason, errorMsg)
}
