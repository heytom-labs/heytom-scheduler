package service

import "task-scheduler/internal/domain"

// SimpleDiscoveryFactory is a simple implementation of DiscoveryFactory
type SimpleDiscoveryFactory struct {
	discovery domain.ServiceDiscovery
}

// NewSimpleDiscoveryFactory creates a new simple discovery factory
func NewSimpleDiscoveryFactory(discovery domain.ServiceDiscovery) DiscoveryFactory {
	return &SimpleDiscoveryFactory{
		discovery: discovery,
	}
}

// GetDiscovery returns the configured service discovery
func (f *SimpleDiscoveryFactory) GetDiscovery(discoveryType domain.ServiceDiscoveryType) (domain.ServiceDiscovery, error) {
	return f.discovery, nil
}
