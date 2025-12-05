package discovery

import (
	"fmt"
	"sync"

	"task-scheduler/internal/domain"
)

// Factory creates service discovery instances
type Factory struct {
	mu        sync.RWMutex
	instances map[domain.ServiceDiscoveryType]domain.ServiceDiscovery
	config    *FactoryConfig
}

// FactoryConfig contains configuration for creating discovery instances
type FactoryConfig struct {
	ConsulAddress string
	EtcdEndpoints []string
	EtcdPrefix    string
	KubeConfig    string
	KubeNamespace string
}

// NewFactory creates a new discovery factory
func NewFactory(config *FactoryConfig) *Factory {
	if config == nil {
		config = &FactoryConfig{}
	}

	return &Factory{
		instances: make(map[domain.ServiceDiscoveryType]domain.ServiceDiscovery),
		config:    config,
	}
}

// GetDiscovery returns a service discovery instance for the given type
func (f *Factory) GetDiscovery(discoveryType domain.ServiceDiscoveryType) (domain.ServiceDiscovery, error) {
	f.mu.RLock()
	instance, exists := f.instances[discoveryType]
	f.mu.RUnlock()

	if exists {
		return instance, nil
	}

	// Create new instance
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if instance, exists := f.instances[discoveryType]; exists {
		return instance, nil
	}

	var err error
	switch discoveryType {
	case domain.ServiceDiscoveryStatic:
		instance = NewStaticDiscovery()
	case domain.ServiceDiscoveryConsul:
		instance, err = NewConsulDiscovery(f.config.ConsulAddress)
	case domain.ServiceDiscoveryEtcd:
		instance, err = NewEtcdDiscovery(f.config.EtcdEndpoints, f.config.EtcdPrefix)
	case domain.ServiceDiscoveryKubernetes:
		instance, err = NewKubernetesDiscovery(f.config.KubeConfig, f.config.KubeNamespace)
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", discoveryType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create discovery instance: %w", err)
	}

	f.instances[discoveryType] = instance
	return instance, nil
}

// RegisterInstance allows registering a pre-configured discovery instance
func (f *Factory) RegisterInstance(discoveryType domain.ServiceDiscoveryType, instance domain.ServiceDiscovery) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.instances[discoveryType] = instance
}
