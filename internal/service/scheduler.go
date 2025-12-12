package service

import (
	"context"

	pb "heytom-scheduler/api/scheduler/v1"
	"heytom-scheduler/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SchedulerService 任务调度服务实现
type SchedulerService struct {
	pb.UnimplementedSchedulerServer

	taskUc      *biz.TaskUsecase
	executionUc *biz.ExecutionUsecase
	log         *log.Helper
}

// NewSchedulerService 创建调度服务实例
func NewSchedulerService(taskUc *biz.TaskUsecase, executionUc *biz.ExecutionUsecase, logger log.Logger) *SchedulerService {
	return &SchedulerService{
		taskUc:      taskUc,
		executionUc: executionUc,
		log:         log.NewHelper(logger),
	}
}

// CreateTask 创建任务
func (s *SchedulerService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskReply, error) {
	s.log.WithContext(ctx).Infof("CreateTask: %s", req.Name)

	task, err := s.taskUc.CreateTask(ctx, &biz.Task{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Schedule:    req.Schedule,
		Handler:     req.Handler,
		Payload:     req.Payload,
		Timeout:     req.Timeout,
		Metadata:    req.Metadata,
	})
	if err != nil {
		return nil, err
	}

	return toTaskReply(task), nil
}

// GetTask 获取任务详情
func (s *SchedulerService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskReply, error) {
	task, err := s.taskUc.GetTask(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toTaskReply(task), nil
}

// UpdateTask 更新任务
func (s *SchedulerService) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskReply, error) {
	s.log.WithContext(ctx).Infof("UpdateTask: %d", req.Id)

	task, err := s.taskUc.UpdateTask(ctx, &biz.Task{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Schedule:    req.Schedule,
		Payload:     req.Payload,
		Timeout:     req.Timeout,
		Metadata:    req.Metadata,
	})
	if err != nil {
		return nil, err
	}

	return toTaskReply(task), nil
}

// DeleteTask 删除任务
func (s *SchedulerService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	s.log.WithContext(ctx).Infof("DeleteTask: %d", req.Id)

	if err := s.taskUc.DeleteTask(ctx, req.Id); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ListTasks 任务列表查询
func (s *SchedulerService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksReply, error) {
	tasks, total, err := s.taskUc.ListTasks(ctx, &biz.TaskListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Type:     req.Type,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}

	taskReplies := make([]*pb.TaskReply, 0, len(tasks))
	for _, task := range tasks {
		taskReplies = append(taskReplies, toTaskReply(task))
	}

	return &pb.ListTasksReply{
		Tasks:    taskReplies,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ExecuteTask 立即执行任务
func (s *SchedulerService) ExecuteTask(ctx context.Context, req *pb.ExecuteTaskRequest) (*pb.TaskExecutionReply, error) {
	s.log.WithContext(ctx).Infof("ExecuteTask: %d", req.Id)

	executionID, err := s.taskUc.ExecuteTask(ctx, req.Id, req.Payload)
	if err != nil {
		return nil, err
	}

	return &pb.TaskExecutionReply{
		ExecutionId: executionID,
		Message:     "Task execution started",
	}, nil
}

// PauseTask 暂停任务
func (s *SchedulerService) PauseTask(ctx context.Context, req *pb.PauseTaskRequest) (*pb.TaskReply, error) {
	s.log.WithContext(ctx).Infof("PauseTask: %d", req.Id)

	task, err := s.taskUc.PauseTask(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return toTaskReply(task), nil
}

// ResumeTask 恢复任务
func (s *SchedulerService) ResumeTask(ctx context.Context, req *pb.ResumeTaskRequest) (*pb.TaskReply, error) {
	s.log.WithContext(ctx).Infof("ResumeTask: %d", req.Id)

	task, err := s.taskUc.ResumeTask(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return toTaskReply(task), nil
}

// GetTaskExecutions 获取任务执行历史
func (s *SchedulerService) GetTaskExecutions(ctx context.Context, req *pb.GetTaskExecutionsRequest) (*pb.ListExecutionsReply, error) {
	executions, total, err := s.executionUc.ListExecutions(ctx, &biz.ExecutionListFilter{
		TaskID:   req.TaskId,
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}

	executionReplies := make([]*pb.ExecutionReply, 0, len(executions))
	for _, execution := range executions {
		executionReplies = append(executionReplies, toExecutionReply(execution))
	}

	return &pb.ListExecutionsReply{
		Executions: executionReplies,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetExecution 获取单次执行详情
func (s *SchedulerService) GetExecution(ctx context.Context, req *pb.GetExecutionRequest) (*pb.ExecutionReply, error) {
	execution, err := s.executionUc.GetExecution(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toExecutionReply(execution), nil
}

// CancelExecution 取消执行中的任务
func (s *SchedulerService) CancelExecution(ctx context.Context, req *pb.CancelExecutionRequest) (*pb.ExecutionReply, error) {
	s.log.WithContext(ctx).Infof("CancelExecution: %d", req.Id)

	execution, err := s.executionUc.CancelExecution(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return toExecutionReply(execution), nil
}

// toTaskReply 转换为 TaskReply
func toTaskReply(task *biz.Task) *pb.TaskReply {
	reply := &pb.TaskReply{
		Id:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		Type:           task.Type,
		Status:         task.Status,
		Schedule:       task.Schedule,
		Handler:        task.Handler,
		Payload:        task.Payload,
		Timeout:        task.Timeout,
		Metadata:       task.Metadata,
		CreatedAt:      timestamppb.New(task.CreatedAt),
		UpdatedAt:      timestamppb.New(task.UpdatedAt),
		ExecutionCount: task.ExecutionCount,
		SuccessCount:   task.SuccessCount,
		FailedCount:    task.FailedCount,
	}

	if task.NextRunTime != nil {
		reply.NextRunTime = timestamppb.New(*task.NextRunTime)
	}

	return reply
}

// toExecutionReply 转换为 ExecutionReply
func toExecutionReply(execution *biz.TaskExecution) *pb.ExecutionReply {
	reply := &pb.ExecutionReply{
		Id:         execution.ID,
		TaskId:     execution.TaskID,
		TaskName:   execution.TaskName,
		Status:     execution.Status,
		NodeId:     execution.NodeID,
		Duration:   execution.Duration,
		Result:     execution.Result,
		Error:      execution.Error,
		RetryCount: execution.RetryCount,
		Payload:    execution.Payload,
	}

	if execution.StartTime != nil {
		reply.StartTime = timestamppb.New(*execution.StartTime)
	}
	if execution.EndTime != nil {
		reply.EndTime = timestamppb.New(*execution.EndTime)
	}

	return reply
}
