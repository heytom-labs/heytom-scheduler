# Task Scheduler System

一个用Go语言编写的分布式任务管理和执行平台，支持多种执行模式、存储后端和服务发现机制。

[English](README.en.md) | 简体中文

## 特性

- ✅ **多种执行模式**: 支持立即执行、定时执行、固定间隔和Cron表达式
- ✅ **分布式部署**: 支持多节点部署，自动任务分配和故障转移
- ✅ **多协议支持**: 提供HTTP和gRPC两种API协议
- ✅ **灵活存储**: 支持MySQL、PostgreSQL和MongoDB
- ✅ **服务发现**: 支持静态地址、Consul、Etcd和Kubernetes服务发现
- ✅ **重试机制**: 可配置的重试策略，支持指数退避
- ✅ **并发控制**: 子任务并发数量限制
- ✅ **报警通知**: 支持Email、Webhook和SMS多种通知渠道
- ✅ **监控指标**: 集成Prometheus指标导出
- ✅ **管理后台**: 基于Vue 3的Web管理界面

## 快速开始

### 使用Docker Compose（推荐）

最快的方式是使用Docker Compose启动完整的开发环境：

```bash
# 克隆仓库
git clone https://github.com/yourusername/task-scheduler.git
cd task-scheduler

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f task-scheduler-1

# 访问服务
# - Web 管理界面: http://localhost:8080/ (已集成到后端服务)
# - HTTP API: http://localhost:8080/api
# - gRPC API: localhost:9090
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3001 (admin/admin)
# - Consul UI: http://localhost:8500
```

> **注意**: Web 管理界面现已集成到后端服务中，共用同一个 HTTP 端口（8080）。

### 本地开发

#### 前置要求

- Go 1.25+
- Node.js 20+
- MySQL 8.0+ / PostgreSQL 14+ / MongoDB 6.0+
- Redis 7.0+
- (可选) Consul / Etcd

#### 后端服务

```bash
# 安装依赖
go mod download

# 复制配置文件
cp config.yaml.example config.yaml

# 编辑配置文件，设置数据库连接等
vim config.yaml

# 运行服务
go run cmd/scheduler/main.go -config config.yaml

# 或者构建后运行
go build -o task-scheduler cmd/scheduler/main.go
./task-scheduler -config config.yaml
```

#### 前端管理后台

```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build
```

## 项目结构

```
.
├── cmd/                          # 应用程序入口
│   └── scheduler/               # 调度器主服务
├── internal/                     # 私有应用代码
│   ├── api/                     # API层
│   │   ├── http/               # HTTP处理器
│   │   └── grpc/               # gRPC处理器
│   ├── app/                     # 应用程序初始化
│   ├── config/                  # 配置管理
│   ├── discovery/               # 服务发现实现
│   ├── domain/                  # 领域模型和接口
│   ├── lock/                    # 分布式锁实现
│   ├── node/                    # 节点管理
│   ├── notification/            # 通知系统
│   ├── scheduler/               # 调度器核心
│   ├── service/                 # 业务服务层
│   └── storage/                 # 存储层实现
├── pkg/                         # 公共库
│   ├── logger/                 # 日志工具
│   └── metrics/                # 指标收集
├── api/                         # API定义
│   ├── task_scheduler.proto    # gRPC协议定义
│   └── pb/                     # 生成的protobuf代码
├── web/                         # Vue管理后台
│   ├── src/                    # 源代码
│   └── dist/                   # 构建输出
├── k8s/                         # Kubernetes部署配置
├── docs/                        # 文档
├── scripts/                     # 脚本文件
├── docker-compose.yml           # Docker Compose配置
├── Dockerfile                   # 后端Docker镜像
└── config.yaml.example          # 配置文件示例
```

## 配置说明

详细的配置说明请参考 [配置文档](docs/configuration.md)。

主要配置项：

```yaml
server:
  http_port: 8080              # HTTP服务端口
  grpc_port: 9090              # gRPC服务端口
  metrics_port: 9091           # 指标端口

scheduler:
  node_id: node-1              # 节点ID
  worker_pool_size: 10         # 工作线程池大小

database:
  type: mysql                  # 数据库类型: mysql/postgres/mongodb
  host: localhost
  port: 3306
  name: task_scheduler
  user: scheduler
  password: password

lock:
  type: redis                  # 锁类型: redis/etcd
  redis:
    host: localhost
    port: 6379

discovery:
  type: consul                 # 服务发现: static/consul/etcd/kubernetes
```

## API文档

### HTTP API

详细的HTTP API文档请参考 [HTTP API文档](docs/api-http.md)。

主要端点：

- `POST /tasks` - 创建任务
- `GET /tasks/:id` - 获取任务详情
- `GET /tasks` - 列出任务
- `DELETE /tasks/:id` - 取消任务
- `GET /tasks/:id/status` - 获取任务状态
- `GET /health` - 健康检查
- `GET /metrics` - Prometheus指标

### gRPC API

详细的gRPC API文档请参考 [gRPC API文档](docs/api-grpc.md)。

Protocol Buffers定义在 `api/task_scheduler.proto`。

## 部署

### Docker部署

```bash
# 构建镜像
docker build -t task-scheduler:latest .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 9091:9091 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  task-scheduler:latest
```

### Kubernetes部署

详细的Kubernetes部署指南请参考 [Kubernetes部署文档](docs/deployment-kubernetes.md)。

快速部署：

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 部署配置和密钥
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmap.yaml

# 部署RBAC
kubectl apply -f k8s/rbac.yaml

# 部署数据库和Redis
kubectl apply -f k8s/mysql-deployment.yaml
kubectl apply -f k8s/redis-deployment.yaml

# 部署应用
kubectl apply -f k8s/task-scheduler-deployment.yaml
kubectl apply -f k8s/admin-ui-deployment.yaml

# 部署Ingress
kubectl apply -f k8s/ingress.yaml

# 部署自动扩缩容
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/pdb.yaml
```

## 监控和可观测性

系统集成了Prometheus指标导出，可以使用Grafana进行可视化。

主要指标：

- `task_scheduler_tasks_created_total` - 创建的任务总数
- `task_scheduler_tasks_completed_total` - 完成的任务总数
- `task_scheduler_tasks_failed_total` - 失败的任务总数
- `task_scheduler_task_execution_duration_seconds` - 任务执行时长
- `task_scheduler_active_tasks` - 活跃任务数
- `task_scheduler_node_health` - 节点健康状态

详细的监控配置请参考 [监控文档](docs/monitoring-observability.md)。

## 开发指南

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/scheduler/...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 代码规范

项目遵循标准的Go代码规范：

```bash
# 格式化代码
go fmt ./...

# 运行linter
golangci-lint run

# 运行静态分析
go vet ./...
```

### 生成gRPC代码

```bash
# 安装protoc和Go插件
# 参考: https://grpc.io/docs/languages/go/quickstart/

# 生成代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/task_scheduler.proto
```

## 架构设计

系统采用分层架构设计：

```
┌─────────────────────────────────────┐
│         客户端层                      │
│  HTTP Client | gRPC Client | Web UI │
└─────────────────────────────────────┘
                 │
┌─────────────────────────────────────┐
│         协议层                        │
│  HTTP Handler | gRPC Handler        │
└─────────────────────────────────────┘
                 │
┌─────────────────────────────────────┐
│         服务层                        │
│  Task Service | Schedule Service    │
│  Callback Service | Retry Service   │
└─────────────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐
│存储层   │  │调度层   │  │通知层   │
└────────┘  └────────┘  └────────┘
```

详细的架构文档请参考 [架构文档](docs/architecture.md)。

## 文档

### 核心文档

- [快速开始](QUICKSTART.md) - 5分钟快速上手指南
- [配置说明](docs/configuration.md) - 详细的配置选项说明
- [架构设计](docs/architecture.md) - 系统架构和设计文档

### API文档

- [HTTP API](docs/api-http.md) - RESTful API接口文档
- [gRPC API](docs/api-grpc.md) - gRPC接口文档

### 部署文档

- [Docker部署](docs/deployment-docker.md) - Docker和Docker Compose部署指南
- [Kubernetes部署](docs/deployment-kubernetes.md) - Kubernetes部署指南
- [Docker快速参考](DOCKER.md) - Docker配置快速参考

### 运维文档

- [运维指南](docs/operations.md) - 日常运维操作指南
- [故障排查](docs/troubleshooting.md) - 常见问题排查和解决方案
- [监控和可观测性](docs/monitoring-observability.md) - 监控配置和最佳实践

### 高级主题

- [分布式锁实现](docs/distributed-lock-implementation.md) - 分布式锁的实现细节
- [多节点任务分配](docs/multi-node-task-allocation.md) - 多节点部署和任务分配

### 开发文档

- [贡献指南](CONTRIBUTING.md) - 如何为项目做贡献

## 常见问题

### 如何配置多节点部署？

参考 [多节点部署文档](docs/multi-node-task-allocation.md)。

### 如何实现分布式锁？

参考 [分布式锁实现文档](docs/distributed-lock-implementation.md)。

### 如何配置报警通知？

在任务创建时配置 `alert_policy`：

```json
{
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
  }
}
```

### 遇到问题怎么办？

请查看 [故障排查指南](docs/troubleshooting.md)。

## 贡献

欢迎贡献！请阅读 [贡献指南](CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 联系方式

- 问题反馈: [GitHub Issues](https://github.com/yourusername/task-scheduler/issues)
- 邮件: support@example.com

## 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - HTTP框架
- [gRPC](https://grpc.io/) - RPC框架
- [GORM](https://gorm.io/) - ORM库
- [robfig/cron](https://github.com/robfig/cron) - Cron解析
- [go-redis](https://github.com/redis/go-redis) - Redis客户端
- [Consul](https://www.consul.io/) - 服务发现
- [Prometheus](https://prometheus.io/) - 监控系统
- [Vue.js](https://vuejs.org/) - 前端框架
- [Element Plus](https://element-plus.org/) - UI组件库
