package discovery

import (
	"context"
	"fmt"
	"sync"

	"task-scheduler/internal/domain"
)

// StaticDiscovery implements service discovery using static addresses
type StaticDiscovery struct {
	mu       sync.RWMutex
	services map[string][]string // serviceName -> addresses
}

// NewStaticDiscovery creates a new static service discovery instance
func NewStaticDiscovery() *StaticDiscovery {
	return &StaticDiscovery{
		services: make(map[string][]string),
	}
}

// Discover returns the registered addresses for a service
func (s *StaticDiscovery) Discover(ctx context.Context, serviceName string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addresses, ok := s.services[serviceName]
	if !ok || len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses found for service: %s", serviceName)
	}

	// Return a copy to avoid external modifications
	result := make([]string, len(addresses))
	copy(result, addresses)
	return result, nil
}

// Register registers a service instance with an address
func (s *StaticDiscovery) Register(ctx context.Context, serviceName string, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	if s.services[serviceName] == nil {
		s.services[serviceName] = []string{}
	}

	// Check if address already exists
	for _, addr := range s.services[serviceName] {
		if addr == address {
			return nil // Already registered
		}
	}

	s.services[serviceName] = append(s.services[serviceName], address)
	return nil
}

// Deregister removes a service instance
func (s *StaticDiscovery) Deregister(ctx context.Context, serviceName string, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	addresses, ok := s.services[serviceName]
	if !ok {
		return nil // Service not found, nothing to do
	}

	// Remove the address
	newAddresses := []string{}
	for _, addr := range addresses {
		if addr != address {
			newAddresses = append(newAddresses, addr)
		}
	}

	if len(newAddresses) == 0 {
		delete(s.services, serviceName)
	} else {
		s.services[serviceName] = newAddresses
	}

	return nil
}

var _ domain.ServiceDiscovery = (*StaticDiscovery)(nil)
