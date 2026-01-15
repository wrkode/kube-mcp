package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServeStdio runs the server on STDIO transport.
func ServeStdio(ctx context.Context, server *Server) error {
	transport := &mcp.StdioTransport{}
	return server.Run(ctx, transport)
}
