package kubernetes

import (
	"fmt"
	"sync"

	"k8s.io/client-go/tools/clientcmd"
)

// KubeconfigProvider implements ClientProvider using kubeconfig files.
type KubeconfigProvider struct {
	factory        *ClientFactory
	kubeconfigPath string
	contexts       map[string]*ClientSet
	mu             sync.RWMutex
}

// NewKubeconfigProvider creates a new kubeconfig-based provider.
func NewKubeconfigProvider(factory *ClientFactory, kubeconfigPath string) (*KubeconfigProvider, error) {
	return &KubeconfigProvider{
		factory:        factory,
		kubeconfigPath: kubeconfigPath,
		contexts:       make(map[string]*ClientSet),
	}, nil
}

// GetClientSet returns a ClientSet for the given context.
func (p *KubeconfigProvider) GetClientSet(ctx string) (*ClientSet, error) {
	p.mu.RLock()
	if clientSet, ok := p.contexts[ctx]; ok {
		p.mu.RUnlock()
		return clientSet, nil
	}
	p.mu.RUnlock()

	// Create new client set
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if clientSet, ok := p.contexts[ctx]; ok {
		return clientSet, nil
	}

	clientSet, err := p.factory.CreateKubeconfigClientSet(p.kubeconfigPath, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set for context %s: %w", ctx, err)
	}

	p.contexts[ctx] = clientSet
	return clientSet, nil
}

// ListContexts returns all available contexts from the kubeconfig.
func (p *KubeconfigProvider) ListContexts() ([]string, error) {
	expandedPath := expandKubeconfigPath(p.kubeconfigPath)
	config, err := clientcmd.LoadFromFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	contexts := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

// GetCurrentContext returns the current context name.
func (p *KubeconfigProvider) GetCurrentContext() (string, error) {
	expandedPath := expandKubeconfigPath(p.kubeconfigPath)
	config, err := clientcmd.LoadFromFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config.CurrentContext, nil
}

// InClusterProvider implements ClientProvider using in-cluster configuration.
type InClusterProvider struct {
	factory   *ClientFactory
	clientSet *ClientSet
	mu        sync.RWMutex
}

// NewInClusterProvider creates a new in-cluster provider.
func NewInClusterProvider(factory *ClientFactory) (*InClusterProvider, error) {
	clientSet, err := factory.CreateInClusterClientSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster client set: %w", err)
	}

	return &InClusterProvider{
		factory:   factory,
		clientSet: clientSet,
	}, nil
}

// GetClientSet returns the in-cluster ClientSet.
func (p *InClusterProvider) GetClientSet(ctx string) (*ClientSet, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.clientSet, nil
}

// ListContexts returns a single context for in-cluster mode.
func (p *InClusterProvider) ListContexts() ([]string, error) {
	return []string{"in-cluster"}, nil
}

// GetCurrentContext returns the in-cluster context name.
func (p *InClusterProvider) GetCurrentContext() (string, error) {
	return "in-cluster", nil
}

// SingleClusterProvider implements ClientProvider for a single cluster.
type SingleClusterProvider struct {
	factory        *ClientFactory
	context        string
	kubeconfigPath string
	clientSet      *ClientSet
	mu             sync.RWMutex
}

// NewSingleClusterProvider creates a new single-cluster provider.
func NewSingleClusterProvider(factory *ClientFactory, kubeconfigPath, context string) (*SingleClusterProvider, error) {
	clientSet, err := factory.CreateKubeconfigClientSet(kubeconfigPath, context)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %w", err)
	}

	return &SingleClusterProvider{
		factory:        factory,
		context:        context,
		kubeconfigPath: kubeconfigPath,
		clientSet:      clientSet,
	}, nil
}

// GetClientSet returns the single ClientSet.
func (p *SingleClusterProvider) GetClientSet(ctx string) (*ClientSet, error) {
	if ctx != "" && ctx != p.context {
		return nil, fmt.Errorf("context %s not available in single-cluster mode", ctx)
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.clientSet, nil
}

// ListContexts returns the single context.
func (p *SingleClusterProvider) ListContexts() ([]string, error) {
	return []string{p.context}, nil
}

// GetCurrentContext returns the single context name.
func (p *SingleClusterProvider) GetCurrentContext() (string, error) {
	return p.context, nil
}

// NewProvider creates a ClientProvider based on the provider type.
func NewProvider(providerType, kubeconfigPath, context string, factory *ClientFactory) (ClientProvider, error) {
	switch providerType {
	case "kubeconfig":
		return NewKubeconfigProvider(factory, kubeconfigPath)
	case "in-cluster":
		return NewInClusterProvider(factory)
	case "single":
		if context == "" {
			return nil, fmt.Errorf("context required for single-cluster provider")
		}
		return NewSingleClusterProvider(factory, kubeconfigPath, context)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}
