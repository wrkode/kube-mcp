package kubernetes

import (
	"context"
	"fmt"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RBACChecker checks RBAC permissions for Kubernetes operations.
type RBACChecker struct {
	clientSet *ClientSet
}

// NewRBACChecker creates a new RBAC checker.
func NewRBACChecker(clientSet *ClientSet) *RBACChecker {
	return &RBACChecker{
		clientSet: clientSet,
	}
}

// CheckAccess checks if the current user has access to perform an action on a resource.
func (r *RBACChecker) CheckAccess(ctx context.Context, gvr schema.GroupVersionResource, namespace, verb string) (bool, error) {
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Group:     gvr.Group,
				Version:   gvr.Version,
				Resource:  gvr.Resource,
				Namespace: namespace,
				Verb:      verb,
			},
		},
	}

	result, err := r.clientSet.Typed.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	return result.Status.Allowed, nil
}

// CheckAccessGVK checks if the current user has access using a GVK.
func (r *RBACChecker) CheckAccessGVK(ctx context.Context, gvk schema.GroupVersionKind, namespace, verb string) (bool, error) {
	// Convert GVK to GVR using REST mapper
	mapping, err := r.clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return false, fmt.Errorf("failed to map GVK to GVR: %w", err)
	}

	return r.CheckAccess(ctx, mapping.Resource, namespace, verb)
}

// CheckMultipleAccess checks multiple access permissions and returns a map of results.
func (r *RBACChecker) CheckMultipleAccess(ctx context.Context, checks []AccessCheck) (map[string]bool, error) {
	results := make(map[string]bool)

	for _, check := range checks {
		allowed, err := r.CheckAccess(ctx, check.GVR, check.Namespace, check.Verb)
		if err != nil {
			return nil, fmt.Errorf("failed to check access for %s/%s/%s: %w", check.GVR.Group, check.GVR.Version, check.GVR.Resource, err)
		}
		results[check.Key] = allowed
	}

	return results, nil
}

// AccessCheck represents a single access check request.
type AccessCheck struct {
	Key       string
	GVR       schema.GroupVersionResource
	Namespace string
	Verb      string
}
