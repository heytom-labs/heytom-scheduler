package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config.yaml in current directory and /etc/task-scheduler/
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/task-scheduler/")
		v.AddConfigPath("$HOME/.task-scheduler/")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional, only return error if file was explicitly specified
		if configPath != "" {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// If no config file specified and not found, continue with defaults and env vars
		fmt.Fprintf(os.Stderr, "Warning: No config file found, using defaults and environment variables\n")
	}

	// Enable environment variable override
	v.SetEnvPrefix("SCHEDULER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "30s")

	// Database defaults
	v.SetDefault("database.type", "mysql")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.username", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "task_scheduler")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "1h")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Etcd defaults
	v.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	v.SetDefault("etcd.username", "")
	v.SetDefault("etcd.password", "")
	v.SetDefault("etcd.timeout", "5s")

	// Discovery defaults
	v.SetDefault("discovery.type", "static")
	v.SetDefault("discovery.consul.address", "localhost:8500")
	v.SetDefault("discovery.consul.token", "")
	v.SetDefault("discovery.etcd.endpoints", []string{"localhost:2379"})
	v.SetDefault("discovery.kubernetes.namespace", "default")
	v.SetDefault("discovery.kubernetes.kubeconfig", "")

	// Scheduler defaults
	v.SetDefault("scheduler.node_id", "node-1")
	v.SetDefault("scheduler.worker_pool_size", 10)
	v.SetDefault("scheduler.heartbeat_interval", "10s")
	v.SetDefault("scheduler.lock_ttl", "30s")
	v.SetDefault("scheduler.lock_type", "redis")

	// Notification defaults
	v.SetDefault("notification.email.enabled", false)
	v.SetDefault("notification.email.smtp_host", "")
	v.SetDefault("notification.email.smtp_port", 587)
	v.SetDefault("notification.email.username", "")
	v.SetDefault("notification.email.password", "")
	v.SetDefault("notification.email.from", "")
	v.SetDefault("notification.webhook.enabled", false)
	v.SetDefault("notification.webhook.url", "")
	v.SetDefault("notification.webhook.timeout", "10s")
	v.SetDefault("notification.sms.enabled", false)
	v.SetDefault("notification.sms.provider", "")
	v.SetDefault("notification.sms.api_key", "")
	v.SetDefault("notification.sms.from", "")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_path", "stdout")
}
