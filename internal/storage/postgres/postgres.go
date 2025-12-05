package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLRepository implements TaskRepository using PostgreSQL
type PostgreSQLRepository struct {
	db *gorm.DB
}

// taskModel is the database model for tasks
type taskModel struct {
	ID               string `gorm:"primaryKey;size:255"`
	Name             string `gorm:"size:255;not null"`
	Description      string `gorm:"type:text"`
	ParentID         sql.NullString
	ExecutionMode    string `gorm:"size:50;not null"`
	ScheduleConfig   string `gorm:"type:text"`
	CallbackConfig   string `gorm:"type:text"`
	RetryPolicy      string `gorm:"type:text"`
	ConcurrencyLimit int
	AlertPolicy      string `gorm:"type:text"`
	Status           string `gorm:"size:50;not null;index"`
	RetryCount       int
	NodeID           sql.NullString `gorm:"index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	StartedAt        sql.NullTime
	CompletedAt      sql.NullTime
	Metadata         string `gorm:"type:text"`
}

// TableName specifies the table name for taskModel
func (taskModel) TableName() string {
	return "tasks"
}

// Config holds PostgreSQL connection configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

// NewPostgreSQLRepository creates a new PostgreSQL repository instance
func NewPostgreSQLRepository(cfg *Config) (*PostgreSQLRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Auto-migrate the schema
	if err := db.AutoMigrate(&taskModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &PostgreSQLRepository{db: db}, nil
}

// Create creates a new task
func (r *PostgreSQLRepository) Create(ctx context.Context, task *domain.Task) error {
	model, err := r.domainToModel(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// Get retrieves a task by ID
func (r *PostgreSQLRepository) Get(ctx context.Context, taskID string) (*domain.Task, error) {
	var model taskModel
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task, err := r.modelToDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to task: %w", err)
	}

	return task, nil
}

// Update updates an existing task
func (r *PostgreSQLRepository) Update(ctx context.Context, task *domain.Task) error {
	model, err := r.domainToModel(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to model: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&taskModel{}).Where("id = ?", task.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return nil
}

// Delete deletes a task by ID
func (r *PostgreSQLRepository) Delete(ctx context.Context, taskID string) error {
	// Start a transaction to handle cascade delete
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all subtasks first
		if err := tx.Where("parent_id = ?", taskID).Delete(&taskModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete subtasks: %w", err)
		}

		// Delete the task itself
		result := tx.Where("id = ?", taskID).Delete(&taskModel{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete task: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("task not found: %s", taskID)
		}

		return nil
	})
}

// List retrieves tasks matching the filter
func (r *PostgreSQLRepository) List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	query := r.db.WithContext(ctx).Model(&taskModel{})

	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}

	if filter.ParentID != nil {
		query = query.Where("parent_id = ?", *filter.ParentID)
	}

	if filter.NodeID != nil {
		query = query.Where("node_id = ?", *filter.NodeID)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var models []taskModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*domain.Task, 0, len(models))
	for _, model := range models {
		task, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetSubTasks retrieves all subtasks of a parent task
func (r *PostgreSQLRepository) GetSubTasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	var models []taskModel
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get subtasks: %w", err)
	}

	tasks := make([]*domain.Task, 0, len(models))
	for _, model := range models {
		task, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// domainToModel converts a domain task to a database model
func (r *PostgreSQLRepository) domainToModel(task *domain.Task) (*taskModel, error) {
	model := &taskModel{
		ID:               task.ID,
		Name:             task.Name,
		Description:      task.Description,
		ExecutionMode:    string(task.ExecutionMode),
		ConcurrencyLimit: task.ConcurrencyLimit,
		Status:           string(task.Status),
		RetryCount:       task.RetryCount,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}

	if task.ParentID != nil {
		model.ParentID = sql.NullString{String: *task.ParentID, Valid: true}
	}

	if task.NodeID != nil {
		model.NodeID = sql.NullString{String: *task.NodeID, Valid: true}
	}

	if task.StartedAt != nil {
		model.StartedAt = sql.NullTime{Time: *task.StartedAt, Valid: true}
	}

	if task.CompletedAt != nil {
		model.CompletedAt = sql.NullTime{Time: *task.CompletedAt, Valid: true}
	}

	// Serialize complex fields to JSON
	if task.ScheduleConfig != nil {
		data, err := json.Marshal(task.ScheduleConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schedule config: %w", err)
		}
		model.ScheduleConfig = string(data)
	}

	if task.CallbackConfig != nil {
		data, err := json.Marshal(task.CallbackConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal callback config: %w", err)
		}
		model.CallbackConfig = string(data)
	}

	if task.RetryPolicy != nil {
		data, err := json.Marshal(task.RetryPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal retry policy: %w", err)
		}
		model.RetryPolicy = string(data)
	}

	if task.AlertPolicy != nil {
		data, err := json.Marshal(task.AlertPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal alert policy: %w", err)
		}
		model.AlertPolicy = string(data)
	}

	if task.Metadata != nil {
		data, err := json.Marshal(task.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		model.Metadata = string(data)
	}

	return model, nil
}

// modelToDomain converts a database model to a domain task
func (r *PostgreSQLRepository) modelToDomain(model *taskModel) (*domain.Task, error) {
	task := &domain.Task{
		ID:               model.ID,
		Name:             model.Name,
		Description:      model.Description,
		ExecutionMode:    domain.ExecutionMode(model.ExecutionMode),
		ConcurrencyLimit: model.ConcurrencyLimit,
		Status:           domain.TaskStatus(model.Status),
		RetryCount:       model.RetryCount,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}

	if model.ParentID.Valid {
		task.ParentID = &model.ParentID.String
	}

	if model.NodeID.Valid {
		task.NodeID = &model.NodeID.String
	}

	if model.StartedAt.Valid {
		task.StartedAt = &model.StartedAt.Time
	}

	if model.CompletedAt.Valid {
		task.CompletedAt = &model.CompletedAt.Time
	}

	// Deserialize complex fields from JSON
	if model.ScheduleConfig != "" {
		var config domain.ScheduleConfig
		if err := json.Unmarshal([]byte(model.ScheduleConfig), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schedule config: %w", err)
		}
		task.ScheduleConfig = &config
	}

	if model.CallbackConfig != "" {
		var config domain.CallbackConfig
		if err := json.Unmarshal([]byte(model.CallbackConfig), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal callback config: %w", err)
		}
		task.CallbackConfig = &config
	}

	if model.RetryPolicy != "" {
		var policy domain.RetryPolicy
		if err := json.Unmarshal([]byte(model.RetryPolicy), &policy); err != nil {
			return nil, fmt.Errorf("failed to unmarshal retry policy: %w", err)
		}
		task.RetryPolicy = &policy
	}

	if model.AlertPolicy != "" {
		var policy domain.AlertPolicy
		if err := json.Unmarshal([]byte(model.AlertPolicy), &policy); err != nil {
			return nil, fmt.Errorf("failed to unmarshal alert policy: %w", err)
		}
		task.AlertPolicy = &policy
	}

	if model.Metadata != "" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		task.Metadata = metadata
	}

	return task, nil
}

// Close closes the database connection
func (r *PostgreSQLRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
