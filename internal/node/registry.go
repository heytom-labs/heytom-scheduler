package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"task-scheduler/pkg/logger"

	"go.uber.org/zap"
)

// NodeStatus represents the status of a scheduler node
type NodeStatus string

const (
	NodeStatusHealthy   NodeStatus = "healthy"
	NodeStatusUnhealthy NodeStatus = "unhealthy"
	NodeStatusUnknown   NodeStatus = "unknown"
)

// Node represents a scheduler node
type Node struct {
	ID            string
	Address       string
	Status        NodeStatus
	LastHeartbeat time.Time
	Metadata      map[string]string
}

// Registry manages scheduler nodes
type Registry interface {
	// Register registers a node
	Register(ctx context.Context, node *Node) error

	// Deregister removes a node from the registry
	Deregister(ctx context.Context, nodeID string) error

	// UpdateHeartbeat updates the heartbeat timestamp for a node
	UpdateHeartbeat(ctx context.Context, nodeID string) error

	// GetNode retrieves a node by ID
	GetNode(ctx context.Context, nodeID string) (*Node, error)

	// ListNodes lists all registered nodes
	ListNodes(ctx context.Context) ([]*Node, error)

	// ListHealthyNodes lists all healthy nodes
	ListHealthyNodes(ctx context.Context) ([]*Node, error)

	// CheckHealth checks the health of all nodes
	CheckHealth(ctx context.Context) error
}

// inMemoryRegistry is an in-memory implementation of Registry
type inMemoryRegistry struct {
	mu                sync.RWMutex
	nodes             map[string]*Node
	heartbeatTimeout  time.Duration
	healthCheckTicker *time.Ticker
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// NewInMemoryRegistry creates a new in-memory node registry
func NewInMemoryRegistry(heartbeatTimeout time.Duration) Registry {
	ctx, cancel := context.WithCancel(context.Background())

	registry := &inMemoryRegistry{
		nodes:            make(map[string]*Node),
		heartbeatTimeout: heartbeatTimeout,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start health check goroutine
	registry.healthCheckTicker = time.NewTicker(heartbeatTimeout / 2)
	registry.wg.Add(1)
	go registry.healthCheckLoop()

	return registry
}

// Register registers a node
func (r *inMemoryRegistry) Register(ctx context.Context, node *Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	node.Status = NodeStatusHealthy
	node.LastHeartbeat = time.Now()

	r.nodes[node.ID] = node

	logger.Info("Node registered",
		zap.String("nodeID", node.ID),
		zap.String("address", node.Address))

	return nil
}

// Deregister removes a node from the registry
func (r *inMemoryRegistry) Deregister(ctx context.Context, nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	delete(r.nodes, nodeID)

	logger.Info("Node deregistered", zap.String("nodeID", nodeID))

	return nil
}

// UpdateHeartbeat updates the heartbeat timestamp for a node
func (r *inMemoryRegistry) UpdateHeartbeat(ctx context.Context, nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	node.LastHeartbeat = time.Now()
	node.Status = NodeStatusHealthy

	logger.Debug("Node heartbeat updated", zap.String("nodeID", nodeID))

	return nil
}

// GetNode retrieves a node by ID
func (r *inMemoryRegistry) GetNode(ctx context.Context, nodeID string) (*Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	// Return a copy to prevent external modification
	nodeCopy := *node
	return &nodeCopy, nil
}

// ListNodes lists all registered nodes
func (r *inMemoryRegistry) ListNodes(ctx context.Context) ([]*Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes, nil
}

// ListHealthyNodes lists all healthy nodes
func (r *inMemoryRegistry) ListHealthyNodes(ctx context.Context) ([]*Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range r.nodes {
		if node.Status == NodeStatusHealthy {
			nodeCopy := *node
			nodes = append(nodes, &nodeCopy)
		}
	}

	return nodes, nil
}

// CheckHealth checks the health of all nodes
func (r *inMemoryRegistry) CheckHealth(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	for _, node := range r.nodes {
		timeSinceHeartbeat := now.Sub(node.LastHeartbeat)

		if timeSinceHeartbeat > r.heartbeatTimeout {
			if node.Status != NodeStatusUnhealthy {
				node.Status = NodeStatusUnhealthy
				logger.Warn("Node marked as unhealthy",
					zap.String("nodeID", node.ID),
					zap.Duration("timeSinceHeartbeat", timeSinceHeartbeat))
			}
		} else {
			if node.Status != NodeStatusHealthy {
				node.Status = NodeStatusHealthy
				logger.Info("Node marked as healthy",
					zap.String("nodeID", node.ID))
			}
		}
	}

	return nil
}

// healthCheckLoop periodically checks node health
func (r *inMemoryRegistry) healthCheckLoop() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.healthCheckTicker.C:
			if err := r.CheckHealth(r.ctx); err != nil {
				logger.Error("Health check failed", zap.Error(err))
			}
		}
	}
}

// Close stops the registry and cleans up resources
func (r *inMemoryRegistry) Close() error {
	r.cancel()
	r.healthCheckTicker.Stop()
	r.wg.Wait()
	return nil
}
