package discovery

import (
	"context"
	"fmt"

	"task-scheduler/internal/domain"

	"github.com/hashicorp/consul/api"
)

// ConsulDiscovery implements service discovery using Consul
type ConsulDiscovery struct {
	client *api.Client
}

// NewConsulDiscovery creates a new Consul service discovery instance
func NewConsulDiscovery(address string) (*ConsulDiscovery, error) {
	config := api.DefaultConfig()
	if address != "" {
		config.Address = address
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulDiscovery{
		client: client,
	}, nil
}

// Discover discovers service instances from Consul
func (c *ConsulDiscovery) Discover(ctx context.Context, serviceName string) ([]string, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	addresses := make([]string, 0, len(services))
	for _, service := range services {
		addr := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

// Register registers a service instance with Consul
func (c *ConsulDiscovery) Register(ctx context.Context, serviceName string, address string) error {
	// Parse address to get host and port
	// For simplicity, we'll use the address as-is
	// In production, you'd want to parse host:port properly
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", serviceName, address),
		Name:    serviceName,
		Address: address,
	}

	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// Deregister removes a service instance from Consul
func (c *ConsulDiscovery) Deregister(ctx context.Context, serviceName string, address string) error {
	serviceID := fmt.Sprintf("%s-%s", serviceName, address)
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}

var _ domain.ServiceDiscovery = (*ConsulDiscovery)(nil)
