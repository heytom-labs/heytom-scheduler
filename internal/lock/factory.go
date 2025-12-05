package lock

import (
	"fmt"

	"task-scheduler/internal/domain"

	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// LockType represents the type of distributed lock
type LockType string

const (
	LockTypeRedis LockType = "redis"
	LockTypeEtcd  LockType = "etcd"
)

// Config contains configuration for distributed lock
type Config struct {
	Type   LockType
	Prefix string

	// Redis configuration
	RedisClient *redis.Client

	// Etcd configuration
	EtcdClient *clientv3.Client
}

// NewDistributedLock creates a new distributed lock based on configuration
func NewDistributedLock(config *Config) (domain.DistributedLock, error) {
	if config == nil {
		return nil, fmt.Errorf("lock config cannot be nil")
	}

	switch config.Type {
	case LockTypeRedis:
		if config.RedisClient == nil {
			return nil, fmt.Errorf("redis client is required for redis lock")
		}
		return NewRedisLock(config.RedisClient, config.Prefix), nil

	case LockTypeEtcd:
		if config.EtcdClient == nil {
			return nil, fmt.Errorf("etcd client is required for etcd lock")
		}
		return NewEtcdLock(config.EtcdClient, config.Prefix)

	default:
		return nil, fmt.Errorf("unsupported lock type: %s", config.Type)
	}
}
