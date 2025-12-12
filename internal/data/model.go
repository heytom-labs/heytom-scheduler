package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	pb "heytom-scheduler/api/scheduler/v1"
)

// Metadata 元数据类型（JSON存储）
type Metadata map[string]string

// Scan 实现 sql.Scanner 接口
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现 driver.Valuer 接口
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// TaskType 任务类型（数据库存储为字符串）
type TaskType pb.TaskType

// Scan 实现 sql.Scanner 接口
func (t *TaskType) Scan(value interface{}) error {
	if value == nil {
		*t = TaskType(pb.TaskType_TASK_TYPE_UNSPECIFIED)
		return nil
	}
	str, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TaskType")
	}
	*t = TaskType(parseTaskTypeFromString(string(str)))
	return nil
}

// Value 实现 driver.Valuer 接口
func (t TaskType) Value() (driver.Value, error) {
	return pb.TaskType(t).String(), nil
}

// TaskStatus 任务状态（数据库存储为字符串）
type TaskStatus pb.TaskStatus

// Scan 实现 sql.Scanner 接口
func (s *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		*s = TaskStatus(pb.TaskStatus_TASK_STATUS_UNSPECIFIED)
		return nil
	}
	str, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TaskStatus")
	}
	*s = TaskStatus(parseTaskStatusFromString(string(str)))
	return nil
}

// Value 实现 driver.Valuer 接口
func (s TaskStatus) Value() (driver.Value, error) {
	return pb.TaskStatus(s).String(), nil
}

// ExecutionStatus 执行状态（数据库存储为字符串）
type ExecutionStatus pb.ExecutionStatus

// Scan 实现 sql.Scanner 接口
func (s *ExecutionStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ExecutionStatus(pb.ExecutionStatus_EXECUTION_STATUS_UNSPECIFIED)
		return nil
	}
	str, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ExecutionStatus")
	}
	*s = ExecutionStatus(parseExecutionStatusFromString(string(str)))
	return nil
}

// Value 实现 driver.Valuer 接口
func (s ExecutionStatus) Value() (driver.Value, error) {
	return pb.ExecutionStatus(s).String(), nil
}

// parseTaskTypeFromString 从字符串解析任务类型
func parseTaskTypeFromString(s string) pb.TaskType {
	switch s {
	case "IMMEDIATE":
		return pb.TaskType_IMMEDIATE
	case "SCHEDULED":
		return pb.TaskType_SCHEDULED
	case "CRON":
		return pb.TaskType_CRON
	case "INTERVAL":
		return pb.TaskType_INTERVAL
	default:
		return pb.TaskType_TASK_TYPE_UNSPECIFIED
	}
}

// parseTaskStatusFromString 从字符串解析任务状态
func parseTaskStatusFromString(s string) pb.TaskStatus {
	switch s {
	case "PENDING":
		return pb.TaskStatus_PENDING
	case "RUNNING":
		return pb.TaskStatus_RUNNING
	case "PAUSED":
		return pb.TaskStatus_PAUSED
	case "COMPLETED":
		return pb.TaskStatus_COMPLETED
	case "FAILED":
		return pb.TaskStatus_FAILED
	case "CANCELLED":
		return pb.TaskStatus_CANCELLED
	default:
		return pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

// parseExecutionStatusFromString 从字符串解析执行状态
func parseExecutionStatusFromString(s string) pb.ExecutionStatus {
	switch s {
	case "QUEUED":
		return pb.ExecutionStatus_QUEUED
	case "EXECUTING":
		return pb.ExecutionStatus_EXECUTING
	case "SUCCESS":
		return pb.ExecutionStatus_SUCCESS
	case "EXECUTION_FAILED":
		return pb.ExecutionStatus_EXECUTION_FAILED
	case "TIMEOUT":
		return pb.ExecutionStatus_TIMEOUT
	case "EXECUTION_CANCELLED":
		return pb.ExecutionStatus_EXECUTION_CANCELLED
	default:
		return pb.ExecutionStatus_EXECUTION_STATUS_UNSPECIFIED
	}
}

// Task 任务模型
type Task struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"`
	Name           string     `gorm:"type:varchar(255);not null;index"`
	Description    string     `gorm:"type:text"`
	Type           TaskType   `gorm:"type:varchar(20);not null;index"`
	Status         TaskStatus `gorm:"type:varchar(20);not null;index;default:'PENDING'"`
	Schedule       string     `gorm:"type:varchar(255)"` // Cron表达式、时间戳或间隔秒数
	Handler        string     `gorm:"type:varchar(255);not null"`
	Payload        string     `gorm:"type:text"` // JSON格式
	Timeout        int32      `gorm:"type:int;default:300"`
	Metadata       Metadata   `gorm:"type:json"`
	NextRunTime    *time.Time `gorm:"type:datetime;index"`
	ExecutionCount int64      `gorm:"type:bigint;default:0"`
	SuccessCount   int64      `gorm:"type:bigint;default:0"`
	FailedCount    int64      `gorm:"type:bigint;default:0"`
	CreatedAt      time.Time  `gorm:"type:datetime;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"type:datetime;not null;autoUpdateTime"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// TaskExecution 任务执行记录模型
type TaskExecution struct {
	ID         int64           `gorm:"primaryKey;autoIncrement"`
	TaskID     int64           `gorm:"type:bigint;not null;index"`
	TaskName   string          `gorm:"type:varchar(255);not null"`
	Status     ExecutionStatus `gorm:"type:varchar(20);not null;index"`
	NodeID     string          `gorm:"type:varchar(100);index"` // 执行节点ID
	StartTime  *time.Time      `gorm:"type:datetime"`
	EndTime    *time.Time      `gorm:"type:datetime"`
	Duration   int32           `gorm:"type:int"` // 执行耗时（毫秒）
	Result     string          `gorm:"type:text"`
	Error      string          `gorm:"type:text"`
	RetryCount int32           `gorm:"type:int;default:0"`
	Payload    string          `gorm:"type:text"` // JSON格式
	CreatedAt  time.Time       `gorm:"type:datetime;not null;autoCreateTime"`
}

// TableName 指定表名
func (TaskExecution) TableName() string {
	return "task_executions"
}
