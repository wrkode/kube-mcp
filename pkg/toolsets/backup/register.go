package backup

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

	// backup.backups_list
	tools = append(tools, mcpHelpers.NewTool("backup.backups_list", "List Velero backups").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// backup.backup_get
	tools = append(tools, mcpHelpers.NewTool("backup.backup_get", "Get backup details").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "Backup name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("raw", "boolean", "Return raw object if true", false).
		WithReadOnly().
		Build())

	// backup.backup_create
	tools = append(tools, mcpHelpers.NewTool("backup.backup_create", "Create a new backup").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "Backup name (optional, auto-generated if not provided)", false).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("ttl", "string", "Time to live (e.g., '720h0m0s')", false).
		WithParameter("included_namespaces", "array", "Namespaces to include", false).
		WithParameter("excluded_namespaces", "array", "Namespaces to exclude", false).
		WithParameter("label_selector", "object", "Label selector map", false).
		WithParameter("snapshot_volumes", "boolean", "Snapshot volumes", false).
		WithParameter("include_cluster_resources", "boolean", "Include cluster resources", false).
		WithParameter("confirm", "boolean", "Must be true to create", true).
		WithDestructive().
		Build())

	// backup.restores_list
	if t.hasRestore {
		tools = append(tools, mcpHelpers.NewTool("backup.restores_list", "List Velero restores").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithReadOnly().
			Build())

		tools = append(tools, mcpHelpers.NewTool("backup.restore_create", "Create a new restore").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("name", "string", "Restore name (optional, auto-generated if not provided)", false).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("backup_name", "string", "Backup name to restore from", true).
			WithParameter("included_namespaces", "array", "Namespaces to include", false).
			WithParameter("excluded_namespaces", "array", "Namespaces to exclude", false).
			WithParameter("confirm", "boolean", "Must be true to create", true).
			WithDestructive().
			Build())
	}

	// backup.locations_list
	if t.hasBackupStorageLocation {
		tools = append(tools, mcpHelpers.NewTool("backup.locations_list", "List backup storage locations").
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

	// Register backup.backups_list
	type BackupsListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[BackupsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleBackupsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("backup.backups_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[BackupsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "backup.backups_list",
		Description: "List Velero backups",
	}, wrappedHandler)

	// Register backup.backup_get
	type BackupGetArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[BackupGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleBackupGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("backup.backup_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[BackupGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "backup.backup_get",
		Description: "Get backup details",
	}, wrappedHandler)

	// Register backup.backup_create
	type BackupCreateArgs struct {
		Context                 string            `json:"context"`
		Name                    string            `json:"name"`
		Namespace               string            `json:"namespace"`
		TTL                     string            `json:"ttl,omitempty"`
		IncludedNamespaces      []string          `json:"included_namespaces,omitempty"`
		ExcludedNamespaces      []string          `json:"excluded_namespaces,omitempty"`
		LabelSelector           map[string]string `json:"label_selector,omitempty"`
		SnapshotVolumes         *bool             `json:"snapshot_volumes,omitempty"`
		IncludeClusterResources *bool             `json:"include_cluster_resources,omitempty"`
		Confirm                 bool              `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[BackupCreateArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleBackupCreate(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("backup.backup_create", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[BackupCreateArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "backup.backup_create",
		Description: "Create a new backup",
	}, wrappedHandler)

	// Register backup.restores_list
	if t.hasRestore {
		type RestoresListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
			Limit         int    `json:"limit"`
			Continue      string `json:"continue"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[RestoresListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleRestoresList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("backup.restores_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[RestoresListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "backup.restores_list",
			Description: "List Velero restores",
		}, wrappedHandler)

		// Register backup.restore_create
		type RestoreCreateArgs struct {
			Context            string   `json:"context"`
			Name               string   `json:"name"`
			Namespace          string   `json:"namespace"`
			BackupName         string   `json:"backup_name"`
			IncludedNamespaces []string `json:"included_namespaces,omitempty"`
			ExcludedNamespaces []string `json:"excluded_namespaces,omitempty"`
			Confirm            bool     `json:"confirm"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[RestoreCreateArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleRestoreCreate(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("backup.restore_create", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[RestoreCreateArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "backup.restore_create",
			Description: "Create a new restore",
		}, wrappedHandler)
	}

	// Register backup.locations_list
	if t.hasBackupStorageLocation {
		type LocationsListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[LocationsListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleLocationsList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("backup.locations_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[LocationsListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "backup.locations_list",
			Description: "List backup storage locations",
		}, wrappedHandler)
	}

	return nil
}
