package notification

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	"go.uber.org/zap"
)

// AlertTrigger evaluates whether an alert should be triggered
type AlertTrigger struct {
	logger *zap.Logger
}

// NewAlertTrigger creates a new alert trigger evaluator
func NewAlertTrigger(logger *zap.Logger) *AlertTrigger {
	return &AlertTrigger{
		logger: logger,
	}
}

// ShouldTriggerAlert evaluates if an alert should be sent based on the task and policy
func (at *AlertTrigger) ShouldTriggerAlert(task *domain.Task, reason AlertReason) bool {
	if task.AlertPolicy == nil {
		return false
	}

	policy := task.AlertPolicy

	switch reason {
	case AlertReasonFailure:
		return policy.EnableFailureAlert && task.Status == domain.TaskStatusFailed

	case AlertReasonRetryThreshold:
		return policy.RetryThreshold > 0 && task.RetryCount >= policy.RetryThreshold

	case AlertReasonTimeout:
		if policy.TimeoutThreshold <= 0 || task.StartedAt == nil {
			return false
		}
		elapsed := time.Since(*task.StartedAt)
		return elapsed >= policy.TimeoutThreshold

	default:
		return false
	}
}

// AlertReason defines why an alert is triggered
type AlertReason string

const (
	AlertReasonFailure        AlertReason = "failure"         // 任务失败
	AlertReasonRetryThreshold AlertReason = "retry_threshold" // 重试超过阈值
	AlertReasonTimeout        AlertReason = "timeout"         // 执行超时
)

// CreateAlert creates an alert from a task and reason
func (at *AlertTrigger) CreateAlert(task *domain.Task, reason AlertReason, errorMsg *string) *domain.Alert {
	reasonText := at.formatReason(reason)

	alert := &domain.Alert{
		TaskID:     task.ID,
		TaskName:   task.Name,
		Reason:     reasonText,
		Timestamp:  time.Now().Format(time.RFC3339),
		RetryCount: task.RetryCount,
		ErrorMsg:   errorMsg,
		Metadata:   task.Metadata,
	}

	return alert
}

func (at *AlertTrigger) formatReason(reason AlertReason) string {
	switch reason {
	case AlertReasonFailure:
		return "任务执行失败"
	case AlertReasonRetryThreshold:
		return "重试次数超过阈值"
	case AlertReasonTimeout:
		return "任务执行超时"
	default:
		return string(reason)
	}
}

// NotificationService manages alert notifications
type NotificationService struct {
	notifiers map[domain.ChannelType]domain.Notifier
	trigger   *AlertTrigger
	logger    *zap.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(logger *zap.Logger) *NotificationService {
	return &NotificationService{
		notifiers: make(map[domain.ChannelType]domain.Notifier),
		trigger:   NewAlertTrigger(logger),
		logger:    logger,
	}
}

// RegisterNotifier registers a notifier for a specific channel type
func (ns *NotificationService) RegisterNotifier(channelType domain.ChannelType, notifier domain.Notifier) {
	ns.notifiers[channelType] = notifier
}

// SendAlert sends an alert through the configured channels
func (ns *NotificationService) SendAlert(ctx context.Context, task *domain.Task, reason AlertReason, errorMsg *string) error {
	if task.AlertPolicy == nil || len(task.AlertPolicy.Channels) == 0 {
		return nil
	}

	// Check if alert should be triggered
	if !ns.trigger.ShouldTriggerAlert(task, reason) {
		return nil
	}

	// Create alert
	alert := ns.trigger.CreateAlert(task, reason, errorMsg)

	// Send to all configured channels
	var lastErr error
	for _, channel := range task.AlertPolicy.Channels {
		notifier, ok := ns.notifiers[channel.Type]
		if !ok {
			ns.logger.Warn("notifier not found for channel type", zap.String("type", string(channel.Type)))
			continue
		}

		if err := notifier.Send(ctx, alert); err != nil {
			ns.logger.Error("failed to send alert", zap.String("channel", string(channel.Type)), zap.Error(err))
			lastErr = err
			// Continue sending to other channels even if one fails
		} else {
			ns.logger.Info("alert sent successfully", zap.String("channel", string(channel.Type)), zap.String("taskID", task.ID))
		}
	}

	// Return last error but don't fail the task
	return lastErr
}

// SendAlertBatch sends multiple alerts through the configured channels
func (ns *NotificationService) SendAlertBatch(ctx context.Context, alerts []*domain.Alert, channels []domain.NotificationChannel) error {
	if len(channels) == 0 {
		return nil
	}

	var lastErr error
	for _, channel := range channels {
		notifier, ok := ns.notifiers[channel.Type]
		if !ok {
			ns.logger.Warn("notifier not found for channel type", zap.String("type", string(channel.Type)))
			continue
		}

		if err := notifier.SendBatch(ctx, alerts); err != nil {
			ns.logger.Error("failed to send alert batch", zap.String("channel", string(channel.Type)), zap.Error(err))
			lastErr = err
		} else {
			ns.logger.Info("alert batch sent successfully", zap.String("channel", string(channel.Type)), zap.Int("count", len(alerts)))
		}
	}

	return lastErr
}

// EvaluateAndSendAlert evaluates if an alert should be sent and sends it
func (ns *NotificationService) EvaluateAndSendAlert(ctx context.Context, task *domain.Task, reason AlertReason, errorMsg *string) {
	if err := ns.SendAlert(ctx, task, reason, errorMsg); err != nil {
		// Log but don't propagate error - alert failures should not affect task execution
		ns.logger.Error("alert sending failed but task execution continues", zap.String("taskID", task.ID), zap.Error(err))
	}
}

// FormatAlertMessage formats an alert into a human-readable message
func FormatAlertMessage(alert *domain.Alert) string {
	msg := fmt.Sprintf("【任务报警】\n")
	msg += fmt.Sprintf("任务ID: %s\n", alert.TaskID)
	msg += fmt.Sprintf("任务名称: %s\n", alert.TaskName)
	msg += fmt.Sprintf("报警原因: %s\n", alert.Reason)
	msg += fmt.Sprintf("发生时间: %s\n", alert.Timestamp)
	msg += fmt.Sprintf("重试次数: %d\n", alert.RetryCount)

	if alert.ErrorMsg != nil {
		msg += fmt.Sprintf("错误信息: %s\n", *alert.ErrorMsg)
	}

	if len(alert.Metadata) > 0 {
		msg += "元数据:\n"
		for k, v := range alert.Metadata {
			msg += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	return msg
}
