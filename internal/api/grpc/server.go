package grpc

import (
	"context"
	"fmt"
	"net"

	"task-scheduler/api/pb"
	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer *grpc.Server
	handler    *Handler
	port       int
	listener   net.Listener
}

// NewServer creates a new gRPC server
func NewServer(port int, taskService domain.TaskService, scheduleService domain.ScheduleService) *Server {
	handler := NewHandler(taskService)
	grpcServer := grpc.NewServer()

	// Register the service
	pb.RegisterTaskSchedulerServiceServer(grpcServer, handler)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		handler:    handler,
		port:       port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	logger.Info("gRPC server starting", zap.String("address", addr))

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("gRPC server stopping")

	// Use graceful stop with context
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		// Force stop if context times out
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		// Graceful stop completed
	}

	if s.listener != nil {
		s.listener.Close()
	}

	return nil
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	return s.port
}
