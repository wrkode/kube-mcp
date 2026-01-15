package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"github.com/wrkode/kube-mcp/pkg/toolsets/backup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// BackupTestSuite tests Backup/Restore toolset functionality.
type BackupTestSuite struct {
	EnvtestSuite
	toolset *backup.Toolset
}

// SetupTest sets up the test.
func (s *BackupTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	s.toolset = backup.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "Backup toolset should be enabled")

	logger := observability.NewLogger(observability.LogLevelInfo, false)
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestBackupsList tests backup.backups_list operation.
func (s *BackupTestSuite) TestBackupsList() {
	ctx := context.Background()
	namespace := "velero"

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	backupGVR := schema.GroupVersionResource{
		Group:    "velero.io",
		Version:  "v1",
		Resource: "backups",
	}

	backup := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "velero.io/v1",
			"kind":       "Backup",
			"metadata": map[string]interface{}{
				"name":      "test-backup",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"phase": "Completed",
			},
		},
	}
	_, err = s.clientSet.Dynamic.Resource(backupGVR).Namespace(namespace).Create(ctx, backup, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create Backup")

	args := struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}{
		Namespace: namespace,
		Limit:     10,
	}

	result, err := s.toolset.TestHandleBackupsList(ctx, args)
	s.Require().NoError(err, "Backups list should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Require().False(result.IsError, "Result should not be an error")

	var resultData map[string]interface{}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Result should be TextContent")
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "Should parse result JSON")

	items, ok := resultData["items"].([]interface{})
	s.Require().True(ok, "Result should have items array")
	s.Require().Greater(len(items), 0, "Should have at least one backup")
}

// TestBackupTestSuite runs the backup test suite.
func TestBackupTestSuite(t *testing.T) {
	suite.Run(t, new(BackupTestSuite))
}
