package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the task scheduler
type Handler struct {
	taskService domain.TaskService
}

// NewHandler creates a new HTTP handler
func NewHandler(taskService domain.TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

// SetupRoutes configures the HTTP routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", h.CreateTask)
			tasks.GET("", h.ListTasks)
			tasks.GET("/:id", h.GetTask)
			tasks.DELETE("/:id", h.CancelTask)
			tasks.GET("/:id/status", h.GetTaskStatus)
		}
	}

	// Health check
	router.GET("/health", h.HealthCheck)
}

// CreateTaskRequest represents the HTTP request for creating a task
type CreateTaskRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Description      string                 `json:"description"`
	ParentID         *string                `json:"parent_id,omitempty"`
	ExecutionMode    domain.ExecutionMode   `json:"execution_mode" binding:"required"`
	ScheduleConfig   *ScheduleConfigRequest `json:"schedule_config,omitempty"`
	CallbackConfig   *CallbackConfigRequest `json:"callback_config,omitempty"`
	RetryPolicy      *RetryPolicyRequest    `json:"retry_policy,omitempty"`
	ConcurrencyLimit int                    `json:"concurrency_limit"`
	AlertPolicy      *AlertPolicyRequest    `json:"alert_policy,omitempty"`
	Metadata         map[string]string      `json:"metadata,omitempty"`
}

// ScheduleConfigRequest represents schedule configuration in HTTP request
type ScheduleConfigRequest struct {
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
	Interval      *string    `json:"interval,omitempty"` // Duration string like "5m", "1h"
	CronExpr      *string    `json:"cron_expr,omitempty"`
}

// CallbackConfigRequest represents callback configuration in HTTP request
type CallbackConfigRequest struct {
	Protocol      domain.CallbackProtocol     `json:"protocol" binding:"required"`
	URL           string                      `json:"url" binding:"required"`
	Method        string                      `json:"method,omitempty"`
	GRPCService   *string                     `json:"grpc_service,omitempty"`
	GRPCMethod    *string                     `json:"grpc_method,omitempty"`
	Headers       map[string]string           `json:"headers,omitempty"`
	Timeout       string                      `json:"timeout,omitempty"` // Duration string
	IsAsync       bool                        `json:"is_async"`
	ServiceName   *string                     `json:"service_name,omitempty"`
	DiscoveryType domain.ServiceDiscoveryType `json:"discovery_type,omitempty"`
}

// RetryPolicyRequest represents retry policy in HTTP request
type RetryPolicyRequest struct {
	MaxRetries    int     `json:"max_retries"`
	RetryInterval string  `json:"retry_interval"` // Duration string
	BackoffFactor float64 `json:"backoff_factor"`
}

// AlertPolicyRequest represents alert policy in HTTP request
type AlertPolicyRequest struct {
	EnableFailureAlert bool                         `json:"enable_failure_alert"`
	RetryThreshold     int                          `json:"retry_threshold"`
	TimeoutThreshold   string                       `json:"timeout_threshold"` // Duration string
	Channels           []NotificationChannelRequest `json:"channels,omitempty"`
}

// NotificationChannelRequest represents notification channel in HTTP request
type NotificationChannelRequest struct {
	Type   domain.ChannelType `json:"type" binding:"required"`
	Config map[string]string  `json:"config,omitempty"`
}

// TaskResponse represents the HTTP response for a task
type TaskResponse struct {
	ID               string                  `json:"id"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	ParentID         *string                 `json:"parent_id,omitempty"`
	ExecutionMode    domain.ExecutionMode    `json:"execution_mode"`
	ScheduleConfig   *ScheduleConfigResponse `json:"schedule_config,omitempty"`
	CallbackConfig   *CallbackConfigResponse `json:"callback_config,omitempty"`
	RetryPolicy      *RetryPolicyResponse    `json:"retry_policy,omitempty"`
	ConcurrencyLimit int                     `json:"concurrency_limit"`
	AlertPolicy      *AlertPolicyResponse    `json:"alert_policy,omitempty"`
	Status           domain.TaskStatus       `json:"status"`
	RetryCount       int                     `json:"retry_count"`
	NodeID           *string                 `json:"node_id,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	StartedAt        *time.Time              `json:"started_at,omitempty"`
	CompletedAt      *time.Time              `json:"completed_at,omitempty"`
	Metadata         map[string]string       `json:"metadata,omitempty"`
}

// ScheduleConfigResponse represents schedule configuration in HTTP response
type ScheduleConfigResponse struct {
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
	Interval      *string    `json:"interval,omitempty"`
	CronExpr      *string    `json:"cron_expr,omitempty"`
}

// CallbackConfigResponse represents callback configuration in HTTP response
type CallbackConfigResponse struct {
	Protocol      domain.CallbackProtocol     `json:"protocol"`
	URL           string                      `json:"url"`
	Method        string                      `json:"method,omitempty"`
	GRPCService   *string                     `json:"grpc_service,omitempty"`
	GRPCMethod    *string                     `json:"grpc_method,omitempty"`
	Headers       map[string]string           `json:"headers,omitempty"`
	Timeout       string                      `json:"timeout,omitempty"`
	IsAsync       bool                        `json:"is_async"`
	ServiceName   *string                     `json:"service_name,omitempty"`
	DiscoveryType domain.ServiceDiscoveryType `json:"discovery_type,omitempty"`
}

// RetryPolicyResponse represents retry policy in HTTP response
type RetryPolicyResponse struct {
	MaxRetries    int     `json:"max_retries"`
	RetryInterval string  `json:"retry_interval"`
	BackoffFactor float64 `json:"backoff_factor"`
}

// AlertPolicyResponse represents alert policy in HTTP response
type AlertPolicyResponse struct {
	EnableFailureAlert bool                          `json:"enable_failure_alert"`
	RetryThreshold     int                           `json:"retry_threshold"`
	TimeoutThreshold   string                        `json:"timeout_threshold"`
	Channels           []NotificationChannelResponse `json:"channels,omitempty"`
}

// NotificationChannelResponse represents notification channel in HTTP response
type NotificationChannelResponse struct {
	Type   domain.ChannelType `json:"type"`
	Config map[string]string  `json:"config,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// CreateTask handles POST /api/v1/tasks
func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Convert HTTP request to domain request
	domainReq, err := h.convertCreateTaskRequest(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Create task
	task, err := h.taskService.CreateTask(c.Request.Context(), domainReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response
	resp := h.convertTaskToResponse(task)
	c.JSON(http.StatusCreated, resp)
}

// GetTask handles GET /api/v1/tasks/:id
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "task ID is required",
		})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := h.convertTaskToResponse(task)
	c.JSON(http.StatusOK, resp)
}

// ListTasks handles GET /api/v1/tasks
func (h *Handler) ListTasks(c *gin.Context) {
	// Parse query parameters
	filter := &domain.TaskFilter{}

	if status := c.Query("status"); status != "" {
		taskStatus := domain.TaskStatus(status)
		filter.Status = &taskStatus
	}

	if parentID := c.Query("parent_id"); parentID != "" {
		filter.ParentID = &parentID
	}

	if nodeID := c.Query("node_id"); nodeID != "" {
		filter.NodeID = &nodeID
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	tasks, err := h.taskService.ListTasks(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response
	resp := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		resp[i] = h.convertTaskToResponse(task)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": resp,
		"count": len(resp),
	})
}

// CancelTask handles DELETE /api/v1/tasks/:id
func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "task ID is required",
		})
		return
	}

	err := h.taskService.CancelTask(c.Request.Context(), taskID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task cancelled successfully",
		"task_id": taskID,
	})
}

// GetTaskStatus handles GET /api/v1/tasks/:id/status
func (h *Handler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "task ID is required",
		})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":      task.ID,
		"status":       task.Status,
		"retry_count":  task.RetryCount,
		"created_at":   task.CreatedAt,
		"updated_at":   task.UpdatedAt,
		"started_at":   task.StartedAt,
		"completed_at": task.CompletedAt,
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}

// convertCreateTaskRequest converts HTTP request to domain request
func (h *Handler) convertCreateTaskRequest(req *CreateTaskRequest) (*domain.CreateTaskRequest, error) {
	domainReq := &domain.CreateTaskRequest{
		Name:             req.Name,
		Description:      req.Description,
		ParentID:         req.ParentID,
		ExecutionMode:    req.ExecutionMode,
		ConcurrencyLimit: req.ConcurrencyLimit,
		Metadata:         req.Metadata,
	}

	// Convert schedule config
	if req.ScheduleConfig != nil {
		scheduleConfig := &domain.ScheduleConfig{
			ScheduledTime: req.ScheduleConfig.ScheduledTime,
			CronExpr:      req.ScheduleConfig.CronExpr,
		}

		if req.ScheduleConfig.Interval != nil {
			duration, err := time.ParseDuration(*req.ScheduleConfig.Interval)
			if err != nil {
				return nil, errors.New("invalid interval format: " + err.Error())
			}
			scheduleConfig.Interval = &duration
		}

		domainReq.ScheduleConfig = scheduleConfig
	}

	// Convert callback config
	if req.CallbackConfig != nil {
		callbackConfig := &domain.CallbackConfig{
			Protocol:      req.CallbackConfig.Protocol,
			URL:           req.CallbackConfig.URL,
			Method:        req.CallbackConfig.Method,
			GRPCService:   req.CallbackConfig.GRPCService,
			GRPCMethod:    req.CallbackConfig.GRPCMethod,
			Headers:       req.CallbackConfig.Headers,
			IsAsync:       req.CallbackConfig.IsAsync,
			ServiceName:   req.CallbackConfig.ServiceName,
			DiscoveryType: req.CallbackConfig.DiscoveryType,
		}

		if req.CallbackConfig.Timeout != "" {
			timeout, err := time.ParseDuration(req.CallbackConfig.Timeout)
			if err != nil {
				return nil, errors.New("invalid timeout format: " + err.Error())
			}
			callbackConfig.Timeout = timeout
		}

		domainReq.CallbackConfig = callbackConfig
	}

	// Convert retry policy
	if req.RetryPolicy != nil {
		retryInterval, err := time.ParseDuration(req.RetryPolicy.RetryInterval)
		if err != nil {
			return nil, errors.New("invalid retry interval format: " + err.Error())
		}

		domainReq.RetryPolicy = &domain.RetryPolicy{
			MaxRetries:    req.RetryPolicy.MaxRetries,
			RetryInterval: retryInterval,
			BackoffFactor: req.RetryPolicy.BackoffFactor,
		}
	}

	// Convert alert policy
	if req.AlertPolicy != nil {
		alertPolicy := &domain.AlertPolicy{
			EnableFailureAlert: req.AlertPolicy.EnableFailureAlert,
			RetryThreshold:     req.AlertPolicy.RetryThreshold,
		}

		if req.AlertPolicy.TimeoutThreshold != "" {
			timeout, err := time.ParseDuration(req.AlertPolicy.TimeoutThreshold)
			if err != nil {
				return nil, errors.New("invalid timeout threshold format: " + err.Error())
			}
			alertPolicy.TimeoutThreshold = timeout
		}

		// Convert channels
		if len(req.AlertPolicy.Channels) > 0 {
			channels := make([]domain.NotificationChannel, len(req.AlertPolicy.Channels))
			for i, ch := range req.AlertPolicy.Channels {
				channels[i] = domain.NotificationChannel{
					Type:   ch.Type,
					Config: ch.Config,
				}
			}
			alertPolicy.Channels = channels
		}

		domainReq.AlertPolicy = alertPolicy
	}

	return domainReq, nil
}

// convertTaskToResponse converts domain task to HTTP response
func (h *Handler) convertTaskToResponse(task *domain.Task) TaskResponse {
	resp := TaskResponse{
		ID:               task.ID,
		Name:             task.Name,
		Description:      task.Description,
		ParentID:         task.ParentID,
		ExecutionMode:    task.ExecutionMode,
		ConcurrencyLimit: task.ConcurrencyLimit,
		Status:           task.Status,
		RetryCount:       task.RetryCount,
		NodeID:           task.NodeID,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
		StartedAt:        task.StartedAt,
		CompletedAt:      task.CompletedAt,
		Metadata:         task.Metadata,
	}

	// Convert schedule config
	if task.ScheduleConfig != nil {
		scheduleConfig := &ScheduleConfigResponse{
			ScheduledTime: task.ScheduleConfig.ScheduledTime,
			CronExpr:      task.ScheduleConfig.CronExpr,
		}

		if task.ScheduleConfig.Interval != nil {
			interval := task.ScheduleConfig.Interval.String()
			scheduleConfig.Interval = &interval
		}

		resp.ScheduleConfig = scheduleConfig
	}

	// Convert callback config
	if task.CallbackConfig != nil {
		callbackConfig := &CallbackConfigResponse{
			Protocol:      task.CallbackConfig.Protocol,
			URL:           task.CallbackConfig.URL,
			Method:        task.CallbackConfig.Method,
			GRPCService:   task.CallbackConfig.GRPCService,
			GRPCMethod:    task.CallbackConfig.GRPCMethod,
			Headers:       task.CallbackConfig.Headers,
			IsAsync:       task.CallbackConfig.IsAsync,
			ServiceName:   task.CallbackConfig.ServiceName,
			DiscoveryType: task.CallbackConfig.DiscoveryType,
		}

		if task.CallbackConfig.Timeout > 0 {
			timeout := task.CallbackConfig.Timeout.String()
			callbackConfig.Timeout = timeout
		}

		resp.CallbackConfig = callbackConfig
	}

	// Convert retry policy
	if task.RetryPolicy != nil {
		resp.RetryPolicy = &RetryPolicyResponse{
			MaxRetries:    task.RetryPolicy.MaxRetries,
			RetryInterval: task.RetryPolicy.RetryInterval.String(),
			BackoffFactor: task.RetryPolicy.BackoffFactor,
		}
	}

	// Convert alert policy
	if task.AlertPolicy != nil {
		alertPolicy := &AlertPolicyResponse{
			EnableFailureAlert: task.AlertPolicy.EnableFailureAlert,
			RetryThreshold:     task.AlertPolicy.RetryThreshold,
			TimeoutThreshold:   task.AlertPolicy.TimeoutThreshold.String(),
		}

		// Convert channels
		if len(task.AlertPolicy.Channels) > 0 {
			channels := make([]NotificationChannelResponse, len(task.AlertPolicy.Channels))
			for i, ch := range task.AlertPolicy.Channels {
				channels[i] = NotificationChannelResponse{
					Type:   ch.Type,
					Config: ch.Config,
				}
			}
			alertPolicy.Channels = channels
		}

		resp.AlertPolicy = alertPolicy
	}

	return resp
}

// handleError handles errors and returns appropriate HTTP responses
func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "task not found",
		})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_input",
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrInvalidExecutionMode):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_execution_mode",
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrInvalidScheduleConfig):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_schedule_config",
			Message: err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "an internal error occurred",
		})
	}
}
