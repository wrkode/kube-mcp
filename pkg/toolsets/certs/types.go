package certs

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CertificateStatus represents the status of a certificate.
type CertificateStatus string

const (
	CertificateStatusReady    CertificateStatus = "Ready"
	CertificateStatusPending  CertificateStatus = "Pending"
	CertificateStatusIssuing  CertificateStatus = "Issuing"
	CertificateStatusFailed   CertificateStatus = "Failed"
	CertificateStatusRenewing CertificateStatus = "Renewing"
	CertificateStatusUnknown  CertificateStatus = "Unknown"
)

// CertificateSummary represents a normalized certificate summary.
type CertificateSummary struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Status      CertificateStatus `json:"status"`
	Issuer      string            `json:"issuer,omitempty"`
	SecretName  string            `json:"secret_name,omitempty"`
	DNSNames    []string          `json:"dns_names,omitempty"`
	NotAfter    *string           `json:"not_after,omitempty"`  // RFC3339 string
	NotBefore   *string           `json:"not_before,omitempty"` // RFC3339 string
	Message     string            `json:"message,omitempty"`
	LastUpdated *string           `json:"last_updated,omitempty"` // RFC3339 string
}

// CertificateDetails represents detailed certificate information.
type CertificateDetails struct {
	CertificateSummary
	Conditions   []map[string]interface{} `json:"conditions,omitempty"`
	RenewalTime  *string                  `json:"renewal_time,omitempty"`
	IssuerRef    map[string]interface{}   `json:"issuer_ref,omitempty"`
	KeyAlgorithm string                   `json:"key_algorithm,omitempty"`
	KeySize      *int                     `json:"key_size,omitempty"`
}

// IssuerSummary represents an issuer summary.
type IssuerSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`           // "Issuer" or "ClusterIssuer"
	Type      string `json:"type,omitempty"` // "ACME", "CA", "Vault", etc.
	Ready     bool   `json:"ready"`
	Message   string `json:"message,omitempty"`
}

// GVKs for Cert-Manager CRDs
var (
	CertificateGVK = schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "Certificate",
	}
	IssuerGVK = schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "Issuer",
	}
	ClusterIssuerGVK = schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "ClusterIssuer",
	}
	CertificateRequestGVK = schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "CertificateRequest",
	}
	OrderGVK = schema.GroupVersionKind{
		Group:   "acme.cert-manager.io",
		Version: "v1",
		Kind:    "Order",
	}
	ChallengeGVK = schema.GroupVersionKind{
		Group:   "acme.cert-manager.io",
		Version: "v1",
		Kind:    "Challenge",
	}
)

// Cert-Manager renewal annotation
// Note: cert-manager does not have a standard annotation for immediate renewal.
// This uses a best-effort approach. For guaranteed renewal, users should use
// cert-manager CLI or delete/recreate the CertificateRequest.
const (
	CertManagerRenewAnnotation = "cert-manager.io/renew"
)
