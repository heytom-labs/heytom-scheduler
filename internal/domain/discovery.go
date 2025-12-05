package domain

import "context"

// ServiceDiscovery defines the interface for service discovery
type ServiceDiscovery interface {
	// Discover discovers service instances by name
	Discover(ctx context.Context, serviceName string) ([]string, error)

	// Register registers a service instance
	Register(ctx context.Context, serviceName string, address string) error

	// Deregister deregisters a service instance
	Deregister(ctx context.Context, serviceName string, address string) error
}
