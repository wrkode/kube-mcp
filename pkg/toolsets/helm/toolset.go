package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Toolset implements the Helm toolset for chart and release management.
type Toolset struct {
	provider kubernetes.ClientProvider
	settings *cli.EnvSettings
	logger   *observability.Logger
	metrics  *observability.Metrics
}

// NewToolset creates a new Helm toolset.
func NewToolset(provider kubernetes.ClientProvider, settings *cli.EnvSettings) *Toolset {
	return &Toolset{
		provider: provider,
		settings: settings,
	}
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

		duration := time.Since(start)
		t.logger.LogToolInvocation(ctx, toolName, cluster, duration, err)
		success := err == nil && (result == nil || !result.IsError)
		t.metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		return result, out, err
	}
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "helm"
}

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	return []*mcp.Tool{
		mcpHelpers.NewTool("helm_install", "Install a Helm chart").
			WithParameter("name", "string", "Release name", true).
			WithParameter("chart", "string", "Chart name or path", true).
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("values", "object", "Chart values", false).
			WithParameter("version", "string", "Chart version", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("helm_releases_list", "List Helm releases").
			WithParameter("namespace", "string", "Namespace (empty for all)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("helm_uninstall", "Uninstall a Helm release").
			WithParameter("name", "string", "Release name", true).
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
	}
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	// Register helm_install
	type HelmInstallArgs struct {
		Name      string                 `json:"name"`
		Chart     string                 `json:"chart"`
		Namespace string                 `json:"namespace"`
		Values    map[string]interface{} `json:"values"`
		Version   string                 `json:"version"`
		Context   string                 `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[HelmInstallArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleInstall(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("helm_install", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[HelmInstallArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "helm_install",
		Description: "Install a Helm chart",
	}, wrappedHandler)

	// Register helm_releases_list
	type HelmReleasesListArgs struct {
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[HelmReleasesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("helm_releases_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[HelmReleasesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "helm_releases_list",
		Description: "List Helm releases",
	}, wrappedHandler)

	// Register helm_uninstall
	type HelmUninstallArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[HelmUninstallArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleUninstall(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("helm_uninstall", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[HelmUninstallArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "helm_uninstall",
		Description: "Uninstall a Helm release",
	}, wrappedHandler)

	return nil
}

// getActionConfig creates an action configuration for the given context.
func (t *Toolset) getActionConfig(ctxName, namespace string) (*action.Configuration, error) {
	clientSet, err := t.provider.GetClientSet(ctxName)
	if err != nil {
		return nil, fmt.Errorf("failed to get client set: %w", err)
	}

	// Create REST client getter from config
	restGetter := &restClientGetter{
		restConfig: clientSet.Config,
		namespace:  namespace,
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		restGetter,
		namespace,
		"secret", // storage driver
		func(format string, v ...interface{}) {
			// Log function - can be customized
		},
	); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	return actionConfig, nil
}

// restClientGetter implements genericclioptions.RESTClientGetter
type restClientGetter struct {
	restConfig *rest.Config
	namespace  string
}

func (r *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.restConfig, nil
}

func (r *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	// Create a minimal kubeconfig from rest config
	config := api.NewConfig()
	config.Clusters["default"] = &api.Cluster{
		Server: r.restConfig.Host,
	}
	config.Contexts["default"] = &api.Context{
		Cluster:  "default",
		AuthInfo: "default",
	}
	config.CurrentContext = "default"
	config.AuthInfos["default"] = &api.AuthInfo{}

	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
}

func (r *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(r.restConfig)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(discoveryClient), nil
}

func (r *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDiscoveryRESTMapper(groupResources), nil
}

// handleInstall handles the helm_install tool.
func (t *Toolset) handleInstall(ctx context.Context, args struct {
	Name      string                 `json:"name"`
	Chart     string                 `json:"chart"`
	Namespace string                 `json:"namespace"`
	Values    map[string]interface{} `json:"values"`
	Version   string                 `json:"version"`
	Context   string                 `json:"context"`
}) (*mcp.CallToolResult, error) {

	actionConfig, err := t.getActionConfig(args.Context, args.Namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(err), nil
	}

	installAction := action.NewInstall(actionConfig)
	installAction.ReleaseName = args.Name
	installAction.Namespace = args.Namespace
	installAction.Version = args.Version

	cp, err := installAction.ChartPathOptions.LocateChart(args.Chart, t.settings)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to locate chart: %w", err)), nil
	}

	chrt, err := loader.Load(cp)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to load chart: %w", err)), nil
	}

	release, err := installAction.RunWithContext(ctx, chrt, args.Values)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to install chart: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":      release.Name,
		"namespace": release.Namespace,
		"status":    release.Info.Status.String(),
		"version":   release.Version,
	})
}

// handleList handles the helm_releases_list tool.
func (t *Toolset) handleList(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {

	actionConfig, err := t.getActionConfig(args.Context, args.Namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(err), nil
	}

	listAction := action.NewList(actionConfig)
	listAction.AllNamespaces = args.Namespace == ""

	releases, err := listAction.Run()
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list releases: %w", err)), nil
	}

	releaseList := make([]map[string]any, 0, len(releases))
	for _, release := range releases {
		releaseList = append(releaseList, map[string]any{
			"name":      release.Name,
			"namespace": release.Namespace,
			"status":    release.Info.Status.String(),
			"version":   release.Version,
			"chart":     release.Chart.Metadata.Name + "-" + release.Chart.Metadata.Version,
		})
	}

	return mcpHelpers.NewJSONResult(map[string]any{"releases": releaseList})
}

// handleUninstall handles the helm_uninstall tool.
func (t *Toolset) handleUninstall(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {

	actionConfig, err := t.getActionConfig(args.Context, args.Namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(err), nil
	}

	uninstallAction := action.NewUninstall(actionConfig)
	release, err := uninstallAction.Run(args.Name)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to uninstall release: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":    release.Release.Name,
		"status":  "uninstalled",
		"message": release.Info,
	})
}
