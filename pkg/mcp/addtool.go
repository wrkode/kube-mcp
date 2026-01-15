package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddTool wraps the SDK's generic AddTool function to normalize tool names for n8n compatibility.
// IMPORTANT: Toolsets should use this function instead of calling mcp.AddTool directly
// from the SDK package. This ensures tool names are normalized when normalizeToolNames
// is enabled in the configuration.
//
// When normalizeToolNames is enabled, dots in tool names are replaced with underscores.
// For example: "autoscaling.hpa_explain" becomes "autoscaling_hpa_explain"
//
// This is a generic wrapper that matches the SDK's AddTool signature.
func AddTool[In, Out any](server *mcp.Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	// If server is wrapped by our Server, normalize the tool name
	if srv, ok := getServerFromSDK(server); ok && srv.normalizeToolNames {
		// Create a copy of the tool to avoid modifying the original
		normalizedTool := *tool
		originalName := tool.Name
		normalizedTool.Name = srv.normalizeToolName(tool.Name)

		// Store mapping for reverse lookup
		if normalizedTool.Name != originalName {
			srv.nameMapping[normalizedTool.Name] = originalName
		}

		// Register with normalized name
		mcp.AddTool(server, &normalizedTool, handler)
		return
	}

	// No normalization needed, call SDK directly
	mcp.AddTool(server, tool, handler)
}

// getServerFromSDK attempts to find our Server wrapper from the SDK server.
// This is a workaround since we can't easily track the relationship.
// For now, we'll use a global registry approach.
var serverRegistry = make(map[*mcp.Server]*Server)

// registerServerMapping registers a mapping between SDK server and our Server wrapper.
func registerServerMapping(sdkServer *mcp.Server, ourServer *Server) {
	serverRegistry[sdkServer] = ourServer
}

// getServerFromSDK retrieves our Server wrapper from the SDK server.
func getServerFromSDK(sdkServer *mcp.Server) (*Server, bool) {
	srv, ok := serverRegistry[sdkServer]
	return srv, ok
}
