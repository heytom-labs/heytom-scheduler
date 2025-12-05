package lock

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	return client
}

func TestRedisLock_Lock(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	lock := NewRedisLock(client, "test:")
	ctx := context.Background()

	// Clean up before test
	client.Del(ctx, "test:mykey")

	// Test acquiring lock
	acquired, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Test acquiring same lock again (should fail)
	acquired2, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired2)

	// Clean up
	err = lock.Unlock(ctx, "mykey")
	require.NoError(t, err)
}

func TestRedisLock_Unlock(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	lock := NewRedisLock(client, "test:")
	ctx := context.Background()

	// Clean up before test
	client.Del(ctx, "test:mykey")

	// Acquire lock
	acquired, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	require.True(t, acquired)

	// Unlock
	err = lock.Unlock(ctx, "mykey")
	require.NoError(t, err)

	// Should be able to acquire again
	acquired2, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired2)

	// Clean up
	lock.Unlock(ctx, "mykey")
}

func TestRedisLock_Refresh(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	lock := NewRedisLock(client, "test:")
	ctx := context.Background()

	// Clean up before test
	client.Del(ctx, "test:mykey")

	// Acquire lock with short TTL
	acquired, err := lock.Lock(ctx, "mykey", 2*time.Second)
	require.NoError(t, err)
	require.True(t, acquired)

	// Refresh the lock
	err = lock.Refresh(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)

	// Wait for original TTL to expire
	time.Sleep(3 * time.Second)

	// Lock should still exist due to refresh
	acquired2, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired2, "Lock should still be held after refresh")

	// Clean up
	lock.Unlock(ctx, "mykey")
}

func TestRedisLock_Expiration(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	lock := NewRedisLock(client, "test:")
	ctx := context.Background()

	// Clean up before test
	client.Del(ctx, "test:mykey")

	// Acquire lock with very short TTL
	acquired, err := lock.Lock(ctx, "mykey", 1*time.Second)
	require.NoError(t, err)
	require.True(t, acquired)

	// Wait for lock to expire
	time.Sleep(2 * time.Second)

	// Should be able to acquire again
	acquired2, err := lock.Lock(ctx, "mykey", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired2, "Lock should be acquirable after expiration")

	// Clean up
	lock.Unlock(ctx, "mykey")
}
