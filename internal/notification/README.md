# Notification System

The notification system provides alert capabilities for the task scheduler. It supports multiple notification channels and integrates seamlessly with the task execution flow.

## Features

- **Multiple Notification Channels**: Email, Webhook, SMS (placeholder)
- **Alert Triggers**: Failure, Retry Threshold, Timeout
- **Flexible Configuration**: Per-task alert policies
- **Non-blocking**: Alert failures don't affect task execution
- **Extensible**: Easy to add new notification channels

## Architecture

### Components

1. **NotificationService**: Main service that manages alert notifications
2. **AlertTrigger**: Evaluates whether alerts should be triggered
3. **Notifiers**: Channel-specific implementations (Email, Webhook, SMS)
4. **NotifierFactory**: Creates notifiers from channel configurations
5. **SchedulerNotificationAdapter**: Bridges notification service with scheduler

### Alert Reasons

- `AlertReasonFailure`: Task execution failed
- `AlertReasonRetryThreshold`: Retry count exceeded threshold
- `AlertReasonTimeout`: Task execution exceeded timeout threshold

## Usage

### Basic Setup

```go
import (
    "task-scheduler/internal/notification"
    "task-scheduler/pkg/logger"
)

// Create notification service
notificationService := notification.NewNotificationService(logger.Get())

// Register notifiers
webhookNotifier := notification.NewWebhookNotifier(webhookConfig, logger.Get())
notificationService.RegisterNotifier(domain.ChannelTypeWebhook, webhookNotifier)

emailNotifier := notification.NewEmailNotifier(emailConfig, logger.Get())
notificationService.RegisterNotifier(domain.ChannelTypeEmail, emailNotifier)
```

### Task Configuration

```go
task := &domain.Task{
    ID:   "task-123",
    Name: "My Task",
    AlertPolicy: &domain.AlertPolicy{
        EnableFailureAlert: true,
        RetryThreshold:     3,
        TimeoutThreshold:   5 * time.Minute,
        Channels: []domain.NotificationChannel{
            {
                Type: domain.ChannelTypeEmail,
                Config: map[string]string{
                    "smtp_host": "smtp.example.com",
                    "smtp_port": "587",
                    "username":  "user@example.com",
                    "password":  "password",
                    "from":      "alerts@example.com",
                },
            },
            {
                Type: domain.ChannelTypeWebhook,
                Config: map[string]string{
                    "url":    "https://example.com/webhook",
                    "method": "POST",
                },
            },
        },
    },
    Metadata: map[string]string{
        "email_recipient": "admin@example.com",
    },
}
```

### Sending Alerts

```go
ctx := context.Background()
errorMsg := "Task execution failed"

// Evaluate and send alert
notificationService.EvaluateAndSendAlert(
    ctx,
    task,
    notification.AlertReasonFailure,
    &errorMsg,
)
```

### Integration with Scheduler

```go
// Create notification adapter
adapter := notification.NewSchedulerNotificationAdapter(notificationService)

// Set in scheduler (if scheduler supports SetNotificationService)
scheduler.SetNotificationService(adapter)
```

## Notification Channels

### Email

Sends alerts via SMTP.

**Configuration:**
- `smtp_host`: SMTP server hostname
- `smtp_port`: SMTP server port
- `username`: SMTP username
- `password`: SMTP password
- `from`: Sender email address

**Metadata:**
- `email_recipient`: Recipient email address (required)

### Webhook

Sends alerts via HTTP POST request.

**Configuration:**
- `url`: Webhook URL (required)
- `method`: HTTP method (default: POST)
- `timeout`: Request timeout (e.g., "10s")
- `header_*`: Custom headers (e.g., `header_Authorization: Bearer token`)

### SMS

Placeholder implementation for SMS notifications.

**Configuration:**
- `api_key`: SMS provider API key
- `api_url`: SMS provider API URL
- `from`: Sender phone number

**Metadata:**
- `phone_number`: Recipient phone number (required)

## Alert Message Format

Alerts include the following information:
- Task ID
- Task Name
- Alert Reason
- Timestamp
- Retry Count
- Error Message (if available)
- Metadata

Example formatted message:
```
【任务报警】
任务ID: task-123
任务名称: My Task
报警原因: 任务执行失败
发生时间: 2024-01-15T10:30:00Z
重试次数: 3
错误信息: Task execution failed due to network timeout
元数据:
  environment: production
  priority: high
```

## Extending with New Channels

To add a new notification channel:

1. Implement the `domain.Notifier` interface:
```go
type MyNotifier struct {
    logger *zap.Logger
}

func (n *MyNotifier) Send(ctx context.Context, alert *domain.Alert) error {
    // Implementation
}

func (n *MyNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
    // Implementation
}
```

2. Add channel type to `domain.ChannelType`:
```go
const (
    ChannelTypeMyChannel ChannelType = "my_channel"
)
```

3. Update `NotifierFactory.CreateNotifier()`:
```go
case domain.ChannelTypeMyChannel:
    return NewMyNotifierFromChannel(channel, nf.logger)
```

## Error Handling

- Alert sending failures are logged but don't affect task execution
- Failed alerts to one channel don't prevent sending to other channels
- All errors are logged with appropriate context

## Best Practices

1. **Configure appropriate thresholds**: Set `RetryThreshold` and `TimeoutThreshold` based on your task requirements
2. **Use multiple channels**: Configure multiple notification channels for critical tasks
3. **Include metadata**: Add relevant metadata to tasks for better alert context
4. **Test configurations**: Verify notification configurations before deploying to production
5. **Monitor alert delivery**: Check logs to ensure alerts are being delivered successfully

## Examples

See `examples/notification_example.go` for a complete working example.
