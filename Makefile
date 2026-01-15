# kube-mcp Makefile
#
# This Makefile follows the well-documented Makefile pattern described at:
# https://suva.sh/posts/well-documented-makefiles/
#
# The help target uses awk to parse ## comments and ##@ group headers to generate
# a nicely formatted help output. Run 'make' or 'make help' to see all available targets.

.DEFAULT_GOAL := help
SHELL := /usr/bin/env bash

# Project metadata
PACKAGE := github.com/wrkode/kube-mcp
BINARY_NAME := kube-mcp
MAIN_PACKAGE := ./cmd/kube-mcp

# Version information
GIT_COMMIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LD_FLAGS := -s -w
# Note: Version is currently hardcoded in main.go, so we don't wire it via -X flags
# If a pkg/version package is added later, use: -X $(PACKAGE)/pkg/version.Version=$(GIT_VERSION)

# Build output directory
OUTPUT_DIR := _output
BINARY_PATH := $(OUTPUT_DIR)/bin/$(BINARY_NAME)

# Cross-compilation targets
OSES := darwin linux windows
ARCHS := amd64 arm64

# Docker image configuration (optional)
IMAGE_REPO ?= ghcr.io/wrkode/kube-mcp
IMAGE_TAG ?= $(GIT_VERSION)

# Tools directory
TOOLS_DIR := _tools
GOLANGCI_LINT := $(TOOLS_DIR)/bin/golangci-lint
SETUP_ENVTEST := $(shell go env GOPATH)/bin/setup-envtest

# Test coverage
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Clean targets
CLEAN_TARGETS := $(OUTPUT_DIR) $(TOOLS_DIR) $(COVERAGE_FILE) $(COVERAGE_HTML) envtest.env

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean up build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(CLEAN_TARGETS)
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	@rm -f *.test *.test.exe
	@echo "Clean complete"

##@ Build

.PHONY: build
build: tidy format vet ## Build the kube-mcp binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)/bin
	@go build -ldflags "$(LD_FLAGS)" -o $(BINARY_PATH) $(MAIN_PACKAGE)
	@ln -sf $(shell pwd)/$(BINARY_PATH) $(BINARY_NAME)
	@echo "Binary built: $(BINARY_PATH)"

.PHONY: build-all-platforms
build-all-platforms: tidy format vet ## Build kube-mcp for all supported platforms
	@echo "Building for all platforms..."
	@mkdir -p $(OUTPUT_DIR)/bin
	@for os in $(OSES); do \
		for arch in $(ARCHS); do \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build -ldflags "$(LD_FLAGS)" -o $(OUTPUT_DIR)/bin/$(BINARY_NAME)-$$os-$$arch$(if $(findstring windows,$$os),.exe,) $(MAIN_PACKAGE); \
		done; \
	done
	@echo "Cross-compilation complete. Binaries in $(OUTPUT_DIR)/bin/"

.PHONY: run
run: build ## Build and run kube-mcp
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME) $(ARGS)

.PHONY: install
install: build ## Install kube-mcp to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $$(go env GOPATH)/bin..."
	@cp $(BINARY_PATH) $$(go env GOPATH)/bin/$(BINARY_NAME)
	@echo "Installation complete"

##@ Test

.PHONY: test
test: ## Run all unit tests
	@echo "Running unit tests..."
	@go test -v ./cmd/... ./pkg/...

.PHONY: test-race
test-race: ## Run unit tests with race detector
	@echo "Running unit tests with race detector..."
	@go test -race -v ./cmd/... ./pkg/...

.PHONY: test-integration
test-integration: tools envtest-setup ## Run integration tests with envtest
	@echo "Running integration tests..."
	@if [ -f envtest.env ]; then \
		source envtest.env && go test -v ./test/integration/...; \
	else \
		echo "Warning: envtest.env not found. Run 'make envtest-setup' first."; \
		exit 1; \
	fi

.PHONY: test-all
test-all: test test-integration ## Run all unit and integration tests

.PHONY: cover
cover: ## Generate test coverage report
	@echo "Generating coverage report..."
	@go test -coverprofile=$(COVERAGE_FILE) ./cmd/... ./pkg/...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

##@ Lint

.PHONY: format
format: ## Format the Go code
	@echo "Formatting Go code..."
	@go fmt ./...

.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	@go mod tidy

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: golangci-lint
golangci-lint: ## Install golangci-lint if not already installed
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		echo "Installing golangci-lint..."; \
		mkdir -p $(TOOLS_DIR)/bin; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR)/bin latest; \
	fi
	@echo "golangci-lint is available at $(GOLANGCI_LINT)"

.PHONY: lint
lint: golangci-lint ## Run golangci-lint
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run --verbose --print-resources-usage

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with --fix
	@echo "Running golangci-lint with --fix..."
	@$(GOLANGCI_LINT) run --fix

##@ Tools

.PHONY: tools
tools: golangci-lint setup-envtest ## Install all required tools

.PHONY: setup-envtest
setup-envtest: ## Install setup-envtest
	@if [ ! -f $(SETUP_ENVTEST) ]; then \
		echo "Installing setup-envtest..."; \
		go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest; \
	fi
	@echo "setup-envtest is available at $(SETUP_ENVTEST)"

.PHONY: envtest-setup
envtest-setup: setup-envtest ## Setup envtest binaries for integration tests
	@echo "Setting up envtest binaries..."
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && \
		setup-envtest use -p env 1.31.x > envtest.env
	@echo "envtest binaries configured. Source envtest.env before running tests."

##@ CI

.PHONY: ci
ci: vet test test-integration ## Run full CI pipeline (vet + unit + integration tests)
	@echo "CI pipeline complete"

##@ Local Development

.PHONY: dev
dev: build ## Build kube-mcp for local development
	@echo "Development build complete: $(BINARY_PATH)"
	@echo "Run with: ./$(BINARY_NAME) --config ./examples/config.toml"

.PHONY: check
check: format tidy vet lint ## Run all code quality checks (format, tidy, vet, lint)

##@ Release / Artifacts

.PHONY: version-update
version-update: ## Update version in main.go, Chart.yaml, values.yaml, and README.md (requires VERSION_TAG env var)
	@if [ -z "$(VERSION_TAG)" ]; then \
		echo "Error: VERSION_TAG environment variable is required"; \
		echo "Usage: make version-update VERSION_TAG=v1.0.1"; \
		exit 1; \
	fi
	@echo "Updating version to $(VERSION_TAG)..."
	@VERSION="$${VERSION_TAG#v}"; \
	MAIN_GO="cmd/kube-mcp/main.go"; \
	CHART_YAML="charts/kube-mcp/Chart.yaml"; \
	VALUES_YAML="charts/kube-mcp/values.yaml"; \
	CHART_README="charts/kube-mcp/README.md"; \
	\
	echo "Updating $$MAIN_GO..."; \
	sed -i.bak "s/version = \"[^\"]*\"/version = \"$$VERSION\"/" $$MAIN_GO; \
	rm -f $$MAIN_GO.bak; \
	\
	echo "Updating $$CHART_YAML..."; \
	if command -v yq &> /dev/null; then \
		yq eval ".version = \"$$VERSION\"" -i $$CHART_YAML; \
		yq eval ".appVersion = \"$$VERSION\"" -i $$CHART_YAML; \
	else \
		sed -i.bak "s/^version:.*/version: $$VERSION/" $$CHART_YAML; \
		sed -i.bak "s/^appVersion:.*/appVersion: \"$$VERSION\"/" $$CHART_YAML; \
		rm -f $$CHART_YAML.bak; \
	fi; \
	\
	echo "Updating $$VALUES_YAML..."; \
	if command -v yq &> /dev/null; then \
		yq eval ".image.tag = \"$$VERSION\"" -i $$VALUES_YAML; \
	else \
		sed -i.bak "s/^  tag:.*/  tag: \"$$VERSION\"/" $$VALUES_YAML; \
		rm -f $$VALUES_YAML.bak; \
	fi; \
	\
	if [ -f $$CHART_README ]; then \
		echo "Updating $$CHART_README..."; \
		sed -i.bak "s/kube-mcp-1\.0\.0/kube-mcp-$$VERSION/g" $$CHART_README; \
		sed -i.bak "s/version 1\.0\.0/version $$VERSION/g" $$CHART_README; \
		sed -i.bak "s/--version 1\.0\.0/--version $$VERSION/g" $$CHART_README; \
		sed -i.bak "s/tag: \"1\.0\.0\"/tag: \"$$VERSION\"/g" $$CHART_README; \
		rm -f $$CHART_README.bak; \
	fi; \
	\
	echo "Version update complete:"; \
	echo "  - $$MAIN_GO: version=\"$$VERSION\""; \
	echo "  - $$CHART_YAML: version=$$VERSION, appVersion=$$VERSION"; \
	echo "  - $$VALUES_YAML: image.tag=$$VERSION"

.PHONY: docker-build
docker-build: ## Build a Docker image for kube-mcp
	@if [ ! -f Dockerfile ]; then \
		echo "Error: Dockerfile not found. Skipping docker build."; \
		exit 1; \
	fi
	@echo "Building Docker image $(IMAGE_REPO):$(IMAGE_TAG)..."
	@docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .
	@docker tag $(IMAGE_REPO):$(IMAGE_TAG) $(IMAGE_REPO):latest
	@echo "Docker image built: $(IMAGE_REPO):$(IMAGE_TAG)"

.PHONY: docker-push
docker-push: ## Push the Docker image
	@echo "Pushing Docker image $(IMAGE_REPO):$(IMAGE_TAG)..."
	@docker push $(IMAGE_REPO):$(IMAGE_TAG)
	@docker push $(IMAGE_REPO):latest
	@echo "Docker image pushed"

.PHONY: release
release: build-all-platforms ## Build binaries for release
	@echo "Release build complete. Binaries in $(OUTPUT_DIR)/bin/"
	@ls -lh $(OUTPUT_DIR)/bin/

.PHONY: release-prep
release-prep: version-update ## Prepare for release (update versions, then build)
	@echo "Release preparation complete"

.PHONY: version
version: ## Print version information
	@echo "Package: $(PACKAGE)"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Git Version: $(GIT_VERSION)"
	@echo "Git Commit: $(GIT_COMMIT_HASH)"
	@echo "Build Time: $(BUILD_TIME)"


