package kiali

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/config"
)

// KialiClientTestSuite tests Kiali client functionality with HTTP mocks.
type KialiClientTestSuite struct {
	suite.Suite
	server  *httptest.Server
	client  *KialiClient
	handler http.HandlerFunc
}

// SetupTest sets up the test.
func (s *KialiClientTestSuite) SetupTest() {
	// Default handler - will be overridden in individual tests
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handler.ServeHTTP(w, r)
	}))

	// Create Kiali client
	cfg := &config.KialiConfig{
		Enabled: true,
		URL:     s.server.URL,
		Token:   "test-token",
		Timeout: config.Duration(30 * time.Second),
		TLS: config.TLSConfig{
			Enabled: false,
		},
	}

	client, err := NewKialiClient(cfg)
	s.Require().NoError(err, "Failed to create Kiali client")
	s.client = client
}

// TearDownTest cleans up after the test.
func (s *KialiClientTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

// TestGetMeshGraph tests GetMeshGraph with successful response.
func (s *KialiClientTestSuite) TestGetMeshGraph() {
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("GET", r.Method, "Should use GET method")
		s.Equal("/api/namespaces/test-ns/graph", r.URL.Path, "Path should match")
		s.Equal("Bearer test-token", r.Header.Get("Authorization"), "Should include auth token")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]any{
			"graph": map[string]any{
				"nodes": []map[string]any{},
				"edges": []map[string]any{},
			},
		}
		json.NewEncoder(w).Encode(response)
	})

	ctx := context.Background()
	result, err := s.client.GetMeshGraph(ctx, "test-ns")
	s.Require().NoError(err, "GetMeshGraph should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Contains(result, "graph", "Result should contain graph")
}

// TestGetMeshGraphError tests GetMeshGraph with error response.
func (s *KialiClientTestSuite) TestGetMeshGraphError() {
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
	})

	ctx := context.Background()
	_, err := s.client.GetMeshGraph(ctx, "test-ns")
	s.Require().Error(err, "GetMeshGraph should fail on server error")
}

// TestGetMeshGraphTimeout tests timeout handling.
func (s *KialiClientTestSuite) TestGetMeshGraphTimeout() {
	// Handler that sleeps longer than timeout
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Create client with very short timeout
	cfg := &config.KialiConfig{
		Enabled: true,
		URL:     s.server.URL,
		Token:   "test-token",
		Timeout: config.Duration(1 * time.Millisecond), // Very short timeout
		TLS: config.TLSConfig{
			Enabled: false,
		},
	}

	client, err := NewKialiClient(cfg)
	s.Require().NoError(err, "Failed to create Kiali client")

	ctx := context.Background()
	_, err = client.GetMeshGraph(ctx, "test-ns")
	s.Require().Error(err, "Should timeout")
}

// TestGetMetrics tests GetMetrics.
func (s *KialiClientTestSuite) TestGetMetrics() {
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/api/namespaces/test-ns/services/test-svc/metrics", r.URL.Path, "Path should match")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"metrics": []map[string]any{}})
	})

	ctx := context.Background()
	result, err := s.client.GetMetrics(ctx, "test-ns", "test-svc")
	s.Require().NoError(err, "GetMetrics should succeed")
	s.Require().NotNil(result, "Result should not be nil")
}

// TestKialiClientTestSuite runs the Kiali client test suite.
func TestKialiClientTestSuite(t *testing.T) {
	suite.Run(t, new(KialiClientTestSuite))
}
