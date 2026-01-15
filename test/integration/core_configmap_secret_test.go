package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreConfigMapSecretTestSuite tests ConfigMap and Secret operations.
type CoreConfigMapSecretTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreConfigMapSecretTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestConfigMapsGetData tests getting ConfigMap data.
func (s *CoreConfigMapSecretTestSuite) TestConfigMapsGetData() {
	ctx := context.Background()
	namespace := "test-ns-cm"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: namespace,
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	_, err = s.clientSet.Typed.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create ConfigMap")

	// Test get all data
	args := struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Keys      []string `json:"keys"`
		Context   string   `json:"context"`
	}{
		Name:      "test-configmap",
		Namespace: namespace,
		Keys:      nil,
		Context:   "",
	}

	result, err := s.toolset.TestHandleConfigMapsGetData(ctx, args)
	s.Require().NoError(err, "configmaps_get_data should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var data map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &data)
	s.Require().NoError(err, "should parse JSON result")
	s.Equal("test-configmap", data["name"])
	s.Equal(namespace, data["namespace"])

	dataMap, ok := data["data"].(map[string]interface{})
	s.Require().True(ok, "data should be a map")
	s.Equal("value1", dataMap["key1"])
	s.Equal("value2", dataMap["key2"])

	// Test get specific keys
	args.Keys = []string{"key1"}
	result, err = s.toolset.TestHandleConfigMapsGetData(ctx, args)
	s.Require().NoError(err, "configmaps_get_data should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	textContent, ok = result.Content[0].(*mcp.TextContent)
	s.Require().True(ok)
	err = json.Unmarshal([]byte(textContent.Text), &data)
	s.Require().NoError(err)
	dataMap, ok = data["data"].(map[string]interface{})
	s.Require().True(ok)
	s.Equal("value1", dataMap["key1"])
	s.NotContains(dataMap, "key2")
}

// TestConfigMapsSetData tests setting ConfigMap data.
func (s *CoreConfigMapSecretTestSuite) TestConfigMapsSetData() {
	ctx := context.Background()
	namespace := "test-ns-cm-set"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-set",
			Namespace: namespace,
		},
		Data: map[string]string{
			"existing": "value",
		},
	}
	_, err = s.clientSet.Typed.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create ConfigMap")

	// Test replace data
	args := struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Data      map[string]string `json:"data"`
		Merge     bool              `json:"merge"`
		Context   string            `json:"context"`
	}{
		Name:      "test-configmap-set",
		Namespace: namespace,
		Data: map[string]string{
			"newkey": "newvalue",
		},
		Merge:   false,
		Context: "",
	}

	result, err := s.toolset.TestHandleConfigMapsSetData(ctx, args)
	s.Require().NoError(err, "configmaps_set_data should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify ConfigMap was updated
	updated, err := s.clientSet.Typed.CoreV1().ConfigMaps(namespace).Get(ctx, "test-configmap-set", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Equal("newvalue", updated.Data["newkey"])
	s.NotContains(updated.Data, "existing")

	// Test merge data
	args.Data = map[string]string{
		"merged": "mergedvalue",
	}
	args.Merge = true

	result, err = s.toolset.TestHandleConfigMapsSetData(ctx, args)
	s.Require().NoError(err)
	s.Require().False(result.IsError)

	updated, err = s.clientSet.Typed.CoreV1().ConfigMaps(namespace).Get(ctx, "test-configmap-set", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Equal("newvalue", updated.Data["newkey"])
	s.Equal("mergedvalue", updated.Data["merged"])
}

// TestSecretsGetData tests getting Secret data.
func (s *CoreConfigMapSecretTestSuite) TestSecretsGetData() {
	ctx := context.Background()
	namespace := "test-ns-secret"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}
	_, err = s.clientSet.Typed.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create Secret")

	// Test get all data (encoded)
	args := struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Keys      []string `json:"keys"`
		Decode    bool     `json:"decode"`
		Context   string   `json:"context"`
	}{
		Name:      "test-secret",
		Namespace: namespace,
		Keys:      nil,
		Decode:    false,
		Context:   "",
	}

	result, err := s.toolset.TestHandleSecretsGetData(ctx, args)
	s.Require().NoError(err, "secrets_get_data should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify encoded result
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok)
	var data map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &data)
	s.Require().NoError(err)
	s.Equal(false, data["decoded"])

	dataMap, ok := data["data"].(map[string]interface{})
	s.Require().True(ok)
	encodedValue1 := dataMap["key1"].(string)
	decoded, err := base64.StdEncoding.DecodeString(encodedValue1)
	s.Require().NoError(err)
	s.Equal("value1", string(decoded))

	// Test get decoded data
	args.Decode = true
	result, err = s.toolset.TestHandleSecretsGetData(ctx, args)
	s.Require().NoError(err)
	s.Require().False(result.IsError)

	textContent, ok = result.Content[0].(*mcp.TextContent)
	s.Require().True(ok)
	err = json.Unmarshal([]byte(textContent.Text), &data)
	s.Require().NoError(err)
	s.Equal(true, data["decoded"])

	dataMap, ok = data["data"].(map[string]interface{})
	s.Require().True(ok)
	s.Equal("value1", dataMap["key1"])
	s.Equal("value2", dataMap["key2"])
}

// TestSecretsSetData tests setting Secret data.
func (s *CoreConfigMapSecretTestSuite) TestSecretsSetData() {
	ctx := context.Background()
	namespace := "test-ns-secret-set"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret-set",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"existing": []byte("value"),
		},
	}
	_, err = s.clientSet.Typed.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create Secret")

	// Test set data with encoding (default)
	args := struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Data      map[string]string `json:"data"`
		Merge     bool              `json:"merge"`
		Encode    bool              `json:"encode"`
		Context   string            `json:"context"`
	}{
		Name:      "test-secret-set",
		Namespace: namespace,
		Data: map[string]string{
			"newkey": "newvalue",
		},
		Merge:   false,
		Encode:  true,
		Context: "",
	}

	result, err := s.toolset.TestHandleSecretsSetData(ctx, args)
	s.Require().NoError(err, "secrets_set_data should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify Secret was updated
	updated, err := s.clientSet.Typed.CoreV1().Secrets(namespace).Get(ctx, "test-secret-set", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Equal("newvalue", string(updated.Data["newkey"]))
	s.NotContains(updated.Data, "existing")

	// Test merge with base64 encoded data
	encodedValue := base64.StdEncoding.EncodeToString([]byte("mergedvalue"))
	args.Data = map[string]string{
		"merged": encodedValue,
	}
	args.Merge = true
	args.Encode = false

	result, err = s.toolset.TestHandleSecretsSetData(ctx, args)
	s.Require().NoError(err)
	s.Require().False(result.IsError)

	updated, err = s.clientSet.Typed.CoreV1().Secrets(namespace).Get(ctx, "test-secret-set", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Equal("newvalue", string(updated.Data["newkey"]))
	s.Equal("mergedvalue", string(updated.Data["merged"]))
}

// TestCoreConfigMapSecretSuite runs the ConfigMap/Secret test suite.
func TestCoreConfigMapSecretSuite(t *testing.T) {
	suite.Run(t, new(CoreConfigMapSecretTestSuite))
}

