package main

import (
	"flag"
	"fmt"
	"os"

	"task-scheduler/internal/app"
	"task-scheduler/internal/config"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger based on config
	if err := logger.InitWithFormat(cfg.Log.Format, cfg.Log.Level); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Task Scheduler System starting...",
		zap.String("node_id", cfg.Scheduler.NodeID),
		zap.Int("http_port", cfg.Server.HTTPPort),
		zap.Int("grpc_port", cfg.Server.GRPCPort),
	)

	// Create and run application
	application, err := app.New(cfg)
	if err != nil {
		logger.Error("Failed to create application", zap.Error(err))
		os.Exit(1)
	}

	// Run application with graceful shutdown
	if err := application.Run(); err != nil {
		logger.Error("Application error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Task Scheduler System shutdown complete")
}
