package data

import (
	"context"
	"time"

	pb "heytom-scheduler/api/scheduler/v1"
	"heytom-scheduler/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type taskRepo struct {
	data *Data
	log  *log.Helper
}

// NewTaskRepo 创建任务仓储实例
func NewTaskRepo(data *Data, logger log.Logger) biz.TaskRepo {
	return &taskRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateTask 创建任务
func (r *taskRepo) CreateTask(ctx context.Context, task *biz.Task) (*biz.Task, error) {
	dbTask := &Task{
		Name:        task.Name,
		Description: task.Description,
		Type:        TaskType(task.Type),
		Status:      TaskStatus(task.Status),
		Schedule:    task.Schedule,
		Handler:     task.Handler,
		Payload:     task.Payload,
		Timeout:     task.Timeout,
		Metadata:    task.Metadata,
		NextRunTime: task.NextRunTime,
	}

	if err := r.data.db.WithContext(ctx).Create(dbTask).Error; err != nil {
		return nil, err
	}

	return r.toBusinessTask(dbTask), nil
}

// GetTask 获取任务详情
func (r *taskRepo) GetTask(ctx context.Context, id int64) (*biz.Task, error) {
	var task Task
	if err := r.data.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toBusinessTask(&task), nil
}

// UpdateTask 更新任务
func (r *taskRepo) UpdateTask(ctx context.Context, task *biz.Task) (*biz.Task, error) {
	dbTask := &Task{
		ID:          task.ID,
		Name:        task.Name,
		Description: task.Description,
		Schedule:    task.Schedule,
		Payload:     task.Payload,
		Timeout:     task.Timeout,
		Metadata:    task.Metadata,
	}

	if err := r.data.db.WithContext(ctx).Model(&Task{}).Where("id = ?", task.ID).Updates(dbTask).Error; err != nil {
		return nil, err
	}

	return r.GetTask(ctx, task.ID)
}

// DeleteTask 删除任务
func (r *taskRepo) DeleteTask(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&Task{}, id).Error
}

// ListTasks 任务列表查询
func (r *taskRepo) ListTasks(ctx context.Context, filter *biz.TaskListFilter) ([]*biz.Task, int64, error) {
	var tasks []Task
	var total int64

	query := r.data.db.WithContext(ctx).Model(&Task{})

	// 状态筛选
	if filter.Status != pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		query = query.Where("status = ?", filter.Status)
	}

	// 类型筛选
	if filter.Type != pb.TaskType_TASK_TYPE_UNSPECIFIED {
		query = query.Where("type = ?", filter.Type)
	}

	// 关键词搜索
	if filter.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(int(offset)).Limit(int(filter.PageSize)).Order("id DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, r.toBusinessTask(&task))
	}

	return result, total, nil
}

// UpdateTaskStatus 更新任务状态
func (r *taskRepo) UpdateTaskStatus(ctx context.Context, id int64, status pb.TaskStatus) error {
	return r.data.db.WithContext(ctx).Model(&Task{}).Where("id = ?", id).Update("status", TaskStatus(status)).Error
}

// UpdateTaskNextRunTime 更新任务下次执行时间
func (r *taskRepo) UpdateTaskNextRunTime(ctx context.Context, id int64, nextRunTime time.Time) error {
	return r.data.db.WithContext(ctx).Model(&Task{}).Where("id = ?", id).Update("next_run_time", nextRunTime).Error
}

// IncrementExecutionCount 增加执行次数
func (r *taskRepo) IncrementExecutionCount(ctx context.Context, id int64, success bool) error {
	updates := map[string]interface{}{
		"execution_count": gorm.Expr("execution_count + ?", 1),
	}
	if success {
		updates["success_count"] = gorm.Expr("success_count + ?", 1)
	} else {
		updates["failed_count"] = gorm.Expr("failed_count + ?", 1)
	}
	return r.data.db.WithContext(ctx).Model(&Task{}).Where("id = ?", id).Updates(updates).Error
}

// toBusinessTask 转换为业务模型
func (r *taskRepo) toBusinessTask(task *Task) *biz.Task {
	return &biz.Task{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		Type:           pb.TaskType(task.Type),
		Status:         pb.TaskStatus(task.Status),
		Schedule:       task.Schedule,
		Handler:        task.Handler,
		Payload:        task.Payload,
		Timeout:        task.Timeout,
		Metadata:       task.Metadata,
		NextRunTime:    task.NextRunTime,
		ExecutionCount: task.ExecutionCount,
		SuccessCount:   task.SuccessCount,
		FailedCount:    task.FailedCount,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}
