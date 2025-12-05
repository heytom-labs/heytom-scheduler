# 分布式锁和多节点支持实现文档

## 概述

本文档描述了任务调度系统中分布式锁和多节点支持的实现。该实现满足需求7.1、7.2、7.3、7.4和7.5。

## 实现的组件

### 1. 分布式锁 (Distributed Lock)

#### 1.1 Redis分布式锁 (`internal/lock/redis_lock.go`)

基于Redis的分布式锁实现，使用Redis的`SETNX`命令实现原子性锁获取。

**特性:**
- 使用`SET key value NX EX ttl`实现原子性锁获取
- 支持TTL自动过期
- 支持锁刷新(Refresh)
- 线程安全

**接口实现:**
```go
Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
Unlock(ctx context.Context, key string) error
Refresh(ctx context.Context, key string, ttl time.Duration) error
```

#### 1.2 Etcd分布式锁 (`internal/lock/etcd_lock.go`)

基于Etcd的分布式锁实现，使用Etcd的concurrency包提供的Mutex。

**特性:**
- 使用Etcd的Session和Mutex机制
- 自动心跳保持会话活跃
- 支持锁超时
- 支持锁刷新

#### 1.3 锁工厂 (`internal/lock/factory.go`)

提供统一的接口创建不同类型的分布式锁。

**支持的锁类型:**
- Redis
- Etcd

#### 1.4 锁自动续期 (`internal/lock/renewer.go`)

自动续期管理器，定期刷新锁的TTL，防止锁过期。

**特性:**
- 后台goroutine定期刷新锁
- 支持多个锁同时续期
- 优雅停止机制

### 2. 节点注册和管理 (Node Registry)

#### 2.1 节点注册表 (`internal/node/registry.go`)

管理调度节点的注册、注销和健康状态。

**特性:**
- 节点注册和注销
- 心跳更新
- 健康检查
- 自动标记不健康节点

**节点状态:**
- `healthy`: 节点健康
- `unhealthy`: 节点不健康
- `unknown`: 节点状态未知

#### 2.2 心跳管理器 (`internal/node/heartbeat.go`)

定期发送心跳到注册表，保持节点活跃状态。

**特性:**
- 定期发送心跳
- 可配置心跳间隔
- 优雅启动和停止

### 3. 任务分配 (Task Allocation)

#### 3.1 任务分配器 (`internal/node/allocator.go`)

负责将任务分配给健康的节点。

**分配策略:**
- `round_robin`: 轮询分配
- `random`: 随机分配
- `least_loaded`: 最少负载分配

**特性:**
- 使用分布式锁确保任务唯一分配
- 自动选择健康节点
- 记录任务与节点的绑定关系
- 支持任务释放

### 4. 故障转移 (Failover)

#### 4.1 故障转移管理器 (`internal/node/failover.go`)

检测节点故障并自动转移任务到健康节点。

**特性:**
- 定期检查节点健康状态
- 自动检测不健康节点
- 自动转移失败节点的任务
- 支持手动强制转移任务
- 详细的转移日志

**工作流程:**
1. 定期检查所有节点健康状态
2. 识别不健康节点
3. 获取不健康节点上的所有任务
4. 释放旧锁
5. 将任务重置为pending状态
6. 重新分配任务到健康节点

## 使用示例

### 创建Redis分布式锁

```go
import (
    "github.com/redis/go-redis/v9"
    "task-scheduler/internal/lock"
)

// 创建Redis客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建分布式锁
distributedLock := lock.NewRedisLock(redisClient, "lock:")

// 获取锁
acquired, err := distributedLock.Lock(ctx, "task:123", 5*time.Minute)
if err != nil {
    // 处理错误
}

if acquired {
    // 执行任务
    defer distributedLock.Unlock(ctx, "task:123")
}
```

### 创建节点注册表和心跳

```go
import "task-scheduler/internal/node"

// 创建节点注册表
registry := node.NewInMemoryRegistry(30 * time.Second)

// 注册节点
myNode := &node.Node{
    ID:      "node-1",
    Address: "localhost:8080",
}
registry.Register(ctx, myNode)

// 创建心跳管理器
heartbeat := node.NewHeartbeatManager(registry, "node-1", 10*time.Second)
heartbeat.Start()
defer heartbeat.Stop()
```

### 创建任务分配器

```go
import "task-scheduler/internal/node"

// 创建任务分配器
allocator := node.NewAllocator(
    registry,
    taskRepository,
    distributedLock,
    node.AllocationStrategyLeastLoaded,
)

// 分配任务
selectedNode, err := allocator.AllocateTask(ctx, task)
if err != nil {
    // 处理错误
}

// 任务完成后释放
defer allocator.ReleaseTask(ctx, task.ID)
```

### 创建故障转移管理器

```go
import "task-scheduler/internal/node"

// 创建故障转移管理器
failover := node.NewFailoverManager(
    registry,
    taskRepository,
    allocator,
    30 * time.Second, // 检查间隔
)

// 启动故障转移
failover.Start()
defer failover.Stop()
```

## 测试

所有组件都包含完整的单元测试：

- `internal/lock/redis_lock_test.go`: Redis锁测试
- `internal/node/registry_test.go`: 节点注册表测试
- `internal/node/failover_test.go`: 故障转移测试

运行测试：
```bash
# 运行所有节点相关测试
go test ./internal/node/... -v

# 运行锁相关测试（需要Redis）
go test ./internal/lock/... -v
```

## 配置建议

### 生产环境配置

```yaml
distributed_lock:
  type: redis  # 或 etcd
  prefix: "scheduler:lock:"
  
node:
  heartbeat_interval: 10s
  heartbeat_timeout: 30s
  
failover:
  check_interval: 30s
  
allocator:
  strategy: least_loaded  # round_robin, random, least_loaded
```

### 性能调优

1. **心跳间隔**: 建议设置为10-30秒，过短会增加网络开销
2. **锁TTL**: 建议设置为任务预期执行时间的2-3倍
3. **故障检查间隔**: 建议设置为心跳超时时间的1-2倍
4. **分配策略**: 
   - 小规模集群使用`round_robin`
   - 大规模集群使用`least_loaded`

## 满足的需求

- ✅ 需求7.1: 多节点部署时确保任务唯一执行（通过分布式锁）
- ✅ 需求7.2: 节点故障时自动转移任务（通过故障转移管理器）
- ✅ 需求7.3: 新节点自动加入集群（通过节点注册）
- ✅ 需求7.4: 节点离开时重新分配任务（通过健康检查和故障转移）
- ✅ 需求7.5: 记录任务与节点的绑定关系（在Task.NodeID字段）

## 未来改进

1. 支持更多分布式锁后端（如Zookeeper）
2. 实现更智能的任务分配策略（基于节点负载、地理位置等）
3. 添加节点权重和优先级
4. 实现任务亲和性（某些任务优先分配到特定节点）
5. 添加分布式追踪支持
