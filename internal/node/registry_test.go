package node

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	ctx := context.Background()

	node := &Node{
		ID:      "node1",
		Address: "localhost:8080",
	}

	err := registry.Register(ctx, node)
	require.NoError(t, err)

	// Verify node was registered
	retrieved, err := registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, "node1", retrieved.ID)
	assert.Equal(t, "localhost:8080", retrieved.Address)
	assert.Equal(t, NodeStatusHealthy, retrieved.Status)
}

func TestRegistry_Deregister(t *testing.T) {
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	ctx := context.Background()

	node := &Node{
		ID:      "node1",
		Address: "localhost:8080",
	}

	err := registry.Register(ctx, node)
	require.NoError(t, err)

	// Deregister
	err = registry.Deregister(ctx, "node1")
	require.NoError(t, err)

	// Verify node was removed
	_, err = registry.GetNode(ctx, "node1")
	assert.Error(t, err)
}

func TestRegistry_UpdateHeartbeat(t *testing.T) {
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	ctx := context.Background()

	node := &Node{
		ID:      "node1",
		Address: "localhost:8080",
	}

	err := registry.Register(ctx, node)
	require.NoError(t, err)

	// Get initial heartbeat
	retrieved1, err := registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	initialHeartbeat := retrieved1.LastHeartbeat

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Update heartbeat
	err = registry.UpdateHeartbeat(ctx, "node1")
	require.NoError(t, err)

	// Verify heartbeat was updated
	retrieved2, err := registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	assert.True(t, retrieved2.LastHeartbeat.After(initialHeartbeat))
}

func TestRegistry_ListHealthyNodes(t *testing.T) {
	registry := NewInMemoryRegistry(30 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	ctx := context.Background()

	// Register multiple nodes
	node1 := &Node{ID: "node1", Address: "localhost:8080"}
	node2 := &Node{ID: "node2", Address: "localhost:8081"}

	registry.Register(ctx, node1)
	registry.Register(ctx, node2)

	// List healthy nodes
	nodes, err := registry.ListHealthyNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, nodes, 2)
}

func TestRegistry_HealthCheck(t *testing.T) {
	// Use short timeout for testing
	registry := NewInMemoryRegistry(1 * time.Second)
	defer registry.(*inMemoryRegistry).Close()

	ctx := context.Background()

	node := &Node{
		ID:      "node1",
		Address: "localhost:8080",
	}

	err := registry.Register(ctx, node)
	require.NoError(t, err)

	// Node should be healthy initially
	retrieved, err := registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, NodeStatusHealthy, retrieved.Status)

	// Wait for heartbeat timeout
	time.Sleep(2 * time.Second)

	// Run health check
	err = registry.CheckHealth(ctx)
	require.NoError(t, err)

	// Node should now be unhealthy
	retrieved, err = registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, NodeStatusUnhealthy, retrieved.Status)

	// Update heartbeat
	err = registry.UpdateHeartbeat(ctx, "node1")
	require.NoError(t, err)

	// Run health check again
	err = registry.CheckHealth(ctx)
	require.NoError(t, err)

	// Node should be healthy again
	retrieved, err = registry.GetNode(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, NodeStatusHealthy, retrieved.Status)
}
