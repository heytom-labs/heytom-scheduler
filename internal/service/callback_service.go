package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"task-scheduler/internal/discovery"
	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"
	"task-scheduler/pkg/metrics"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// CallbackServiceImpl implements the CallbackService interface
type CallbackServiceImpl struct {
	taskService      domain.TaskService
	httpClient       *http.Client
	discoveryFactory DiscoveryFactory
	loadBalancer     discovery.LoadBalancer
}

// DiscoveryFactory creates service discovery instances based on type
type DiscoveryFactory interface {
	GetDiscovery(discoveryType domain.ServiceDiscoveryType) (domain.ServiceDiscovery, error)
}

// CallbackServiceConfig contains configuration for callback service
type CallbackServiceConfig struct {
	DefaultTimeout   time.Duration
	MaxRetries       int
	LoadBalancerType discovery.LoadBalancerStrategy
}

// NewCallbackService creates a new callback service instance
func NewCallbackService(
	taskService domain.TaskService,
	discoveryFactory DiscoveryFactory,
	config *CallbackServiceConfig,
) *CallbackServiceImpl {
	if config == nil {
		config = &CallbackServiceConfig{
			DefaultTimeout:   30 * time.Second,
			MaxRetries:       3,
			LoadBalancerType: discovery.RoundRobin,
		}
	}

	return &CallbackServiceImpl{
		taskService: taskService,
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
		discoveryFactory: discoveryFactory,
		loadBalancer:     discovery.NewLoadBalancer(config.LoadBalancerType),
	}
}

// ExecuteCallback executes a synchronous callback
func (c *CallbackServiceImpl) ExecuteCallback(ctx context.Context, task *domain.Task, result *domain.ExecutionResult) error {
	if task.CallbackConfig == nil {
		return nil // No callback configured
	}

	startTime := time.Now()
	protocol := string(task.CallbackConfig.Protocol)

	logger.InfoContext(ctx, "Executing callback",
		zap.String("task_id", task.ID),
		zap.String("protocol", protocol),
		zap.Bool("is_async", task.CallbackConfig.IsAsync))

	// Resolve service address
	address, err := c.resolveAddress(ctx, task.CallbackConfig)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to resolve callback address",
			zap.String("task_id", task.ID),
			zap.Error(err))
		metrics.RecordCallbackExecution(protocol, "address_resolution_failed")
		return fmt.Errorf("failed to resolve callback address: %w", err)
	}

	// Execute callback based on protocol
	var response *domain.CallbackResponse
	switch task.CallbackConfig.Protocol {
	case domain.CallbackProtocolHTTP:
		response, err = c.executeHTTPCallback(ctx, task.CallbackConfig, address, result)
	case domain.CallbackProtocolGRPC:
		response, err = c.executeGRPCCallback(ctx, task.CallbackConfig, address, result)
	default:
		logger.ErrorContext(ctx, "Unsupported callback protocol",
			zap.String("task_id", task.ID),
			zap.String("protocol", protocol))
		metrics.RecordCallbackExecution(protocol, "unsupported_protocol")
		return fmt.Errorf("unsupported callback protocol: %s", task.CallbackConfig.Protocol)
	}

	duration := time.Since(startTime).Seconds()
	metrics.RecordCallbackExecutionDuration(protocol, duration)

	if err != nil {
		logger.ErrorContext(ctx, "Callback execution failed",
			zap.String("task_id", task.ID),
			zap.String("protocol", protocol),
			zap.Error(err))
		metrics.RecordCallbackExecution(protocol, "failed")
		return fmt.Errorf("callback execution failed: %w", err)
	}

	metrics.RecordCallbackExecution(protocol, "success")
	logger.InfoContext(ctx, "Callback executed successfully",
		zap.String("task_id", task.ID),
		zap.String("protocol", protocol),
		zap.Duration("duration", time.Since(startTime)))

	// Update task status based on callback response
	if response != nil {
		err = c.taskService.UpdateTaskStatus(ctx, task.ID, response.Status)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update task status",
				zap.String("task_id", task.ID),
				zap.Error(err))
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	return nil
}

// ExecuteAsyncCallback executes an asynchronous callback
func (c *CallbackServiceImpl) ExecuteAsyncCallback(ctx context.Context, task *domain.Task) error {
	if task.CallbackConfig == nil {
		return nil // No callback configured
	}

	// For async callbacks, we don't wait for the response
	// Execute in a goroutine
	go func() {
		// Create a new context with timeout for the async operation
		asyncCtx, cancel := context.WithTimeout(context.Background(), task.CallbackConfig.Timeout)
		defer cancel()

		address, err := c.resolveAddress(asyncCtx, task.CallbackConfig)
		if err != nil {
			// Log error but don't fail the task
			fmt.Printf("async callback address resolution failed for task %s: %v\n", task.ID, err)
			return
		}

		// Create execution result for async callback
		result := &domain.ExecutionResult{
			TaskID:    task.ID,
			Status:    task.Status,
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}

		switch task.CallbackConfig.Protocol {
		case domain.CallbackProtocolHTTP:
			_, err = c.executeHTTPCallback(asyncCtx, task.CallbackConfig, address, result)
		case domain.CallbackProtocolGRPC:
			_, err = c.executeGRPCCallback(asyncCtx, task.CallbackConfig, address, result)
		default:
			fmt.Printf("unsupported callback protocol for task %s: %s\n", task.ID, task.CallbackConfig.Protocol)
			return
		}

		if err != nil {
			fmt.Printf("async callback execution failed for task %s: %v\n", task.ID, err)
		}
	}()

	return nil
}

// HandleCallbackResponse handles the response from an async callback
func (c *CallbackServiceImpl) HandleCallbackResponse(ctx context.Context, taskID string, response *domain.CallbackResponse) error {
	if response == nil {
		return fmt.Errorf("callback response is nil")
	}

	// Update task status
	err := c.taskService.UpdateTaskStatus(ctx, taskID, response.Status)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// resolveAddress resolves the callback address using service discovery
func (c *CallbackServiceImpl) resolveAddress(ctx context.Context, config *domain.CallbackConfig) (string, error) {
	// If service discovery is not configured, use the URL directly
	if config.ServiceName == nil || *config.ServiceName == "" {
		return config.URL, nil
	}

	// Get service discovery instance
	sd, err := c.discoveryFactory.GetDiscovery(config.DiscoveryType)
	if err != nil {
		return "", fmt.Errorf("failed to get service discovery: %w", err)
	}

	// Discover service instances
	addresses, err := sd.Discover(ctx, *config.ServiceName)
	if err != nil {
		return "", fmt.Errorf("service discovery failed: %w", err)
	}

	// Use load balancer to select an address
	address, err := c.loadBalancer.Select(addresses)
	if err != nil {
		return "", fmt.Errorf("load balancer selection failed: %w", err)
	}

	return address, nil
}

// executeHTTPCallback executes an HTTP callback
func (c *CallbackServiceImpl) executeHTTPCallback(
	ctx context.Context,
	config *domain.CallbackConfig,
	address string,
	result *domain.ExecutionResult,
) (*domain.CallbackResponse, error) {
	// Prepare request body
	body, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build URL
	url := address
	if config.URL != "" && config.URL != address {
		url = fmt.Sprintf("%s%s", address, config.URL)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, config.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP callback returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var callbackResp domain.CallbackResponse
	if err := json.NewDecoder(resp.Body).Decode(&callbackResp); err != nil {
		// If response parsing fails, consider it successful but without status update
		return nil, nil
	}

	return &callbackResp, nil
}

// executeGRPCCallback executes a gRPC callback
func (c *CallbackServiceImpl) executeGRPCCallback(
	ctx context.Context,
	config *domain.CallbackConfig,
	address string,
	result *domain.ExecutionResult,
) (*domain.CallbackResponse, error) {
	if config.GRPCService == nil || config.GRPCMethod == nil {
		return nil, fmt.Errorf("gRPC service and method must be specified")
	}

	// Create gRPC connection
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer conn.Close()

	// Prepare metadata
	md := metadata.New(config.Headers)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// For simplicity, we'll use a generic invoke
	// In production, you'd want to use generated protobuf code
	method := fmt.Sprintf("/%s/%s", *config.GRPCService, *config.GRPCMethod)

	var response domain.CallbackResponse
	err = conn.Invoke(ctx, method, result, &response)
	if err != nil {
		return nil, fmt.Errorf("gRPC invocation failed: %w", err)
	}

	return &response, nil
}

var _ domain.CallbackService = (*CallbackServiceImpl)(nil)
