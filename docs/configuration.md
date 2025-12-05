# 配置说明

本文档详细说明Task Scheduler System的所有配置选项。

## 配置文件格式

系统使用YAML格式的配置文件。默认配置文件路径为 `config.yaml`，可以通过命令行参数 `-config` 指定。

```bash
./task-scheduler -config /path/to/config.yaml
```

## 完整配置示例

```yaml
# 服务器配置
server:
  http_port: 8080                    # HTTP服务端口
  grpc_port: 9090                    # gRPC服务端口
  metrics_port: 9091                 # Prometheus指标端口
  shutdown_timeout: 30s              # 优雅关闭超时时间

# 调度器配置
scheduler:
  node_id: node-1                    # 节点唯一标识符
  worker_pool_size: 10               # 工作线程池大小
  task_poll_interval: 5s             # 任务轮询间隔
  heartbeat_interval: 10s            # 心跳间隔
  heartbeat_timeout: 30s             # 心跳超时时间

# 数据库配置
database:
  type: mysql                        # 数据库类型: mysql/postgres/mongodb
  host: localhost                    # 数据库主机
  port: 3306                         # 数据库端口
  name: task_scheduler               # 数据库名称
  user: scheduler                    # 数据库用户
  password: password                 # 数据库密码
  max_open_conns: 25                 # 最大打开连接数
  max_idle_conns: 5                  # 最大空闲连接数
  conn_max_lifetime: 5m              # 连接最大生命周期

# 分布式锁配置
lock:
  type: redis                        # 锁类型: redis/etcd
  redis:
    host: localhost                  # Redis主机
    port: 6379                       # Redis端口
    password: ""                     # Redis密码
    db: 0                            # Redis数据库编号
    ttl: 30s                         # 锁TTL
    retry_interval: 1s               # 重试间隔
    max_retries: 3                   # 最大重试次数
  etcd:
    endpoints:                       # Etcd端点列表
      - localhost:2379
    dial_timeout: 5s                 # 连接超时
    username: ""                     # 用户名
    password: ""                     # 密码
    ttl: 30s                         # 锁TTL

# 服务发现配置
discovery:
  type: consul                       # 服务发现类型: static/consul/etcd/kubernetes
  static:
    addresses:                       # 静态地址列表
      - localhost:8080
      - localhost:8081
  consul:
    address: localhost:8500          # Consul地址
    scheme: http                     # 协议: http/https
    datacenter: dc1                  # 数据中心
    token: ""                        # ACL Token
  etcd:
    endpoints:                       # Etcd端点列表
      - localhost:2379
    dial_timeout: 5s                 # 连接超时
    username: ""                     # 用户名
    password: ""                     # 密码
  kubernetes:
    namespace: default               # Kubernetes命名空间
    label_selector: app=task-scheduler  # 标签选择器

# 日志配置
log:
  level: info                        # 日志级别: debug/info/warn/error
  format: json                       # 日志格式: json/console
  output: stdout                     # 输出: stdout/stderr/file
  file_path: /var/log/task-scheduler.log  # 文件路径（当output=file时）

# 指标配置
metrics:
  enabled: true                      # 是否启用指标
  path: /metrics                     # 指标路径
```

## 配置项详解

### Server配置

#### http_port

- **类型**: 整数
- **默认值**: 8080
- **说明**: HTTP API服务监听端口

#### grpc_port

- **类型**: 整数
- **默认值**: 9090
- **说明**: gRPC API服务监听端口

#### metrics_port

- **类型**: 整数
- **默认值**: 9091
- **说明**: Prometheus指标导出端口

#### shutdown_timeout

- **类型**: 时间间隔
- **默认值**: 30s
- **说明**: 优雅关闭时等待现有请求完成的超时时间

### Scheduler配置

#### node_id

- **类型**: 字符串
- **必填**: 是
- **说明**: 节点的唯一标识符，在多节点部署时必须唯一
- **示例**: `node-1`, `scheduler-prod-01`

#### worker_pool_size

- **类型**: 整数
- **默认值**: 10
- **说明**: 工作线程池大小，决定可以并发执行的任务数量
- **建议**: 根据CPU核心数和任务类型调整，通常设置为CPU核心数的2-4倍

#### task_poll_interval

- **类型**: 时间间隔
- **默认值**: 5s
- **说明**: 从数据库轮询待执行任务的间隔
- **建议**: 根据任务创建频率调整，频繁创建任务时可以缩短间隔

#### heartbeat_interval

- **类型**: 时间间隔
- **默认值**: 10s
- **说明**: 节点向注册中心发送心跳的间隔

#### heartbeat_timeout

- **类型**: 时间间隔
- **默认值**: 30s
- **说明**: 节点心跳超时时间，超过此时间未收到心跳则认为节点故障

### Database配置

#### type

- **类型**: 字符串
- **可选值**: `mysql`, `postgres`, `mongodb`
- **必填**: 是
- **说明**: 数据库类型

#### host

- **类型**: 字符串
- **必填**: 是
- **说明**: 数据库主机地址

#### port

- **类型**: 整数
- **必填**: 是
- **说明**: 数据库端口
- **默认端口**: MySQL: 3306, PostgreSQL: 5432, MongoDB: 27017

#### name

- **类型**: 字符串
- **必填**: 是
- **说明**: 数据库名称

#### user

- **类型**: 字符串
- **必填**: 是
- **说明**: 数据库用户名

#### password

- **类型**: 字符串
- **必填**: 是
- **说明**: 数据库密码
- **安全建议**: 使用环境变量或密钥管理系统

#### max_open_conns

- **类型**: 整数
- **默认值**: 25
- **说明**: 最大打开连接数
- **建议**: 根据数据库服务器配置和并发需求调整

#### max_idle_conns

- **类型**: 整数
- **默认值**: 5
- **说明**: 最大空闲连接数
- **建议**: 通常设置为max_open_conns的20-30%

#### conn_max_lifetime

- **类型**: 时间间隔
- **默认值**: 5m
- **说明**: 连接的最大生命周期
- **建议**: 设置为数据库服务器的连接超时时间的一半

### Lock配置

#### type

- **类型**: 字符串
- **可选值**: `redis`, `etcd`
- **必填**: 是
- **说明**: 分布式锁实现类型

#### Redis配置

##### host

- **类型**: 字符串
- **必填**: 是（当type=redis时）
- **说明**: Redis主机地址

##### port

- **类型**: 整数
- **默认值**: 6379
- **说明**: Redis端口

##### password

- **类型**: 字符串
- **说明**: Redis密码，如果Redis未设置密码则留空

##### db

- **类型**: 整数
- **默认值**: 0
- **说明**: Redis数据库编号（0-15）

##### ttl

- **类型**: 时间间隔
- **默认值**: 30s
- **说明**: 锁的生存时间
- **建议**: 设置为任务平均执行时间的2-3倍

##### retry_interval

- **类型**: 时间间隔
- **默认值**: 1s
- **说明**: 获取锁失败后的重试间隔

##### max_retries

- **类型**: 整数
- **默认值**: 3
- **说明**: 获取锁的最大重试次数

#### Etcd配置

##### endpoints

- **类型**: 字符串数组
- **必填**: 是（当type=etcd时）
- **说明**: Etcd端点列表
- **示例**: `["localhost:2379", "localhost:2380"]`

##### dial_timeout

- **类型**: 时间间隔
- **默认值**: 5s
- **说明**: 连接超时时间

##### username

- **类型**: 字符串
- **说明**: Etcd用户名（如果启用了认证）

##### password

- **类型**: 字符串
- **说明**: Etcd密码（如果启用了认证）

##### ttl

- **类型**: 时间间隔
- **默认值**: 30s
- **说明**: 锁的生存时间

### Discovery配置

#### type

- **类型**: 字符串
- **可选值**: `static`, `consul`, `etcd`, `kubernetes`
- **必填**: 是
- **说明**: 服务发现类型

#### Static配置

##### addresses

- **类型**: 字符串数组
- **必填**: 是（当type=static时）
- **说明**: 静态服务地址列表
- **示例**: `["localhost:8080", "192.168.1.10:8080"]`

#### Consul配置

##### address

- **类型**: 字符串
- **必填**: 是（当type=consul时）
- **说明**: Consul服务器地址
- **示例**: `localhost:8500`

##### scheme

- **类型**: 字符串
- **可选值**: `http`, `https`
- **默认值**: `http`
- **说明**: 连接协议

##### datacenter

- **类型**: 字符串
- **默认值**: `dc1`
- **说明**: Consul数据中心名称

##### token

- **类型**: 字符串
- **说明**: Consul ACL Token（如果启用了ACL）

#### Kubernetes配置

##### namespace

- **类型**: 字符串
- **默认值**: `default`
- **说明**: Kubernetes命名空间

##### label_selector

- **类型**: 字符串
- **必填**: 是（当type=kubernetes时）
- **说明**: Pod标签选择器
- **示例**: `app=task-scheduler`

### Log配置

#### level

- **类型**: 字符串
- **可选值**: `debug`, `info`, `warn`, `error`
- **默认值**: `info`
- **说明**: 日志级别

#### format

- **类型**: 字符串
- **可选值**: `json`, `console`
- **默认值**: `json`
- **说明**: 日志格式
- **建议**: 生产环境使用json格式，开发环境使用console格式

#### output

- **类型**: 字符串
- **可选值**: `stdout`, `stderr`, `file`
- **默认值**: `stdout`
- **说明**: 日志输出目标

#### file_path

- **类型**: 字符串
- **说明**: 日志文件路径（当output=file时必填）
- **示例**: `/var/log/task-scheduler.log`

### Metrics配置

#### enabled

- **类型**: 布尔值
- **默认值**: `true`
- **说明**: 是否启用Prometheus指标导出

#### path

- **类型**: 字符串
- **默认值**: `/metrics`
- **说明**: 指标导出路径

## 环境变量

配置文件中可以使用环境变量，格式为 `${VAR_NAME}`：

```yaml
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  password: ${DB_PASSWORD}
```

设置环境变量：

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_PASSWORD=secret
```

## 配置验证

启动时系统会自动验证配置的有效性。如果配置无效，会输出详细的错误信息并退出。

常见配置错误：

1. **必填字段缺失**: 确保所有必填字段都已配置
2. **端口冲突**: 确保http_port、grpc_port和metrics_port不冲突
3. **数据库连接失败**: 检查数据库地址、端口、用户名和密码
4. **Redis/Etcd连接失败**: 检查连接信息和网络连通性

## 配置最佳实践

### 开发环境

```yaml
log:
  level: debug
  format: console
  output: stdout

scheduler:
  worker_pool_size: 5
  task_poll_interval: 10s

database:
  max_open_conns: 10
  max_idle_conns: 2
```

### 生产环境

```yaml
log:
  level: info
  format: json
  output: stdout

scheduler:
  worker_pool_size: 20
  task_poll_interval: 5s

database:
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 5m

# 使用环境变量存储敏感信息
database:
  password: ${DB_PASSWORD}

lock:
  redis:
    password: ${REDIS_PASSWORD}
```

### 高负载环境

```yaml
scheduler:
  worker_pool_size: 50
  task_poll_interval: 2s

database:
  max_open_conns: 100
  max_idle_conns: 20

lock:
  redis:
    ttl: 60s
    max_retries: 5
```

## 动态配置

某些配置项支持运行时动态修改（通过API或管理后台）：

- 日志级别
- 工作线程池大小
- 任务轮询间隔

其他配置项需要重启服务才能生效。

## 配置模板

项目提供了多个配置模板：

- `config.yaml.example` - 基础配置模板
- `config.dev.yaml` - 开发环境配置
- `config.prod.yaml` - 生产环境配置（需要创建）

## 故障排查

### 配置文件未找到

```
Error: Failed to load configuration: config file not found
```

解决方案：
- 检查配置文件路径是否正确
- 使用 `-config` 参数指定配置文件路径

### 数据库连接失败

```
Error: Failed to connect to database: dial tcp: connection refused
```

解决方案：
- 检查数据库是否运行
- 检查数据库地址和端口是否正确
- 检查网络连通性
- 检查用户名和密码是否正确

### Redis连接失败

```
Error: Failed to connect to Redis: dial tcp: connection refused
```

解决方案：
- 检查Redis是否运行
- 检查Redis地址和端口是否正确
- 检查Redis密码是否正确
- 检查网络连通性

## 参考资源

- [Viper配置库文档](https://github.com/spf13/viper)
- [YAML语法](https://yaml.org/)
