package notification

import (
	"context"
	"fmt"
	"net/smtp"

	"task-scheduler/internal/domain"

	"go.uber.org/zap"
)

// EmailNotifier sends alerts via email
type EmailNotifier struct {
	smtpHost string
	smtpPort string
	username string
	password string
	from     string
	logger   *zap.Logger
}

// EmailConfig contains email configuration
type EmailConfig struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
	From     string
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config EmailConfig, logger *zap.Logger) *EmailNotifier {
	return &EmailNotifier{
		smtpHost: config.SMTPHost,
		smtpPort: config.SMTPPort,
		username: config.Username,
		password: config.Password,
		from:     config.From,
		logger:   logger,
	}
}

// Send sends a single alert via email
func (en *EmailNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	// Get recipient from alert metadata or use default
	to := en.getRecipient(alert)
	if to == "" {
		return fmt.Errorf("no recipient specified for email alert")
	}

	subject := fmt.Sprintf("任务报警: %s", alert.TaskName)
	body := FormatAlertMessage(alert)

	return en.sendEmail(to, subject, body)
}

// SendBatch sends multiple alerts via email
func (en *EmailNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
	for _, alert := range alerts {
		if err := en.Send(ctx, alert); err != nil {
			en.logger.Error("failed to send email alert", zap.String("taskID", alert.TaskID), zap.Error(err))
			// Continue with other alerts
		}
	}
	return nil
}

func (en *EmailNotifier) getRecipient(alert *domain.Alert) string {
	if alert.Metadata != nil {
		if recipient, ok := alert.Metadata["email_recipient"]; ok {
			return recipient
		}
	}
	return ""
}

func (en *EmailNotifier) sendEmail(to, subject, body string) error {
	// Build email message
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	// Setup authentication
	auth := smtp.PlainAuth("", en.username, en.password, en.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", en.smtpHost, en.smtpPort)
	recipients := []string{to}

	err := smtp.SendMail(addr, auth, en.from, recipients, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	en.logger.Info("email sent successfully", zap.String("to", to), zap.String("subject", subject))
	return nil
}

// NewEmailNotifierFromChannel creates an email notifier from channel config
func NewEmailNotifierFromChannel(channel domain.NotificationChannel, logger *zap.Logger) (*EmailNotifier, error) {
	config := EmailConfig{
		SMTPHost: channel.Config["smtp_host"],
		SMTPPort: channel.Config["smtp_port"],
		Username: channel.Config["username"],
		Password: channel.Config["password"],
		From:     channel.Config["from"],
	}

	// Validate required fields
	if config.SMTPHost == "" || config.SMTPPort == "" {
		return nil, fmt.Errorf("smtp_host and smtp_port are required")
	}

	return NewEmailNotifier(config, logger), nil
}
