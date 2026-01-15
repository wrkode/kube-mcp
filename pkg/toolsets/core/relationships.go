package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleResourcesRelationships handles the resources_relationships tool.
// It finds resource owners and/or dependents based on owner references.
func (t *Toolset) handleResourcesRelationships(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Direction string `json:"direction"` // "owners", "dependents", or "both" (default: "both")
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	gvk := schema.GroupVersionKind{
		Group:   args.Group,
		Version: args.Version,
		Kind:    args.Kind,
	}

	// Map GVK to GVR
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource

	// Get the resource
	resource, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get resource: %w", err)), nil
	}

	direction := args.Direction
	if direction == "" {
		direction = "both"
	}

	result := map[string]interface{}{
		"name":      resource.GetName(),
		"namespace": resource.GetNamespace(),
		"kind":      resource.GetKind(),
		"gvk":       gvk.String(),
	}

	// Find owners (resources that own this resource)
	if direction == "owners" || direction == "both" {
		owners := t.findOwners(ctx, clientSet, resource)
		result["owners"] = owners
	}

	// Find dependents (resources owned by this resource)
	if direction == "dependents" || direction == "both" {
		dependents := t.findDependents(ctx, clientSet, resource, gvk)
		result["dependents"] = dependents
	}

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(result)
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

// findOwners finds all resources that own the given resource.
func (t *Toolset) findOwners(ctx context.Context, clientSet *kubernetes.ClientSet, resource *unstructured.Unstructured) []map[string]interface{} {
	owners := []map[string]interface{}{}

	ownerRefs := resource.GetOwnerReferences()
	for _, ownerRef := range ownerRefs {
		ownerGVK := schema.FromAPIVersionAndKind(ownerRef.APIVersion, ownerRef.Kind)
		mapping, err := clientSet.RESTMapper.RESTMapping(ownerGVK.GroupKind(), ownerGVK.Version)
		if err != nil {
			continue
		}

		ownerGVR := mapping.Resource
		ownerNamespace := resource.GetNamespace()
		if mapping.Scope.Name() == "Cluster" {
			ownerNamespace = ""
		}

		owner, err := clientSet.Dynamic.Resource(ownerGVR).Namespace(ownerNamespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			// Owner might not exist anymore
			owners = append(owners, map[string]interface{}{
				"name":       ownerRef.Name,
				"kind":       ownerRef.Kind,
				"api_version": ownerRef.APIVersion,
				"uid":        ownerRef.UID,
				"namespace":  ownerNamespace,
				"exists":     false,
			})
			continue
		}

		owners = append(owners, map[string]interface{}{
			"name":       owner.GetName(),
			"kind":       owner.GetKind(),
			"api_version": owner.GetAPIVersion(),
			"uid":        string(owner.GetUID()),
			"namespace":  owner.GetNamespace(),
			"exists":     true,
		})
	}

	return owners
}

// findDependents finds all resources owned by the given resource.
func (t *Toolset) findDependents(ctx context.Context, clientSet *kubernetes.ClientSet, resource *unstructured.Unstructured, ownerGVK schema.GroupVersionKind) []map[string]interface{} {
	dependents := []map[string]interface{}{}
	resourceUID := resource.GetUID()

	// Get all API resources
	apiResources, err := clientSet.Discovery.ServerPreferredResources()
	if err != nil {
		return dependents
	}

	// Search through all resource types for dependents
	for _, apiResourceList := range apiResources {
		for _, apiResource := range apiResourceList.APIResources {
			// Skip subresources
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			// Parse GVK
			gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
			if err != nil {
				continue
			}

			gvk := gv.WithKind(apiResource.Kind)
			mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
			if err != nil {
				continue
			}

			gvr := mapping.Resource
			namespace := resource.GetNamespace()

			// List resources in the namespace (or cluster-wide)
			var list *unstructured.UnstructuredList
			if mapping.Scope.Name() == "Namespaced" && namespace != "" {
				list, err = clientSet.Dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
			} else if mapping.Scope.Name() == "Cluster" {
				list, err = clientSet.Dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
			} else {
				continue
			}

			if err != nil {
				continue
			}

			// Check each resource for owner reference
			for _, item := range list.Items {
				for _, ownerRef := range item.GetOwnerReferences() {
					if ownerRef.UID == resourceUID {
						dependents = append(dependents, map[string]interface{}{
							"name":       item.GetName(),
							"kind":       item.GetKind(),
							"api_version": item.GetAPIVersion(),
							"uid":        string(item.GetUID()),
							"namespace":  item.GetNamespace(),
						})
						break
					}
				}
			}
		}
	}

	return dependents
}

