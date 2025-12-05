package discovery

import (
	"context"
	"fmt"

	"task-scheduler/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesDiscovery implements service discovery using Kubernetes
type KubernetesDiscovery struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewKubernetesDiscovery creates a new Kubernetes service discovery instance
func NewKubernetesDiscovery(kubeconfig string, namespace string) (*KubernetesDiscovery, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		// Use kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &KubernetesDiscovery{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// Discover discovers service instances from Kubernetes
func (k *KubernetesDiscovery) Discover(ctx context.Context, serviceName string) ([]string, error) {
	// Verify service exists
	_, err := k.clientset.CoreV1().Services(k.namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	// Get endpoints for the service
	endpoints, err := k.clientset.CoreV1().Endpoints(k.namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints for service %s: %w", serviceName, err)
	}

	addresses := []string{}
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			for _, port := range subset.Ports {
				address := fmt.Sprintf("%s:%d", addr.IP, port.Port)
				addresses = append(addresses, address)
			}
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no endpoints found for service: %s", serviceName)
	}

	return addresses, nil
}

// Register is not applicable for Kubernetes (services are managed by K8s)
func (k *KubernetesDiscovery) Register(ctx context.Context, serviceName string, address string) error {
	return fmt.Errorf("register operation not supported for Kubernetes discovery")
}

// Deregister is not applicable for Kubernetes (services are managed by K8s)
func (k *KubernetesDiscovery) Deregister(ctx context.Context, serviceName string, address string) error {
	return fmt.Errorf("deregister operation not supported for Kubernetes discovery")
}

var _ domain.ServiceDiscovery = (*KubernetesDiscovery)(nil)
