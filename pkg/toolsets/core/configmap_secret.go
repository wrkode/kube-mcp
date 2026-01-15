package core

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleConfigMapsGetData handles the configmaps_get_data tool.
func (t *Toolset) handleConfigMapsGetData(ctx context.Context, args struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"` // Optional: specific keys to retrieve
	Context   string   `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	configMap, err := clientSet.Typed.CoreV1().ConfigMaps(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ConfigMap: %w", err)), nil
	}

	result := map[string]interface{}{
		"name":      configMap.Name,
		"namespace": configMap.Namespace,
	}

	// If specific keys requested, return only those
	if len(args.Keys) > 0 {
		filteredData := make(map[string]string)
		for _, key := range args.Keys {
			if value, exists := configMap.Data[key]; exists {
				filteredData[key] = value
			}
		}
	result["data"] = filteredData
	result["keys_requested"] = args.Keys
	result["keys_found"] = len(filteredData)
	} else {
		result["data"] = configMap.Data
		result["binary_data"] = make(map[string]string) // ConfigMaps don't expose binaryData in Data field
	}

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(result)
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

// handleConfigMapsSetData handles the configmaps_set_data tool.
func (t *Toolset) handleConfigMapsSetData(ctx context.Context, args struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	Merge     bool              `json:"merge"` // If true, merge with existing data; if false, replace
	Context   string            `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Check RBAC
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "update", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	// Get existing ConfigMap
	configMap, err := clientSet.Typed.CoreV1().ConfigMaps(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ConfigMap: %w", err)), nil
	}

	// Merge or replace data
	if args.Merge {
		// Merge with existing data
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		for k, v := range args.Data {
			configMap.Data[k] = v
		}
	} else {
		// Replace data
		configMap.Data = args.Data
	}

	// Update ConfigMap
	updated, err := clientSet.Typed.CoreV1().ConfigMaps(args.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to update ConfigMap: %w", err)), nil
	}

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(map[string]interface{}{
		"name":      updated.Name,
		"namespace": updated.Namespace,
		"status":    "updated",
		"data_keys": len(updated.Data),
	})
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

// handleSecretsGetData handles the secrets_get_data tool.
func (t *Toolset) handleSecretsGetData(ctx context.Context, args struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"` // Optional: specific keys to retrieve
	Decode    bool     `json:"decode"` // If true, base64 decode the values
	Context   string   `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	secret, err := clientSet.Typed.CoreV1().Secrets(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Secret: %w", err)), nil
	}

	result := map[string]interface{}{
		"name":      secret.Name,
		"namespace": secret.Namespace,
		"type":      string(secret.Type),
	}

	// Handle data retrieval
	data := make(map[string]string)
	if len(args.Keys) > 0 {
		// Specific keys requested
		for _, key := range args.Keys {
			if value, exists := secret.Data[key]; exists {
				if args.Decode {
					data[key] = string(value)
				} else {
					data[key] = base64.StdEncoding.EncodeToString(value)
				}
			}
		}
		result["keys_requested"] = args.Keys
		result["keys_found"] = len(data)
	} else {
		// All keys
		for k, v := range secret.Data {
			if args.Decode {
				data[k] = string(v)
			} else {
				data[k] = base64.StdEncoding.EncodeToString(v)
			}
		}
	}

	result["data"] = data
	result["decoded"] = args.Decode

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(result)
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

// handleSecretsSetData handles the secrets_set_data tool.
func (t *Toolset) handleSecretsSetData(ctx context.Context, args struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"` // Values should be base64 encoded or plain text (will be encoded)
	Merge     bool              `json:"merge"` // If true, merge with existing data; if false, replace
	Encode    bool              `json:"encode"` // If true, base64 encode the provided values; if false, assume already encoded
	Context   string            `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Check RBAC
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "update", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	// Get existing Secret
	secret, err := clientSet.Typed.CoreV1().Secrets(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Secret: %w", err)), nil
	}

	// Prepare data map
	secretData := make(map[string][]byte)
	if args.Merge && secret.Data != nil {
		// Copy existing data
		for k, v := range secret.Data {
			secretData[k] = v
		}
	}

	// Process new data
	for k, v := range args.Data {
		if args.Encode {
			// Encode plain text to base64
			secretData[k] = []byte(v)
		} else {
			// Assume already base64 encoded, decode then store as bytes
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to decode base64 value for key %q: %w", k, err)), nil
			}
			secretData[k] = decoded
		}
	}

	secret.Data = secretData

	// Update Secret
	updated, err := clientSet.Typed.CoreV1().Secrets(args.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to update Secret: %w", err)), nil
	}

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(map[string]interface{}{
		"name":      updated.Name,
		"namespace": updated.Namespace,
		"status":    "updated",
		"data_keys": len(updated.Data),
	})
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

