package kubernetes

import (
	"fmt"
)

// ClusterTarget represents a cluster targeting configuration.
type ClusterTarget struct {
	// Context is the kubeconfig context to target.
	// If empty, uses the provider's default context.
	Context string `json:"context,omitempty"`
}

// ClusterClient represents a client bound to a specific cluster context.
type ClusterClient interface {
	// GetClientSet returns the ClientSet for this cluster.
	GetClientSet() *ClientSet
	// GetContext returns the context name.
	GetContext() string
}

// clusterClientImpl implements ClusterClient.
type clusterClientImpl struct {
	clientSet *ClientSet
	context   string
}

// GetClientSet returns the ClientSet.
func (c *clusterClientImpl) GetClientSet() *ClientSet {
	return c.clientSet
}

// GetContext returns the context name.
func (c *clusterClientImpl) GetContext() string {
	return c.context
}

// WithContext returns a ClusterClient for the specified context.
// If context is empty, uses the provider's default context.
func (p *KubeconfigProvider) WithContext(ctx string) (ClusterClient, error) {
	clientSet, err := p.GetClientSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client set for context %s: %w", ctx, err)
	}

	// If context is empty, get the default context name
	if ctx == "" {
		defaultCtx, err := p.GetCurrentContext()
		if err != nil {
			defaultCtx = ""
		}
		ctx = defaultCtx
	}

	return &clusterClientImpl{
		clientSet: clientSet,
		context:   ctx,
	}, nil
}

// WithContext returns a ClusterClient for the specified context.
// For in-cluster provider, context is ignored.
func (p *InClusterProvider) WithContext(ctx string) (ClusterClient, error) {
	clientSet, err := p.GetClientSet("")
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster client set: %w", err)
	}

	return &clusterClientImpl{
		clientSet: clientSet,
		context:   "in-cluster",
	}, nil
}

// WithContext returns a ClusterClient for the specified context.
// For single-cluster provider, only the configured context is allowed.
func (p *SingleClusterProvider) WithContext(ctx string) (ClusterClient, error) {
	if ctx != "" && ctx != p.context {
		return nil, fmt.Errorf("context %s not available in single-cluster mode (configured: %s)", ctx, p.context)
	}

	clientSet, err := p.GetClientSet("")
	if err != nil {
		return nil, fmt.Errorf("failed to get client set: %w", err)
	}

	return &clusterClientImpl{
		clientSet: clientSet,
		context:   p.context,
	}, nil
}
