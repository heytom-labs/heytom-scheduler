package lock

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"
	"task-scheduler/pkg/metrics"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// redisLock implements distributed lock using Redis
type redisLock struct {
	client *redis.Client
	prefix string
}

// NewRedisLock creates a new Redis-based distributed lock
func NewRedisLock(client *redis.Client, prefix string) domain.DistributedLock {
	if prefix == "" {
		prefix = "lock:"
	}
	return &redisLock{
		client: client,
		prefix: prefix,
	}
}

// Lock acquires a distributed lock
func (r *redisLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	startTime := time.Now()
	fullKey := r.prefix + key

	logger.DebugContext(ctx, "Attempting to acquire lock",
		zap.String("key", key),
		zap.Duration("ttl", ttl))

	// Use SET with NX (only set if not exists) and EX (expiration)
	result, err := r.client.SetNX(ctx, fullKey, "locked", ttl).Result()

	duration := time.Since(startTime).Seconds()
	metrics.RecordLockAcquisitionDuration(duration)

	if err != nil {
		logger.ErrorContext(ctx, "Failed to acquire lock",
			zap.String("key", key),
			zap.Error(err))
		metrics.RecordLockAcquisition("error")
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result {
		logger.DebugContext(ctx, "Lock acquired",
			zap.String("key", key),
			zap.Duration("duration", time.Since(startTime)))
		metrics.RecordLockAcquisition("success")
	} else {
		logger.DebugContext(ctx, "Lock already held",
			zap.String("key", key))
		metrics.RecordLockAcquisition("already_held")
	}

	return result, nil
}

// Unlock releases a distributed lock
func (r *redisLock) Unlock(ctx context.Context, key string) error {
	fullKey := r.prefix + key

	result, err := r.client.Del(ctx, fullKey).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result == 0 {
		// Key didn't exist, but this is not necessarily an error
		// The lock might have expired
		return nil
	}

	return nil
}

// Refresh refreshes the TTL of a lock
func (r *redisLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.prefix + key

	// Check if key exists first
	exists, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check lock existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("lock does not exist")
	}

	// Refresh the TTL
	result, err := r.client.Expire(ctx, fullKey, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to refresh lock TTL: %w", err)
	}

	if !result {
		return fmt.Errorf("failed to refresh lock TTL: key does not exist")
	}

	return nil
}
