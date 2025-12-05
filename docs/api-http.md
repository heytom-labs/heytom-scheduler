# HTTP API文档

本文档详细说明Task Scheduler System的HTTP API接口。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **认证**: 暂未实现（计划支持JWT）

## 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 1001,
  "message": "error description",
  "data": null
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 任务不存在 |
| 1003 | 数据库错误 |
| 1004 | 内部服务器错误 |
| 1005 | 任务已取消 |
| 1006 | 无效的Cron表达式 |

## API端点

### 1. 创建任务

创建一个新的任务。

**请求**

```
POST /tasks
```

**请求体**

```json
{
  "name": "示例任务",
  "description": "这是一个示例任务",
  "execution_mode": "immediate",
  "schedule_config": {
    "scheduled_time": "2024-01-01T10:00:00Z",
    "interval": "1h",
    "cron_expr": "0 0 * * *"
  },
  "callback_config": {
    "protocol": "http",
    "url": "http://example.com/callback",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer token"
    },
    "timeout": "30s",
    "is_async": false,
    "service_name": "my-service",
    "discovery_type": "consul"
  },
  "retry_policy": {
    "max_retries": 3,
    "retry_interval": "10s",
    "backoff_factor": 2.0
  },
  "concurrency_limit": 5,
  "alert_policy": {
    "enable_failure_alert": true,
    "retry_threshold": 3,
    "timeout_threshold": "5m",
    "channels": [
      {
        "type": "email",
        "config": {
          "to": "admin@example.com"
        }
      }
    ]
  },
  "sub_tasks": [
    {
      "name": "子任务1",
      "description": "第一个子任务",
      "execution_mode": "immediate",
      "callback_config": { ... }
    }
  ],
  "metadata": {
    "user_id": "123",
    "project": "demo"
  }
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 任务名称 |
| description | string | 否 | 任务描述 |
| execution_mode | string | 是 | 执行模式: immediate/scheduled/interval/cron |
| schedule_config | object | 否 | 调度配置（根据execution_mode决定） |
| callback_config | object | 否 | 回调配置 |
| retry_policy | object | 否 | 重试策略 |
| concurrency_limit | int | 否 | 子任务并发限制（0表示无限制） |
| alert_policy | object | 否 | 报警策略 |
| sub_tasks | array | 否 | 子任务列表 |
| metadata | object | 否 | 元数据 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "示例任务",
    "status": "pending",
    "created_at": "2024-01-01T09:00:00Z"
  }
}
```

**示例**

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试任务",
    "execution_mode": "immediate",
    "callback_config": {
      "protocol": "http",
      "url": "http://example.com/callback",
      "method": "POST",
      "timeout": "30s"
    }
  }'
```

### 2. 获取任务详情

获取指定任务的详细信息。

**请求**

```
GET /tasks/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 任务ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "示例任务",
    "description": "这是一个示例任务",
    "parent_id": null,
    "execution_mode": "immediate",
    "schedule_config": null,
    "callback_config": { ... },
    "retry_policy": { ... },
    "concurrency_limit": 0,
    "alert_policy": null,
    "status": "success",
    "retry_count": 0,
    "node_id": "node-1",
    "created_at": "2024-01-01T09:00:00Z",
    "updated_at": "2024-01-01T09:01:00Z",
    "started_at": "2024-01-01T09:00:05Z",
    "completed_at": "2024-01-01T09:01:00Z",
    "metadata": { ... },
    "sub_tasks": [
      {
        "id": "...",
        "name": "子任务1",
        "status": "success"
      }
    ]
  }
}
```

**示例**

```bash
curl http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000
```

### 3. 列出任务

获取任务列表，支持分页和过滤。

**请求**

```
GET /tasks
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码（从1开始），默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| status | string | 否 | 按状态过滤: pending/running/success/failed/cancelled |
| execution_mode | string | 否 | 按执行模式过滤 |
| parent_id | string | 否 | 按父任务ID过滤 |
| node_id | string | 否 | 按节点ID过滤 |
| created_after | string | 否 | 创建时间起始（ISO 8601格式） |
| created_before | string | 否 | 创建时间结束（ISO 8601格式） |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tasks": [
      {
        "id": "...",
        "name": "任务1",
        "status": "success",
        "created_at": "2024-01-01T09:00:00Z"
      },
      {
        "id": "...",
        "name": "任务2",
        "status": "running",
        "created_at": "2024-01-01T09:05:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

**示例**

```bash
# 获取所有任务
curl http://localhost:8080/tasks

# 获取第2页，每页10条
curl http://localhost:8080/tasks?page=2&page_size=10

# 获取运行中的任务
curl http://localhost:8080/tasks?status=running

# 获取今天创建的任务
curl "http://localhost:8080/tasks?created_after=2024-01-01T00:00:00Z&created_before=2024-01-02T00:00:00Z"
```

### 4. 取消任务

取消一个待执行或正在执行的任务。

**请求**

```
DELETE /tasks/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 任务ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "cancelled"
  }
}
```

**示例**

```bash
curl -X DELETE http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000
```

### 5. 获取任务状态

获取任务的当前状态。

**请求**

```
GET /tasks/:id/status
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 任务ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "running",
    "retry_count": 1,
    "node_id": "node-1",
    "started_at": "2024-01-01T09:00:05Z",
    "updated_at": "2024-01-01T09:00:30Z"
  }
}
```

**示例**

```bash
curl http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000/status
```

### 6. 获取任务执行历史

获取任务的执行历史记录。

**请求**

```
GET /tasks/:id/history
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 任务ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "history": [
      {
        "id": 1,
        "task_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "failed",
        "output": "",
        "error": "connection timeout",
        "start_time": "2024-01-01T09:00:05Z",
        "end_time": "2024-01-01T09:00:35Z",
        "created_at": "2024-01-01T09:00:35Z"
      },
      {
        "id": 2,
        "task_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "success",
        "output": "task completed",
        "error": null,
        "start_time": "2024-01-01T09:01:00Z",
        "end_time": "2024-01-01T09:01:30Z",
        "created_at": "2024-01-01T09:01:30Z"
      }
    ]
  }
}
```

**示例**

```bash
curl http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000/history
```

### 7. 健康检查

检查服务健康状态。

**请求**

```
GET /health
```

**响应**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T09:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "scheduler": "ok"
  }
}
```

**示例**

```bash
curl http://localhost:8080/health
```

### 8. 获取指标

获取Prometheus格式的指标数据。

**请求**

```
GET /metrics
```

**响应**

```
# HELP task_scheduler_tasks_created_total Total number of tasks created
# TYPE task_scheduler_tasks_created_total counter
task_scheduler_tasks_created_total 1234

# HELP task_scheduler_tasks_completed_total Total number of tasks completed
# TYPE task_scheduler_tasks_completed_total counter
task_scheduler_tasks_completed_total{status="success"} 1000
task_scheduler_tasks_completed_total{status="failed"} 50

# HELP task_scheduler_task_execution_duration_seconds Task execution duration
# TYPE task_scheduler_task_execution_duration_seconds histogram
task_scheduler_task_execution_duration_seconds_bucket{le="1"} 500
task_scheduler_task_execution_duration_seconds_bucket{le="5"} 800
task_scheduler_task_execution_duration_seconds_bucket{le="10"} 950
task_scheduler_task_execution_duration_seconds_bucket{le="+Inf"} 1000
task_scheduler_task_execution_duration_seconds_sum 3500
task_scheduler_task_execution_duration_seconds_count 1000
```

**示例**

```bash
curl http://localhost:8080/metrics
```

## WebSocket API

### 任务状态实时推送

订阅任务状态变化的实时推送。

**连接**

```
ws://localhost:8080/ws/tasks/:id
```

**消息格式**

```json
{
  "type": "status_update",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "running",
    "updated_at": "2024-01-01T09:00:30Z"
  }
}
```

**示例（JavaScript）**

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/tasks/550e8400-e29b-41d4-a716-446655440000');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Status update:', message.data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket connection closed');
};
```

## 错误处理

### 参数验证错误

```json
{
  "code": 1001,
  "message": "validation error: name is required",
  "data": {
    "field": "name",
    "error": "required"
  }
}
```

### 任务不存在

```json
{
  "code": 1002,
  "message": "task not found",
  "data": null
}
```

### 内部服务器错误

```json
{
  "code": 1004,
  "message": "internal server error",
  "data": null
}
```

## 速率限制

当前版本暂未实现速率限制。计划在未来版本中添加：

- 每个IP每分钟最多100个请求
- 超过限制返回429 Too Many Requests

## 认证和授权

当前版本暂未实现认证和授权。计划在未来版本中添加：

- JWT Token认证
- 基于角色的访问控制（RBAC）
- API Key认证

## 最佳实践

### 1. 使用幂等性

创建任务时使用客户端生成的唯一ID，避免重复创建：

```json
{
  "id": "client-generated-uuid",
  "name": "任务名称",
  ...
}
```

### 2. 合理设置超时

根据任务的实际执行时间设置合理的超时时间：

```json
{
  "callback_config": {
    "timeout": "30s"
  }
}
```

### 3. 使用重试策略

对于可能失败的任务，配置合理的重试策略：

```json
{
  "retry_policy": {
    "max_retries": 3,
    "retry_interval": "10s",
    "backoff_factor": 2.0
  }
}
```

### 4. 监控任务状态

使用WebSocket实时监控任务状态，而不是频繁轮询：

```javascript
const ws = new WebSocket(`ws://localhost:8080/ws/tasks/${taskId}`);
```

### 5. 分页查询

查询任务列表时使用分页，避免一次性加载大量数据：

```bash
curl "http://localhost:8080/tasks?page=1&page_size=20"
```

## 示例代码

### Python

```python
import requests
import json

# 创建任务
def create_task():
    url = "http://localhost:8080/tasks"
    payload = {
        "name": "Python测试任务",
        "execution_mode": "immediate",
        "callback_config": {
            "protocol": "http",
            "url": "http://example.com/callback",
            "method": "POST",
            "timeout": "30s"
        }
    }
    response = requests.post(url, json=payload)
    return response.json()

# 获取任务状态
def get_task_status(task_id):
    url = f"http://localhost:8080/tasks/{task_id}/status"
    response = requests.get(url)
    return response.json()

# 使用示例
result = create_task()
print(f"Task created: {result['data']['id']}")

status = get_task_status(result['data']['id'])
print(f"Task status: {status['data']['status']}")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

// 创建任务
async function createTask() {
  const response = await axios.post('http://localhost:8080/tasks', {
    name: 'JavaScript测试任务',
    execution_mode: 'immediate',
    callback_config: {
      protocol: 'http',
      url: 'http://example.com/callback',
      method: 'POST',
      timeout: '30s'
    }
  });
  return response.data;
}

// 获取任务状态
async function getTaskStatus(taskId) {
  const response = await axios.get(`http://localhost:8080/tasks/${taskId}/status`);
  return response.data;
}

// 使用示例
(async () => {
  const result = await createTask();
  console.log(`Task created: ${result.data.id}`);
  
  const status = await getTaskStatus(result.data.id);
  console.log(`Task status: ${status.data.status}`);
})();
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CreateTaskRequest struct {
    Name          string `json:"name"`
    ExecutionMode string `json:"execution_mode"`
}

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func createTask() (*Response, error) {
    url := "http://localhost:8080/tasks"
    payload := CreateTaskRequest{
        Name:          "Go测试任务",
        ExecutionMode: "immediate",
    }
    
    data, _ := json.Marshal(payload)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result Response
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}

func main() {
    result, err := createTask()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Task created: %+v\n", result.Data)
}
```

## 参考资源

- [RESTful API设计指南](https://restfulapi.net/)
- [HTTP状态码](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [WebSocket协议](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
