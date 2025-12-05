package node

import (
	"context"
	"sync"
	"time"

	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// HeartbeatManager manages heartbeat for a node
type HeartbeatManager struct {
	registry Registry
	nodeID   string
	interval time.Duration

	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewHeartbeatManager creates a new heartbeat manager
func NewHeartbeatManager(registry Registry, nodeID string, interval time.Duration) *HeartbeatManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &HeartbeatManager{
		registry: registry,
		nodeID:   nodeID,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts sending heartbeats
func (h *HeartbeatManager) Start() error {
	h.ticker = time.NewTicker(h.interval)

	h.wg.Add(1)
	go h.heartbeatLoop()

	logger.Info("Heartbeat manager started",
		zap.String("nodeID", h.nodeID),
		zap.Duration("interval", h.interval))

	return nil
}

// Stop stops sending heartbeats
func (h *HeartbeatManager) Stop() error {
	h.cancel()

	if h.ticker != nil {
		h.ticker.Stop()
	}

	h.wg.Wait()

	logger.Info("Heartbeat manager stopped", zap.String("nodeID", h.nodeID))

	return nil
}

// heartbeatLoop sends periodic heartbeats
func (h *HeartbeatManager) heartbeatLoop() {
	defer h.wg.Done()

	// Send initial heartbeat
	if err := h.sendHeartbeat(); err != nil {
		logger.Error("Failed to send initial heartbeat",
			zap.String("nodeID", h.nodeID),
			zap.Error(err))
	}

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.ticker.C:
			if err := h.sendHeartbeat(); err != nil {
				logger.Error("Failed to send heartbeat",
					zap.String("nodeID", h.nodeID),
					zap.Error(err))
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the registry
func (h *HeartbeatManager) sendHeartbeat() error {
	ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
	defer cancel()

	if err := h.registry.UpdateHeartbeat(ctx, h.nodeID); err != nil {
		return err
	}

	logger.Debug("Heartbeat sent", zap.String("nodeID", h.nodeID))

	return nil
}
