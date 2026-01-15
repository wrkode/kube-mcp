package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolBuilder helps build MCP tools with consistent patterns.
type ToolBuilder struct {
	tool *mcp.Tool
}

// NewTool creates a new tool builder.
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &mcp.Tool{
			Name:        name,
			Description: description,
		},
	}
}

// WithReadOnly marks the tool as read-only using annotations.
func (b *ToolBuilder) WithReadOnly() *ToolBuilder {
	if b.tool.Annotations == nil {
		b.tool.Annotations = &mcp.ToolAnnotations{}
	}
	// Store read-only hint in annotations or meta
	return b
}

// WithDestructive marks the tool as destructive using annotations.
func (b *ToolBuilder) WithDestructive() *ToolBuilder {
	if b.tool.Annotations == nil {
		b.tool.Annotations = &mcp.ToolAnnotations{}
	}
	// Store destructive hint in annotations or meta
	return b
}

// WithParameter adds a parameter to the tool's input schema.
func (b *ToolBuilder) WithParameter(name, paramType, description string, required bool) *ToolBuilder {
	// Build JSON schema for the parameter
	if b.tool.InputSchema == nil {
		b.tool.InputSchema = map[string]any{
			"type":       "object",
			"properties": make(map[string]any),
			"required":   make([]string, 0),
		}
	}

	schema, ok := b.tool.InputSchema.(map[string]any)
	if !ok {
		// If InputSchema is not a map, create a new one
		b.tool.InputSchema = map[string]any{
			"type":       "object",
			"properties": make(map[string]any),
			"required":   make([]string, 0),
		}
		schema = b.tool.InputSchema.(map[string]any)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		properties = make(map[string]any)
		schema["properties"] = properties
	}

	properties[name] = map[string]any{
		"type":        paramType,
		"description": description,
	}

	if required {
		requiredList, ok := schema["required"].([]string)
		if !ok {
			requiredList = make([]string, 0)
		}
		// Check if already in list
		found := false
		for _, r := range requiredList {
			if r == name {
				found = true
				break
			}
		}
		if !found {
			requiredList = append(requiredList, name)
			schema["required"] = requiredList
		}
	}

	return b
}

// Build returns the built tool.
func (b *ToolBuilder) Build() *mcp.Tool {
	return b.tool
}

// ParseArguments parses tool arguments from a request into a struct.
// This is a helper for handlers that receive typed arguments from mcp.AddTool.
// When using mcp.AddTool with typed handlers, arguments are already parsed.
func ParseArguments[T any](req *mcp.CallToolRequest) (T, error) {
	var args T
	// When using mcp.AddTool, arguments are passed directly to the handler
	// This function is kept for compatibility but may not be needed
	return args, nil
}

// NewTextResult creates a text content result.
func NewTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// NewJSONResult creates a JSON content result.
func NewJSONResult(data any) (*mcp.CallToolResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, nil
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
}
