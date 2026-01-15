package certs

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// Tools returns all tools in this toolset (only if enabled).
func (t *Toolset) Tools() []*mcp.Tool {
	if !t.enabled {
		return []*mcp.Tool{}
	}

	tools := []*mcp.Tool{}

	// certs.certificates_list
	tools = append(tools, mcpHelpers.NewTool("certs.certificates_list", "List Cert-Manager certificates").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// certs.certificate_get
	tools = append(tools, mcpHelpers.NewTool("certs.certificate_get", "Get certificate details").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "Certificate name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("raw", "boolean", "Return raw object if true", false).
		WithReadOnly().
		Build())

	// certs.issuers_list
	if t.hasIssuer || t.hasClusterIssuer {
		tools = append(tools, mcpHelpers.NewTool("certs.issuers_list", "List Cert-Manager issuers and cluster issuers").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces, ignored for ClusterIssuer)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithReadOnly().
			Build())
	}

	// certs.status_explain
	tools = append(tools, mcpHelpers.NewTool("certs.status_explain", "Explain certificate status and provide diagnosis hints").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "Certificate name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithReadOnly().
		Build())

	// certs.renew
	tools = append(tools, mcpHelpers.NewTool("certs.renew", "Trigger certificate renewal").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "Certificate name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("confirm", "boolean", "Must be true to renew", true).
		WithDestructive().
		Build())

	// certs.acme_challenges_list (optional)
	if t.hasChallenge {
		tools = append(tools, mcpHelpers.NewTool("certs.acme_challenges_list", "List ACME challenges").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithReadOnly().
			Build())
	}

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	// Register certs.certificates_list
	type CertificatesListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[CertificatesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleCertificatesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("certs.certificates_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[CertificatesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "certs.certificates_list",
		Description: "List Cert-Manager certificates",
	}, wrappedHandler)

	// Register certs.certificate_get
	type CertificateGetArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[CertificateGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleCertificateGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("certs.certificate_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[CertificateGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "certs.certificate_get",
		Description: "Get certificate details",
	}, wrappedHandler)

	// Register certs.renew
	type CertificateRenewArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[CertificateRenewArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleCertificateRenew(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("certs.renew", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[CertificateRenewArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "certs.renew",
		Description: "Trigger certificate renewal",
	}, wrappedHandler)

	// Register certs.issuers_list
	if t.hasIssuer || t.hasClusterIssuer {
		type IssuersListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
			Limit         int    `json:"limit"`
			Continue      string `json:"continue"`
			Raw           bool   `json:"raw"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[IssuersListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleIssuersList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("certs.issuers_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[IssuersListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "certs.issuers_list",
			Description: "List Cert-Manager issuers and cluster issuers",
		}, wrappedHandler)
	}

	// Register certs.status_explain
	type StatusExplainArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[StatusExplainArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleStatusExplain(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("certs.status_explain", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[StatusExplainArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "certs.status_explain",
		Description: "Explain certificate status and provide diagnosis hints",
	}, wrappedHandler)

	// Register certs.acme_challenges_list
	if t.hasChallenge || t.hasOrder {
		type ACMEChallengesListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
			Limit         int    `json:"limit"`
			Continue      string `json:"continue"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[ACMEChallengesListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleACMEChallengesList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("certs.acme_challenges_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[ACMEChallengesListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "certs.acme_challenges_list",
			Description: "List ACME challenges",
		}, wrappedHandler)
	}

	return nil
}
