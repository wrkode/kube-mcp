package capi

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ClusterSummary represents a normalized Cluster summary.
type ClusterSummary struct {
	Name                string           `json:"name"`
	Namespace           string           `json:"namespace"`
	Ready               bool             `json:"ready"`
	ControlPlaneReady   bool             `json:"control_plane_ready"`
	InfrastructureReady bool             `json:"infrastructure_ready"`
	KubernetesVersion   string           `json:"kubernetes_version,omitempty"`
	ControlPlaneRef     *ObjectReference `json:"control_plane_ref,omitempty"`
	InfrastructureRef   *ObjectReference `json:"infrastructure_ref,omitempty"`
	Conditions          []Condition      `json:"conditions,omitempty"`
}

// ObjectReference represents a reference to another object.
type ObjectReference struct {
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
}

// Condition represents a normalized condition.
type Condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Time    string `json:"time,omitempty"` // RFC3339 string
}

// MachineSummary represents a normalized Machine summary.
type MachineSummary struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Phase      string      `json:"phase,omitempty"`
	Ready      bool        `json:"ready"`
	NodeRef    string      `json:"node_ref,omitempty"`
	Version    string      `json:"version,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// MachineDeploymentSummary represents a normalized MachineDeployment summary.
type MachineDeploymentSummary struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	ReplicasDesired int         `json:"replicas_desired"`
	ReplicasReady   int         `json:"replicas_ready"`
	ReplicasUpdated int         `json:"replicas_updated"`
	Paused          bool        `json:"paused,omitempty"`
	Conditions      []Condition `json:"conditions,omitempty"`
}

// RolloutStatus represents the rollout status of a cluster.
type RolloutStatus struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
	Counts  struct {
		MachinesDesired int `json:"machines_desired"`
		MachinesReady   int `json:"machines_ready"`
		MachinesUpdated int `json:"machines_updated"`
	} `json:"counts"`
	Blockers []string `json:"blockers,omitempty"`
}

// GVKs for CAPI CRDs
var (
	ClusterGVK = schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "Cluster",
	}
	MachineGVK = schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "Machine",
	}
	MachineDeploymentGVK = schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "MachineDeployment",
	}
)
