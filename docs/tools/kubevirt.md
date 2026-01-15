# KubeVirt Toolset

## Overview

The KubeVirt toolset provides tools for managing VirtualMachine resources in Kubernetes clusters where KubeVirt is installed. This toolset is **optional** and only available when KubeVirt CRDs are detected in the cluster.

## Dependencies

- **KubeVirt CRDs**: Requires KubeVirt CRDs to be installed in the cluster
  - `kubevirt.io/v1/VirtualMachine` (required)
  - `cdi.kubevirt.io/v1beta1/DataSource` (optional, enhances functionality)
  - `instancetypes.kubevirt.io/v1beta1/VirtualMachineInstancetype` (optional, enhances functionality)

## Feature Gating

- **Tool Registration**: If KubeVirt CRDs are not detected, **no tools from this toolset are registered**
- **CRD Detection**: The toolset automatically detects available CRDs on startup
- **Error Handling**: Tools return structured errors when CRDs are missing or operations fail

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- Each cluster is checked independently for KubeVirt CRDs

## Tools

### kubevirt_vm_create

**Description**: Create a VirtualMachine resource from a manifest.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `manifest` | object | Yes | - | Complete VirtualMachine manifest |

#### Output Schema

**Success**:
```json
{
  "name": "my-vm",
  "namespace": "default",
  "status": "created"
}
```

**Error**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to create VirtualMachine: VirtualMachine.kubevirt.io \"my-vm\" is invalid",
    "details": "VirtualMachine.kubevirt.io \"my-vm\" is invalid: spec.template.spec.domain.devices.disks: Required value",
    "cluster": "dev-cluster",
    "tool": "kubevirt_vm_create"
  }
}
```

#### Example Call

```json
{
  "tool": "kubevirt_vm_create",
  "params": {
    "context": "dev-cluster",
    "manifest": {
      "apiVersion": "kubevirt.io/v1",
      "kind": "VirtualMachine",
      "metadata": {
        "name": "my-vm",
        "namespace": "default"
      },
      "spec": {
        "running": false,
        "template": {
          "metadata": {
            "labels": {
              "kubevirt.io/vm": "my-vm"
            }
          },
          "spec": {
            "domain": {
              "devices": {
                "disks": [
                  {
                    "name": "disk0",
                    "disk": {
                      "bus": "virtio"
                    }
                  }
                ]
              },
              "resources": {
                "requests": {
                  "memory": "1Gi"
                }
              }
            },
            "volumes": [
              {
                "name": "disk0",
                "containerDisk": {
                  "image": "quay.io/kubevirt/fedora-cloud-container-disk-demo"
                }
              }
            ]
          }
        }
      }
    }
  }
}
```

---

### kubevirt_vm_start

**Description**: Start a VirtualMachine by setting `spec.running` to `true`.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | VirtualMachine name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

```json
{
  "name": "my-vm",
  "action": "start",
  "status": "success"
}
```

#### Example Call

```json
{
  "tool": "kubevirt_vm_start",
  "params": {
    "context": "dev-cluster",
    "name": "my-vm",
    "namespace": "default"
  }
}
```

---

### kubevirt_vm_stop

**Description**: Stop a VirtualMachine by setting `spec.running` to `false`.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | VirtualMachine name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

```json
{
  "name": "my-vm",
  "action": "stop",
  "status": "success"
}
```

#### Example Call

```json
{
  "tool": "kubevirt_vm_stop",
  "params": {
    "context": "dev-cluster",
    "name": "my-vm",
    "namespace": "default"
  }
}
```

---

### kubevirt_vm_restart

**Description**: Restart a VirtualMachine by toggling `spec.running`.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | VirtualMachine name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

```json
{
  "name": "my-vm",
  "action": "restart",
  "status": "success"
}
```

#### Example Call

```json
{
  "tool": "kubevirt_vm_restart",
  "params": {
    "context": "dev-cluster",
    "name": "my-vm",
    "namespace": "default"
  }
}
```

---

### kubevirt_datasources_list

**Description**: List KubeVirt DataSource resources. Requires CDI (Containerized Data Importer) CRDs.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt (CDI CRDs)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace (empty for all namespaces) |

#### Output Schema

```json
{
  "datasources": [
    {
      "name": "fedora-ds",
      "namespace": "default",
      "kind": "DataSource",
      "apiVersion": "cdi.kubevirt.io/v1beta1"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "kubevirt_datasources_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

---

### kubevirt_instancetypes_list

**Description**: List KubeVirt InstanceType resources. Requires InstanceTypes CRDs.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: KubeVirt (InstanceTypes CRDs)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace (empty for all namespaces) |

#### Output Schema

```json
{
  "instancetypes": [
    {
      "name": "small",
      "namespace": "default",
      "kind": "VirtualMachineInstancetype",
      "apiVersion": "instancetypes.kubevirt.io/v1beta1"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "kubevirt_instancetypes_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

## Error Handling

### Toolset Not Available

If KubeVirt CRDs are not detected, the toolset is not registered and tools will not appear in the tool list. This is not an error condition - it's expected behavior when KubeVirt is not installed.

### Common Errors

- **KubernetesError**: Standard Kubernetes API errors (resource not found, validation errors, etc.)
- **FeatureDisabled**: Returned if toolset is disabled in configuration (rare)

### Example Error Responses

**VM Not Found**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to start VM: virtualmachines.kubevirt.io \"my-vm\" not found",
    "details": "virtualmachines.kubevirt.io \"my-vm\" not found",
    "cluster": "dev-cluster",
    "tool": "kubevirt_vm_start"
  }
}
```

**Invalid Manifest**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to create VirtualMachine: VirtualMachine.kubevirt.io \"my-vm\" is invalid",
    "details": "VirtualMachine.kubevirt.io \"my-vm\" is invalid: spec.template.spec.domain.devices.disks: Required value",
    "cluster": "dev-cluster",
    "tool": "kubevirt_vm_create"
  }
}
```

