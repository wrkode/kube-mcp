package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpServer "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
)

// Server provides HTTP transport for MCP.
type Server struct {
	mcpServer  *mcpServer.Server
	httpServer *http.Server
	config     *config.HTTPConfig
	oauth      *OAuthMiddleware
	logger     *observability.Logger
	metrics    *observability.Metrics
}

// NewServer creates a new HTTP server for MCP.
func NewServer(mcpServer *mcpServer.Server, cfg *config.HTTPConfig, logger *observability.Logger, metrics *observability.Metrics, clientSet *kubernetes.ClientSet, securityCfg *config.SecurityConfig) (*Server, error) {
	s := &Server{
		mcpServer: mcpServer,
		config:    cfg,
		logger:    logger,
		metrics:   metrics,
	}

	// Setup OAuth middleware if enabled
	if cfg.OAuth.Enabled {
		var err error
		s.oauth, err = NewOAuthMiddleware(&cfg.OAuth, clientSet, securityCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create OAuth middleware: %w", err)
		}
	}

	// Setup router
	router := mux.NewRouter()
	s.setupRoutes(router)

	// Apply observability middleware to all routes
	var rootHandler http.Handler = router
	if s.logger != nil && s.metrics != nil {
		rootHandler = observability.HTTPMiddleware(s.logger, s.metrics, router)
	}

	s.httpServer = &http.Server{
		Addr:         cfg.Address,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// setupRoutes configures HTTP routes.
func (s *Server) setupRoutes(router *mux.Router) {
	// CORS middleware
	if s.config.CORS.Enabled {
		router.Use(s.corsMiddleware)
	}

	// OAuth middleware for protected routes
	var mcpHandler http.Handler = s.mcpHandler()
	if s.oauth != nil {
		mcpHandler = s.oauth.Middleware(mcpHandler)
	}

	// MCP endpoint
	router.Handle("/mcp", mcpHandler).Methods("POST", "OPTIONS")

	// Health check endpoint
	router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Prometheus metrics endpoint
	if s.metrics != nil {
		router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	}

	// Well-known endpoints
	router.HandleFunc("/.well-known/mcp", s.wellKnownHandler).Methods("GET")
}

// mcpHandler creates the MCP HTTP handler.
func (s *Server) mcpHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return s.mcpServer.GetSDKServer()
	}, &mcp.StreamableHTTPOptions{})
}

// healthHandler handles health check requests.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// wellKnownHandler handles .well-known/mcp requests.
func (s *Server) wellKnownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Get implementation info - we store it in our wrapper
	json.NewEncoder(w).Encode(map[string]any{
		"name":     "kube-mcp",
		"version":  "1.0.0",
		"endpoint": "/mcp",
	})
}

// corsMiddleware handles CORS.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if s.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", joinStrings(s.config.CORS.AllowedMethods, ","))
		w.Header().Set("Access-Control-Allow-Headers", joinStrings(s.config.CORS.AllowedHeaders, ","))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowedOrigin checks if an origin is allowed.
func (s *Server) isAllowedOrigin(origin string) bool {
	if !s.config.CORS.Enabled {
		return false
	}

	if len(s.config.CORS.AllowedOrigins) == 0 {
		return true // Allow all if none specified
	}

	for _, allowed := range s.config.CORS.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP implements http.Handler for compatibility.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}
