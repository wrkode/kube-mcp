# kube-mcp Usage Guide

## Overview

This guide provides practical examples and patterns for using kube-mcp effectively. It covers common workflows, best practices, and advanced usage patterns.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Resource Management](#resource-management)
4. [Monitoring and Debugging](#monitoring-and-debugging)
5. [Advanced Features](#advanced-features)
6. [Integration Patterns](#integration-patterns)
7. [Troubleshooting](#troubleshooting)

## Getting Started

### Basic Tool Call Pattern

All kube-mcp tools follow a consistent pattern:

```json
{
  "tool": "toolset_tool_name",
  "params": {
    "param1": "value1",
    "param2": "value2"
  }
}
```

### Response Format

**Success Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "{ ... tool-specific data ... }"
    }
  ],
  "isError": false
}
```

**Error Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "{ \"error\": { \"type\": \"ErrorType\", \"message\": \"...\" } }"
    }
  ],
  "isError": true
}
```

## Common Workflows

### 1. Listing Resources

**List all pods:**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": ""
  }
}
```

**List pods in specific namespace:**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": "production"
  }
}
```

**List pods with label selector:**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": "default",
    "label_selector": "app=nginx"
  }
}
```

**List deployments:**
```json
{
  "tool": "resources_list",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default"
  }
}
```

### 2. Getting Resource Details

**Get pod details:**
```json
{
  "tool": "pods_get",
  "params": {
    "name": "nginx-deployment-abc123",
    "namespace": "default"
  }
}
```

**Get deployment details:**
```json
{
  "tool": "resources_get",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default"
  }
}
```

**Get resource with describe output:**
```json
{
  "tool": "resources_describe",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default"
  }
}
```

### 3. Creating Resources

**Create a deployment:**
```json
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "nginx-deployment",
        "namespace": "default"
      },
      "spec": {
        "replicas": 3,
        "selector": {
          "matchLabels": {
            "app": "nginx"
          }
        },
        "template": {
          "metadata": {
            "labels": {
              "app": "nginx"
            }
          },
          "spec": {
            "containers": [
              {
                "name": "nginx",
                "image": "nginx:1.21"
              }
            ]
          }
        }
      }
    }
  }
}
```

**Validate before applying (dry-run):**
```json
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* deployment manifest */ },
    "dry_run": true
  }
}
```

### 4. Updating Resources

**Update using server-side apply:**
```json
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "nginx-deployment",
        "namespace": "default"
      },
      "spec": {
        "replicas": 5  // Update replica count
      }
    }
  }
}
```

**Patch specific fields:**
```json
{
  "tool": "resources_patch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "patch_type": "merge",
    "patch_data": {
      "spec": {
        "replicas": 5
      }
    }
  }
}
```

**JSON Patch:**
```json
{
  "tool": "resources_patch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "patch_type": "json",
    "patch_data": [
      {
        "op": "replace",
        "path": "/spec/replicas",
        "value": 5
      }
    ]
  }
}
```

### 5. Scaling Resources

**Get current scale:**
```json
{
  "tool": "resources_scale",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default"
  }
}
```

**Scale to specific replica count:**
```json
{
  "tool": "resources_scale",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "replicas": 5
  }
}
```

**Scale to zero:**
```json
{
  "tool": "resources_scale",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "replicas": 0
  }
}
```

### 6. Deleting Resources

**Delete a pod:**
```json
{
  "tool": "pods_delete",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default"
  }
}
```

**Delete a resource:**
```json
{
  "tool": "resources_delete",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default"
  }
}
```

**Validate deletion (dry-run):**
```json
{
  "tool": "resources_delete",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "dry_run": true
  }
}
```

## Resource Management

### Comparing Resources

**Diff current vs desired state:**
```json
{
  "tool": "resources_diff",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "manifest": {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "nginx-deployment",
        "namespace": "default"
      },
      "spec": {
        "replicas": 5  // Desired: 5, Current: 3
      }
    },
    "format": "unified"
  }
}
```

**JSON diff format:**
```json
{
  "tool": "resources_diff",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "manifest": { /* desired manifest */ },
    "format": "json"
  }
}
```

### Validating Resources

**Validate a manifest:**
```json
{
  "tool": "resources_validate",
  "params": {
    "manifest": {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "nginx-deployment"
      },
      "spec": { /* deployment spec */ }
    }
  }
}
```

### Watching Resources

**Watch for changes:**
```json
{
  "tool": "resources_watch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "timeout": 30
  }
}
```

**Watch with label selector:**
```json
{
  "tool": "resources_watch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "label_selector": "app=nginx",
    "timeout": 60
  }
}
```

### Finding Resource Relationships

**Find owners and dependents:**
```json
{
  "tool": "resources_relationships",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "ReplicaSet",
    "name": "nginx-deployment-abc123",
    "namespace": "default"
  }
}
```

## Monitoring and Debugging

### Pod Logs

**Get pod logs:**
```json
{
  "tool": "pods_logs",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default"
  }
}
```

**Get logs with tail:**
```json
{
  "tool": "pods_logs",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default",
    "tail_lines": 100
  }
}
```

**Get logs since time:**
```json
{
  "tool": "pods_logs",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default",
    "since": "1h"
  }
}
```

**Follow logs (streaming):**
```json
{
  "tool": "pods_logs",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default",
    "follow": true
  }
}
```

### Pod Execution

**Execute command in pod:**
```json
{
  "tool": "pods_exec",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default",
    "container": "nginx",
    "command": ["sh", "-c", "echo 'Hello from pod'"]
  }
}
```

### Metrics

**Get pod metrics:**
```json
{
  "tool": "pods_top",
  "params": {
    "namespace": "default"
  }
}
```

**Get node metrics:**
```json
{
  "tool": "nodes_top",
  "params": {}
}
```

**Get node summary:**
```json
{
  "tool": "nodes_summary",
  "params": {}
}
```

### Events

**List events:**
```json
{
  "tool": "events_list",
  "params": {
    "namespace": "default"
  }
}
```

**List events for specific resource:**
```json
{
  "tool": "events_list",
  "params": {
    "namespace": "default",
    "field_selector": "involvedObject.name=nginx-deployment"
  }
}
```

## Advanced Features

### ConfigMap and Secret Management

**Get ConfigMap data:**
```json
{
  "tool": "configmaps_get_data",
  "params": {
    "name": "app-config",
    "namespace": "default"
  }
}
```

**Set ConfigMap data:**
```json
{
  "tool": "configmaps_set_data",
  "params": {
    "name": "app-config",
    "namespace": "default",
    "data": {
      "key1": "value1",
      "key2": "value2"
    },
    "merge": true
  }
}
```

**Get Secret data:**
```json
{
  "tool": "secrets_get_data",
  "params": {
    "name": "app-secret",
    "namespace": "default"
  }
}
```

**Set Secret data:**
```json
{
  "tool": "secrets_set_data",
  "params": {
    "name": "app-secret",
    "namespace": "default",
    "data": {
      "password": "base64encodedvalue"
    }
  }
}
```

### Port Forwarding

**Set up port forwarding:**
```json
{
  "tool": "pods_port_forward",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default",
    "local_port": 8080,
    "pod_port": 80
  }
}
```

Note: Active port forwarding requires HTTP transport or kubectl.

### Pagination

**List resources with pagination:**
```json
{
  "tool": "resources_list",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "limit": 10,
    "continue": ""  // Use continue token from previous response
  }
}
```

## Integration Patterns

### Error Handling

Always check for errors in responses:

```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "{\"error\":{\"type\":\"KubernetesError\",\"message\":\"...\"}}"
    }
  ]
}
```

Common error types:
- `KubernetesError` - Kubernetes API errors
- `UnknownContext` - Cluster context not found
- `MetricsUnavailable` - Metrics server not available
- `ScalingNotSupported` - Resource doesn't support scaling

### Multi-Cluster Operations

**Switch between clusters:**
```json
// Dev cluster
{
  "tool": "pods_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}

// Prod cluster
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",
    "namespace": "production"
  }
}
```

### Batch Operations

**Pattern: List then operate**
```json
// 1. List resources
{
  "tool": "resources_list",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "label_selector": "app=nginx"
  }
}

// 2. Operate on each resource
{
  "tool": "resources_scale",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment-1",
    "namespace": "default",
    "replicas": 3
  }
}
```

### Workflow: Deploy and Verify

```json
// 1. Validate manifest
{
  "tool": "resources_validate",
  "params": {
    "manifest": { /* deployment manifest */ }
  }
}

// 2. Dry-run apply
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* deployment manifest */ },
    "dry_run": true
  }
}

// 3. Apply for real
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* deployment manifest */ }
  }
}

// 4. Watch for ready
{
  "tool": "resources_watch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "timeout": 60
  }
}

// 5. Verify pods
{
  "tool": "pods_list",
  "params": {
    "namespace": "default",
    "label_selector": "app=nginx"
  }
}
```

## Troubleshooting

### Common Issues

**1. Resource Not Found**
- Verify namespace exists
- Check resource name spelling
- Ensure resource exists in cluster

**2. Permission Denied**
- Check RBAC permissions
- Verify service account permissions (in-cluster)
- Review security configuration

**3. Context Not Found**
- List available contexts: `config_contexts_list`
- Verify context name spelling
- Check kubeconfig path

**4. Metrics Unavailable**
- Ensure metrics-server is installed
- Verify metrics.k8s.io API is available
- Check pod/node metrics permissions

### Debugging Tips

**1. Use describe for detailed info:**
```json
{
  "tool": "resources_describe",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default"
  }
}
```

**2. Check events:**
```json
{
  "tool": "events_list",
  "params": {
    "namespace": "default",
    "field_selector": "involvedObject.name=nginx-deployment"
  }
}
```

**3. Validate before applying:**
```json
{
  "tool": "resources_validate",
  "params": {
    "manifest": { /* your manifest */ }
  }
}
```

**4. Use dry-run:**
```json
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* your manifest */ },
    "dry_run": true
  }
}
```

## Best Practices

1. **Always validate manifests** before applying
2. **Use dry-run** for destructive operations
3. **Specify context** explicitly in multi-cluster scenarios
4. **Handle errors** appropriately in your client code
5. **Use label selectors** for efficient filtering
6. **Enable RBAC checks** in production
7. **Monitor with events** and metrics
8. **Use pagination** for large result sets
9. **Watch resources** for real-time updates
10. **Check relationships** to understand dependencies

## Summary

kube-mcp provides a comprehensive set of tools for Kubernetes management. This guide covers:
- Common workflows and patterns
- Resource management operations
- Monitoring and debugging techniques
- Advanced features and integrations
- Troubleshooting tips

For detailed tool documentation, see [Tools Reference](TOOLS.md) and individual toolset documentation in [tools/](tools/).

