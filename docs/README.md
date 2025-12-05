# Task Scheduler System 文档索引

欢迎查阅Task Scheduler System的完整文档。

## 快速导航

### 新手入门

- [快速开始指南](../QUICKSTART.md) - 5分钟快速上手
- [项目README](../README.md) - 项目概述和特性介绍
- [Docker快速参考](../DOCKER.md) - Docker环境快速参考

### API文档

- [HTTP API文档](api-http.md) - RESTful API完整参考
- [gRPC API文档](api-grpc.md) - gRPC API完整参考

### 部署指南

- [Docker部署指南](deployment-docker.md) - 使用Docker和Docker Compose部署
- [Kubernetes部署指南](deployment-kubernetes.md) - 在Kubernetes集群中部署

### 配置和运维

- [配置说明](configuration.md) - 所有配置选项的详细说明
- [运维指南](operations.md) - 日常运维操作和最佳实践
- [故障排查指南](troubleshooting.md) - 常见问题的排查和解决方案
- [监控和可观测性](monitoring-observability.md) - 监控配置和指标说明

### 架构和设计

- [架构概述](architecture.md) - 系统架构和组件说明
- [分布式锁实现](distributed-lock-implementation.md) - 分布式锁的实现细节
- [多节点任务分配](multi-node-task-allocation.md) - 多节点部署和任务分配机制

### 开发指南

- [贡献指南](../CONTRIBUTING.md) - 如何为项目做贡献
- [代码规范](../CONTRIBUTING.md#代码规范) - 代码风格和规范
- [开发环境设置](../CONTRIBUTING.md#开发环境设置) - 本地开发环境配置

## 按主题浏览

### 部署相关

1. [快速开始](../QUICKSTART.md) - 最快的方式启动系统
2. [Docker部署](deployment-docker.md) - 适合开发和小规模生产环境
3. [Kubernetes部署](deployment-kubernetes.md) - 适合大规模生产环境
4. [配置说明](configuration.md) - 配置所有选项

### 使用相关

1. [HTTP API](api-http.md) - 使用HTTP接口创建和管理任务
2. [gRPC API](api-grpc.md) - 使用gRPC接口创建和管理任务
3. [配置说明](configuration.md) - 配置任务调度、重试、报警等

### 运维相关

1. [运维指南](operations.md) - 日常运维操作
2. [监控和可观测性](monitoring-observability.md) - 配置监控和告警
3. [故障排查](troubleshooting.md) - 解决常见问题
4. [备份和恢复](operations.md#备份和恢复) - 数据备份和灾难恢复

### 开发相关

1. [贡献指南](../CONTRIBUTING.md) - 开始贡献代码
2. [架构概述](architecture.md) - 了解系统架构
3. [分布式锁实现](distributed-lock-implementation.md) - 深入了解分布式锁
4. [多节点任务分配](multi-node-task-allocation.md) - 深入了解任务分配

## 文档结构

```
docs/
├── README.md                              # 本文件 - 文档索引
├── api-http.md                           # HTTP API文档
├── api-grpc.md                           # gRPC API文档
├── architecture.md                       # 架构概述
├── configuration.md                      # 配置说明
├── deployment-docker.md                  # Docker部署指南
├── deployment-kubernetes.md              # Kubernetes部署指南
├── distributed-lock-implementation.md    # 分布式锁实现
├── monitoring-observability.md           # 监控和可观测性
├── multi-node-task-allocation.md         # 多节点任务分配
├── operations.md                         # 运维指南
└── troubleshooting.md                    # 故障排查指南
```

## 常见使用场景

### 场景1: 首次部署

1. 阅读 [快速开始指南](../QUICKSTART.md)
2. 选择部署方式：
   - 开发环境：[Docker部署](deployment-docker.md)
   - 生产环境：[Kubernetes部署](deployment-kubernetes.md)
3. 配置系统：[配置说明](configuration.md)
4. 设置监控：[监控和可观测性](monitoring-observability.md)

### 场景2: 集成到现有系统

1. 阅读 [HTTP API文档](api-http.md) 或 [gRPC API文档](api-grpc.md)
2. 了解任务执行模式和回调机制
3. 配置服务发现（如需要）
4. 实现回调接口

### 场景3: 生产环境运维

1. 阅读 [运维指南](operations.md)
2. 配置监控告警：[监控和可观测性](monitoring-observability.md)
3. 设置备份：[备份和恢复](operations.md#备份和恢复)
4. 熟悉故障排查：[故障排查指南](troubleshooting.md)

### 场景4: 性能优化

1. 查看 [性能调优](operations.md#性能调优)
2. 分析监控指标：[监控和可观测性](monitoring-observability.md)
3. 优化配置：[配置说明](configuration.md)
4. 考虑水平扩展：[多节点任务分配](multi-node-task-allocation.md)

### 场景5: 贡献代码

1. 阅读 [贡献指南](../CONTRIBUTING.md)
2. 了解系统架构：[架构概述](architecture.md)
3. 设置开发环境：[开发环境设置](../CONTRIBUTING.md#开发环境设置)
4. 遵循代码规范：[代码规范](../CONTRIBUTING.md#代码规范)

## 获取帮助

如果您在文档中找不到需要的信息，可以：

1. 搜索 [GitHub Issues](https://github.com/yourusername/task-scheduler/issues)
2. 查看 [故障排查指南](troubleshooting.md)
3. 创建新的 [Issue](https://github.com/yourusername/task-scheduler/issues/new)
4. 加入 [讨论区](https://github.com/yourusername/task-scheduler/discussions)
5. 发送邮件至 support@example.com

## 文档贡献

我们欢迎文档改进！如果您发现文档中的错误或希望改进文档，请：

1. Fork项目
2. 修改文档
3. 提交Pull Request

详见 [贡献指南](../CONTRIBUTING.md)。

## 版本说明

本文档适用于Task Scheduler System v1.0.0及以上版本。

不同版本的文档可能有所不同，请确保查看与您使用的版本对应的文档。

## 许可证

本文档采用 [MIT许可证](../LICENSE)。

