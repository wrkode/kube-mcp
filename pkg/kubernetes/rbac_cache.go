package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RBACAuthorizer provides RBAC authorization checking with caching.
type RBACAuthorizer interface {
	// Allowed checks if a user is allowed to perform an action.
	Allowed(ctx context.Context, user string, verb string, gvr schema.GroupVersionResource, namespace string) (bool, error)
}

// rbacCacheEntry represents a cached RBAC check result.
type rbacCacheEntry struct {
	allowed   bool
	expiresAt time.Time
}

// rbacAuthorizerImpl implements RBACAuthorizer with caching.
type rbacAuthorizerImpl struct {
	clientSet *ClientSet
	cache     map[string]*rbacCacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
}

// NewRBACAuthorizer creates a new RBAC authorizer with caching.
func NewRBACAuthorizer(clientSet *ClientSet, ttlSeconds int) RBACAuthorizer {
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Second // Default TTL
	}

	return &rbacAuthorizerImpl{
		clientSet: clientSet,
		cache:     make(map[string]*rbacCacheEntry),
		ttl:       ttl,
	}
}

// cacheKey generates a cache key for an RBAC check.
func cacheKey(user, verb string, gvr schema.GroupVersionResource, namespace string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", user, verb, gvr.Group, gvr.Resource, namespace)
}

// Allowed checks if a user is allowed to perform an action.
func (r *rbacAuthorizerImpl) Allowed(ctx context.Context, user string, verb string, gvr schema.GroupVersionResource, namespace string) (bool, error) {
	key := cacheKey(user, verb, gvr, namespace)

	// Check cache
	r.mu.RLock()
	if entry, ok := r.cache[key]; ok {
		if time.Now().Before(entry.expiresAt) {
			allowed := entry.allowed
			r.mu.RUnlock()
			return allowed, nil
		}
		// Entry expired, remove it
		delete(r.cache, key)
	}
	r.mu.RUnlock()

	// Perform actual RBAC check
	allowed, err := r.checkRBAC(ctx, user, verb, gvr, namespace)
	if err != nil {
		return false, err
	}

	// Cache the result
	r.mu.Lock()
	r.cache[key] = &rbacCacheEntry{
		allowed:   allowed,
		expiresAt: time.Now().Add(r.ttl),
	}
	r.mu.Unlock()

	return allowed, nil
}

// checkRBAC performs the actual RBAC check using SelfSubjectAccessReview.
func (r *rbacAuthorizerImpl) checkRBAC(ctx context.Context, user string, verb string, gvr schema.GroupVersionResource, namespace string) (bool, error) {
	review := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     gvr.Group,
				Resource:  gvr.Resource,
			},
		},
	}

	result, err := r.clientSet.Typed.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to create self subject access review: %w", err)
	}

	return result.Status.Allowed, nil
}

// ClearCache clears the RBAC cache.
func (r *rbacAuthorizerImpl) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]*rbacCacheEntry)
}
