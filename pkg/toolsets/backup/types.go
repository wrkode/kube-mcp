package backup

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// BackupPhase represents the phase of a backup.
type BackupPhase string

const (
	BackupPhaseNew             BackupPhase = "New"
	BackupPhaseInProgress      BackupPhase = "InProgress"
	BackupPhaseCompleted       BackupPhase = "Completed"
	BackupPhaseFailed          BackupPhase = "Failed"
	BackupPhasePartiallyFailed BackupPhase = "PartiallyFailed"
	BackupPhaseUnknown         BackupPhase = "Unknown"
)

// BackupSummary represents a normalized Velero backup summary.
type BackupSummary struct {
	Name                string      `json:"name"`
	Namespace           string      `json:"namespace"`
	Phase               BackupPhase `json:"phase"`
	StartTimestamp      *string     `json:"start_timestamp,omitempty"`      // RFC3339 string
	CompletionTimestamp *string     `json:"completion_timestamp,omitempty"` // RFC3339 string
	Expiration          *string     `json:"expiration,omitempty"`           // RFC3339 string
	Errors              int         `json:"errors,omitempty"`
	Warnings            int         `json:"warnings,omitempty"`
	Message             string      `json:"message,omitempty"`
}

// BackupDetails represents detailed backup information.
type BackupDetails struct {
	BackupSummary
	IncludedNamespaces      []string                 `json:"included_namespaces,omitempty"`
	ExcludedNamespaces      []string                 `json:"excluded_namespaces,omitempty"`
	IncludedResources       []string                 `json:"included_resources,omitempty"`
	ExcludedResources       []string                 `json:"excluded_resources,omitempty"`
	LabelSelector           map[string]string        `json:"label_selector,omitempty"`
	SnapshotVolumes         *bool                    `json:"snapshot_volumes,omitempty"`
	IncludeClusterResources *bool                    `json:"include_cluster_resources,omitempty"`
	StorageLocation         string                   `json:"storage_location,omitempty"`
	VolumeSnapshots         []string                 `json:"volume_snapshots,omitempty"`
	Conditions              []map[string]interface{} `json:"conditions,omitempty"`
}

// RestoreSummary represents a normalized Velero restore summary.
type RestoreSummary struct {
	Name                string      `json:"name"`
	Namespace           string      `json:"namespace"`
	BackupName          string      `json:"backup_name"`
	Phase               BackupPhase `json:"phase"`
	StartTimestamp      *string     `json:"start_timestamp,omitempty"`
	CompletionTimestamp *string     `json:"completion_timestamp,omitempty"`
	Errors              int         `json:"errors,omitempty"`
	Warnings            int         `json:"warnings,omitempty"`
	Message             string      `json:"message,omitempty"`
}

// BackupStorageLocationSummary represents a backup storage location summary.
type BackupStorageLocationSummary struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Provider   string `json:"provider,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	Region     string `json:"region,omitempty"`
	AccessMode string `json:"access_mode,omitempty"`
	Phase      string `json:"phase,omitempty"`
}

// GVKs for Velero CRDs
var (
	BackupGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "Backup",
	}
	RestoreGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "Restore",
	}
	BackupStorageLocationGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "BackupStorageLocation",
	}
	ScheduleGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "Schedule",
	}
)
