package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClientFactory creates Kubernetes clients from a REST config.
type ClientFactory struct {
	qps     float32
	burst   int
	timeout time.Duration
}

// NewClientFactory creates a new client factory with the given settings.
func NewClientFactory(qps float32, burst int, timeout time.Duration) *ClientFactory {
	return &ClientFactory{
		qps:     qps,
		burst:   burst,
		timeout: timeout,
	}
}

// CreateClientSet creates a ClientSet from a REST config.
func (f *ClientFactory) CreateClientSet(config *rest.Config) (*ClientSet, error) {
	// Apply QPS and burst settings
	config.QPS = f.qps
	config.Burst = f.burst
	config.Timeout = f.timeout

	// Create typed client
	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create typed client: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Create REST mapper
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	// Create metrics client (may fail if metrics server is not available)
	metricsClient, _ := metricsclientset.NewForConfig(config)

	return &ClientSet{
		Typed:      typedClient,
		Dynamic:    dynamicClient,
		Discovery:  discoveryClient,
		Metrics:    metricsClient,
		Config:     config,
		RESTMapper: mapper,
	}, nil
}

// CreateInClusterClientSet creates a ClientSet using in-cluster configuration.
func (f *ClientFactory) CreateInClusterClientSet() (*ClientSet, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	return f.CreateClientSet(config)
}

// expandKubeconfigPath expands ~ to the user's home directory in kubeconfig paths.
func expandKubeconfigPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// CreateKubeconfigClientSet creates a ClientSet from a kubeconfig file and context.
func (f *ClientFactory) CreateKubeconfigClientSet(kubeconfigPath, context string) (*ClientSet, error) {
	// Expand ~ in kubeconfig path
	expandedPath := expandKubeconfigPath(kubeconfigPath)
	
	// Use clientcmd to load kubeconfig
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: expandedPath,
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return f.CreateClientSet(restConfig)
}
