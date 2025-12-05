package grpc

import (
	"context"
	"errors"
	"time"

	"task-scheduler/api/pb"
	"task-scheduler/internal/domain"
	"task-scheduler/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the gRPC TaskSchedulerService
type Handler struct {
	pb.UnimplementedTaskSchedulerServiceServer
	taskService domain.TaskService
}

// NewHandler creates a new gRPC handler
func NewHandler(taskService domain.TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

// CreateTask implements the CreateTask RPC method
func (h *Handler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "task name is required")
	}

	// Convert proto request to domain request
	domainReq, err := h.convertCreateTaskRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Create task
	task, err := h.taskService.CreateTask(ctx, domainReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	// Convert to proto response
	protoTask := h.convertTaskToProto(task)
	return &pb.CreateTaskResponse{
		Task: protoTask,
	}, nil
}

// GetTask implements the GetTask RPC method
func (h *Handler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task ID is required")
	}

	task, err := h.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, h.handleError(err)
	}

	protoTask := h.convertTaskToProto(task)
	return &pb.GetTaskResponse{
		Task: protoTask,
	}, nil
}

// ListTasks implements the ListTasks RPC method
func (h *Handler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	// Convert proto filter to domain filter
	filter := &domain.TaskFilter{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	if req.Status != pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		domainStatus := h.convertProtoStatusToDomain(req.Status)
		filter.Status = &domainStatus
	}

	if req.ParentId != "" {
		filter.ParentID = &req.ParentId
	}

	if req.NodeId != "" {
		filter.NodeID = &req.NodeId
	}

	tasks, err := h.taskService.ListTasks(ctx, filter)
	if err != nil {
		return nil, h.handleError(err)
	}

	// Convert to proto response
	protoTasks := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		protoTasks[i] = h.convertTaskToProto(task)
	}

	return &pb.ListTasksResponse{
		Tasks: protoTasks,
		Count: int32(len(protoTasks)),
	}, nil
}

// CancelTask implements the CancelTask RPC method
func (h *Handler) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task ID is required")
	}

	err := h.taskService.CancelTask(ctx, req.TaskId)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &pb.CancelTaskResponse{
		Message: "task cancelled successfully",
		TaskId:  req.TaskId,
	}, nil
}

// GetTaskStatus implements the GetTaskStatus RPC method
func (h *Handler) GetTaskStatus(ctx context.Context, req *pb.GetTaskStatusRequest) (*pb.GetTaskStatusResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task ID is required")
	}

	task, err := h.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &pb.GetTaskStatusResponse{
		TaskId:      task.ID,
		Status:      h.convertDomainStatusToProto(task.Status),
		RetryCount:  int32(task.RetryCount),
		CreatedAt:   task.CreatedAt.Unix(),
		UpdatedAt:   task.UpdatedAt.Unix(),
		StartedAt:   timeToUnix(task.StartedAt),
		CompletedAt: timeToUnix(task.CompletedAt),
	}, nil
}

// convertCreateTaskRequest converts proto request to domain request
func (h *Handler) convertCreateTaskRequest(req *pb.CreateTaskRequest) (*domain.CreateTaskRequest, error) {
	domainReq := &domain.CreateTaskRequest{
		Name:             req.Name,
		Description:      req.Description,
		ExecutionMode:    h.convertProtoExecutionModeToDomain(req.ExecutionMode),
		ConcurrencyLimit: int(req.ConcurrencyLimit),
		Metadata:         req.Metadata,
	}

	if req.ParentId != "" {
		domainReq.ParentID = &req.ParentId
	}

	// Convert schedule config
	if req.ScheduleConfig != nil {
		scheduleConfig := &domain.ScheduleConfig{
			CronExpr: stringPtr(req.ScheduleConfig.CronExpr),
		}

		if req.ScheduleConfig.ScheduledTime > 0 {
			t := time.Unix(req.ScheduleConfig.ScheduledTime, 0)
			scheduleConfig.ScheduledTime = &t
		}

		if req.ScheduleConfig.IntervalSeconds > 0 {
			interval := time.Duration(req.ScheduleConfig.IntervalSeconds) * time.Second
			scheduleConfig.Interval = &interval
		}

		domainReq.ScheduleConfig = scheduleConfig
	}

	// Convert callback config
	if req.CallbackConfig != nil {
		callbackConfig := &domain.CallbackConfig{
			Protocol:      h.convertProtoCallbackProtocolToDomain(req.CallbackConfig.Protocol),
			URL:           req.CallbackConfig.Url,
			Method:        req.CallbackConfig.Method,
			Headers:       req.CallbackConfig.Headers,
			IsAsync:       req.CallbackConfig.IsAsync,
			DiscoveryType: h.convertProtoServiceDiscoveryToDomain(req.CallbackConfig.DiscoveryType),
		}

		if req.CallbackConfig.GrpcService != "" {
			callbackConfig.GRPCService = &req.CallbackConfig.GrpcService
		}

		if req.CallbackConfig.GrpcMethod != "" {
			callbackConfig.GRPCMethod = &req.CallbackConfig.GrpcMethod
		}

		if req.CallbackConfig.TimeoutSeconds > 0 {
			callbackConfig.Timeout = time.Duration(req.CallbackConfig.TimeoutSeconds) * time.Second
		}

		if req.CallbackConfig.ServiceName != "" {
			callbackConfig.ServiceName = &req.CallbackConfig.ServiceName
		}

		domainReq.CallbackConfig = callbackConfig
	}

	// Convert retry policy
	if req.RetryPolicy != nil {
		domainReq.RetryPolicy = &domain.RetryPolicy{
			MaxRetries:    int(req.RetryPolicy.MaxRetries),
			RetryInterval: time.Duration(req.RetryPolicy.RetryIntervalSeconds) * time.Second,
			BackoffFactor: req.RetryPolicy.BackoffFactor,
		}
	}

	// Convert alert policy
	if req.AlertPolicy != nil {
		alertPolicy := &domain.AlertPolicy{
			EnableFailureAlert: req.AlertPolicy.EnableFailureAlert,
			RetryThreshold:     int(req.AlertPolicy.RetryThreshold),
			TimeoutThreshold:   time.Duration(req.AlertPolicy.TimeoutThresholdSeconds) * time.Second,
		}

		// Convert channels
		if len(req.AlertPolicy.Channels) > 0 {
			channels := make([]domain.NotificationChannel, len(req.AlertPolicy.Channels))
			for i, ch := range req.AlertPolicy.Channels {
				channels[i] = domain.NotificationChannel{
					Type:   h.convertProtoChannelTypeToDomain(ch.Type),
					Config: ch.Config,
				}
			}
			alertPolicy.Channels = channels
		}

		domainReq.AlertPolicy = alertPolicy
	}

	return domainReq, nil
}

// convertTaskToProto converts domain task to proto task
func (h *Handler) convertTaskToProto(task *domain.Task) *pb.Task {
	protoTask := &pb.Task{
		Id:               task.ID,
		Name:             task.Name,
		Description:      task.Description,
		ParentId:         stringValue(task.ParentID),
		ExecutionMode:    h.convertDomainExecutionModeToProto(task.ExecutionMode),
		ConcurrencyLimit: int32(task.ConcurrencyLimit),
		Status:           h.convertDomainStatusToProto(task.Status),
		RetryCount:       int32(task.RetryCount),
		NodeId:           stringValue(task.NodeID),
		CreatedAt:        task.CreatedAt.Unix(),
		UpdatedAt:        task.UpdatedAt.Unix(),
		StartedAt:        timeToUnix(task.StartedAt),
		CompletedAt:      timeToUnix(task.CompletedAt),
		Metadata:         task.Metadata,
	}

	// Convert schedule config
	if task.ScheduleConfig != nil {
		scheduleConfig := &pb.ScheduleConfig{
			CronExpr: stringValue(task.ScheduleConfig.CronExpr),
		}

		if task.ScheduleConfig.ScheduledTime != nil {
			scheduleConfig.ScheduledTime = task.ScheduleConfig.ScheduledTime.Unix()
		}

		if task.ScheduleConfig.Interval != nil {
			scheduleConfig.IntervalSeconds = int64(task.ScheduleConfig.Interval.Seconds())
		}

		protoTask.ScheduleConfig = scheduleConfig
	}

	// Convert callback config
	if task.CallbackConfig != nil {
		callbackConfig := &pb.CallbackConfig{
			Protocol:      h.convertDomainCallbackProtocolToProto(task.CallbackConfig.Protocol),
			Url:           task.CallbackConfig.URL,
			Method:        task.CallbackConfig.Method,
			Headers:       task.CallbackConfig.Headers,
			IsAsync:       task.CallbackConfig.IsAsync,
			DiscoveryType: h.convertDomainServiceDiscoveryToProto(task.CallbackConfig.DiscoveryType),
		}

		if task.CallbackConfig.GRPCService != nil {
			callbackConfig.GrpcService = *task.CallbackConfig.GRPCService
		}

		if task.CallbackConfig.GRPCMethod != nil {
			callbackConfig.GrpcMethod = *task.CallbackConfig.GRPCMethod
		}

		if task.CallbackConfig.Timeout > 0 {
			callbackConfig.TimeoutSeconds = int64(task.CallbackConfig.Timeout.Seconds())
		}

		if task.CallbackConfig.ServiceName != nil {
			callbackConfig.ServiceName = *task.CallbackConfig.ServiceName
		}

		protoTask.CallbackConfig = callbackConfig
	}

	// Convert retry policy
	if task.RetryPolicy != nil {
		protoTask.RetryPolicy = &pb.RetryPolicy{
			MaxRetries:           int32(task.RetryPolicy.MaxRetries),
			RetryIntervalSeconds: int64(task.RetryPolicy.RetryInterval.Seconds()),
			BackoffFactor:        task.RetryPolicy.BackoffFactor,
		}
	}

	// Convert alert policy
	if task.AlertPolicy != nil {
		alertPolicy := &pb.AlertPolicy{
			EnableFailureAlert:      task.AlertPolicy.EnableFailureAlert,
			RetryThreshold:          int32(task.AlertPolicy.RetryThreshold),
			TimeoutThresholdSeconds: int64(task.AlertPolicy.TimeoutThreshold.Seconds()),
		}

		// Convert channels
		if len(task.AlertPolicy.Channels) > 0 {
			channels := make([]*pb.NotificationChannel, len(task.AlertPolicy.Channels))
			for i, ch := range task.AlertPolicy.Channels {
				channels[i] = &pb.NotificationChannel{
					Type:   h.convertDomainChannelTypeToProto(ch.Type),
					Config: ch.Config,
				}
			}
			alertPolicy.Channels = channels
		}

		protoTask.AlertPolicy = alertPolicy
	}

	return protoTask
}

// Enum conversion functions
func (h *Handler) convertProtoExecutionModeToDomain(mode pb.ExecutionMode) domain.ExecutionMode {
	switch mode {
	case pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE:
		return domain.ExecutionModeImmediate
	case pb.ExecutionMode_EXECUTION_MODE_SCHEDULED:
		return domain.ExecutionModeScheduled
	case pb.ExecutionMode_EXECUTION_MODE_INTERVAL:
		return domain.ExecutionModeInterval
	case pb.ExecutionMode_EXECUTION_MODE_CRON:
		return domain.ExecutionModeCron
	default:
		return domain.ExecutionModeImmediate
	}
}

func (h *Handler) convertDomainExecutionModeToProto(mode domain.ExecutionMode) pb.ExecutionMode {
	switch mode {
	case domain.ExecutionModeImmediate:
		return pb.ExecutionMode_EXECUTION_MODE_IMMEDIATE
	case domain.ExecutionModeScheduled:
		return pb.ExecutionMode_EXECUTION_MODE_SCHEDULED
	case domain.ExecutionModeInterval:
		return pb.ExecutionMode_EXECUTION_MODE_INTERVAL
	case domain.ExecutionModeCron:
		return pb.ExecutionMode_EXECUTION_MODE_CRON
	default:
		return pb.ExecutionMode_EXECUTION_MODE_UNSPECIFIED
	}
}

func (h *Handler) convertProtoStatusToDomain(status pb.TaskStatus) domain.TaskStatus {
	switch status {
	case pb.TaskStatus_TASK_STATUS_PENDING:
		return domain.TaskStatusPending
	case pb.TaskStatus_TASK_STATUS_RUNNING:
		return domain.TaskStatusRunning
	case pb.TaskStatus_TASK_STATUS_SUCCESS:
		return domain.TaskStatusSuccess
	case pb.TaskStatus_TASK_STATUS_FAILED:
		return domain.TaskStatusFailed
	case pb.TaskStatus_TASK_STATUS_CANCELLED:
		return domain.TaskStatusCancelled
	default:
		return domain.TaskStatusPending
	}
}

func (h *Handler) convertDomainStatusToProto(status domain.TaskStatus) pb.TaskStatus {
	switch status {
	case domain.TaskStatusPending:
		return pb.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusRunning:
		return pb.TaskStatus_TASK_STATUS_RUNNING
	case domain.TaskStatusSuccess:
		return pb.TaskStatus_TASK_STATUS_SUCCESS
	case domain.TaskStatusFailed:
		return pb.TaskStatus_TASK_STATUS_FAILED
	case domain.TaskStatusCancelled:
		return pb.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func (h *Handler) convertProtoCallbackProtocolToDomain(protocol pb.CallbackProtocol) domain.CallbackProtocol {
	switch protocol {
	case pb.CallbackProtocol_CALLBACK_PROTOCOL_HTTP:
		return domain.CallbackProtocolHTTP
	case pb.CallbackProtocol_CALLBACK_PROTOCOL_GRPC:
		return domain.CallbackProtocolGRPC
	default:
		return domain.CallbackProtocolHTTP
	}
}

func (h *Handler) convertDomainCallbackProtocolToProto(protocol domain.CallbackProtocol) pb.CallbackProtocol {
	switch protocol {
	case domain.CallbackProtocolHTTP:
		return pb.CallbackProtocol_CALLBACK_PROTOCOL_HTTP
	case domain.CallbackProtocolGRPC:
		return pb.CallbackProtocol_CALLBACK_PROTOCOL_GRPC
	default:
		return pb.CallbackProtocol_CALLBACK_PROTOCOL_UNSPECIFIED
	}
}

func (h *Handler) convertProtoServiceDiscoveryToDomain(discoveryType pb.ServiceDiscoveryType) domain.ServiceDiscoveryType {
	switch discoveryType {
	case pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_STATIC:
		return domain.ServiceDiscoveryStatic
	case pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_CONSUL:
		return domain.ServiceDiscoveryConsul
	case pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_ETCD:
		return domain.ServiceDiscoveryEtcd
	case pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_KUBERNETES:
		return domain.ServiceDiscoveryKubernetes
	default:
		return domain.ServiceDiscoveryStatic
	}
}

func (h *Handler) convertDomainServiceDiscoveryToProto(discoveryType domain.ServiceDiscoveryType) pb.ServiceDiscoveryType {
	switch discoveryType {
	case domain.ServiceDiscoveryStatic:
		return pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_STATIC
	case domain.ServiceDiscoveryConsul:
		return pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_CONSUL
	case domain.ServiceDiscoveryEtcd:
		return pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_ETCD
	case domain.ServiceDiscoveryKubernetes:
		return pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_KUBERNETES
	default:
		return pb.ServiceDiscoveryType_SERVICE_DISCOVERY_TYPE_UNSPECIFIED
	}
}

func (h *Handler) convertProtoChannelTypeToDomain(channelType pb.ChannelType) domain.ChannelType {
	switch channelType {
	case pb.ChannelType_CHANNEL_TYPE_EMAIL:
		return domain.ChannelTypeEmail
	case pb.ChannelType_CHANNEL_TYPE_WEBHOOK:
		return domain.ChannelTypeWebhook
	case pb.ChannelType_CHANNEL_TYPE_SMS:
		return domain.ChannelTypeSMS
	default:
		return domain.ChannelTypeEmail
	}
}

func (h *Handler) convertDomainChannelTypeToProto(channelType domain.ChannelType) pb.ChannelType {
	switch channelType {
	case domain.ChannelTypeEmail:
		return pb.ChannelType_CHANNEL_TYPE_EMAIL
	case domain.ChannelTypeWebhook:
		return pb.ChannelType_CHANNEL_TYPE_WEBHOOK
	case domain.ChannelTypeSMS:
		return pb.ChannelType_CHANNEL_TYPE_SMS
	default:
		return pb.ChannelType_CHANNEL_TYPE_UNSPECIFIED
	}
}

// handleError converts domain errors to gRPC status errors
func (h *Handler) handleError(err error) error {
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		return status.Error(codes.NotFound, "task not found")
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrInvalidExecutionMode):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrInvalidScheduleConfig):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "an internal error occurred")
	}
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timeToUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
