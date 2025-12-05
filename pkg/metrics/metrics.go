package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Task creation metrics
	TasksCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_tasks_created_total",
			Help: "Total number of tasks created",
		},
		[]string{"execution_mode"},
	)

	// Task execution metrics
	TasksExecutedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_tasks_executed_total",
			Help: "Total number of tasks executed",
		},
		[]string{"status"},
	)

	// Task completion metrics
	TasksCompletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_tasks_completed_total",
			Help: "Total number of tasks completed",
		},
		[]string{"status"},
	)

	// Task execution duration histogram
	TaskExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_task_execution_duration_seconds",
			Help:    "Task execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~102s
		},
		[]string{"execution_mode", "status"},
	)

	// Task failure metrics
	TaskFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_task_failures_total",
			Help: "Total number of task failures",
		},
		[]string{"execution_mode"},
	)

	// Task retry metrics
	TaskRetriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_task_retries_total",
			Help: "Total number of task retries",
		},
		[]string{"execution_mode"},
	)

	// Task retry rate gauge
	TaskRetryRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "task_scheduler_task_retry_rate",
			Help: "Current task retry rate (retries per task)",
		},
		[]string{"execution_mode"},
	)

	// Task failure rate gauge
	TaskFailureRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "task_scheduler_task_failure_rate",
			Help: "Current task failure rate (failures per task)",
		},
		[]string{"execution_mode"},
	)

	// Node health status
	NodeHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "task_scheduler_node_health_status",
			Help: "Node health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"node_id"},
	)

	// Active nodes count
	ActiveNodesCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "task_scheduler_active_nodes_count",
			Help: "Number of active scheduler nodes",
		},
	)

	// Task queue size
	TaskQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "task_scheduler_task_queue_size",
			Help: "Current size of the task queue",
		},
	)

	// Active workers count
	ActiveWorkersCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "task_scheduler_active_workers_count",
			Help: "Number of active worker goroutines",
		},
	)

	// Callback execution metrics
	CallbackExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_callback_executions_total",
			Help: "Total number of callback executions",
		},
		[]string{"protocol", "status"},
	)

	// Callback execution duration
	CallbackExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_callback_execution_duration_seconds",
			Help:    "Callback execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		},
		[]string{"protocol"},
	)

	// Distributed lock metrics
	LockAcquisitionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_lock_acquisitions_total",
			Help: "Total number of distributed lock acquisitions",
		},
		[]string{"status"},
	)

	// Lock acquisition duration
	LockAcquisitionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_lock_acquisition_duration_seconds",
			Help:    "Lock acquisition duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
	)

	// Storage operation metrics
	StorageOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_storage_operations_total",
			Help: "Total number of storage operations",
		},
		[]string{"operation", "status"},
	)

	// Storage operation duration
	StorageOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_storage_operation_duration_seconds",
			Help:    "Storage operation duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"operation"},
	)

	// Notification metrics
	NotificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_notifications_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"channel", "status"},
	)

	// Service discovery metrics
	ServiceDiscoveryLookupsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_service_discovery_lookups_total",
			Help: "Total number of service discovery lookups",
		},
		[]string{"type", "status"},
	)

	// HTTP API metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// gRPC API metrics
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_scheduler_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	// gRPC request duration
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_scheduler_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

// RecordTaskCreated records a task creation event
func RecordTaskCreated(executionMode string) {
	TasksCreatedTotal.WithLabelValues(executionMode).Inc()
}

// RecordTaskExecuted records a task execution event
func RecordTaskExecuted(status string) {
	TasksExecutedTotal.WithLabelValues(status).Inc()
}

// RecordTaskCompleted records a task completion event
func RecordTaskCompleted(status string) {
	TasksCompletedTotal.WithLabelValues(status).Inc()
}

// RecordTaskExecutionDuration records task execution duration
func RecordTaskExecutionDuration(executionMode, status string, durationSeconds float64) {
	TaskExecutionDuration.WithLabelValues(executionMode, status).Observe(durationSeconds)
}

// RecordTaskFailure records a task failure
func RecordTaskFailure(executionMode string) {
	TaskFailuresTotal.WithLabelValues(executionMode).Inc()
}

// RecordTaskRetry records a task retry
func RecordTaskRetry(executionMode string) {
	TaskRetriesTotal.WithLabelValues(executionMode).Inc()
}

// SetTaskRetryRate sets the task retry rate
func SetTaskRetryRate(executionMode string, rate float64) {
	TaskRetryRate.WithLabelValues(executionMode).Set(rate)
}

// SetTaskFailureRate sets the task failure rate
func SetTaskFailureRate(executionMode string, rate float64) {
	TaskFailureRate.WithLabelValues(executionMode).Set(rate)
}

// SetNodeHealthStatus sets the node health status
func SetNodeHealthStatus(nodeID string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	NodeHealthStatus.WithLabelValues(nodeID).Set(status)
}

// SetActiveNodesCount sets the active nodes count
func SetActiveNodesCount(count int) {
	ActiveNodesCount.Set(float64(count))
}

// SetTaskQueueSize sets the task queue size
func SetTaskQueueSize(size int) {
	TaskQueueSize.Set(float64(size))
}

// SetActiveWorkersCount sets the active workers count
func SetActiveWorkersCount(count int) {
	ActiveWorkersCount.Set(float64(count))
}

// RecordCallbackExecution records a callback execution
func RecordCallbackExecution(protocol, status string) {
	CallbackExecutionsTotal.WithLabelValues(protocol, status).Inc()
}

// RecordCallbackExecutionDuration records callback execution duration
func RecordCallbackExecutionDuration(protocol string, durationSeconds float64) {
	CallbackExecutionDuration.WithLabelValues(protocol).Observe(durationSeconds)
}

// RecordLockAcquisition records a lock acquisition attempt
func RecordLockAcquisition(status string) {
	LockAcquisitionsTotal.WithLabelValues(status).Inc()
}

// RecordLockAcquisitionDuration records lock acquisition duration
func RecordLockAcquisitionDuration(durationSeconds float64) {
	LockAcquisitionDuration.Observe(durationSeconds)
}

// RecordStorageOperation records a storage operation
func RecordStorageOperation(operation, status string) {
	StorageOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordStorageOperationDuration records storage operation duration
func RecordStorageOperationDuration(operation string, durationSeconds float64) {
	StorageOperationDuration.WithLabelValues(operation).Observe(durationSeconds)
}

// RecordNotificationSent records a notification sent
func RecordNotificationSent(channel, status string) {
	NotificationsSentTotal.WithLabelValues(channel, status).Inc()
}

// RecordServiceDiscoveryLookup records a service discovery lookup
func RecordServiceDiscoveryLookup(discoveryType, status string) {
	ServiceDiscoveryLookupsTotal.WithLabelValues(discoveryType, status).Inc()
}

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, endpoint, status string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordHTTPRequestDuration records HTTP request duration
func RecordHTTPRequestDuration(method, endpoint string, durationSeconds float64) {
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(durationSeconds)
}

// RecordGRPCRequest records a gRPC request
func RecordGRPCRequest(method, status string) {
	GRPCRequestsTotal.WithLabelValues(method, status).Inc()
}

// RecordGRPCRequestDuration records gRPC request duration
func RecordGRPCRequestDuration(method string, durationSeconds float64) {
	GRPCRequestDuration.WithLabelValues(method).Observe(durationSeconds)
}
