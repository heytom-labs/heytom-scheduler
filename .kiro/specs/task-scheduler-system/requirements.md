# 需求文档

## 简介

任务调度系统是一个分布式任务管理和执行平台，支持创建、调度和监控任务的完整生命周期。系统由Go语言编写的后端服务和Vue编写的管理后台组成，提供灵活的任务调度策略、多协议支持、分布式部署能力和可扩展的存储方案。

## 术语表

- **TaskScheduler**: 任务调度系统，负责管理和执行任务的核心服务
- **Task**: 任务，系统中可被调度和执行的工作单元
- **SubTask**: 子任务，从属于父任务的工作单元
- **ExecutionMode**: 执行模式，定义任务何时执行（立即、定时、间隔、Cron）
- **TaskStatus**: 任务状态，包括成功、失败、进行中
- **RetryPolicy**: 重试策略，定义任务失败后的重试行为
- **CallbackDriver**: 回调驱动，用于任务执行后通知的机制
- **CallbackProtocol**: 回调协议，支持HTTP和gRPC两种回调方式
- **ServiceDiscovery**: 服务发现，用于定位任务回调接口的机制
- **ClientProtocol**: 客户端协议，支持HTTP和gRPC两种通信方式
- **StorageBackend**: 存储后端，用于持久化任务数据的数据库
- **AdminUI**: 管理后台，基于Vue的Web界面
- **SchedulerNode**: 调度节点，TaskScheduler的一个运行实例
- **ConcurrencyLimit**: 并发限制，控制子任务同时执行的最大数量
- **AlertPolicy**: 报警策略，定义何时触发报警通知
- **NotificationChannel**: 通知渠道，用于发送报警消息的方式

## 需求

### 需求 1

**用户故事:** 作为系统用户，我希望能够通过API或Web界面创建任务，以便灵活地管理工作流程

#### 验收标准

1. WHEN 用户通过HTTP接口提交任务创建请求 THEN TaskScheduler SHALL 验证请求参数并创建新任务记录
2. WHEN 用户通过gRPC接口提交任务创建请求 THEN TaskScheduler SHALL 验证请求参数并创建新任务记录
3. WHEN 用户通过AdminUI提交任务创建表单 THEN TaskScheduler SHALL 接收请求并创建新任务记录
4. WHEN 任务创建成功 THEN TaskScheduler SHALL 返回包含任务唯一标识符的响应
5. WHEN 任务创建失败 THEN TaskScheduler SHALL 返回明确的错误信息说明失败原因

### 需求 2

**用户故事:** 作为系统用户，我希望任务能够包含多个子任务，以便组织复杂的工作流程

#### 验收标准

1. WHEN 用户创建任务时指定子任务列表 THEN TaskScheduler SHALL 创建父任务及其所有子任务的关联关系
2. WHEN 查询任务详情 THEN TaskScheduler SHALL 返回任务及其所有子任务的完整信息
3. WHEN 父任务被删除 THEN TaskScheduler SHALL 同时删除所有关联的子任务
4. WHEN 子任务执行状态变化 THEN TaskScheduler SHALL 更新父任务的聚合状态

### 需求 3

**用户故事:** 作为系统用户，我希望支持多种任务执行模式，以便满足不同的调度需求

#### 验收标准

1. WHEN 用户创建ExecutionMode为立即执行的任务 THEN TaskScheduler SHALL 在任务创建后立即将其加入执行队列
2. WHEN 用户创建ExecutionMode为定时执行的任务并指定执行时间 THEN TaskScheduler SHALL 在指定时间到达时执行该任务
3. WHEN 用户创建ExecutionMode为固定间隔的任务并指定间隔时长 THEN TaskScheduler SHALL 按照指定间隔重复执行该任务
4. WHEN 用户创建ExecutionMode为Cron的任务并指定Cron表达式 THEN TaskScheduler SHALL 按照Cron表达式定义的时间规则执行该任务
5. WHEN Cron表达式格式无效 THEN TaskScheduler SHALL 拒绝任务创建并返回格式错误信息

### 需求 4

**用户故事:** 作为系统用户，我希望任务具有明确的执行状态和重试机制，以便处理执行失败的情况

#### 验收标准

1. WHEN 任务开始执行 THEN TaskScheduler SHALL 将TaskStatus设置为进行中
2. WHEN 任务执行成功完成 THEN TaskScheduler SHALL 将TaskStatus设置为成功
3. WHEN 任务执行失败且RetryPolicy设置为不重试 THEN TaskScheduler SHALL 将TaskStatus设置为失败
4. WHEN 任务执行失败且RetryPolicy设置了重试次数 THEN TaskScheduler SHALL 重新执行任务直到成功或达到最大重试次数
5. WHEN 任务达到最大重试次数仍失败 THEN TaskScheduler SHALL 将TaskStatus设置为失败
6. WHEN 查询任务状态 THEN TaskScheduler SHALL 返回当前TaskStatus和已重试次数

### 需求 5

**用户故事:** 作为系统用户，我希望任务回调支持同步和异步模式，以便适应不同的执行场景

#### 验收标准

1. WHEN 任务配置为同步回调模式 THEN TaskScheduler SHALL 等待回调接口返回后再更新TaskStatus
2. WHEN 任务配置为异步回调模式 THEN TaskScheduler SHALL 发送回调请求后立即继续处理
3. WHEN 异步回调任务执行完成 THEN CallbackDriver SHALL 接收执行结果并更新TaskStatus
4. WHEN 回调接口响应超时 THEN TaskScheduler SHALL 根据RetryPolicy决定是否重试
5. WHEN 异步回调返回执行状态 THEN TaskScheduler SHALL 将返回的状态更新到任务记录
6. WHEN 任务配置为HTTP回调 THEN TaskScheduler SHALL 使用HTTP协议调用回调接口
7. WHEN 任务配置为gRPC回调 THEN TaskScheduler SHALL 使用gRPC协议调用回调接口

### 需求 6

**用户故事:** 作为系统集成者，我希望客户端支持HTTP和gRPC两种协议，以便适应不同的技术栈

#### 验收标准

1. WHEN 客户端通过HTTP协议调用任务创建接口 THEN TaskScheduler SHALL 使用相同的业务逻辑处理请求
2. WHEN 客户端通过gRPC协议调用任务创建接口 THEN TaskScheduler SHALL 使用相同的业务逻辑处理请求
3. WHEN 客户端通过HTTP协议查询任务状态 THEN TaskScheduler SHALL 返回与gRPC协议相同格式的数据
4. WHEN 客户端通过gRPC协议查询任务状态 THEN TaskScheduler SHALL 返回与HTTP协议相同格式的数据
5. WHEN 协议层接收到请求 THEN TaskScheduler SHALL 将请求转换为统一的内部表示后调用服务层

### 需求 7

**用户故事:** 作为系统运维人员，我希望服务端支持多节点部署，以便实现高可用和负载均衡

#### 验收标准

1. WHEN 多个SchedulerNode同时运行 THEN TaskScheduler SHALL 确保每个任务只被一个节点执行
2. WHEN 某个SchedulerNode故障 THEN TaskScheduler SHALL 将该节点上的任务转移到其他健康节点
3. WHEN 新的SchedulerNode加入集群 THEN TaskScheduler SHALL 自动发现新节点并参与任务分配
4. WHEN SchedulerNode离开集群 THEN TaskScheduler SHALL 检测到节点离开并重新分配任务
5. WHEN 任务被分配到节点 THEN TaskScheduler SHALL 记录任务与节点的绑定关系

### 需求 8

**用户故事:** 作为系统部署者，我希望支持多种数据库存储，以便根据实际环境选择合适的存储方案

#### 验收标准

1. WHEN 系统配置使用MySQL作为StorageBackend THEN TaskScheduler SHALL 将任务数据持久化到MySQL数据库
2. WHEN 系统配置使用PostgreSQL作为StorageBackend THEN TaskScheduler SHALL 将任务数据持久化到PostgreSQL数据库
3. WHEN 系统配置使用MongoDB作为StorageBackend THEN TaskScheduler SHALL 将任务数据持久化到MongoDB数据库
4. WHEN 切换StorageBackend THEN TaskScheduler SHALL 使用统一的数据访问接口而无需修改业务逻辑
5. WHEN 存储操作失败 THEN TaskScheduler SHALL 返回明确的错误信息并保持数据一致性

### 需求 9

**用户故事:** 作为系统集成者，我希望任务回调接口支持多驱动的服务发现，以便灵活定位回调服务

#### 验收标准

1. WHEN 系统配置使用静态地址作为ServiceDiscovery THEN CallbackDriver SHALL 使用配置的固定地址进行回调
2. WHEN 系统配置使用Consul作为ServiceDiscovery THEN CallbackDriver SHALL 通过Consul查询服务地址后进行回调
3. WHEN 系统配置使用Etcd作为ServiceDiscovery THEN CallbackDriver SHALL 通过Etcd查询服务地址后进行回调
4. WHEN 系统配置使用Kubernetes作为ServiceDiscovery THEN CallbackDriver SHALL 通过Kubernetes服务发现机制查询地址后进行回调
5. WHEN ServiceDiscovery返回多个服务实例 THEN CallbackDriver SHALL 使用负载均衡策略选择一个实例进行回调
6. WHEN ServiceDiscovery查询失败 THEN CallbackDriver SHALL 记录错误并根据RetryPolicy决定是否重试

### 需求 10

**用户故事:** 作为管理员，我希望通过Web界面管理任务，以便直观地监控和操作任务

#### 验收标准

1. WHEN 管理员访问AdminUI THEN AdminUI SHALL 显示任务列表及其当前状态
2. WHEN 管理员在AdminUI中创建任务 THEN AdminUI SHALL 提交请求到TaskScheduler并显示创建结果
3. WHEN 管理员在AdminUI中查看任务详情 THEN AdminUI SHALL 显示任务的完整信息包括子任务和执行历史
4. WHEN 管理员在AdminUI中取消任务 THEN AdminUI SHALL 发送取消请求到TaskScheduler
5. WHEN 任务状态变化 THEN AdminUI SHALL 实时更新显示的任务状态

### 需求 11

**用户故事:** 作为系统用户，我希望控制子任务的并发执行数量，以便避免资源过载和优化执行效率

#### 验收标准

1. WHEN 用户创建任务时指定ConcurrencyLimit THEN TaskScheduler SHALL 记录该任务的最大并发子任务数量
2. WHEN 任务包含多个子任务且设置了ConcurrencyLimit THEN TaskScheduler SHALL 确保同时执行的子任务数量不超过限制值
3. WHEN 正在执行的子任务数量达到ConcurrencyLimit THEN TaskScheduler SHALL 将剩余子任务保持在等待状态
4. WHEN 某个子任务执行完成且有等待中的子任务 THEN TaskScheduler SHALL 从等待队列中选择下一个子任务开始执行
5. WHEN 未设置ConcurrencyLimit THEN TaskScheduler SHALL 允许所有子任务同时执行

### 需求 12

**用户故事:** 作为系统管理员，我希望设置报警通知策略，以便及时了解任务执行异常情况

#### 验收标准

1. WHEN 用户创建任务时配置AlertPolicy THEN TaskScheduler SHALL 保存报警策略配置
2. WHEN 任务执行失败且AlertPolicy配置了失败报警 THEN TaskScheduler SHALL 通过NotificationChannel发送报警消息
3. WHEN 任务重试次数超过AlertPolicy设置的阈值 THEN TaskScheduler SHALL 通过NotificationChannel发送报警消息
4. WHEN 任务执行时间超过AlertPolicy设置的超时阈值 THEN TaskScheduler SHALL 通过NotificationChannel发送报警消息
5. WHEN AlertPolicy配置了多个NotificationChannel THEN TaskScheduler SHALL 向所有配置的渠道发送报警消息
6. WHEN NotificationChannel发送失败 THEN TaskScheduler SHALL 记录发送失败日志但不影响任务执行
7. WHEN 报警消息发送 THEN TaskScheduler SHALL 包含任务标识符、失败原因、发生时间等详细信息
