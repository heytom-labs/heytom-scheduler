package domain

import "context"

// Alert represents an alert notification
type Alert struct {
	TaskID     string            // 任务ID
	TaskName   string            // 任务名称
	Reason     string            // 报警原因
	Timestamp  string            // 发生时间
	RetryCount int               // 重试次数
	ErrorMsg   *string           // 错误信息
	Metadata   map[string]string // 元数据
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Send sends a single alert
	Send(ctx context.Context, alert *Alert) error

	// SendBatch sends multiple alerts
	SendBatch(ctx context.Context, alerts []*Alert) error
}
