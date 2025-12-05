package notification

import (
	"context"
	"fmt"

	"task-scheduler/internal/domain"

	"go.uber.org/zap"
)

// SMSNotifier sends alerts via SMS (placeholder implementation)
type SMSNotifier struct {
	apiKey string
	apiURL string
	from   string
	logger *zap.Logger
}

// SMSConfig contains SMS configuration
type SMSConfig struct {
	APIKey string
	APIURL string
	From   string
}

// NewSMSNotifier creates a new SMS notifier
func NewSMSNotifier(config SMSConfig, logger *zap.Logger) *SMSNotifier {
	return &SMSNotifier{
		apiKey: config.APIKey,
		apiURL: config.APIURL,
		from:   config.From,
		logger: logger,
	}
}

// Send sends a single alert via SMS
func (sn *SMSNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	// Get recipient phone number from alert metadata
	to := sn.getRecipient(alert)
	if to == "" {
		return fmt.Errorf("no recipient phone number specified for SMS alert")
	}

	// Format message (SMS has character limits)
	message := sn.formatSMSMessage(alert)

	// This is a placeholder implementation
	// In a real implementation, you would integrate with an SMS provider API
	// such as Twilio, AWS SNS, Alibaba Cloud SMS, etc.
	sn.logger.Info("SMS alert (placeholder)", zap.String("to", to), zap.String("message", message))

	// TODO: Implement actual SMS sending logic
	// Example with Twilio:
	// return sn.sendViaTwilio(to, message)

	return nil
}

// SendBatch sends multiple alerts via SMS
func (sn *SMSNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
	for _, alert := range alerts {
		if err := sn.Send(ctx, alert); err != nil {
			sn.logger.Error("failed to send SMS alert", zap.String("taskID", alert.TaskID), zap.Error(err))
			// Continue with other alerts
		}
	}
	return nil
}

func (sn *SMSNotifier) getRecipient(alert *domain.Alert) string {
	if alert.Metadata != nil {
		if phone, ok := alert.Metadata["phone_number"]; ok {
			return phone
		}
	}
	return ""
}

func (sn *SMSNotifier) formatSMSMessage(alert *domain.Alert) string {
	// Keep SMS messages short due to character limits
	msg := fmt.Sprintf("【任务报警】%s: %s (任务ID: %s, 时间: %s)",
		alert.TaskName,
		alert.Reason,
		alert.TaskID,
		alert.Timestamp)

	// Truncate if too long (typical SMS limit is 160 characters)
	if len(msg) > 160 {
		msg = msg[:157] + "..."
	}

	return msg
}

// NewSMSNotifierFromChannel creates an SMS notifier from channel config
func NewSMSNotifierFromChannel(channel domain.NotificationChannel, logger *zap.Logger) (*SMSNotifier, error) {
	config := SMSConfig{
		APIKey: channel.Config["api_key"],
		APIURL: channel.Config["api_url"],
		From:   channel.Config["from"],
	}

	// Validate required fields
	if config.APIKey == "" || config.APIURL == "" {
		return nil, fmt.Errorf("api_key and api_url are required for SMS channel")
	}

	return NewSMSNotifier(config, logger), nil
}
