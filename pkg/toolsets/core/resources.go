package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleResourcesList handles the resources_list tool.
func (t *Toolset) handleResourcesList(ctx context.Context, args struct {
	Group         string `json:"group"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	FieldSelector string `json:"field_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
	Context       string `json:"context"`
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

	// Build list options with selectors and pagination
	listOptions := metav1.ListOptions{}
	if args.LabelSelector != "" {
		listOptions.LabelSelector = args.LabelSelector
	}
	if args.FieldSelector != "" {
		listOptions.FieldSelector = args.FieldSelector
	}
	if args.Limit > 0 {
		listOptions.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOptions.Continue = args.Continue
	}

	var list *unstructured.UnstructuredList

	if args.Namespace == "" && mapping.Scope.Name() == "Namespaced" {
		// List all namespaces
		namespaces, err := clientSet.Typed.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list namespaces: %w", err)), nil
		}

		allItems := make([]unstructured.Unstructured, 0)
		for _, ns := range namespaces.Items {
			items, err := clientSet.Dynamic.Resource(gvr).Namespace(ns.Name).List(ctx, listOptions)
			if err != nil {
				continue
			}
			allItems = append(allItems, items.Items...)
			// Handle pagination - if we have a continue token, we'd need to handle it per namespace
			// For simplicity, pagination across all namespaces is not fully supported
		}

		list = &unstructured.UnstructuredList{Items: allItems}
	} else {
		list, err = clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list resources: %w", err)), nil
		}
	}

	resources := make([]map[string]any, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, map[string]any{
			"name":       item.GetName(),
			"namespace":  item.GetNamespace(),
			"kind":       item.GetKind(),
			"apiVersion": item.GetAPIVersion(),
		})
	}

	result := map[string]any{"resources": resources}
	if continueToken := list.GetContinue(); continueToken != "" {
		result["continue"] = continueToken
		result["has_more"] = true
	}

	return mcpHelpers.NewJSONResult(result)
}

// handleResourcesGet handles the resources_get tool.
func (t *Toolset) handleResourcesGet(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
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

	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource
	resource, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get resource: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(resource.Object)
}

// handleResourcesApply handles the resources_apply tool using server-side apply.
func (t *Toolset) handleResourcesApply(ctx context.Context, args struct {
	Manifest     map[string]any `json:"manifest"`
	FieldManager string         `json:"field_manager"`
	DryRun       bool           `json:"dry_run"`
	Context      string         `json:"context"`
}) (*mcp.CallToolResult, error) {

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Convert manifest to unstructured
	obj := &unstructured.Unstructured{Object: args.Manifest}

	gvk := obj.GroupVersionKind()
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource
	namespace := obj.GetNamespace()

	// Check if resource exists to determine verb (create vs update)
	// For server-side apply, we check for both create and update permissions
	// In practice, apply requires both, but we'll check update as it's more restrictive
	verb := "update"
	_, getErr := clientSet.Dynamic.Resource(gvr).Namespace(namespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if getErr != nil {
		// Resource doesn't exist, check create permission
		verb = "create"
	}

	// Check RBAC before apply (even in dry-run mode to validate permissions)
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, verb, gvr, namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	fieldManager := args.FieldManager
	if fieldManager == "" {
		fieldManager = "kube-mcp"
	}

	// Build apply options with dry-run support
	applyOptions := metav1.ApplyOptions{
		FieldManager: fieldManager,
	}
	if args.DryRun {
		applyOptions.DryRun = []string{metav1.DryRunAll}
	}

	// Server-side apply
	applied, err := clientSet.Dynamic.Resource(gvr).Namespace(namespace).
		Apply(ctx, obj.GetName(), obj, applyOptions)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to apply resource: %w", err)), nil
	}

	status := "applied"
	if args.DryRun {
		status = "dry-run-applied"
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":      applied.GetName(),
		"namespace": applied.GetNamespace(),
		"kind":      applied.GetKind(),
		"status":    status,
		"dry_run":   args.DryRun,
	})
}

// handleResourcesDelete handles the resources_delete tool.
func (t *Toolset) handleResourcesDelete(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	DryRun    bool   `json:"dry_run"`
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

	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource

	// Check RBAC before deletion (even in dry-run mode to validate permissions)
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "delete", gvr, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	// Build delete options with dry-run support
	deleteOptions := metav1.DeleteOptions{}
	if args.DryRun {
		deleteOptions.DryRun = []string{metav1.DryRunAll}
	}

	err = clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Delete(ctx, args.Name, deleteOptions)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to delete resource: %w", err)), nil
	}

	message := fmt.Sprintf("Resource %s/%s deleted successfully", args.Namespace, args.Name)
	if args.DryRun {
		message = fmt.Sprintf("Resource %s/%s would be deleted (dry-run)", args.Namespace, args.Name)
	}

	return mcpHelpers.NewTextResult(message), nil
}

// handleResourcesScale is now implemented in scale.go
// This function is kept for backward compatibility but delegates to the scale implementation
