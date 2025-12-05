package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-scheduler/internal/config"
	"task-scheduler/internal/discovery"
	"task-scheduler/internal/domain"
	"task-scheduler/internal/lock"
	"task-scheduler/internal/node"
	"task-scheduler/internal/notification"
	"task-scheduler/internal/scheduler"
	"task-scheduler/internal/service"
	"task-scheduler/internal/storage/mongodb"
	"task-scheduler/internal/storage/mysql"
	"task-scheduler/internal/storage/postgres"
	"task-scheduler/pkg/logger"

	grpcapi "task-scheduler/internal/api/grpc"
	httpapi "task-scheduler/internal/api/http"

	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// App represents the application
type App struct {
	config *config.Config

	// Storage
	repository domain.TaskRepository

	// Services
	taskService     domain.TaskService
	scheduleService domain.ScheduleService
	callbackService domain.CallbackService
	retryService    *service.RetryService

	// Scheduler
	scheduler domain.Scheduler

	// Distributed components
	distributedLock domain.DistributedLock
	nodeRegistry    node.Registry
	nodeAllocator   *node.Allocator
	failoverManager *node.FailoverManager

	// Service discovery
	serviceDiscovery domain.ServiceDiscovery

	// Notification
	notifier domain.Notifier

	// API servers
	httpServer *httpapi.Server
	grpcServer *grpcapi.Server

	// External clients
	redisClient *redis.Client
	etcdClient  *clientv3.Client
}

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
	}

	if err := app.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize app: %w", err)
	}

	return app, nil
}

// initialize initializes all application components
func (a *App) initialize() error {
	ctx := context.Background()

	// Initialize storage
	if err := a.initStorage(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize Redis client (for distributed lock)
	if err := a.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize Etcd client if needed
	if a.config.Scheduler.LockType == "etcd" || a.config.Discovery.Type == "etcd" {
		if err := a.initEtcd(); err != nil {
			return fmt.Errorf("failed to initialize Etcd: %w", err)
		}
	}

	// Initialize distributed lock
	if err := a.initDistributedLock(); err != nil {
		return fmt.Errorf("failed to initialize distributed lock: %w", err)
	}

	// Initialize service discovery
	if err := a.initServiceDiscovery(); err != nil {
		return fmt.Errorf("failed to initialize service discovery: %w", err)
	}

	// Initialize notification
	if err := a.initNotification(); err != nil {
		return fmt.Errorf("failed to initialize notification: %w", err)
	}

	// Initialize services
	if err := a.initServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize scheduler
	if err := a.initScheduler(); err != nil {
		return fmt.Errorf("failed to initialize scheduler: %w", err)
	}

	// Initialize node components
	if err := a.initNodeComponents(ctx); err != nil {
		return fmt.Errorf("failed to initialize node components: %w", err)
	}

	// Initialize API servers
	if err := a.initAPIServers(); err != nil {
		return fmt.Errorf("failed to initialize API servers: %w", err)
	}

	return nil
}

// initStorage initializes the storage layer
func (a *App) initStorage() error {
	var err error

	switch a.config.Database.Type {
	case "mysql":
		cfg := &mysql.Config{
			Host:            a.config.Database.Host,
			Port:            a.config.Database.Port,
			User:            a.config.Database.Username,
			Password:        a.config.Database.Password,
			Database:        a.config.Database.Database,
			MaxOpenConns:    a.config.Database.MaxOpenConns,
			MaxIdleConns:    a.config.Database.MaxIdleConns,
			ConnMaxLifetime: a.config.Database.ConnMaxLifetime,
		}
		a.repository, err = mysql.NewMySQLRepository(cfg)
	case "postgres":
		cfg := &postgres.Config{
			Host:            a.config.Database.Host,
			Port:            a.config.Database.Port,
			User:            a.config.Database.Username,
			Password:        a.config.Database.Password,
			Database:        a.config.Database.Database,
			MaxOpenConns:    a.config.Database.MaxOpenConns,
			MaxIdleConns:    a.config.Database.MaxIdleConns,
			ConnMaxLifetime: a.config.Database.ConnMaxLifetime,
			SSLMode:         "disable",
		}
		a.repository, err = postgres.NewPostgreSQLRepository(cfg)
	case "mongodb":
		uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
			a.config.Database.Username,
			a.config.Database.Password,
			a.config.Database.Host,
			a.config.Database.Port,
		)
		cfg := &mongodb.Config{
			URI:            uri,
			Database:       a.config.Database.Database,
			Collection:     "tasks",
			ConnectTimeout: 10 * time.Second,
			MaxPoolSize:    100,
			MinPoolSize:    10,
		}
		a.repository, err = mongodb.NewMongoDBRepository(cfg)
	default:
		return fmt.Errorf("unsupported database type: %s", a.config.Database.Type)
	}

	if err != nil {
		return err
	}

	logger.Info("Storage initialized", zap.String("type", a.config.Database.Type))
	return nil
}

// initRedis initializes Redis client
func (a *App) initRedis() error {
	a.redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", a.config.Redis.Host, a.config.Redis.Port),
		Password: a.config.Redis.Password,
		DB:       a.config.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis client initialized")
	return nil
}

// initEtcd initializes Etcd client
func (a *App) initEtcd() error {
	var err error
	a.etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   a.config.Etcd.Endpoints,
		Username:    a.config.Etcd.Username,
		Password:    a.config.Etcd.Password,
		DialTimeout: a.config.Etcd.Timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create Etcd client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := a.etcdClient.Status(ctx, a.config.Etcd.Endpoints[0]); err != nil {
		return fmt.Errorf("failed to connect to Etcd: %w", err)
	}

	logger.Info("Etcd client initialized")
	return nil
}

// initDistributedLock initializes distributed lock
func (a *App) initDistributedLock() error {
	switch a.config.Scheduler.LockType {
	case "redis":
		a.distributedLock = lock.NewRedisLock(a.redisClient, "scheduler:")
	case "etcd":
		if a.etcdClient == nil {
			return fmt.Errorf("Etcd client not initialized")
		}
		var err error
		a.distributedLock, err = lock.NewEtcdLock(a.etcdClient, "scheduler:")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported lock type: %s", a.config.Scheduler.LockType)
	}

	logger.Info("Distributed lock initialized", zap.String("type", a.config.Scheduler.LockType))
	return nil
}

// initServiceDiscovery initializes service discovery
func (a *App) initServiceDiscovery() error {
	var err error

	switch a.config.Discovery.Type {
	case "static":
		a.serviceDiscovery = discovery.NewStaticDiscovery()
	case "consul":
		a.serviceDiscovery, err = discovery.NewConsulDiscovery(a.config.Discovery.Consul.Address)
	case "etcd":
		a.serviceDiscovery, err = discovery.NewEtcdDiscovery(a.config.Discovery.Etcd.Endpoints, "/services")
	case "kubernetes":
		a.serviceDiscovery, err = discovery.NewKubernetesDiscovery(a.config.Discovery.Kubernetes.KubeConfig, a.config.Discovery.Kubernetes.Namespace)
	default:
		// Default to static if not specified
		a.serviceDiscovery = discovery.NewStaticDiscovery()
	}

	if err != nil {
		return err
	}

	logger.Info("Service discovery initialized", zap.String("type", a.config.Discovery.Type))
	return nil
}

// initNotification initializes notification system
func (a *App) initNotification() error {
	notifiers := []domain.Notifier{}

	if a.config.Notification.Email.Enabled {
		emailCfg := notification.EmailConfig{
			SMTPHost: a.config.Notification.Email.SMTPHost,
			SMTPPort: fmt.Sprintf("%d", a.config.Notification.Email.SMTPPort),
			Username: a.config.Notification.Email.Username,
			Password: a.config.Notification.Email.Password,
			From:     a.config.Notification.Email.From,
		}
		emailNotifier := notification.NewEmailNotifier(emailCfg, logger.Get())
		notifiers = append(notifiers, emailNotifier)
	}

	if a.config.Notification.Webhook.Enabled {
		webhookCfg := notification.WebhookConfig{
			URL:     a.config.Notification.Webhook.URL,
			Timeout: a.config.Notification.Webhook.Timeout,
		}
		webhookNotifier := notification.NewWebhookNotifier(webhookCfg, logger.Get())
		notifiers = append(notifiers, webhookNotifier)
	}

	if a.config.Notification.SMS.Enabled {
		smsCfg := notification.SMSConfig{
			APIKey: a.config.Notification.SMS.APIKey,
			APIURL: "", // Provider-specific URL would be set based on provider
			From:   a.config.Notification.SMS.From,
		}
		smsNotifier := notification.NewSMSNotifier(smsCfg, logger.Get())
		notifiers = append(notifiers, smsNotifier)
	}

	if len(notifiers) == 0 {
		// Create a no-op notifier if none are enabled
		a.notifier = notification.NewNoOpNotifier()
	} else if len(notifiers) == 1 {
		a.notifier = notifiers[0]
	} else {
		a.notifier = notification.NewMultiNotifier(notifiers...)
	}

	logger.Info("Notification system initialized", zap.Int("enabled_channels", len(notifiers)))
	return nil
}

// initServices initializes business services
func (a *App) initServices() error {
	// Task service
	a.taskService = service.NewTaskService(a.repository)

	// Schedule service
	a.scheduleService = service.NewScheduleService()

	// Retry service
	a.retryService = service.NewRetryService(a.repository)

	// Callback service - using a simple factory that returns the configured discovery
	discoveryFactory := service.NewSimpleDiscoveryFactory(a.serviceDiscovery)
	a.callbackService = service.NewCallbackService(a.taskService, discoveryFactory, nil)

	logger.Info("Services initialized")
	return nil
}

// initScheduler initializes the scheduler
func (a *App) initScheduler() error {
	schedulerCfg := &scheduler.Config{
		WorkerCount: a.config.Scheduler.WorkerPoolSize,
		QueueSize:   1000,
	}

	a.scheduler = scheduler.NewScheduler(
		schedulerCfg,
		a.repository,
		a.scheduleService,
		a.callbackService,
		a.distributedLock,
		a.config.Scheduler.NodeID,
	)

	// Set notification service (if needed)
	// Note: The scheduler has its own notification handling

	logger.Info("Scheduler initialized", zap.Int("worker_pool_size", a.config.Scheduler.WorkerPoolSize))
	return nil
}

// initNodeComponents initializes node registry and allocator
func (a *App) initNodeComponents(ctx context.Context) error {
	// Node registry
	a.nodeRegistry = node.NewInMemoryRegistry(a.config.Scheduler.HeartbeatInterval * 3)

	// Register this node
	nodeInfo := &node.Node{
		ID:            a.config.Scheduler.NodeID,
		Address:       fmt.Sprintf("localhost:%d", a.config.Server.GRPCPort),
		Status:        node.NodeStatusHealthy,
		LastHeartbeat: time.Now(),
		Metadata:      make(map[string]string),
	}
	if err := a.nodeRegistry.Register(ctx, nodeInfo); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// Node allocator
	a.nodeAllocator = node.NewAllocator(
		a.nodeRegistry,
		a.repository,
		a.distributedLock,
		node.AllocationStrategyRoundRobin,
	)

	// Failover manager
	a.failoverManager = node.NewFailoverManager(
		a.nodeRegistry,
		a.repository,
		a.nodeAllocator,
		a.config.Scheduler.HeartbeatInterval,
	)

	logger.Info("Node components initialized", zap.String("node_id", a.config.Scheduler.NodeID))
	return nil
}

// initAPIServers initializes HTTP and gRPC servers
func (a *App) initAPIServers() error {
	// HTTP server
	a.httpServer = httpapi.NewServer(
		a.config.Server.HTTPPort,
		a.taskService,
		a.scheduleService,
	)

	// gRPC server
	a.grpcServer = grpcapi.NewServer(
		a.config.Server.GRPCPort,
		a.taskService,
		a.scheduleService,
	)

	logger.Info("API servers initialized",
		zap.Int("http_port", a.config.Server.HTTPPort),
		zap.Int("grpc_port", a.config.Server.GRPCPort))
	return nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	logger.Info("Starting Task Scheduler System...")

	// Start scheduler
	if err := a.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start failover manager
	go a.failoverManager.Start()

	// Start heartbeat
	go a.startHeartbeat(ctx)

	// Start HTTP server
	go func() {
		if err := a.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		if err := a.grpcServer.Start(); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	logger.Info("Task Scheduler System started successfully")
	return nil
}

// startHeartbeat starts the heartbeat mechanism
func (a *App) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(a.config.Scheduler.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.nodeRegistry.UpdateHeartbeat(ctx, a.config.Scheduler.NodeID); err != nil {
				logger.Error("Failed to send heartbeat", zap.Error(err))
			}
		}
	}
}

// Stop gracefully stops the application
func (a *App) Stop(ctx context.Context) error {
	logger.Info("Stopping Task Scheduler System...")

	// Stop scheduler
	if err := a.scheduler.Stop(ctx); err != nil {
		logger.Error("Failed to stop scheduler", zap.Error(err))
	}

	// Stop HTTP server
	if err := a.httpServer.Stop(ctx); err != nil {
		logger.Error("Failed to stop HTTP server", zap.Error(err))
	}

	// Stop gRPC server
	if err := a.grpcServer.Stop(ctx); err != nil {
		logger.Error("Failed to stop gRPC server", zap.Error(err))
	}

	// Deregister node
	if err := a.nodeRegistry.Deregister(ctx, a.config.Scheduler.NodeID); err != nil {
		logger.Error("Failed to deregister node", zap.Error(err))
	}

	// Close Redis client
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis client", zap.Error(err))
		}
	}

	// Close Etcd client
	if a.etcdClient != nil {
		if err := a.etcdClient.Close(); err != nil {
			logger.Error("Failed to close Etcd client", zap.Error(err))
		}
	}

	logger.Info("Task Scheduler System stopped")
	return nil
}

// Run runs the application with graceful shutdown
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the application
	if err := a.Start(ctx); err != nil {
		return err
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		a.config.Server.ShutdownTimeout,
	)
	defer shutdownCancel()

	// Stop the application
	return a.Stop(shutdownCtx)
}

// HealthCheck performs a health check
func (a *App) HealthCheck(ctx context.Context) error {
	// Check Redis connection
	if err := a.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	// Check Etcd connection if used
	if a.etcdClient != nil {
		if _, err := a.etcdClient.Status(ctx, a.config.Etcd.Endpoints[0]); err != nil {
			return fmt.Errorf("Etcd health check failed: %w", err)
		}
	}

	// Check if node is registered
	nodes, err := a.nodeRegistry.ListNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeFound := false
	for _, n := range nodes {
		if n.ID == a.config.Scheduler.NodeID {
			nodeFound = true
			break
		}
	}

	if !nodeFound {
		return fmt.Errorf("node not registered")
	}

	return nil
}
