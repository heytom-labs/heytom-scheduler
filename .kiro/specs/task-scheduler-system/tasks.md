# 实施计划

- [x] 1. 初始化项目结构和核心接口





  - 创建Go项目目录结构（cmd、internal、pkg、api等）
  - 定义核心数据模型（Task、ExecutionMode、TaskStatus等）
  - 定义Repository、Service、Scheduler等核心接口
  - 设置依赖管理（go.mod）
  - 配置日志框架（zap）
  - _需求: 1.1, 1.2, 2.1_

- [-] 2. 实现存储层


- [x] 2.1 实现Repository接口和MySQL适配器


  - 实现TaskRepository接口
  - 实现MySQL存储适配器
  - 实现数据库连接池管理
  - 实现事务支持
  - _需求: 8.1_

- [ ]* 2.2 编写属性测试验证存储后端数据持久化
  - **Feature: task-scheduler-system, Property 17: 存储后端数据持久化**
  - **验证: 需求 8.1**


- [x] 2.3 实现PostgreSQL和MongoDB适配器


  - 实现PostgreSQL存储适配器
  - 实现MongoDB存储适配器
  - 确保统一的Repository接口
  - _需求: 8.2, 8.3_

- [ ]* 2.4 编写单元测试验证不同存储后端
  - 测试MySQL适配器的CRUD操作
  - 测试PostgreSQL适配器的CRUD操作
  - 测试MongoDB适配器的CRUD操作
  - _需求: 8.1, 8.2, 8.3_

- [-] 3. 实现任务服务层


- [x] 3.1 实现TaskService核心业务逻辑


  - 实现CreateTask方法（包含输入验证）
  - 实现GetTask、ListTasks方法
  - 实现CancelTask方法
  - 实现UpdateTaskStatus方法
  - 实现父子任务关联逻辑
  - _需求: 1.1, 1.2, 2.1, 2.2_

- [ ]* 3.2 编写属性测试验证任务创建和ID唯一性
  - **Feature: task-scheduler-system, Property 2: 任务创建响应包含唯一ID**
  - **验证: 需求 1.4**

- [ ]* 3.3 编写属性测试验证父子任务级联删除
  - **Feature: task-scheduler-system, Property 3: 父子任务级联删除**
  - **验证: 需求 2.3**

- [ ]* 3.4 编写属性测试验证子任务状态聚合
  - **Feature: task-scheduler-system, Property 4: 子任务状态聚合**
  - **验证: 需求 2.4**


- [x] 4. 实现调度服务层




- [x] 4.1 实现ScheduleService和Cron解析


  - 集成robfig/cron库
  - 实现Cron表达式验证和解析
  - 实现GetNextExecutionTime方法
  - 实现定时任务、间隔任务的时间计算
  - _需求: 3.2, 3.3, 3.4, 3.5_

- [ ]* 4.2 编写属性测试验证不同执行模式
  - **Feature: task-scheduler-system, Property 5: 立即执行任务入队**
  - **验证: 需求 3.1**

- [ ]* 4.3 编写属性测试验证定时和间隔任务
  - **Feature: task-scheduler-system, Property 6: 定时任务按时执行**
  - **Feature: task-scheduler-system, Property 7: 间隔任务重复执行**
  - **验证: 需求 3.2, 3.3**

- [ ]* 4.4 编写属性测试验证Cron任务执行
  - **Feature: task-scheduler-system, Property 8: Cron任务按规则执行**
  - **验证: 需求 3.4**



- [x] 4.5 实现任务调度器核心逻辑





  - 实现Scheduler接口
  - 实现任务队列管理
  - 实现任务执行器（worker pool）
  - 实现任务状态转换逻辑
  - _需求: 3.1, 4.1, 4.2_

- [ ]* 4.6 编写属性测试验证任务状态转换
  - **Feature: task-scheduler-system, Property 9: 任务状态转换正确性**
  - **验证: 需求 4.1, 4.2, 4.3**

- [x] 5. 实现重试机制




- [x] 5.1 实现RetryPolicy逻辑


  - 实现重试次数控制
  - 实现重试间隔计算（支持指数退避）
  - 集成到任务执行流程
  - _需求: 4.4, 4.5, 4.6_

- [ ]* 5.2 编写属性测试验证重试机制
  - **Feature: task-scheduler-system, Property 10: 重试机制正确性**
  - **验证: 需求 4.4, 4.5**

- [ ]* 5.3 编写单元测试验证重试策略
  - 测试不重试场景
  - 测试固定次数重试
  - 测试指数退避计算
  - _需求: 4.3, 4.4, 4.5_

- [x] 6. 实现回调服务





- [x] 6.1 实现CallbackService基础功能


  - 实现同步回调逻辑
  - 实现异步回调逻辑
  - 实现HTTP客户端封装
  - 实现gRPC客户端封装
  - 实现超时控制
  - _需求: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

- [ ]* 6.2 编写属性测试验证同步和异步回调
  - **Feature: task-scheduler-system, Property 11: 同步回调等待返回**
  - **Feature: task-scheduler-system, Property 12: 异步回调立即继续**
  - **验证: 需求 5.1, 5.2**

- [ ]* 6.2.1 编写属性测试验证回调协议支持
  - **Feature: task-scheduler-system, Property 12.1: 回调协议支持**
  - **验证: 需求 5.6, 5.7**

- [x] 6.3 实现服务发现层


  - 定义ServiceDiscovery接口
  - 实现静态地址服务发现
  - 实现Consul服务发现
  - 实现Etcd服务发现
  - 实现Kubernetes服务发现
  - 实现负载均衡策略（轮询、随机）
  - _需求: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ]* 6.4 编写属性测试验证服务发现
  - **Feature: task-scheduler-system, Property 18: 服务发现地址解析**
  - **验证: 需求 9.1, 9.2, 9.3, 9.4**

- [ ]* 6.5 编写属性测试验证负载均衡
  - **Feature: task-scheduler-system, Property 19: 多实例负载均衡**
  - **验证: 需求 9.5**

- [x] 7. 实现分布式锁和多节点支持




- [x] 7.1 实现分布式锁


  - 定义DistributedLock接口
  - 实现基于Redis的分布式锁
  - 实现基于Etcd的分布式锁
  - 实现锁的自动续期机制
  - _需求: 7.1_



- [x] 7.2 实现多节点任务分配





  - 实现节点注册和心跳检测
  - 实现任务获取和锁定逻辑
  - 实现任务与节点的绑定记录
  - _需求: 7.1, 7.3, 7.5_

- [ ]* 7.3 编写属性测试验证分布式任务唯一执行
  - **Feature: task-scheduler-system, Property 14: 分布式任务唯一执行**
  - **验证: 需求 7.1**

- [ ]* 7.4 编写属性测试验证任务节点绑定
  - **Feature: task-scheduler-system, Property 16: 任务节点绑定记录**


  - **验证: 需求 7.5**

- [x] 7.5 实现节点故障检测和任务转移





  - 实现节点健康检查
  - 实现故障节点检测
  - 实现任务自动转移逻辑
  - _需求: 7.2, 7.4_

- [ ]* 7.6 编写属性测试验证节点故障转移
  - **Feature: task-scheduler-system, Property 15: 节点故障任务转移**
  - **验证: 需求 7.2**

- [x] 8. 实现并发控制




- [x] 8.1 实现子任务并发限制


  - 实现ConcurrencyLimit配置
  - 实现子任务并发数控制
  - 实现子任务等待队列
  - 实现子任务调度逻辑
  - _需求: 11.1, 11.2, 11.3, 11.4, 11.5_

- [ ]* 8.2 编写属性测试验证并发限制
  - **Feature: task-scheduler-system, Property 20: 并发限制控制**
  - **验证: 需求 11.2, 11.3**

- [ ]* 8.3 编写属性测试验证子任务调度队列
  - **Feature: task-scheduler-system, Property 21: 子任务调度队列**
  - **验证: 需求 11.4**

- [ ]* 8.4 编写单元测试验证并发控制边界情况
  - 测试未设置并发限制的情况
  - 测试并发限制为1的情况
  - 测试并发限制大于子任务数的情况
  - _需求: 11.5_

- [x] 9. 实现报警通知系统




- [x] 9.1 实现通知层基础架构


  - 定义Notifier接口
  - 定义AlertPolicy数据模型
  - 实现报警触发条件判断
  - _需求: 12.1_

- [x] 9.2 实现多种通知渠道


  - 实现Email通知渠道
  - 实现Webhook通知渠道
  - 实现SMS通知渠道（可选）
  - 实现报警消息格式化
  - _需求: 12.2, 12.5, 12.7_

- [ ]* 9.3 编写属性测试验证报警触发
  - **Feature: task-scheduler-system, Property 22: 报警策略触发**
  - **验证: 需求 12.2, 12.3, 12.4**

- [ ]* 9.4 编写属性测试验证多渠道报警
  - **Feature: task-scheduler-system, Property 23: 多渠道报警发送**
  - **验证: 需求 12.5**

- [ ]* 9.5 编写属性测试验证报警消息完整性
  - **Feature: task-scheduler-system, Property 24: 报警消息完整性**
  - **验证: 需求 12.7**


- [x] 9.6 集成报警到任务执行流程

  - 在任务失败时触发报警
  - 在重试超过阈值时触发报警
  - 在执行超时时触发报警
  - 确保报警发送失败不影响任务执行
  - _需求: 12.2, 12.3, 12.4, 12.6_

- [-] 10. 实现HTTP协议层


- [x] 10.1 实现HTTP Handler和路由



  - 集成Gin框架
  - 实现CreateTask HTTP接口
  - 实现GetTask、ListTasks HTTP接口
  - 实现CancelTask HTTP接口
  - 实现GetTaskStatus HTTP接口
  - 实现请求参数验证
  - 实现统一的错误响应格式
  - _需求: 1.1, 1.5_

- [ ]* 10.2 编写单元测试验证HTTP接口
  - 测试各个HTTP端点
  - 测试参数验证
  - 测试错误处理
  - _需求: 1.1, 1.5_

- [ ] 11. 实现gRPC协议层




- [x] 11.1 定义gRPC Protocol Buffers


  - 定义TaskSchedulerService服务
  - 定义所有请求和响应消息
  - 生成Go代码
  - _需求: 1.2_



- [x] 11.2 实现gRPC Handler

  - 实现CreateTask gRPC方法
  - 实现GetTask、ListTasks gRPC方法
  - 实现CancelTask gRPC方法
  - 实现GetTaskStatus gRPC方法
  - 实现gRPC错误处理
  - _需求: 1.2, 1.5_

- [ ]* 11.3 编写属性测试验证协议一致性
  - **Feature: task-scheduler-system, Property 1: 协议无关的任务创建**
  - **Feature: task-scheduler-system, Property 13: 协议返回格式一致性**
  - **验证: 需求 1.1, 1.2, 6.1, 6.2, 6.3, 6.4**

- [ ]* 11.4 编写单元测试验证gRPC接口
  - 测试各个gRPC方法
  - 测试参数验证
  - 测试错误处理
  - _需求: 1.2, 1.5_

- [x] 12. 实现配置管理和启动流程






- [x] 12.1 实现配置管理

  - 集成viper配置库
  - 定义配置文件结构（YAML）
  - 实现配置加载和验证
  - 支持环境变量覆盖
  - _需求: 8.1, 8.2, 8.3, 9.1, 9.2, 9.3, 9.4_


- [x] 12.2 实现应用启动和优雅关闭

  - 实现main函数和启动流程
  - 实现依赖注入和组件初始化
  - 实现优雅关闭（graceful shutdown）
  - 实现健康检查接口
  - _需求: 7.3, 7.4_

- [ ] 13. Checkpoint - 确保所有后端测试通过
  - 确保所有测试通过，如有问题请询问用户

- [x] 14. 实现Vue管理后台





- [x] 14.1 初始化Vue项目


  - 使用Vite创建Vue 3 + TypeScript项目
  - 集成Element Plus UI组件库
  - 配置Vue Router
  - 配置Pinia状态管理
  - 配置Axios HTTP客户端
  - _需求: 10.1_



- [x] 14.2 实现任务列表页面

  - 实现任务列表组件
  - 实现任务状态显示
  - 实现任务筛选和搜索
  - 实现分页功能

  - _需求: 10.1_

- [x] 14.3 实现任务创建页面

  - 实现任务创建表单
  - 实现执行模式选择
  - 实现子任务配置
  - 实现重试策略配置
  - 实现并发限制配置

  - 实现报警策略配置
  - 实现表单验证
  - _需求: 10.2, 1.3_

- [x] 14.4 实现任务详情页面



  - 实现任务详情展示
  - 实现子任务列表展示
  - 实现执行历史展示
  - 实现任务取消功能
  - _需求: 10.3, 10.4_




- [ ] 14.5 实现实时状态更新
  - 实现WebSocket连接
  - 实现任务状态实时推送
  - 实现UI自动刷新
  - _需求: 10.5_

- [ ]* 14.6 编写前端单元测试
  - 测试关键组件
  - 测试状态管理
  - 测试API调用
  - _需求: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 15. 实现监控和可观测性





- [x] 15.1 集成Prometheus指标


  - 实现任务创建/执行/完成计数器
  - 实现任务执行延迟直方图
  - 实现失败率和重试率指标
  - 实现节点健康状态指标
  - 暴露/metrics端点
  - _需求: 所有需求的监控_

- [x] 15.2 实现结构化日志


  - 配置zap日志格式（JSON）
  - 实现分布式追踪ID
  - 在关键操作点添加日志
  - _需求: 所有需求的日志_
- [x] 16. 编写部署文档和配置




- [ ] 16. 编写部署文档和配置

- [x] 16.1 创建Docker镜像







  - 编写Dockerfile（后端）
  - 编写Dockerfile（前端）
  - 编写docker-compose.yml（开发环境）
  - _需求: 所有需求_
- [x] 16.2 创建Kubernetes部署配置




- [x] 16.2 创建Kubernetes部署配置



  - 编写Deployment配置
  - 编写Service配置
  - 编写ConfigMap配置
  - 编写数据库部署配置
  - _需求: 7.1, 7.2, 7.3, 7.4_
-

- [x] 16.3 编写部署和运维文档






  - 编写README.md
  - 编写部署指南
  - 编写配置说明
  - 编写API文档
  - _需求: 所有需求_

- [x] 17. Final Checkpoint - 确保所有测试通过




  - 确保所有测试通过，如有问题请询问用户
