package biz

import (
	"context"

	pb "heytom-scheduler/api/scheduler/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// ExecutionUsecase 执行记录用例
type ExecutionUsecase struct {
	repo ExecutionRepo
	log  *log.Helper
}

// NewExecutionUsecase 创建执行记录用例实例
func NewExecutionUsecase(repo ExecutionRepo, logger log.Logger) *ExecutionUsecase {
	return &ExecutionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// GetExecution 获取执行记录详情
func (uc *ExecutionUsecase) GetExecution(ctx context.Context, id int64) (*TaskExecution, error) {
	return uc.repo.GetExecution(ctx, id)
}

// ListExecutions 执行记录列表查询
func (uc *ExecutionUsecase) ListExecutions(ctx context.Context, filter *ExecutionListFilter) ([]*TaskExecution, int64, error) {
	return uc.repo.ListExecutions(ctx, filter)
}

// CancelExecution 取消执行中的任务
func (uc *ExecutionUsecase) CancelExecution(ctx context.Context, id int64) (*TaskExecution, error) {
	uc.log.WithContext(ctx).Infof("CancelExecution: %d", id)

	if err := uc.repo.UpdateExecutionStatus(ctx, id, pb.ExecutionStatus_EXECUTION_CANCELLED); err != nil {
		return nil, err
	}

	return uc.repo.GetExecution(ctx, id)
}
