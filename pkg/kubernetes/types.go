package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClientSet represents a set of Kubernetes clients for a specific context.
type ClientSet struct {
	// Typed client for core Kubernetes resources
	Typed kubernetes.Interface

	// Dynamic client for arbitrary resources (including CRDs)
	Dynamic dynamic.Interface

	// Discovery client for API discovery
	Discovery discovery.DiscoveryInterface

	// Metrics client for metrics.k8s.io API
	Metrics metricsclientset.Interface

	// REST config used to create clients
	Config *rest.Config

	// REST mapper for GVK to GVR mapping
	RESTMapper meta.RESTMapper
}

// ClientProvider provides Kubernetes clients for different contexts.
type ClientProvider interface {
	// GetClientSet returns a ClientSet for the given context.
	// If context is empty, returns the default/current context.
	GetClientSet(ctx string) (*ClientSet, error)

	// ListContexts returns all available contexts.
	ListContexts() ([]string, error)

	// GetCurrentContext returns the current/default context name.
	GetCurrentContext() (string, error)
}

// GVK represents a GroupVersionKind.
type GVK struct {
	Group   string
	Version string
	Kind    string
}

// String returns the string representation of GVK.
func (g GVK) String() string {
	if g.Group == "" {
		return g.Version + "/" + g.Kind
	}
	return g.Group + "/" + g.Version + "/" + g.Kind
}

// ParseGVK parses a GVK string into a GVK struct.
func ParseGVK(s string) GVK {
	// Simple parsing - format: group/version/kind or version/kind
	// This is a placeholder implementation
	// In production, use proper parsing logic
	return GVK{}
}

// ToSchemaGVK converts to schema.GroupVersionKind.
func (g GVK) ToSchemaGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.Group,
		Version: g.Version,
		Kind:    g.Kind,
	}
}
