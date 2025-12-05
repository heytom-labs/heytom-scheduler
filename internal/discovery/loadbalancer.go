package discovery

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// LoadBalancerStrategy defines the load balancing strategy
type LoadBalancerStrategy string

const (
	// RoundRobin distributes requests evenly across instances
	RoundRobin LoadBalancerStrategy = "round_robin"
	// Random selects a random instance
	Random LoadBalancerStrategy = "random"
)

// LoadBalancer selects an instance from a list of addresses
type LoadBalancer interface {
	// Select chooses an address from the list
	Select(addresses []string) (string, error)
}

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	counter uint64
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		counter: 0,
	}
}

// Select selects an address using round-robin strategy
func (r *RoundRobinLoadBalancer) Select(addresses []string) (string, error) {
	if len(addresses) == 0 {
		return "", fmt.Errorf("no addresses available")
	}

	index := atomic.AddUint64(&r.counter, 1) - 1
	return addresses[index%uint64(len(addresses))], nil
}

// RandomLoadBalancer implements random load balancing
type RandomLoadBalancer struct {
	rng *rand.Rand
}

// NewRandomLoadBalancer creates a new random load balancer
func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select selects an address using random strategy
func (r *RandomLoadBalancer) Select(addresses []string) (string, error) {
	if len(addresses) == 0 {
		return "", fmt.Errorf("no addresses available")
	}

	index := r.rng.Intn(len(addresses))
	return addresses[index], nil
}

// NewLoadBalancer creates a load balancer based on the strategy
func NewLoadBalancer(strategy LoadBalancerStrategy) LoadBalancer {
	switch strategy {
	case RoundRobin:
		return NewRoundRobinLoadBalancer()
	case Random:
		return NewRandomLoadBalancer()
	default:
		return NewRoundRobinLoadBalancer()
	}
}
