package mcp

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP SDK server and provides toolset management.
type Server struct {
	sdkServer          *mcp.Server
	registry           *ToolRegistry
	implementation     *mcp.Implementation
	normalizeToolNames bool
	nameMapping        map[string]string // normalized -> original name mapping
}

// NewServer creates a new MCP server.
func NewServer(name, version string, normalizeToolNames bool) *Server {
	impl := &mcp.Implementation{
		Name:    name,
		Version: version,
	}

	sdkServer := mcp.NewServer(impl, nil)
	registry := NewToolRegistry()

	srv := &Server{
		sdkServer:          sdkServer,
		registry:           registry,
		implementation:     impl,
		normalizeToolNames: normalizeToolNames,
		nameMapping:        make(map[string]string),
	}

	// Register mapping for AddTool wrapper
	registerServerMapping(sdkServer, srv)

	return srv
}

// normalizeToolName replaces dots with underscores in tool names for n8n compatibility.
func (s *Server) normalizeToolName(name string) string {
	if !s.normalizeToolNames {
		return name
	}
	normalized := strings.ReplaceAll(name, ".", "_")
	if normalized != name {
		s.nameMapping[normalized] = name
	}
	return normalized
}

// GetOriginalToolName returns the original tool name from a normalized name.
func (s *Server) GetOriginalToolName(normalizedName string) string {
	if original, ok := s.nameMapping[normalizedName]; ok {
		return original
	}
	return normalizedName
}

// RegisterToolset registers a toolset with the server.
func (s *Server) RegisterToolset(toolset Toolset) error {
	if err := s.registry.RegisterToolset(toolset); err != nil {
		return err
	}

	// Register tools with the SDK server
	// Note: Tool name normalization requires toolsets to use mcp.AddTool wrapper
	// For now, we'll normalize tool names after registration by intercepting
	// the tool list. This is a workaround until all toolsets are updated.
	return toolset.RegisterTools(s.sdkServer)
}

// GetSDKServer returns the underlying MCP SDK server.
func (s *Server) GetSDKServer() *mcp.Server {
	return s.sdkServer
}

// Run runs the server on the given transport.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.sdkServer.Run(ctx, transport)
}

// ListTools returns all registered tools from the registry.
func (s *Server) ListTools() []*mcp.Tool {
	return s.registry.ListTools()
}
