package discovery

import (
	"context"
	"fmt"
	"time"

	"task-scheduler/internal/domain"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdDiscovery implements service discovery using Etcd
type EtcdDiscovery struct {
	client *clientv3.Client
	prefix string
}

// NewEtcdDiscovery creates a new Etcd service discovery instance
func NewEtcdDiscovery(endpoints []string, prefix string) (*EtcdDiscovery, error) {
	if prefix == "" {
		prefix = "/services"
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdDiscovery{
		client: client,
		prefix: prefix,
	}, nil
}

// Discover discovers service instances from Etcd
func (e *EtcdDiscovery) Discover(ctx context.Context, serviceName string) ([]string, error) {
	key := fmt.Sprintf("%s/%s/", e.prefix, serviceName)

	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

	addresses := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addresses = append(addresses, string(kv.Value))
	}

	return addresses, nil
}

// Register registers a service instance with Etcd
func (e *EtcdDiscovery) Register(ctx context.Context, serviceName string, address string) error {
	key := fmt.Sprintf("%s/%s/%s", e.prefix, serviceName, address)

	_, err := e.client.Put(ctx, key, address)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// Deregister removes a service instance from Etcd
func (e *EtcdDiscovery) Deregister(ctx context.Context, serviceName string, address string) error {
	key := fmt.Sprintf("%s/%s/%s", e.prefix, serviceName, address)

	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}

// Close closes the Etcd client connection
func (e *EtcdDiscovery) Close() error {
	return e.client.Close()
}

var _ domain.ServiceDiscovery = (*EtcdDiscovery)(nil)
