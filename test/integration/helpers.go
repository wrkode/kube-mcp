package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// int32Ptr returns a pointer to an int32.
func int32Ptr(i int32) *int32 {
	return &i
}

// int64Ptr returns a pointer to an int64.
func int64Ptr(i int64) *int64 {
	return &i
}

// stringPtr returns a pointer to a string.
func stringPtr(s string) *string {
	return &s
}

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}

// toMap converts an object to map[string]any.
func toMap(obj interface{}) (map[string]any, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(data, &result)
	return result, err
}

// mustParseTime parses an RFC3339 time string or panics.
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time %q: %v", s, err))
	}
	return t
}

// createUnstructured creates an unstructured resource using the dynamic client.
func createUnstructured(ctx context.Context, clientSet *kubernetes.ClientSet, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return clientSet.Dynamic.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
	}
	return clientSet.Dynamic.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// updateUnstructured updates an unstructured resource using the dynamic client.
func updateUnstructured(ctx context.Context, clientSet *kubernetes.ClientSet, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return clientSet.Dynamic.Resource(gvr).Update(ctx, obj, metav1.UpdateOptions{})
	}
	return clientSet.Dynamic.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

// getUnstructured gets an unstructured resource using the dynamic client.
func getUnstructured(ctx context.Context, clientSet *kubernetes.ClientSet, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return clientSet.Dynamic.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	return clientSet.Dynamic.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// listUnstructured lists unstructured resources using the dynamic client.
func listUnstructured(ctx context.Context, clientSet *kubernetes.ClientSet, gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	if namespace == "" {
		return clientSet.Dynamic.Resource(gvr).List(ctx, opts)
	}
	return clientSet.Dynamic.Resource(gvr).Namespace(namespace).List(ctx, opts)
}
