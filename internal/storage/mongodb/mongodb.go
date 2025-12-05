package mongodb

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBRepository implements TaskRepository using MongoDB
type MongoDBRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// taskDocument is the MongoDB document structure for tasks
type taskDocument struct {
	ID               string                 `bson:"_id"`
	Name             string                 `bson:"name"`
	Description      string                 `bson:"description"`
	ParentID         *string                `bson:"parent_id,omitempty"`
	ExecutionMode    string                 `bson:"execution_mode"`
	ScheduleConfig   map[string]interface{} `bson:"schedule_config,omitempty"`
	CallbackConfig   map[string]interface{} `bson:"callback_config,omitempty"`
	RetryPolicy      map[string]interface{} `bson:"retry_policy,omitempty"`
	ConcurrencyLimit int                    `bson:"concurrency_limit"`
	AlertPolicy      map[string]interface{} `bson:"alert_policy,omitempty"`
	Status           string                 `bson:"status"`
	RetryCount       int                    `bson:"retry_count"`
	NodeID           *string                `bson:"node_id,omitempty"`
	CreatedAt        time.Time              `bson:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at"`
	StartedAt        *time.Time             `bson:"started_at,omitempty"`
	CompletedAt      *time.Time             `bson:"completed_at,omitempty"`
	Metadata         map[string]string      `bson:"metadata,omitempty"`
}

// Config holds MongoDB connection configuration
type Config struct {
	URI            string
	Database       string
	Collection     string
	ConnectTimeout time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
}

// NewMongoDBRepository creates a new MongoDB repository instance
func NewMongoDBRepository(cfg *Config) (*MongoDBRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.Database)
	collection := database.Collection(cfg.Collection)

	// Create indexes
	repo := &MongoDBRepository{
		client:     client,
		database:   database,
		collection: collection,
	}

	if err := repo.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

// createIndexes creates necessary indexes for the collection
func (r *MongoDBRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "parent_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "node_id", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Create creates a new task
func (r *MongoDBRepository) Create(ctx context.Context, task *domain.Task) error {
	doc, err := r.domainToDocument(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to document: %w", err)
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// Get retrieves a task by ID
func (r *MongoDBRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	var doc taskDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task, err := r.documentToDomain(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to task: %w", err)
	}

	return task, nil
}

// Update updates an existing task
func (r *MongoDBRepository) Update(ctx context.Context, task *domain.Task) error {
	doc, err := r.domainToDocument(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to document: %w", err)
	}

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": task.ID}, doc)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return nil
}

// Delete deletes a task by ID
func (r *MongoDBRepository) Delete(ctx context.Context, taskID string) error {
	// Start a session for transaction
	session, err := r.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// Execute transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete all subtasks first
		_, err := r.collection.DeleteMany(sessCtx, bson.M{"parent_id": taskID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete subtasks: %w", err)
		}

		// Delete the task itself
		result, err := r.collection.DeleteOne(sessCtx, bson.M{"_id": taskID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete task: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}

		return nil, nil
	})

	return err
}

// List retrieves tasks matching the filter
func (r *MongoDBRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	query := bson.M{}

	if filter.Status != nil {
		query["status"] = string(*filter.Status)
	}

	if filter.ParentID != nil {
		query["parent_id"] = *filter.ParentID
	}

	if filter.NodeID != nil {
		query["node_id"] = *filter.NodeID
	}

	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		findOptions.SetSkip(int64(filter.Offset))
	}

	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	for cursor.Next(ctx) {
		var doc taskDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		task, err := r.documentToDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return tasks, nil
}

// GetSubTasks retrieves all subtasks of a parent task
func (r *MongoDBRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	for cursor.Next(ctx) {
		var doc taskDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		task, err := r.documentToDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return tasks, nil
}

// domainToDocument converts a domain task to a MongoDB document
func (r *MongoDBRepository) domainToDocument(task *domain.Task) (*taskDocument, error) {
	doc := &taskDocument{
		ID:               task.ID,
		Name:             task.Name,
		Description:      task.Description,
		ParentID:         task.ParentID,
		ExecutionMode:    string(task.ExecutionMode),
		ConcurrencyLimit: task.ConcurrencyLimit,
		Status:           string(task.Status),
		RetryCount:       task.RetryCount,
		NodeID:           task.NodeID,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
		StartedAt:        task.StartedAt,
		CompletedAt:      task.CompletedAt,
		Metadata:         task.Metadata,
	}

	// Convert complex fields to map[string]interface{}
	if task.ScheduleConfig != nil {
		doc.ScheduleConfig = r.scheduleConfigToMap(task.ScheduleConfig)
	}

	if task.CallbackConfig != nil {
		doc.CallbackConfig = r.callbackConfigToMap(task.CallbackConfig)
	}

	if task.RetryPolicy != nil {
		doc.RetryPolicy = r.retryPolicyToMap(task.RetryPolicy)
	}

	if task.AlertPolicy != nil {
		doc.AlertPolicy = r.alertPolicyToMap(task.AlertPolicy)
	}

	return doc, nil
}

// documentToDomain converts a MongoDB document to a domain task
func (r *MongoDBRepository) documentToDomain(doc *taskDocument) (*domain.Task, error) {
	task := &domain.Task{
		ID:               doc.ID,
		Name:             doc.Name,
		Description:      doc.Description,
		ParentID:         doc.ParentID,
		ExecutionMode:    domain.ExecutionMode(doc.ExecutionMode),
		ConcurrencyLimit: doc.ConcurrencyLimit,
		Status:           domain.TaskStatus(doc.Status),
		RetryCount:       doc.RetryCount,
		NodeID:           doc.NodeID,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
		StartedAt:        doc.StartedAt,
		CompletedAt:      doc.CompletedAt,
		Metadata:         doc.Metadata,
	}

	// Convert complex fields from map[string]interface{}
	if doc.ScheduleConfig != nil {
		task.ScheduleConfig = r.mapToScheduleConfig(doc.ScheduleConfig)
	}

	if doc.CallbackConfig != nil {
		task.CallbackConfig = r.mapToCallbackConfig(doc.CallbackConfig)
	}

	if doc.RetryPolicy != nil {
		task.RetryPolicy = r.mapToRetryPolicy(doc.RetryPolicy)
	}

	if doc.AlertPolicy != nil {
		task.AlertPolicy = r.mapToAlertPolicy(doc.AlertPolicy)
	}

	return task, nil
}

// Helper functions for converting complex types
func (r *MongoDBRepository) scheduleConfigToMap(config *domain.ScheduleConfig) map[string]interface{} {
	m := make(map[string]interface{})
	if config.ScheduledTime != nil {
		m["scheduled_time"] = *config.ScheduledTime
	}
	if config.Interval != nil {
		m["interval"] = config.Interval.Nanoseconds()
	}
	if config.CronExpr != nil {
		m["cron_expr"] = *config.CronExpr
	}
	return m
}

func (r *MongoDBRepository) mapToScheduleConfig(m map[string]interface{}) *domain.ScheduleConfig {
	config := &domain.ScheduleConfig{}
	if v, ok := m["scheduled_time"].(time.Time); ok {
		config.ScheduledTime = &v
	}
	if v, ok := m["interval"].(int64); ok {
		interval := time.Duration(v)
		config.Interval = &interval
	}
	if v, ok := m["cron_expr"].(string); ok {
		config.CronExpr = &v
	}
	return config
}

func (r *MongoDBRepository) callbackConfigToMap(config *domain.CallbackConfig) map[string]interface{} {
	m := map[string]interface{}{
		"protocol": string(config.Protocol),
		"url":      config.URL,
		"method":   config.Method,
		"headers":  config.Headers,
		"timeout":  config.Timeout.Nanoseconds(),
		"is_async": config.IsAsync,
	}
	if config.GRPCService != nil {
		m["grpc_service"] = *config.GRPCService
	}
	if config.GRPCMethod != nil {
		m["grpc_method"] = *config.GRPCMethod
	}
	if config.ServiceName != nil {
		m["service_name"] = *config.ServiceName
	}
	m["discovery_type"] = string(config.DiscoveryType)
	return m
}

func (r *MongoDBRepository) mapToCallbackConfig(m map[string]interface{}) *domain.CallbackConfig {
	config := &domain.CallbackConfig{}
	if v, ok := m["protocol"].(string); ok {
		config.Protocol = domain.CallbackProtocol(v)
	}
	if v, ok := m["url"].(string); ok {
		config.URL = v
	}
	if v, ok := m["method"].(string); ok {
		config.Method = v
	}
	if v, ok := m["headers"].(map[string]string); ok {
		config.Headers = v
	}
	if v, ok := m["timeout"].(int64); ok {
		config.Timeout = time.Duration(v)
	}
	if v, ok := m["is_async"].(bool); ok {
		config.IsAsync = v
	}
	if v, ok := m["grpc_service"].(string); ok {
		config.GRPCService = &v
	}
	if v, ok := m["grpc_method"].(string); ok {
		config.GRPCMethod = &v
	}
	if v, ok := m["service_name"].(string); ok {
		config.ServiceName = &v
	}
	if v, ok := m["discovery_type"].(string); ok {
		config.DiscoveryType = domain.ServiceDiscoveryType(v)
	}
	return config
}

func (r *MongoDBRepository) retryPolicyToMap(policy *domain.RetryPolicy) map[string]interface{} {
	return map[string]interface{}{
		"max_retries":    policy.MaxRetries,
		"retry_interval": policy.RetryInterval.Nanoseconds(),
		"backoff_factor": policy.BackoffFactor,
	}
}

func (r *MongoDBRepository) mapToRetryPolicy(m map[string]interface{}) *domain.RetryPolicy {
	policy := &domain.RetryPolicy{}
	if v, ok := m["max_retries"].(int32); ok {
		policy.MaxRetries = int(v)
	} else if v, ok := m["max_retries"].(int64); ok {
		policy.MaxRetries = int(v)
	}
	if v, ok := m["retry_interval"].(int64); ok {
		policy.RetryInterval = time.Duration(v)
	}
	if v, ok := m["backoff_factor"].(float64); ok {
		policy.BackoffFactor = v
	}
	return policy
}

func (r *MongoDBRepository) alertPolicyToMap(policy *domain.AlertPolicy) map[string]interface{} {
	channels := make([]map[string]interface{}, len(policy.Channels))
	for i, ch := range policy.Channels {
		channels[i] = map[string]interface{}{
			"type":   string(ch.Type),
			"config": ch.Config,
		}
	}
	return map[string]interface{}{
		"enable_failure_alert": policy.EnableFailureAlert,
		"retry_threshold":      policy.RetryThreshold,
		"timeout_threshold":    policy.TimeoutThreshold.Nanoseconds(),
		"channels":             channels,
	}
}

func (r *MongoDBRepository) mapToAlertPolicy(m map[string]interface{}) *domain.AlertPolicy {
	policy := &domain.AlertPolicy{}
	if v, ok := m["enable_failure_alert"].(bool); ok {
		policy.EnableFailureAlert = v
	}
	if v, ok := m["retry_threshold"].(int32); ok {
		policy.RetryThreshold = int(v)
	} else if v, ok := m["retry_threshold"].(int64); ok {
		policy.RetryThreshold = int(v)
	}
	if v, ok := m["timeout_threshold"].(int64); ok {
		policy.TimeoutThreshold = time.Duration(v)
	}
	if channels, ok := m["channels"].([]interface{}); ok {
		policy.Channels = make([]domain.NotificationChannel, len(channels))
		for i, ch := range channels {
			if chMap, ok := ch.(map[string]interface{}); ok {
				if t, ok := chMap["type"].(string); ok {
					policy.Channels[i].Type = domain.ChannelType(t)
				}
				if cfg, ok := chMap["config"].(map[string]string); ok {
					policy.Channels[i].Config = cfg
				}
			}
		}
	}
	return policy
}

// Close closes the MongoDB connection
func (r *MongoDBRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
