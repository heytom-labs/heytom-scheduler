# Multi-Node Task Allocation

## Overview

The task scheduler system supports multi-node deployment for high availability and load distribution. This document describes how tasks are allocated across multiple scheduler nodes.

## Components

### 1. Node Registry

The `Registry` interface manages scheduler nodes in the cluster:

- **Register**: Registers a new node in the cluster
- **Deregister**: Removes a node from the cluster
- **UpdateHeartbeat**: Updates the last heartbeat timestamp for a node
- **GetNode**: Retrieves node information by ID
- **ListNodes**: Lists all registered nodes
- **ListHealthyNodes**: Lists only healthy nodes
- **CheckHealth**: Checks the health status of all nodes

The in-memory implementation (`inMemoryRegistry`) automatically monitors node health by checking heartbeat timestamps against a configurable timeout.

### 2. Heartbeat Manager

The `HeartbeatManager` sends periodic heartbeats to the registry to indicate that a node is alive and healthy:

```go
hb := NewHeartbeatManager(registry, nodeID, 10*time.Second)
hb.Start()
defer hb.Stop()
```

### 3. Task Allocator

The `Allocator` handles task allocation to nodes using different strategies:

#### Allocation Strategies

- **Round Robin**: Distributes tasks evenly across nodes in a circular fashion
- **Random**: Randomly selects a node for each task
- **Least Loaded**: Assigns tasks to the node with the fewest running tasks

#### Key Methods

- **AllocateTask**: Allocates a task to a healthy node
  - Selects a node based on the configured strategy
  - Acquires a distributed lock for the task
  - Updates the task with node assignment
  - Sets task status to RUNNING

- **ReleaseTask**: Releases a task's distributed lock after execution

- **GetNodeTasks**: Returns all tasks assigned to a specific node

## Task-Node Binding

When a task is allocated to a node:

1. The task's `NodeID` field is set to the assigned node's ID
2. The task status is updated to `RUNNING`
3. The `StartedAt` timestamp is recorded
4. A distributed lock is acquired to prevent other nodes from executing the same task

## Distributed Locking

The system uses distributed locks (Redis or Etcd) to ensure that each task is executed by only one node:

```go
lockKey := fmt.Sprintf("task:lock:%s", task.ID)
acquired, err := distributedLock.Lock(ctx, lockKey, 5*time.Minute)
```

## Health Monitoring

Nodes are continuously monitored for health:

1. Each node sends periodic heartbeats to the registry
2. The registry checks heartbeat timestamps against a timeout threshold
3. Nodes that miss heartbeats are marked as unhealthy
4. Only healthy nodes receive new task allocations

## Failover

When a node fails:

1. The health check detects the missing heartbeats
2. The node is marked as unhealthy
3. New tasks are only allocated to healthy nodes
4. Existing tasks on the failed node can be transferred to healthy nodes (see `FailoverManager`)

## Usage Example

```go
// Create registry
registry := node.NewInMemoryRegistry(30 * time.Second)

// Register nodes
node1 := &node.Node{
    ID:      "node1",
    Address: "localhost:8080",
}
registry.Register(ctx, node1)

// Start heartbeat
hb := node.NewHeartbeatManager(registry, "node1", 10*time.Second)
hb.Start()
defer hb.Stop()

// Create allocator
allocator := node.NewAllocator(
    registry,
    taskRepository,
    distributedLock,
    node.AllocationStrategyRoundRobin,
)

// Allocate a task
selectedNode, err := allocator.AllocateTask(ctx, task)
if err != nil {
    log.Fatal(err)
}

// Task is now assigned to selectedNode
fmt.Printf("Task %s allocated to node %s\n", task.ID, selectedNode.ID)

// After execution, release the task
allocator.ReleaseTask(ctx, task.ID)
```

## Configuration

Key configuration parameters:

- **Heartbeat Interval**: How often nodes send heartbeats (default: 10s)
- **Heartbeat Timeout**: How long before a node is considered unhealthy (default: 30s)
- **Lock TTL**: Time-to-live for distributed locks (default: 5m)
- **Allocation Strategy**: Round robin, random, or least loaded

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 7.1**: Multiple nodes can run simultaneously, and distributed locks ensure each task is executed by only one node
- **Requirement 7.3**: New nodes are automatically discovered when they register with the registry
- **Requirement 7.5**: Task-node binding is recorded in the task's `NodeID` field

## Testing

The implementation includes comprehensive tests:

- `TestAllocator_AllocateTask`: Verifies task allocation and locking
- `TestAllocator_ReleaseTask`: Verifies lock release
- `TestAllocator_RoundRobinStrategy`: Verifies round-robin distribution
- `TestAllocator_NoHealthyNodes`: Verifies error handling when no nodes are available
- `TestAllocator_GetNodeTasks`: Verifies task-node binding queries
- `TestMultiNodeTaskAllocation_Integration`: End-to-end integration test with heartbeats and failover

## See Also

- [Distributed Lock Implementation](./distributed-lock-implementation.md)
- [Architecture Overview](./architecture.md)
