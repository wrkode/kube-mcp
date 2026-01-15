package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Transport represents an MCP transport mechanism.
type Transport interface {
	mcp.Transport
}

// TransportFactory creates transports.
type TransportFactory interface {
	Create() (Transport, error)
}
