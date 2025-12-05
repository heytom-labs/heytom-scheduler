package domain

import "time"

// ExecutionMode defines how a task should be executed
type ExecutionMode string

const (
	ExecutionModeImmediate ExecutionMode = "immediate" // 立即执行
	ExecutionModeScheduled ExecutionMode = "scheduled" // 定时执行
	ExecutionModeInterval  ExecutionMode = "interval"  // 固定间隔
	ExecutionModeCron      ExecutionMode = "cron"      // Cron表达式
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待中
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusSuccess   TaskStatus = "success"   // 成功
	TaskStatusFailed    TaskStatus = "failed"    // 失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
)

// Task represents a schedulable work unit
type Task struct {
	ID               string            // 任务唯一标识
	Name             string            // 任务名称
	Description      string            // 任务描述
	ParentID         *string           // 父任务ID（子任务时非空）
	ExecutionMode    ExecutionMode     // 执行模式
	ScheduleConfig   *ScheduleConfig   // 调度配置
	CallbackConfig   *CallbackConfig   // 回调配置
	RetryPolicy      *RetryPolicy      // 重试策略
	ConcurrencyLimit int               // 子任务并发限制
	AlertPolicy      *AlertPolicy      // 报警策略
	Status           TaskStatus        // 任务状态
	RetryCount       int               // 已重试次数
	NodeID           *string           // 执行节点ID
	CreatedAt        time.Time         // 创建时间
	UpdatedAt        time.Time         // 更新时间
	StartedAt        *time.Time        // 开始执行时间
	CompletedAt      *time.Time        // 完成时间
	Metadata         map[string]string // 元数据
}

// ScheduleConfig contains scheduling parameters
type ScheduleConfig struct {
	ScheduledTime *time.Time     // 定时执行时间
	Interval      *time.Duration // 间隔时长
	CronExpr      *string        // Cron表达式
}

// RetryPolicy defines retry behavior for failed tasks
type RetryPolicy struct {
	MaxRetries    int           // 最大重试次数（0表示不重试）
	RetryInterval time.Duration // 重试间隔
	BackoffFactor float64       // 退避因子（指数退避）
}

// ServiceDiscoveryType defines the service discovery mechanism
type ServiceDiscoveryType string

const (
	ServiceDiscoveryStatic     ServiceDiscoveryType = "static"     // 静态地址
	ServiceDiscoveryConsul     ServiceDiscoveryType = "consul"     // Consul
	ServiceDiscoveryEtcd       ServiceDiscoveryType = "etcd"       // Etcd
	ServiceDiscoveryKubernetes ServiceDiscoveryType = "kubernetes" // Kubernetes
)

// CallbackProtocol defines the callback protocol type
type CallbackProtocol string

const (
	CallbackProtocolHTTP CallbackProtocol = "http" // HTTP协议
	CallbackProtocolGRPC CallbackProtocol = "grpc" // gRPC协议
)

// CallbackConfig contains callback parameters
type CallbackConfig struct {
	Protocol      CallbackProtocol     // 回调协议类型
	URL           string               // 回调URL（HTTP）或地址（gRPC）
	Method        string               // HTTP方法（仅HTTP使用）
	GRPCService   *string              // gRPC服务名（仅gRPC使用）
	GRPCMethod    *string              // gRPC方法名（仅gRPC使用）
	Headers       map[string]string    // 请求头（HTTP）或元数据（gRPC）
	Timeout       time.Duration        // 超时时间
	IsAsync       bool                 // 是否异步
	ServiceName   *string              // 服务名（用于服务发现）
	DiscoveryType ServiceDiscoveryType // 服务发现类型
}

// ChannelType defines notification channel types
type ChannelType string

const (
	ChannelTypeEmail   ChannelType = "email"   // 邮件
	ChannelTypeWebhook ChannelType = "webhook" // Webhook
	ChannelTypeSMS     ChannelType = "sms"     // 短信
)

// NotificationChannel represents a notification delivery method
type NotificationChannel struct {
	Type   ChannelType       // 渠道类型
	Config map[string]string // 渠道配置
}

// AlertPolicy defines alerting rules
type AlertPolicy struct {
	EnableFailureAlert bool                  // 启用失败报警
	RetryThreshold     int                   // 重试次数阈值
	TimeoutThreshold   time.Duration         // 超时阈值
	Channels           []NotificationChannel // 通知渠道
}

// ExecutionResult contains task execution outcome
type ExecutionResult struct {
	TaskID    string            // 任务ID
	Status    TaskStatus        // 执行状态
	Output    string            // 输出内容
	Error     *string           // 错误信息
	StartTime time.Time         // 开始时间
	EndTime   time.Time         // 结束时间
	Metadata  map[string]string // 元数据
}
