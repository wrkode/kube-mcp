# Security Guide

## Overview

kube-mcp implements multiple layers of security to ensure safe Kubernetes cluster management.

## RBAC Compliance

All operations respect Kubernetes Role-Based Access Control (RBAC):

- **SelfSubjectAccessReview**: Before performing operations, the server checks if the current user/service account has the required permissions
- **Namespace Scoping**: Operations are scoped to the user's accessible namespaces
- **Resource-Level Permissions**: Permissions are checked at the resource level

## Security Modes

### Read-Only Mode

When `security.read_only = true`:
- All write operations are blocked
- Only read operations (list, get) are allowed
- Applies, deletes, and scales are rejected

### Non-Destructive Mode

When `security.non_destructive = true`:
- Read operations are allowed
- Server-side apply operations are allowed
- Delete operations are blocked
- Scale operations are blocked

### Denied GVKs

The `security.denied_gvks` list specifies GroupVersionKinds that cannot be accessed:

```toml
[security]
denied_gvks = [
  "rbac.authorization.k8s.io/v1/ClusterRole",
  "rbac.authorization.k8s.io/v1/ClusterRoleBinding"
]
```

## Authentication

### STDIO Transport
- No authentication required
- Uses the user's kubeconfig credentials
- Suitable for local development

### HTTP Transport
- OAuth2/OIDC authentication supported
- Bearer token authentication
- Token verification against OIDC provider
- Streamable HTTP for efficient bidirectional communication

## Best Practices

1. **Use Read-Only Mode in Production**: Enable read-only mode for production deployments
2. **Restrict GVKs**: Use denied GVKs list to prevent access to sensitive resources
3. **Enable RBAC Checks**: Always enable `require_rbac = true`
4. **Use OAuth for HTTP**: Enable OAuth authentication for HTTP transport
5. **Limit Network Access**: Bind HTTP server to specific interfaces, not 0.0.0.0
6. **Use Service Accounts**: When running in-cluster, use least-privilege service accounts
7. **Audit Logging**: Enable Kubernetes audit logging to track all operations

## Service Account Permissions

When running in-cluster, create a service account with minimal required permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-mcp
  namespace: kube-mcp
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-mcp-reader
rules:
- apiGroups: [""]
  resources: ["pods", "namespaces", "nodes", "events"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-mcp-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-mcp-reader
subjects:
- kind: ServiceAccount
  name: kube-mcp
  namespace: kube-mcp
```

## Network Security

- Use TLS for HTTP transport in production
- Configure firewall rules to restrict access
- Use network policies to limit pod-to-pod communication
- Enable mTLS for service mesh integration

## Configuration Security

- Store sensitive configuration in secrets
- Use environment variables for secrets
- Restrict file permissions on configuration files
- Use configuration encryption for sensitive data

