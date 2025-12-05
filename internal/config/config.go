package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Etcd         EtcdConfig         `mapstructure:"etcd"`
	Discovery    DiscoveryConfig    `mapstructure:"discovery"`
	Scheduler    SchedulerConfig    `mapstructure:"scheduler"`
	Notification NotificationConfig `mapstructure:"notification"`
	Log          LogConfig          `mapstructure:"log"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	HTTPPort        int           `mapstructure:"http_port"`
	GRPCPort        int           `mapstructure:"grpc_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type            string        `mapstructure:"type"` // mysql, postgres, mongodb
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// EtcdConfig contains Etcd configuration
type EtcdConfig struct {
	Endpoints []string      `mapstructure:"endpoints"`
	Username  string        `mapstructure:"username"`
	Password  string        `mapstructure:"password"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// DiscoveryConfig contains service discovery configuration
type DiscoveryConfig struct {
	Type       string       `mapstructure:"type"` // static, consul, etcd, kubernetes
	Consul     ConsulConfig `mapstructure:"consul"`
	Etcd       EtcdConfig   `mapstructure:"etcd"`
	Kubernetes K8sConfig    `mapstructure:"kubernetes"`
}

// ConsulConfig contains Consul-specific configuration
type ConsulConfig struct {
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
}

// K8sConfig contains Kubernetes-specific configuration
type K8sConfig struct {
	Namespace  string `mapstructure:"namespace"`
	KubeConfig string `mapstructure:"kubeconfig"`
}

// SchedulerConfig contains scheduler-specific configuration
type SchedulerConfig struct {
	NodeID            string        `mapstructure:"node_id"`
	WorkerPoolSize    int           `mapstructure:"worker_pool_size"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	LockTTL           time.Duration `mapstructure:"lock_ttl"`
	LockType          string        `mapstructure:"lock_type"` // redis, etcd
}

// NotificationConfig contains notification configuration
type NotificationConfig struct {
	Email   EmailConfig   `mapstructure:"email"`
	Webhook WebhookConfig `mapstructure:"webhook"`
	SMS     SMSConfig     `mapstructure:"sms"`
}

// EmailConfig contains email notification configuration
type EmailConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// WebhookConfig contains webhook notification configuration
type WebhookConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// SMSConfig contains SMS notification configuration
type SMSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	From     string `mapstructure:"from"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level"`  // debug, info, warn, error
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
	}
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
	}

	// Validate database config
	if c.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}
	if c.Database.Type != "mysql" && c.Database.Type != "postgres" && c.Database.Type != "mongodb" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port <= 0 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	// Validate scheduler config
	if c.Scheduler.NodeID == "" {
		return fmt.Errorf("scheduler node ID is required")
	}
	if c.Scheduler.WorkerPoolSize <= 0 {
		return fmt.Errorf("worker pool size must be positive")
	}
	if c.Scheduler.LockType != "redis" && c.Scheduler.LockType != "etcd" {
		return fmt.Errorf("unsupported lock type: %s", c.Scheduler.LockType)
	}

	// Validate discovery config
	if c.Discovery.Type != "" {
		if c.Discovery.Type != "static" && c.Discovery.Type != "consul" &&
			c.Discovery.Type != "etcd" && c.Discovery.Type != "kubernetes" {
			return fmt.Errorf("unsupported discovery type: %s", c.Discovery.Type)
		}
	}

	return nil
}
