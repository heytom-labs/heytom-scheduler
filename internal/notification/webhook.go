package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"task-scheduler/internal/domain"

	"go.uber.org/zap"
)

// WebhookNotifier sends alerts via HTTP webhook
type WebhookNotifier struct {
	url     string
	method  string
	headers map[string]string
	timeout time.Duration
	client  *http.Client
	logger  *zap.Logger
}

// WebhookConfig contains webhook configuration
type WebhookConfig struct {
	URL     string
	Method  string
	Headers map[string]string
	Timeout time.Duration
}

// NewWebhookNotifier creates a new webhook notifier
func NewWebhookNotifier(config WebhookConfig, logger *zap.Logger) *WebhookNotifier {
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &WebhookNotifier{
		url:     config.URL,
		method:  config.Method,
		headers: config.Headers,
		timeout: config.Timeout,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// Send sends a single alert via webhook
func (wn *WebhookNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	// Marshal alert to JSON
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, wn.method, wn.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range wn.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := wn.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	wn.logger.Info("webhook sent successfully", zap.String("url", wn.url), zap.Int("status", resp.StatusCode))
	return nil
}

// SendBatch sends multiple alerts via webhook
func (wn *WebhookNotifier) SendBatch(ctx context.Context, alerts []*domain.Alert) error {
	// Marshal alerts to JSON
	payload, err := json.Marshal(alerts)
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, wn.method, wn.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range wn.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := wn.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	wn.logger.Info("webhook batch sent successfully", zap.String("url", wn.url), zap.Int("count", len(alerts)), zap.Int("status", resp.StatusCode))
	return nil
}

// NewWebhookNotifierFromChannel creates a webhook notifier from channel config
func NewWebhookNotifierFromChannel(channel domain.NotificationChannel, logger *zap.Logger) (*WebhookNotifier, error) {
	url := channel.Config["url"]
	if url == "" {
		return nil, fmt.Errorf("url is required for webhook channel")
	}

	config := WebhookConfig{
		URL:     url,
		Method:  channel.Config["method"],
		Headers: make(map[string]string),
	}

	// Parse timeout if provided
	if timeoutStr := channel.Config["timeout"]; timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err == nil {
			config.Timeout = timeout
		}
	}

	// Parse headers if provided
	for key, value := range channel.Config {
		if len(key) > 7 && key[:7] == "header_" {
			headerName := key[7:]
			config.Headers[headerName] = value
		}
	}

	return NewWebhookNotifier(config, logger), nil
}
