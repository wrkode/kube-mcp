package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Toolset represents a collection of related MCP tools.
type Toolset interface {
	// Name returns the name of the toolset.
	Name() string

	// Tools returns all tools provided by this toolset.
	Tools() []*mcp.Tool

	// RegisterTools registers all tools from this toolset with the MCP server.
	RegisterTools(server *mcp.Server) error
}

// ToolRegistry manages tool registration and dispatch.
type ToolRegistry struct {
	toolsets map[string]Toolset
	tools    map[string]*mcp.Tool
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		toolsets: make(map[string]Toolset),
		tools:    make(map[string]*mcp.Tool),
	}
}

// RegisterToolset registers a toolset.
func (r *ToolRegistry) RegisterToolset(toolset Toolset) error {
	name := toolset.Name()
	if _, exists := r.toolsets[name]; exists {
		return &ErrToolsetExists{Name: name}
	}

	r.toolsets[name] = toolset

	// Store tools (with normalized names if needed)
	for _, tool := range toolset.Tools() {
		r.tools[tool.Name] = tool
	}

	return nil
}

// GetTool returns a tool by name.
func (r *ToolRegistry) GetTool(name string) (*mcp.Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// ListTools returns all registered tools.
// Note: Tool name normalization happens when tools are registered via AddTool wrapper.
func (r *ToolRegistry) ListTools() []*mcp.Tool {
	tools := make([]*mcp.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ErrToolsetExists is returned when trying to register a toolset that already exists.
type ErrToolsetExists struct {
	Name string
}

func (e *ErrToolsetExists) Error() string {
	return "toolset already exists: " + e.Name
}

// ErrToolNotFound is returned when a tool is not found.
type ErrToolNotFound struct {
	Name string
}

func (e *ErrToolNotFound) Error() string {
	return "tool not found: " + e.Name
}
