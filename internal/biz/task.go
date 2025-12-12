package biz

import (
	"context"
	"time"

	pb "heytom-scheduler/api/scheduler/v1"
)

// Task 任务业务模型
type Task struct {
	ID             int64
	Name           string
	Description    string
	Type           pb.TaskType
	Status         pb.TaskStatus
	Schedule       string
	Handler        string
	Payload        string
	Timeout        int32
	Metadata       map[string]string
	NextRunTime    *time.Time
	ExecutionCount int64
	SuccessCount   int64
	FailedCount    int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TaskExecution 任务执行记录业务模型
type TaskExecution struct {
	ID         int64
	TaskID     int64
	TaskName   string
	Status     pb.ExecutionStatus
	NodeID     string
	StartTime  *time.Time
	EndTime    *time.Time
	Duration   int32
	Result     string
	Error      string
	RetryCount int32
	Payload    string
	CreatedAt  time.Time
}

// TaskListFilter 任务列表过滤条件
type TaskListFilter struct {
	Page     int32
	PageSize int32
	Status   pb.TaskStatus
	Type     pb.TaskType
	Keyword  string
}

// ExecutionListFilter 执行记录列表过滤条件
type ExecutionListFilter struct {
	TaskID   int64
	Page     int32
	PageSize int32
	Status   pb.ExecutionStatus
}

// TaskRepo 任务仓储接口
type TaskRepo interface {
	// CreateTask 创建任务
	CreateTask(ctx context.Context, task *Task) (*Task, error)

	// GetTask 获取任务详情
	GetTask(ctx context.Context, id int64) (*Task, error)

	// UpdateTask 更新任务
	UpdateTask(ctx context.Context, task *Task) (*Task, error)

	// DeleteTask 删除任务
	DeleteTask(ctx context.Context, id int64) error

	// ListTasks 任务列表查询
	ListTasks(ctx context.Context, filter *TaskListFilter) ([]*Task, int64, error)

	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(ctx context.Context, id int64, status pb.TaskStatus) error

	// UpdateTaskNextRunTime 更新任务下次执行时间
	UpdateTaskNextRunTime(ctx context.Context, id int64, nextRunTime time.Time) error

	// IncrementExecutionCount 增加执行次数
	IncrementExecutionCount(ctx context.Context, id int64, success bool) error
}

// ExecutionRepo 执行记录仓储接口
type ExecutionRepo interface {
	// CreateExecution 创建执行记录
	CreateExecution(ctx context.Context, execution *TaskExecution) (*TaskExecution, error)

	// GetExecution 获取执行记录详情
	GetExecution(ctx context.Context, id int64) (*TaskExecution, error)

	// UpdateExecution 更新执行记录
	UpdateExecution(ctx context.Context, execution *TaskExecution) (*TaskExecution, error)

	// ListExecutions 执行记录列表查询
	ListExecutions(ctx context.Context, filter *ExecutionListFilter) ([]*TaskExecution, int64, error)

	// UpdateExecutionStatus 更新执行状态
	UpdateExecutionStatus(ctx context.Context, id int64, status pb.ExecutionStatus) error
}
