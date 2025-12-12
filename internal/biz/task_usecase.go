package biz

import (
	"context"
	"time"

	pb "heytom-scheduler/api/scheduler/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// TaskUsecase 任务用例
type TaskUsecase struct {
	repo TaskRepo
	log  *log.Helper
}

// NewTaskUsecase 创建任务用例实例
func NewTaskUsecase(repo TaskRepo, logger log.Logger) *TaskUsecase {
	return &TaskUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateTask 创建任务
func (uc *TaskUsecase) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	uc.log.WithContext(ctx).Infof("CreateTask: %s", task.Name)

	// 设置初始状态
	if task.Status == pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		task.Status = pb.TaskStatus_PENDING
	}

	// 计算下次执行时间
	if task.Type == pb.TaskType_CRON || task.Type == pb.TaskType_INTERVAL {
		nextRunTime := calculateNextRunTime(task.Type, task.Schedule)
		task.NextRunTime = &nextRunTime
	} else if task.Type == pb.TaskType_SCHEDULED {
		scheduledTime, err := time.Parse(time.RFC3339, task.Schedule)
		if err == nil {
			task.NextRunTime = &scheduledTime
		}
	}

	return uc.repo.CreateTask(ctx, task)
}

// GetTask 获取任务详情
func (uc *TaskUsecase) GetTask(ctx context.Context, id int64) (*Task, error) {
	return uc.repo.GetTask(ctx, id)
}

// UpdateTask 更新任务
func (uc *TaskUsecase) UpdateTask(ctx context.Context, task *Task) (*Task, error) {
	uc.log.WithContext(ctx).Infof("UpdateTask: %d", task.ID)
	return uc.repo.UpdateTask(ctx, task)
}

// DeleteTask 删除任务
func (uc *TaskUsecase) DeleteTask(ctx context.Context, id int64) error {
	uc.log.WithContext(ctx).Infof("DeleteTask: %d", id)
	return uc.repo.DeleteTask(ctx, id)
}

// ListTasks 任务列表查询
func (uc *TaskUsecase) ListTasks(ctx context.Context, filter *TaskListFilter) ([]*Task, int64, error) {
	return uc.repo.ListTasks(ctx, filter)
}

// ExecuteTask 立即执行任务
func (uc *TaskUsecase) ExecuteTask(ctx context.Context, taskID int64, payload string) (int64, error) {
	uc.log.WithContext(ctx).Infof("ExecuteTask: %d", taskID)

	// 获取任务
	task, err := uc.repo.GetTask(ctx, taskID)
	if err != nil {
		return 0, err
	}

	// TODO: 创建执行记录并提交到执行队列
	// 这里先简单实现，后续会添加调度逻辑

	return task.ID, nil
}

// PauseTask 暂停任务
func (uc *TaskUsecase) PauseTask(ctx context.Context, id int64) (*Task, error) {
	uc.log.WithContext(ctx).Infof("PauseTask: %d", id)

	if err := uc.repo.UpdateTaskStatus(ctx, id, pb.TaskStatus_PAUSED); err != nil {
		return nil, err
	}

	return uc.repo.GetTask(ctx, id)
}

// ResumeTask 恢复任务
func (uc *TaskUsecase) ResumeTask(ctx context.Context, id int64) (*Task, error) {
	uc.log.WithContext(ctx).Infof("ResumeTask: %d", id)

	if err := uc.repo.UpdateTaskStatus(ctx, id, pb.TaskStatus_PENDING); err != nil {
		return nil, err
	}

	return uc.repo.GetTask(ctx, id)
}

// calculateNextRunTime 计算下次执行时间
func calculateNextRunTime(taskType pb.TaskType, schedule string) time.Time {
	// TODO: 实现真实的计算逻辑
	// 这里先简单返回当前时间
	return time.Now()
}
