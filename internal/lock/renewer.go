package lock

import (
	"context"
	"sync"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// Renewer handles automatic lock renewal
type Renewer struct {
	lock     domain.DistributedLock
	interval time.Duration
	ttl      time.Duration

	mu     sync.RWMutex
	locks  map[string]*lockRenewal
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// lockRenewal represents a lock that needs renewal
type lockRenewal struct {
	key       string
	stopChan  chan struct{}
	stoppedCh chan struct{}
}

// NewRenewer creates a new lock renewer
func NewRenewer(lock domain.DistributedLock, interval, ttl time.Duration) *Renewer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Renewer{
		lock:     lock,
		interval: interval,
		ttl:      ttl,
		locks:    make(map[string]*lockRenewal),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// StartRenewal starts automatic renewal for a lock
func (r *Renewer) StartRenewal(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already renewing
	if _, exists := r.locks[key]; exists {
		return
	}

	renewal := &lockRenewal{
		key:       key,
		stopChan:  make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}

	r.locks[key] = renewal

	r.wg.Add(1)
	go r.renewLoop(renewal)
}

// StopRenewal stops automatic renewal for a lock
func (r *Renewer) StopRenewal(key string) {
	r.mu.Lock()
	renewal, exists := r.locks[key]
	if !exists {
		r.mu.Unlock()
		return
	}
	delete(r.locks, key)
	r.mu.Unlock()

	// Signal stop
	close(renewal.stopChan)

	// Wait for goroutine to finish
	<-renewal.stoppedCh
}

// StopAll stops all lock renewals
func (r *Renewer) StopAll() {
	r.cancel()
	r.wg.Wait()
}

// renewLoop continuously renews a lock
func (r *Renewer) renewLoop(renewal *lockRenewal) {
	defer r.wg.Done()
	defer close(renewal.stoppedCh)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	logger.Debug("Started lock renewal",
		zap.String("key", renewal.key),
		zap.Duration("interval", r.interval))

	for {
		select {
		case <-r.ctx.Done():
			logger.Debug("Lock renewal stopped (context cancelled)",
				zap.String("key", renewal.key))
			return

		case <-renewal.stopChan:
			logger.Debug("Lock renewal stopped",
				zap.String("key", renewal.key))
			return

		case <-ticker.C:
			if err := r.lock.Refresh(r.ctx, renewal.key, r.ttl); err != nil {
				logger.Warn("Failed to refresh lock",
					zap.String("key", renewal.key),
					zap.Error(err))
				// Continue trying to refresh
			} else {
				logger.Debug("Lock refreshed",
					zap.String("key", renewal.key))
			}
		}
	}
}
