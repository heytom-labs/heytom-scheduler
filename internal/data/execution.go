package data

import (
	"context"

	pb "heytom-scheduler/api/scheduler/v1"
	"heytom-scheduler/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type executionRepo struct {
	data *Data
	log  *log.Helper
}

// NewExecutionRepo 创建执行记录仓储实例
func NewExecutionRepo(data *Data, logger log.Logger) biz.ExecutionRepo {
	return &executionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateExecution 创建执行记录
func (r *executionRepo) CreateExecution(ctx context.Context, execution *biz.TaskExecution) (*biz.TaskExecution, error) {
	dbExecution := &TaskExecution{
		TaskID:     execution.TaskID,
		TaskName:   execution.TaskName,
		Status:     ExecutionStatus(execution.Status),
		NodeID:     execution.NodeID,
		StartTime:  execution.StartTime,
		EndTime:    execution.EndTime,
		Duration:   execution.Duration,
		Result:     execution.Result,
		Error:      execution.Error,
		RetryCount: execution.RetryCount,
		Payload:    execution.Payload,
	}

	if err := r.data.db.WithContext(ctx).Create(dbExecution).Error; err != nil {
		return nil, err
	}

	return r.toBusinessExecution(dbExecution), nil
}

// GetExecution 获取执行记录详情
func (r *executionRepo) GetExecution(ctx context.Context, id int64) (*biz.TaskExecution, error) {
	var execution TaskExecution
	if err := r.data.db.WithContext(ctx).First(&execution, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toBusinessExecution(&execution), nil
}

// UpdateExecution 更新执行记录
func (r *executionRepo) UpdateExecution(ctx context.Context, execution *biz.TaskExecution) (*biz.TaskExecution, error) {
	dbExecution := &TaskExecution{
		ID:         execution.ID,
		Status:     ExecutionStatus(execution.Status),
		EndTime:    execution.EndTime,
		Duration:   execution.Duration,
		Result:     execution.Result,
		Error:      execution.Error,
		RetryCount: execution.RetryCount,
	}

	if err := r.data.db.WithContext(ctx).Model(&TaskExecution{}).Where("id = ?", execution.ID).Updates(dbExecution).Error; err != nil {
		return nil, err
	}

	return r.GetExecution(ctx, execution.ID)
}

// ListExecutions 执行记录列表查询
func (r *executionRepo) ListExecutions(ctx context.Context, filter *biz.ExecutionListFilter) ([]*biz.TaskExecution, int64, error) {
	var executions []TaskExecution
	var total int64

	query := r.data.db.WithContext(ctx).Model(&TaskExecution{})

	// 任务ID筛选
	if filter.TaskID > 0 {
		query = query.Where("task_id = ?", filter.TaskID)
	}

	// 状态筛选
	if filter.Status != pb.ExecutionStatus_EXECUTION_STATUS_UNSPECIFIED {
		query = query.Where("status = ?", filter.Status)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(int(offset)).Limit(int(filter.PageSize)).Order("id DESC").Find(&executions).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.TaskExecution, 0, len(executions))
	for _, execution := range executions {
		result = append(result, r.toBusinessExecution(&execution))
	}

	return result, total, nil
}

// UpdateExecutionStatus 更新执行状态
func (r *executionRepo) UpdateExecutionStatus(ctx context.Context, id int64, status pb.ExecutionStatus) error {
	return r.data.db.WithContext(ctx).Model(&TaskExecution{}).Where("id = ?", id).Update("status", ExecutionStatus(status)).Error
}

// toBusinessExecution 转换为业务模型
func (r *executionRepo) toBusinessExecution(execution *TaskExecution) *biz.TaskExecution {
	return &biz.TaskExecution{
		ID:         execution.ID,
		TaskID:     execution.TaskID,
		TaskName:   execution.TaskName,
		Status:     pb.ExecutionStatus(execution.Status),
		NodeID:     execution.NodeID,
		StartTime:  execution.StartTime,
		EndTime:    execution.EndTime,
		Duration:   execution.Duration,
		Result:     execution.Result,
		Error:      execution.Error,
		RetryCount: execution.RetryCount,
		Payload:    execution.Payload,
		CreatedAt:  execution.CreatedAt,
	}
}
