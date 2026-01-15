# Backup/Restore Toolset

## Overview

The Backup/Restore toolset provides tools for managing Velero backups and restores via CRDs. This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **Required CRD**:
  - `velero.io/v1/Backup` (enables toolset)
- **Optional CRDs**:
  - `velero.io/v1/Restore`
  - `velero.io/v1/BackupStorageLocation`
  - `velero.io/v1/Schedule`

## Tools

### backup.backups_list

**Description**: List Velero backups.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (CRDs required)

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `namespace` | string | No | Namespace name (empty for all) |
| `label_selector` | string | No | Label selector |
| `limit` | integer | No | Maximum number of items |
| `continue` | string | No | Pagination token |

#### Output Schema

```json
{
  "items": [
    {
      "name": "backup-1234567890",
      "namespace": "velero",
      "phase": "Completed",
      "start_timestamp": "2024-01-15T10:00:00Z",
      "completion_timestamp": "2024-01-15T10:05:00Z",
      "expiration": "2024-02-15T10:00:00Z",
      "errors": 0,
      "warnings": 0
    }
  ]
}
```

### backup.backup_get

**Description**: Get backup details.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (CRDs required)

### backup.backup_create

**Description**: Create a new backup.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (CRDs required)  
**Requires**: `confirm: true`, RBAC check

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `name` | string | No | Backup name (auto-generated if not provided) |
| `namespace` | string | Yes | Namespace name |
| `ttl` | string | No | Time to live (e.g., "720h0m0s") |
| `included_namespaces` | array | No | Namespaces to include |
| `excluded_namespaces` | array | No | Namespaces to exclude |
| `label_selector` | object | No | Label selector map |
| `snapshot_volumes` | boolean | No | Snapshot volumes |
| `include_cluster_resources` | boolean | No | Include cluster resources |
| `confirm` | boolean | Yes | Must be true to create |

### backup.restores_list

**Description**: List Velero restores (requires Restore CRD).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (Restore CRD required)

### backup.restore_create

**Description**: Create a new restore (requires Restore CRD).

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (Restore CRD required)  
**Requires**: `confirm: true`, RBAC check

### backup.locations_list

**Description**: List backup storage locations (requires BackupStorageLocation CRD).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Backup (BackupStorageLocation CRD required)

## Error Codes

- **FeatureNotInstalled**: Required CRDs not detected
- **KubernetesError**: Kubernetes API errors
