# kube-mcp Dockerfile
#
# This Dockerfile builds a production-ready container image for kube-mcp, a Model Context
# Protocol (MCP) server for Kubernetes management.
#
# The image:
# - Runs as a non-root user (UID 65532)
# - Expects configuration at /etc/kube-mcp/config.toml (mount as volume)
# - Exposes HTTP port 8080 (default) with /health endpoint for health checks
# - Includes CA certificates for Kubernetes API TLS connections
# - Uses distroless base image for minimal attack surface
#
# Usage:
#   docker build -t kube-mcp:latest .
#   docker run -v /path/to/config.toml:/etc/kube-mcp/config.toml:ro \
#              -p 8080:8080 \
#              kube-mcp:latest

# Build arguments
ARG GO_VERSION=1.24
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# Note: Version is hardcoded in main.go, so we don't use -X flags
# CGO_ENABLED=0 creates a fully static binary
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w" \
    -o /workspace/bin/kube-mcp \
    ./cmd/kube-mcp

# Verify binary
RUN ls -lh /workspace/bin/kube-mcp

# =============================================================================
# Stage 2: Runtime
# =============================================================================
# Use distroless static image for minimal attack surface
# The nonroot user has UID 65532
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates from builder (needed for Kubernetes API TLS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data (optional, but useful for logging)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /workspace/bin/kube-mcp /usr/local/bin/kube-mcp

# Set working directory
WORKDIR /

# Note: Config directory /etc/kube-mcp will be created automatically when
# mounting a volume. The nonroot user (UID 65532) can read from /etc.

# Expose default HTTP port
EXPOSE 8080

# Health check endpoint
# Note: distroless images don't have curl, so we use a simple HTTP check
# For Kubernetes deployments, use liveness/readiness probes instead
# HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
#   CMD ["/usr/local/bin/kube-mcp", "--version"] || exit 1

# Default command: run kube-mcp with HTTP transport
# Mount config file: -v /path/to/config.toml:/etc/kube-mcp/config.toml:ro
# Override with: docker run ... kube-mcp:latest --config /custom/path.toml
ENTRYPOINT ["/usr/local/bin/kube-mcp"]
CMD ["--config", "/etc/kube-mcp/config.toml"]

