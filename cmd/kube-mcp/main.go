package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/http"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"github.com/wrkode/kube-mcp/pkg/toolsets/autoscaling"
	"github.com/wrkode/kube-mcp/pkg/toolsets/backup"
	"github.com/wrkode/kube-mcp/pkg/toolsets/capi"
	"github.com/wrkode/kube-mcp/pkg/toolsets/certs"
	configToolset "github.com/wrkode/kube-mcp/pkg/toolsets/config"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	"github.com/wrkode/kube-mcp/pkg/toolsets/gitops"
	"github.com/wrkode/kube-mcp/pkg/toolsets/helm"
	"github.com/wrkode/kube-mcp/pkg/toolsets/kiali"
	"github.com/wrkode/kube-mcp/pkg/toolsets/kubevirt"
	"github.com/wrkode/kube-mcp/pkg/toolsets/net"
	"github.com/wrkode/kube-mcp/pkg/toolsets/policy"
	"github.com/wrkode/kube-mcp/pkg/toolsets/rollouts"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	configPath  = flag.String("config", "", "Path to configuration file")
	confDPath   = flag.String("conf-d", "", "Path to configuration drop-in directory")
	transport   = flag.String("transport", "", "Transport to use (stdio, http). If not specified, uses config.")
	versionFlag = flag.Bool("version", false, "Print version and exit")
)

const (
	version = "1.0.0"
	name    = "kube-mcp"
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s version %s\n", name, version)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfgLoader := config.NewLoader(*configPath, *confDPath)
	cfg, err := cfgLoader.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup hot reload with callback to apply runtime-reloadable settings
	if err := config.SetupReload(cfgLoader, func(cfg *config.Config) error {
		// Apply runtime-reloadable settings here
		// For now, we'll just update the logger level if needed
		// More sophisticated reload logic can be added later
		return nil
	}); err != nil {
		log.Printf("Warning: Failed to setup hot reload: %v", err)
	}

	// Create Kubernetes client factory
	factory := kubernetes.NewClientFactory(
		cfg.Kubernetes.QPS,
		cfg.Kubernetes.Burst,
		cfg.Kubernetes.Timeout.Duration(),
	)

	// Create Kubernetes provider
	provider, err := kubernetes.NewProvider(
		cfg.Kubernetes.Provider,
		cfg.Kubernetes.KubeconfigPath,
		cfg.Kubernetes.Context,
		factory,
	)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes provider: %v", err)
	}

	// Get default client set for CRD discovery
	defaultClientSet, err := provider.GetClientSet("")
	if err != nil {
		log.Fatalf("Failed to get default client set: %v", err)
	}

	// Create CRD discovery
	crdDiscovery := kubernetes.NewCRDDiscovery(defaultClientSet, 5*time.Minute)
	if err := crdDiscovery.DiscoverCRDs(ctx); err != nil {
		log.Printf("Warning: Failed to discover CRDs: %v", err)
	}

	// Initialize observability
	logLevel := observability.LogLevel(cfg.Server.LogLevel)
	if logLevel == "" {
		logLevel = observability.LogLevelInfo
	}
	obsLogger := observability.NewLogger(logLevel, false) // JSON format can be configurable
	obsMetrics := observability.NewMetrics(nil)           // Use default registry

	// Create MCP server
	mcpServer := mcp.NewServer(name, version, cfg.Server.NormalizeToolNames)

	// Register toolsets with observability
	if err := registerToolsets(mcpServer, provider, crdDiscovery, cfg, obsLogger, obsMetrics, defaultClientSet); err != nil {
		log.Fatalf("Failed to register toolsets: %v", err)
	}

	// Determine transport
	transports := cfg.Server.Transports
	if *transport != "" {
		transports = []string{*transport}
	}

	// Start transports with observability
	if err := startTransports(ctx, mcpServer, cfg, transports, obsLogger, obsMetrics, defaultClientSet); err != nil {
		log.Fatalf("Failed to start transports: %v", err)
	}

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}

// registerToolsets registers all toolsets with the MCP server.
func registerToolsets(
	mcpServer *mcp.Server,
	provider kubernetes.ClientProvider,
	crdDiscovery *kubernetes.CRDDiscovery,
	cfg *config.Config,
	logger *observability.Logger,
	metrics *observability.Metrics,
	defaultClientSet *kubernetes.ClientSet,
) error {
	// Config toolset (always enabled)
	cfgToolset := configToolset.NewToolset(provider)
	if err := mcpServer.RegisterToolset(cfgToolset); err != nil {
		return fmt.Errorf("failed to register config toolset: %w", err)
	}

	// Core toolset (always enabled)
	coreToolset := core.NewToolset(provider)
	coreToolset.SetObservability(logger, metrics)

	// Setup RBAC authorizer for core toolset
	if cfg.Security.RequireRBAC {
		rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
		coreToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
	}

	if err := mcpServer.RegisterToolset(coreToolset); err != nil {
		return fmt.Errorf("failed to register core toolset: %w", err)
	}

	// Helm toolset (always enabled)
	helmSettings := cli.New()
	helmToolset := helm.NewToolset(provider, helmSettings)
	helmToolset.SetObservability(logger, metrics)
	if err := mcpServer.RegisterToolset(helmToolset); err != nil {
		return fmt.Errorf("failed to register helm toolset: %w", err)
	}

	// KubeVirt toolset (conditional)
	if cfg.KubeVirt.Enabled {
		kubevirtToolset := kubevirt.NewToolset(provider, crdDiscovery)
		if kubevirtToolset.IsEnabled() {
			kubevirtToolset.SetObservability(logger, metrics)
			if err := mcpServer.RegisterToolset(kubevirtToolset); err != nil {
				return fmt.Errorf("failed to register kubevirt toolset: %w", err)
			}
			log.Println("KubeVirt toolset enabled")
		}
	}

	// Kiali toolset (conditional)
	if cfg.Kiali.Enabled {
		kialiToolset, err := kiali.NewToolset(&cfg.Kiali)
		if err != nil {
			return fmt.Errorf("failed to create kiali toolset: %w", err)
		}
		if kialiToolset.IsEnabled() {
			kialiToolset.SetObservability(logger, metrics)
			if err := mcpServer.RegisterToolset(kialiToolset); err != nil {
				return fmt.Errorf("failed to register kiali toolset: %w", err)
			}
			log.Println("Kiali toolset enabled")
		}
	}

	// GitOps toolset (conditional)
	if cfg.Toolsets.GitOps.Enabled {
		gitopsToolset := gitops.NewToolset(provider, crdDiscovery)
		if gitopsToolset.IsEnabled() {
			gitopsToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				gitopsToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(gitopsToolset); err != nil {
				return fmt.Errorf("failed to register gitops toolset: %w", err)
			}
			log.Println("GitOps toolset enabled")
		}
	}

	// Policy toolset (conditional)
	if cfg.Toolsets.Policy.Enabled {
		policyToolset := policy.NewToolset(provider, crdDiscovery)
		if policyToolset.IsEnabled() {
			policyToolset.SetObservability(logger, metrics)
			if err := mcpServer.RegisterToolset(policyToolset); err != nil {
				return fmt.Errorf("failed to register policy toolset: %w", err)
			}
			log.Println("Policy toolset enabled")
		}
	}

	// CAPI toolset (conditional)
	if cfg.Toolsets.CAPI.Enabled {
		capiToolset := capi.NewToolset(provider, crdDiscovery)
		if capiToolset.IsEnabled() {
			capiToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				capiToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(capiToolset); err != nil {
				return fmt.Errorf("failed to register capi toolset: %w", err)
			}
			log.Println("CAPI toolset enabled")
		}
	}

	// Rollouts toolset (conditional)
	if cfg.Toolsets.Rollouts.Enabled {
		rolloutsToolset := rollouts.NewToolset(provider, crdDiscovery)
		if rolloutsToolset.IsEnabled() {
			rolloutsToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				rolloutsToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(rolloutsToolset); err != nil {
				return fmt.Errorf("failed to register rollouts toolset: %w", err)
			}
			log.Println("Rollouts toolset enabled")
		}
	}

	// Certs toolset (conditional)
	if cfg.Toolsets.Certs.Enabled {
		certsToolset := certs.NewToolset(provider, crdDiscovery)
		if certsToolset.IsEnabled() {
			certsToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				certsToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(certsToolset); err != nil {
				return fmt.Errorf("failed to register certs toolset: %w", err)
			}
			log.Println("Certs toolset enabled")
		}
	}

	// Autoscaling toolset (conditional)
	if cfg.Toolsets.Autoscaling.Enabled {
		autoscalingToolset := autoscaling.NewToolset(provider, crdDiscovery)
		if autoscalingToolset.IsEnabled() {
			autoscalingToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				autoscalingToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(autoscalingToolset); err != nil {
				return fmt.Errorf("failed to register autoscaling toolset: %w", err)
			}
			log.Println("Autoscaling toolset enabled")
		}
	}

	// Backup toolset (conditional)
	if cfg.Toolsets.Backup.Enabled {
		backupToolset := backup.NewToolset(provider, crdDiscovery)
		if backupToolset.IsEnabled() {
			backupToolset.SetObservability(logger, metrics)
			if cfg.Security.RequireRBAC {
				rbacAuthorizer := kubernetes.NewRBACAuthorizer(defaultClientSet, cfg.Security.RBACCacheTTL)
				backupToolset.SetRBACAuthorizer(rbacAuthorizer, cfg.Security.RequireRBAC)
			}
			if err := mcpServer.RegisterToolset(backupToolset); err != nil {
				return fmt.Errorf("failed to register backup toolset: %w", err)
			}
			log.Println("Backup toolset enabled")
		}
	}

	// Network toolset (conditional)
	if cfg.Toolsets.Net.Enabled {
		netToolset := net.NewToolset(provider, crdDiscovery, cfg.Toolsets.Net)
		if netToolset.IsEnabled() {
			netToolset.SetObservability(logger, metrics)
			if err := mcpServer.RegisterToolset(netToolset); err != nil {
				return fmt.Errorf("failed to register net toolset: %w", err)
			}
			log.Println("Network toolset enabled")
		}
	}

	return nil
}

// startTransports starts the configured transports.
func startTransports(
	ctx context.Context,
	mcpServer *mcp.Server,
	cfg *config.Config,
	transports []string,
	logger *observability.Logger,
	metrics *observability.Metrics,
	defaultClientSet *kubernetes.ClientSet,
) error {
	for _, transportName := range transports {
		switch transportName {
		case "stdio":
			log.Println("Starting STDIO transport...")
			go func() {
				if err := mcp.ServeStdio(ctx, mcpServer); err != nil {
					log.Printf("STDIO transport error: %v", err)
				}
			}()

		case "http":
			log.Printf("Starting HTTP transport on %s...", cfg.Server.HTTP.Address)
			httpServer, err := http.NewServer(mcpServer, &cfg.Server.HTTP, logger, metrics, defaultClientSet, &cfg.Security)
			if err != nil {
				return fmt.Errorf("failed to create HTTP server: %w", err)
			}
			go func() {
				if err := httpServer.Start(); err != nil {
					log.Printf("HTTP server error: %v", err)
				}
			}()

		default:
			return fmt.Errorf("unknown transport: %s", transportName)
		}
	}

	return nil
}
