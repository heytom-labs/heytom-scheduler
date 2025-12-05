package lock

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// etcdLock implements distributed lock using Etcd
type etcdLock struct {
	client  *clientv3.Client
	prefix  string
	session *concurrency.Session
	mutexes map[string]*concurrency.Mutex
}

// NewEtcdLock creates a new Etcd-based distributed lock
func NewEtcdLock(client *clientv3.Client, prefix string) (domain.DistributedLock, error) {
	if prefix == "" {
		prefix = "/lock/"
	}

	// Create a session with default TTL
	session, err := concurrency.NewSession(client, concurrency.WithTTL(60))
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd session: %w", err)
	}

	return &etcdLock{
		client:  client,
		prefix:  prefix,
		session: session,
		mutexes: make(map[string]*concurrency.Mutex),
	}, nil
}

// Lock acquires a distributed lock
func (e *etcdLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := e.prefix + key

	// Create a new session with the specified TTL
	session, err := concurrency.NewSession(e.client, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}

	// Create a mutex for this key
	mutex := concurrency.NewMutex(session, fullKey)

	// Try to acquire the lock with context timeout
	lockCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err = mutex.Lock(lockCtx)
	if err != nil {
		if err == context.DeadlineExceeded {
			// Lock is held by another process
			session.Close()
			return false, nil
		}
		session.Close()
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Store the mutex for later unlock
	e.mutexes[key] = mutex

	return true, nil
}

// Unlock releases a distributed lock
func (e *etcdLock) Unlock(ctx context.Context, key string) error {
	mutex, exists := e.mutexes[key]
	if !exists {
		// Lock doesn't exist, might have expired
		return nil
	}

	// Unlock the mutex
	err := mutex.Unlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	// Clean up
	delete(e.mutexes, key)

	return nil
}

// Refresh refreshes the TTL of a lock
func (e *etcdLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	mutex, exists := e.mutexes[key]
	if !exists {
		return fmt.Errorf("lock does not exist")
	}

	// Get the session from the mutex
	// Note: etcd's concurrency.Mutex doesn't directly support TTL refresh
	// We need to keep the session alive
	// The session's keepalive mechanism handles this automatically

	// Verify the lock still exists
	fullKey := e.prefix + key
	resp, err := e.client.Get(ctx, fullKey)
	if err != nil {
		return fmt.Errorf("failed to check lock existence: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("lock does not exist")
	}

	// The session's keepalive will automatically refresh the TTL
	// We just need to verify the mutex is still valid
	_ = mutex

	return nil
}

// Close closes the etcd lock and releases all resources
func (e *etcdLock) Close() error {
	// Unlock all remaining mutexes
	ctx := context.Background()
	for key := range e.mutexes {
		e.Unlock(ctx, key)
	}

	// Close the session
	if e.session != nil {
		return e.session.Close()
	}

	return nil
}
