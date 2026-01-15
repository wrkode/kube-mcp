package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// unmarshalArgs unmarshals args from map[string]interface{} to the target struct type.
func unmarshalArgs[T any](args any) (T, error) {
	var result T
	if args == nil {
		return result, nil
	}

	// If args is already the correct type, return it
	if typed, ok := args.(T); ok {
		return typed, nil
	}

	// If args is a map, unmarshal it
	if argsMap, ok := args.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(argsMap)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal(jsonData, &result)
		return result, err
	}

	// Try to marshal and unmarshal as a fallback
	jsonData, err := json.Marshal(args)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// wrapToolHandler wraps a tool handler with observability.
func (t *Toolset) wrapToolHandler(
	toolName string,
	handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error),
	getCluster func(args any) string,
) func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	if t.logger == nil || t.metrics == nil {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		start := time.Now()
		cluster := getCluster(args)
		if cluster == "" {
			cluster = "default"
		}

		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.logger.Error(ctx, "Panic in tool handler",
					"tool", toolName,
					"panic", r,
					"cluster", cluster,
				)
			}
		}()

		result, out, err := handler(ctx, req, args)

		// Log and record metrics
		duration := time.Since(start)
		t.logger.LogToolInvocation(ctx, toolName, cluster, duration, err)
		success := err == nil && (result == nil || !result.IsError)
		t.metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		return result, out, err
	}
}
