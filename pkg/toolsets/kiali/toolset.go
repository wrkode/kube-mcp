package kiali

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/config"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
)

// Toolset implements the Kiali toolset for service mesh observability.
type Toolset struct {
	client  *KialiClient
	enabled bool
	logger  *observability.Logger
	metrics *observability.Metrics
}

// NewToolset creates a new Kiali toolset.
func NewToolset(cfg *config.KialiConfig) (*Toolset, error) {
	if !cfg.Enabled {
		return &Toolset{enabled: false}, nil
	}

	client, err := NewKialiClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kiali client: %w", err)
	}

	return &Toolset{
		client:  client,
		enabled: true,
	}, nil
}

// SetObservability sets the observability components for the toolset.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

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

// IsEnabled returns whether the Kiali toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "kiali"
}

// Tools returns all tools in this toolset (only if enabled).
func (t *Toolset) Tools() []*mcp.Tool {
	if !t.enabled {
		return []*mcp.Tool{}
	}

	return []*mcp.Tool{
		mcpHelpers.NewTool("kiali_mesh_graph", "Get service mesh graph").
			WithParameter("namespace", "string", "Namespace", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("kiali_istio_config_get", "Get Istio configuration").
			WithParameter("namespace", "string", "Namespace", false).
			WithParameter("object_type", "string", "Object type", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("kiali_metrics", "Get metrics").
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("service", "string", "Service name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("kiali_logs", "Get logs").
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("workload", "string", "Workload name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("kiali_traces", "Get traces").
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("service", "string", "Service name", false).
			WithReadOnly().
			Build(),
	}
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	var handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)
	var wrappedHandler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)

	// Register kiali_mesh_graph
	type MeshGraphArgs struct {
		Namespace string `json:"namespace"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[MeshGraphArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleMeshGraph(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kiali_mesh_graph", handler, func(args any) string {
		return "default"
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kiali_mesh_graph",
		Description: "Get service mesh graph",
	}, wrappedHandler)

	// Register kiali_istio_config_get
	type IstioConfigGetArgs struct {
		Namespace  string `json:"namespace"`
		ObjectType string `json:"object_type"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[IstioConfigGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleIstioConfigGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kiali_istio_config_get", handler, func(args any) string {
		return "default"
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kiali_istio_config_get",
		Description: "Get Istio configuration",
	}, wrappedHandler)

	// Register kiali_metrics
	type MetricsArgs struct {
		Namespace string `json:"namespace"`
		Service   string `json:"service"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[MetricsArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleMetrics(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kiali_metrics", handler, func(args any) string {
		return "default"
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kiali_metrics",
		Description: "Get metrics",
	}, wrappedHandler)

	// Register kiali_logs
	type LogsArgs struct {
		Namespace string `json:"namespace"`
		Workload  string `json:"workload"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[LogsArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleLogs(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kiali_logs", handler, func(args any) string {
		return "default"
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kiali_logs",
		Description: "Get logs",
	}, wrappedHandler)

	// Register kiali_traces
	type TracesArgs struct {
		Namespace string `json:"namespace"`
		Service   string `json:"service"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[TracesArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleTraces(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kiali_traces", handler, func(args any) string {
		return "default"
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kiali_traces",
		Description: "Get traces",
	}, wrappedHandler)

	return nil
}

// handleMeshGraph handles the kiali_mesh_graph tool.
func (t *Toolset) handleMeshGraph(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	graph, err := t.client.GetMeshGraph(ctx, args.Namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get mesh graph: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(graph)
}

// handleIstioConfigGet handles the kiali_istio_config_get tool.
func (t *Toolset) handleIstioConfigGet(ctx context.Context, args struct {
	Namespace  string `json:"namespace"`
	ObjectType string `json:"object_type"`
}) (*mcp.CallToolResult, error) {
	config, err := t.client.GetIstioConfig(ctx, args.Namespace, args.ObjectType)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Istio config: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(config)
}

// handleMetrics handles the kiali_metrics tool.
func (t *Toolset) handleMetrics(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}) (*mcp.CallToolResult, error) {
	metrics, err := t.client.GetMetrics(ctx, args.Namespace, args.Service)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get metrics: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(metrics)
}

// handleLogs handles the kiali_logs tool.
func (t *Toolset) handleLogs(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Workload  string `json:"workload"`
}) (*mcp.CallToolResult, error) {
	logs, err := t.client.GetLogs(ctx, args.Namespace, args.Workload)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get logs: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(logs)
}

// handleTraces handles the kiali_traces tool.
func (t *Toolset) handleTraces(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}) (*mcp.CallToolResult, error) {
	traces, err := t.client.GetTraces(ctx, args.Namespace, args.Service)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get traces: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(traces)
}

// KialiClient provides HTTP client for Kiali API.
type KialiClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewKialiClient creates a new Kiali client with TLS and timeout support.
func NewKialiClient(cfg *config.KialiConfig) (*KialiClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("Kiali URL is required")
	}

	// Validate configuration
	if cfg.Timeout <= 0 {
		cfg.Timeout = config.Duration(30 * time.Second) // Default timeout
	}

	// Setup TLS if configured
	var tlsConfig *tls.Config
	if cfg.TLS.Enabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}

		// Load CA certificate if provided
		if cfg.TLS.CAFile != "" {
			caCert, err := os.ReadFile(cfg.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		// Load client certificate if provided
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	transport := &http.Transport{}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	httpClient := &http.Client{
		Timeout:   cfg.Timeout.Duration(),
		Transport: transport,
	}

	return &KialiClient{
		baseURL:    cfg.URL,
		httpClient: httpClient,
		token:      cfg.Token,
	}, nil
}

// GetMeshGraph gets the service mesh graph.
func (c *KialiClient) GetMeshGraph(ctx context.Context, namespace string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/graph", c.baseURL, namespace)
	return c.get(ctx, url)
}

// GetIstioConfig gets Istio configuration.
func (c *KialiClient) GetIstioConfig(ctx context.Context, namespace, objectType string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/istio", c.baseURL, namespace)
	if objectType != "" {
		url += "/" + objectType
	}
	return c.get(ctx, url)
}

// GetMetrics gets metrics.
func (c *KialiClient) GetMetrics(ctx context.Context, namespace, service string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/services/%s/metrics", c.baseURL, namespace, service)
	return c.get(ctx, url)
}

// GetLogs gets logs.
func (c *KialiClient) GetLogs(ctx context.Context, namespace, workload string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/workloads/%s/logs", c.baseURL, namespace, workload)
	return c.get(ctx, url)
}

// GetTraces gets traces.
func (c *KialiClient) GetTraces(ctx context.Context, namespace, service string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/services/%s/traces", c.baseURL, namespace, service)
	return c.get(ctx, url)
}

// get performs a GET request to the Kiali API with improved error handling.
func (c *KialiClient) get(ctx context.Context, url string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Kiali server at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kiali API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
