# Monitoring and Observability

This document describes the monitoring and observability features implemented in the Task Scheduler System.

## Overview

The system includes comprehensive monitoring and observability features:

1. **Prometheus Metrics** - Detailed metrics for all system operations
2. **Structured Logging** - JSON-formatted logs with distributed tracing support
3. **Health Checks** - HTTP endpoints for health monitoring

## Prometheus Metrics

### Metrics Endpoint

The system exposes Prometheus metrics at:
```
GET /metrics
```

### Available Metrics

#### Task Metrics

- `task_scheduler_tasks_created_total` - Total number of tasks created (by execution_mode)
- `task_scheduler_tasks_executed_total` - Total number of tasks executed (by status)
- `task_scheduler_tasks_completed_total` - Total number of tasks completed (by status)
- `task_scheduler_task_execution_duration_seconds` - Task execution duration histogram (by execution_mode, status)
- `task_scheduler_task_failures_total` - Total number of task failures (by execution_mode)
- `task_scheduler_task_retries_total` - Total number of task retries (by execution_mode)
- `task_scheduler_task_retry_rate` - Current task retry rate (by execution_mode)
- `task_scheduler_task_failure_rate` - Current task failure rate (by execution_mode)

#### Node Metrics

- `task_scheduler_node_health_status` - Node health status (1 = healthy, 0 = unhealthy) (by node_id)
- `task_scheduler_active_nodes_count` - Number of active scheduler nodes

#### Queue Metrics

- `task_scheduler_task_queue_size` - Current size of the task queue
- `task_scheduler_active_workers_count` - Number of active worker goroutines

#### Callback Metrics

- `task_scheduler_callback_executions_total` - Total number of callback executions (by protocol, status)
- `task_scheduler_callback_execution_duration_seconds` - Callback execution duration histogram (by protocol)

#### Lock Metrics

- `task_scheduler_lock_acquisitions_total` - Total number of distributed lock acquisitions (by status)
- `task_scheduler_lock_acquisition_duration_seconds` - Lock acquisition duration histogram

#### Storage Metrics

- `task_scheduler_storage_operations_total` - Total number of storage operations (by operation, status)
- `task_scheduler_storage_operation_duration_seconds` - Storage operation duration histogram (by operation)

#### Notification Metrics

- `task_scheduler_notifications_sent_total` - Total number of notifications sent (by channel, status)

#### Service Discovery Metrics

- `task_scheduler_service_discovery_lookups_total` - Total number of service discovery lookups (by type, status)

#### API Metrics

- `task_scheduler_http_requests_total` - Total number of HTTP requests (by method, endpoint, status)
- `task_scheduler_http_request_duration_seconds` - HTTP request duration histogram (by method, endpoint)
- `task_scheduler_grpc_requests_total` - Total number of gRPC requests (by method, status)
- `task_scheduler_grpc_request_duration_seconds` - gRPC request duration histogram (by method)

## Structured Logging

### Log Format

The system uses structured logging with JSON format in production:

```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "message": "Task execution completed",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "request_id": "660e8400-e29b-41d4-a716-446655440001",
  "task_id": "task-123",
  "status": "success",
  "duration": "2.5s"
}
```

### Distributed Tracing

The system supports distributed tracing with trace IDs:

- **Trace ID**: Unique identifier for a request flow across services
- **Request ID**: Unique identifier for a specific HTTP/gRPC request

#### HTTP Headers

Trace IDs can be provided via HTTP headers:
- `X-Trace-ID` - Trace ID for the request
- `X-Request-ID` - Request ID for the request

If not provided, the system automatically generates new IDs.

#### Context Propagation

Trace IDs are propagated through the system using Go's context:

```go
// Add trace ID to context
ctx = logger.WithTraceID(ctx, traceID)

// Use context-aware logging
logger.InfoContext(ctx, "Processing task", zap.String("task_id", taskID))
```

### Log Levels

Supported log levels:
- `debug` - Detailed debugging information
- `info` - General informational messages
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors that cause the application to exit

### Configuration

Configure logging in `config.yaml`:

```yaml
log:
  level: info          # Log level: debug, info, warn, error
  format: json         # Log format: json, console
  output_path: stdout  # Output path (not yet implemented)
```

## Health Checks

### Health Endpoint

The system provides a health check endpoint:

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:45Z"
}
```

### Application Health Check

The application also provides a comprehensive health check method that verifies:
- Redis connection
- Etcd connection (if used)
- Node registration status

## Integration with Monitoring Tools

### Prometheus

Add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'task-scheduler'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana

Import the provided Grafana dashboard (to be created) for visualizing:
- Task execution rates
- Task success/failure rates
- Task execution duration
- Queue sizes
- Node health status
- API request rates and latencies

### Log Aggregation

For log aggregation, use tools like:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Loki** (with Grafana)
- **Splunk**

The JSON log format makes it easy to parse and index logs.

## Best Practices

1. **Monitor Key Metrics**:
   - Task failure rate
   - Task execution duration
   - Queue size
   - Node health status

2. **Set Up Alerts**:
   - High task failure rate
   - Long task execution times
   - Queue size exceeding threshold
   - Node failures

3. **Use Trace IDs**:
   - Always include trace IDs in external requests
   - Use trace IDs for debugging and troubleshooting

4. **Log Levels**:
   - Use `debug` for development
   - Use `info` for production
   - Use `error` for errors that need attention

5. **Metrics Retention**:
   - Configure appropriate retention policies in Prometheus
   - Archive old metrics for long-term analysis

## Example Queries

### Prometheus Queries

Task success rate:
```promql
rate(task_scheduler_tasks_completed_total{status="success"}[5m]) / 
rate(task_scheduler_tasks_completed_total[5m])
```

Average task execution duration:
```promql
rate(task_scheduler_task_execution_duration_seconds_sum[5m]) / 
rate(task_scheduler_task_execution_duration_seconds_count[5m])
```

Queue size over time:
```promql
task_scheduler_task_queue_size
```

### Log Queries

Find all errors for a specific trace ID:
```json
{
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "level": "error"
}
```

Find slow tasks (execution > 10s):
```json
{
  "message": "Task execution completed",
  "duration": { "$gt": "10s" }
}
```
