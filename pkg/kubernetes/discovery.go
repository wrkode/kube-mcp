package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CRDDiscovery handles discovery and caching of Custom Resource Definitions.
type CRDDiscovery struct {
	clientSet *ClientSet
	cache     map[string]schema.GroupVersionResource
	cacheTime time.Time
	cacheTTL  time.Duration
	mu        sync.RWMutex
}

// NewCRDDiscovery creates a new CRD discovery instance.
func NewCRDDiscovery(clientSet *ClientSet, cacheTTL time.Duration) *CRDDiscovery {
	return &CRDDiscovery{
		clientSet: clientSet,
		cache:     make(map[string]schema.GroupVersionResource),
		cacheTTL:  cacheTTL,
	}
}

// DiscoverCRDs discovers all CRDs in the cluster and caches them.
func (d *CRDDiscovery) DiscoverCRDs(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if cache is still valid
	if !d.cacheTime.IsZero() && time.Since(d.cacheTime) < d.cacheTTL {
		return nil
	}

	// Discover API resources
	apiResources, err := d.clientSet.Discovery.ServerPreferredResources()
	if err != nil {
		return fmt.Errorf("failed to discover API resources: %w", err)
	}

	// Clear cache
	d.cache = make(map[string]schema.GroupVersionResource)

	// Build cache of CRDs
	for _, groupVersionResources := range apiResources {
		gv, err := schema.ParseGroupVersion(groupVersionResources.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range groupVersionResources.APIResources {
			// Skip subresources
			if len(resource.Name) > 0 && resource.Name[len(resource.Name)-1] == '/' {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}

			// Store as group/version/kind
			key := fmt.Sprintf("%s/%s/%s", gv.Group, gv.Version, resource.Kind)
			d.cache[key] = gvr
		}
	}

	d.cacheTime = time.Now()
	return nil
}

// GetGVR returns the GroupVersionResource for a given GVK.
func (d *CRDDiscovery) GetGVR(gvk schema.GroupVersionKind) (schema.GroupVersionResource, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
	gvr, ok := d.cache[key]
	return gvr, ok
}

// HasCRD checks if a CRD exists for the given GVK.
func (d *CRDDiscovery) HasCRD(gvk schema.GroupVersionKind) bool {
	_, ok := d.GetGVR(gvk)
	return ok
}

// ListCRDs returns all discovered CRDs.
func (d *CRDDiscovery) ListCRDs() []schema.GroupVersionKind {
	d.mu.RLock()
	defer d.mu.RUnlock()

	gvks := make([]schema.GroupVersionKind, 0, len(d.cache))
	for key := range d.cache {
		// Parse key format: "group/version/kind"
		var gvk schema.GroupVersionKind
		_, err := fmt.Sscanf(key, "%s/%s/%s", &gvk.Group, &gvk.Version, &gvk.Kind)
		if err == nil {
			gvks = append(gvks, gvk)
		}
	}

	return gvks
}

// CheckKubeVirtCRDs checks if KubeVirt CRDs are available.
func (d *CRDDiscovery) CheckKubeVirtCRDs(ctx context.Context) (bool, error) {
	if err := d.DiscoverCRDs(ctx); err != nil {
		return false, err
	}

	// Check for VirtualMachine CRD
	vmGVK := schema.GroupVersionKind{
		Group:   "kubevirt.io",
		Version: "v1",
		Kind:    "VirtualMachine",
	}

	return d.HasCRD(vmGVK), nil
}

// Refresh forces a refresh of the CRD cache.
func (d *CRDDiscovery) Refresh(ctx context.Context) error {
	d.mu.Lock()
	d.cacheTime = time.Time{} // Invalidate cache
	d.mu.Unlock()

	return d.DiscoverCRDs(ctx)
}
