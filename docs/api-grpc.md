# gRPC API文档

本文档详细说明Task Scheduler System的gRPC API接口。

## 基础信息

- **服务地址**: `localhost:9090`
- **Protocol Buffers**: `api/task_scheduler.proto`
- **认证**: 暂未实现（计划支持mTLS和Token认证）

## Protocol Buffers定义

完整的Protocol Buffers定义位于 `api/task_scheduler.proto`。

### 服务定义

```protobuf
service TaskSchedulerService {
    rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
    rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
    rpc ListTasks(ListTasksRequest) returns (ListTasksResponse);
    rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
    rpc GetTaskStatus(GetTaskStatusRequest) returns (GetTaskStatusResponse);
}
```

## API方法

### 1. CreateTask

创建一个新的任务。

**请求消息**

```protobuf
message CreateTaskRequest {
    string name = 1;
    string description = 2;
    string parent_id = 3;
    ExecutionMode execution_mode = 4;
    ScheduleConfig schedule_config = 5;
    CallbackConfig callback_config = 6;
    RetryPolicy retry_policy = 7;
    int32 concurrency_limit = 8;
    AlertPolicy alert_policy = 9;
    repeated CreateTaskRequest sub_tasks = 10;
    map<string, string> metadata = 11;
}
```

**响应消息**

```protobuf
message CreateTaskResponse {
    int32 code = 1;
    string message = 2;
    Task task = 3;
}
```

**示例（Go）**

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "task-scheduler/api/pb"
)

func main() {
    // 建立连接
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // 创建客户端
    client := pb.NewTaskSchedulerServiceClient(conn)

    // 创建任务
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req := &pb.CreateTaskRequest{
        Name:          "gRPC测试任务",
        Description:   "通过gRPC创建的任务",
        ExecutionMode: pb.ExecutionMode_IMMEDIATE,
        CallbackConfig: &pb.CallbackConfig{
            Protocol: pb.CallbackProtocol_HTTP,
            Url:      "http://example.com/callback",
            Method:   "POST",
            Timeout:  "30s",
            IsAsync:  false,
        },
    }

    resp, err := client.CreateTask(ctx, req)
    if err != nil {
        log.Fatalf("CreateTask failed: %v", err)
    }

    log.Printf("Task created: %s", resp.Task.Id)
}
```

**示例（Python）**

```python
import grpc
from api.pb import task_scheduler_pb2
from api.pb import task_scheduler_pb2_grpc

def create_task():
    # 建立连接
    channel = grpc.insecure_channel('localhost:9090')
    stub = task_scheduler_pb2_grpc.TaskSchedulerServiceStub(channel)
    
    # 创建请求
    request = task_scheduler_pb2.CreateTaskRequest(
        name='Python gRPC测试任务',
        description='通过Python gRPC创建的任务',
        execution_mode=task_scheduler_pb2.IMMEDIATE,
        callback_config=task_scheduler_pb2.CallbackConfig(
            protocol=task_scheduler_pb2.HTTP,
            url='http://example.com/callback',
            method='POST',
            timeout='30s',
            is_async=False
        )
    )
    
    # 调用服务
    response = stub.CreateTask(request)
    print(f'Task created: {response.task.id}')

if __name__ == '__main__':
    create_task()
```

### 2. GetTask

获取指定任务的详细信息。

**请求消息**

```protobuf
message GetTaskRequest {
    string id = 1;
}
```

**响应消息**

```protobuf
message GetTaskResponse {
    int32 code = 1;
    string message = 2;
    Task task = 3;
}
```

**示例（Go）**

```go
req := &pb.GetTaskRequest{
    Id: "550e8400-e29b-41d4-a716-446655440000",
}

resp, err := client.GetTask(ctx, req)
if err != nil {
    log.Fatalf("GetTask failed: %v", err)
}

log.Printf("Task: %+v", resp.Task)
```

### 3. ListTasks

获取任务列表，支持分页和过滤。

**请求消息**

```protobuf
message ListTasksRequest {
    int32 page = 1;
    int32 page_size = 2;
    string status = 3;
    string execution_mode = 4;
    string parent_id = 5;
    string node_id = 6;
    string created_after = 7;
    string created_before = 8;
}
```

**响应消息**

```protobuf
message ListTasksResponse {
    int32 code = 1;
    string message = 2;
    repeated Task tasks = 3;
    int32 total = 4;
    int32 page = 5;
    int32 page_size = 6;
}
```

**示例（Go）**

```go
req := &pb.ListTasksRequest{
    Page:     1,
    PageSize: 20,
    Status:   "running",
}

resp, err := client.ListTasks(ctx, req)
if err != nil {
    log.Fatalf("ListTasks failed: %v", err)
}

log.Printf("Found %d tasks", resp.Total)
for _, task := range resp.Tasks {
    log.Printf("Task: %s - %s", task.Id, task.Name)
}
```

### 4. CancelTask

取消一个待执行或正在执行的任务。

**请求消息**

```protobuf
message CancelTaskRequest {
    string id = 1;
}
```

**响应消息**

```protobuf
message CancelTaskResponse {
    int32 code = 1;
    string message = 2;
    Task task = 3;
}
```

**示例（Go）**

```go
req := &pb.CancelTaskRequest{
    Id: "550e8400-e29b-41d4-a716-446655440000",
}

resp, err := client.CancelTask(ctx, req)
if err != nil {
    log.Fatalf("CancelTask failed: %v", err)
}

log.Printf("Task cancelled: %s", resp.Task.Id)
```

### 5. GetTaskStatus

获取任务的当前状态。

**请求消息**

```protobuf
message GetTaskStatusRequest {
    string id = 1;
}
```

**响应消息**

```protobuf
message GetTaskStatusResponse {
    int32 code = 1;
    string message = 2;
    TaskStatusInfo status_info = 3;
}
```

**示例（Go）**

```go
req := &pb.GetTaskStatusRequest{
    Id: "550e8400-e29b-41d4-a716-446655440000",
}

resp, err := client.GetTaskStatus(ctx, req)
if err != nil {
    log.Fatalf("GetTaskStatus failed: %v", err)
}

log.Printf("Task status: %s", resp.StatusInfo.Status)
```

## 数据类型

### ExecutionMode

```protobuf
enum ExecutionMode {
    IMMEDIATE = 0;  // 立即执行
    SCHEDULED = 1;  // 定时执行
    INTERVAL = 2;   // 固定间隔
    CRON = 3;       // Cron表达式
}
```

### TaskStatus

```protobuf
enum TaskStatus {
    PENDING = 0;    // 等待中
    RUNNING = 1;    // 执行中
    SUCCESS = 2;    // 成功
    FAILED = 3;     // 失败
    CANCELLED = 4;  // 已取消
}
```

### CallbackProtocol

```protobuf
enum CallbackProtocol {
    HTTP = 0;  // HTTP协议
    GRPC = 1;  // gRPC协议
}
```

### ServiceDiscoveryType

```protobuf
enum ServiceDiscoveryType {
    STATIC = 0;      // 静态地址
    CONSUL = 1;      // Consul
    ETCD = 2;        // Etcd
    KUBERNETES = 3;  // Kubernetes
}
```

### ChannelType

```protobuf
enum ChannelType {
    EMAIL = 0;    // 邮件
    WEBHOOK = 1;  // Webhook
    SMS = 2;      // 短信
}
```

## 错误处理

### 错误码

gRPC使用标准的状态码：

| 状态码 | 说明 |
|--------|------|
| OK (0) | 成功 |
| INVALID_ARGUMENT (3) | 参数错误 |
| NOT_FOUND (5) | 任务不存在 |
| ALREADY_EXISTS (6) | 任务已存在 |
| INTERNAL (13) | 内部服务器错误 |
| UNAVAILABLE (14) | 服务不可用 |

### 错误处理示例

```go
resp, err := client.CreateTask(ctx, req)
if err != nil {
    if st, ok := status.FromError(err); ok {
        switch st.Code() {
        case codes.InvalidArgument:
            log.Printf("Invalid argument: %s", st.Message())
        case codes.Internal:
            log.Printf("Internal error: %s", st.Message())
        default:
            log.Printf("Error: %s", st.Message())
        }
    }
    return
}
```

## 元数据（Metadata）

### 请求元数据

可以通过gRPC元数据传递额外信息：

```go
import "google.golang.org/grpc/metadata"

// 添加元数据
md := metadata.Pairs(
    "trace-id", "550e8400-e29b-41d4-a716-446655440000",
    "request-id", "660e8400-e29b-41d4-a716-446655440001",
)
ctx := metadata.NewOutgoingContext(context.Background(), md)

// 发送请求
resp, err := client.CreateTask(ctx, req)
```

### 响应元数据

```go
var header, trailer metadata.MD

resp, err := client.CreateTask(ctx, req,
    grpc.Header(&header),
    grpc.Trailer(&trailer),
)

// 读取响应头
if values := header.Get("server-version"); len(values) > 0 {
    log.Printf("Server version: %s", values[0])
}
```

## 流式RPC

当前版本暂不支持流式RPC。计划在未来版本中添加：

- **Server Streaming**: 用于实时推送任务状态更新
- **Client Streaming**: 用于批量创建任务
- **Bidirectional Streaming**: 用于实时任务管理

## 拦截器（Interceptors）

### 客户端拦截器示例

```go
// 日志拦截器
func loggingInterceptor(
    ctx context.Context,
    method string,
    req, reply interface{},
    cc *grpc.ClientConn,
    invoker grpc.UnaryInvoker,
    opts ...grpc.CallOption,
) error {
    start := time.Now()
    err := invoker(ctx, method, req, reply, cc, opts...)
    log.Printf("Method: %s, Duration: %v, Error: %v", method, time.Since(start), err)
    return err
}

// 使用拦截器
conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(loggingInterceptor),
)
```

## 性能优化

### 连接池

```go
// 创建连接池
var conns []*grpc.ClientConn
for i := 0; i < 10; i++ {
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    conns = append(conns, conn)
}

// 使用连接池
func getClient() pb.TaskSchedulerServiceClient {
    conn := conns[rand.Intn(len(conns))]
    return pb.NewTaskSchedulerServiceClient(conn)
}
```

### Keep-Alive

```go
import "google.golang.org/grpc/keepalive"

conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
)
```

### 压缩

```go
import "google.golang.org/grpc/encoding/gzip"

resp, err := client.CreateTask(
    ctx,
    req,
    grpc.UseCompressor(gzip.Name),
)
```

## 生成客户端代码

### Go

```bash
# 安装protoc和Go插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/task_scheduler.proto
```

### Python

```bash
# 安装grpcio-tools
pip install grpcio-tools

# 生成代码
python -m grpc_tools.protoc \
    -I. \
    --python_out=. \
    --grpc_python_out=. \
    api/task_scheduler.proto
```

### Java

```bash
# 使用Maven插件
# 在pom.xml中添加protobuf-maven-plugin配置

mvn clean compile
```

### Node.js

```bash
# 安装grpc-tools
npm install -g grpc-tools

# 生成代码
grpc_tools_node_protoc \
    --js_out=import_style=commonjs,binary:. \
    --grpc_out=grpc_js:. \
    api/task_scheduler.proto
```

## 测试工具

### grpcurl

使用grpcurl测试gRPC API：

```bash
# 安装grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出服务
grpcurl -plaintext localhost:9090 list

# 列出方法
grpcurl -plaintext localhost:9090 list task_scheduler.TaskSchedulerService

# 调用方法
grpcurl -plaintext -d '{
  "name": "测试任务",
  "execution_mode": "IMMEDIATE"
}' localhost:9090 task_scheduler.TaskSchedulerService/CreateTask
```

### BloomRPC

BloomRPC是一个图形化的gRPC客户端工具：

1. 下载并安装BloomRPC
2. 导入 `api/task_scheduler.proto`
3. 设置服务器地址为 `localhost:9090`
4. 选择方法并填写请求参数
5. 点击发送

## 最佳实践

### 1. 使用超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.CreateTask(ctx, req)
```

### 2. 错误处理

```go
import "google.golang.org/grpc/status"

resp, err := client.CreateTask(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        log.Printf("Error code: %s, Message: %s", st.Code(), st.Message())
    }
    return
}
```

### 3. 重试机制

```go
import "google.golang.org/grpc/codes"

func createTaskWithRetry(client pb.TaskSchedulerServiceClient, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        resp, err := client.CreateTask(context.Background(), req)
        if err == nil {
            return resp, nil
        }
        
        if st, ok := status.FromError(err); ok {
            if st.Code() == codes.Unavailable {
                time.Sleep(time.Second * time.Duration(i+1))
                continue
            }
        }
        return nil, err
    }
    return nil, fmt.Errorf("max retries exceeded")
}
```

### 4. 连接管理

```go
// 创建连接
conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 检查连接状态
state := conn.GetState()
log.Printf("Connection state: %s", state)
```

## 安全性

### TLS/SSL

```go
import "google.golang.org/grpc/credentials"

// 加载TLS证书
creds, err := credentials.NewClientTLSFromFile("server.crt", "")
if err != nil {
    log.Fatal(err)
}

// 使用TLS连接
conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithTransportCredentials(creds),
)
```

### Token认证

```go
type tokenAuth struct {
    token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
    return map[string]string{
        "authorization": "Bearer " + t.token,
    }, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
    return true
}

// 使用Token认证
conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithTransportCredentials(creds),
    grpc.WithPerRPCCredentials(tokenAuth{token: "your-token"}),
)
```

## 参考资源

- [gRPC官方文档](https://grpc.io/docs/)
- [Protocol Buffers文档](https://developers.google.com/protocol-buffers)
- [gRPC Go快速开始](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC Python快速开始](https://grpc.io/docs/languages/python/quickstart/)

