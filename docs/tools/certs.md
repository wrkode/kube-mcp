# Cert-Manager Toolset

## Overview

The Cert-Manager toolset provides tools for managing certificates and certificate issuers via Cert-Manager CRDs. This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **Required CRD**:
  - `cert-manager.io/v1/Certificate` (enables toolset)
- **Optional CRDs**:
  - `cert-manager.io/v1/Issuer`
  - `cert-manager.io/v1/ClusterIssuer`
  - `cert-manager.io/v1/CertificateRequest`
  - `acme.cert-manager.io/v1/Order`
  - `acme.cert-manager.io/v1/Challenge`

## Tools

### certs.certificates_list

**Description**: List Cert-Manager certificates.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (CRDs required)

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
      "name": "my-cert",
      "namespace": "default",
      "status": "Ready",
      "issuer": "letsencrypt-prod",
      "secret_name": "my-cert-tls",
      "dns_names": ["example.com"],
      "not_after": "2024-12-31T23:59:59Z",
      "not_before": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### certs.certificate_get

**Description**: Get certificate details.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (CRDs required)

### certs.issuers_list

**Description**: List Cert-Manager issuers and cluster issuers.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (Issuer/ClusterIssuer CRDs required)

### certs.status_explain

**Description**: Explain certificate status and provide diagnosis hints.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (CRDs required)

#### Output Schema

```json
{
  "ready": true,
  "status": "Ready",
  "issuer_ref": {"name": "letsencrypt-prod"},
  "secret_name": "my-cert-tls",
  "dns_names": ["example.com"],
  "not_after": "2024-12-31T23:59:59Z",
  "renewal_time": "2024-10-01T00:00:00Z",
  "conditions": [],
  "diagnosis_hints": []
}
```

### certs.renew

**Description**: Trigger certificate renewal (best-effort).

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (CRDs required)  
**Requires**: `confirm: true`, RBAC check

**Implementation**: Uses `cert-manager.io/renew` annotation (best-effort). Note: cert-manager does not have a standard annotation for immediate renewal. For guaranteed renewal, use cert-manager CLI or delete/recreate the CertificateRequest.

### certs.acme_challenges_list

**Description**: List ACME challenges (optional, requires Challenge/Order CRDs).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Certs (Challenge/Order CRDs required)

## Error Codes

- **FeatureNotInstalled**: Required CRDs not detected
- **KubernetesError**: Kubernetes API errors
