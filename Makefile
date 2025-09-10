# Defuddle Go Implementation Makefile
# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin
BINARY_NAME = defuddle
CLI_FOLDER = cmd/defuddle

# golangci-lint version management
GOLANGCI_LINT_BINARY := $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION := $(shell $(GOLANGCI_LINT_BINARY) version --format short 2>/dev/null || $(GOLANGCI_LINT_BINARY) version --short 2>/dev/null || echo "not-installed")
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version 2>/dev/null || echo "2.4.0")

# Directories containing independent Go modules.
MODULE_DIRS = .

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(GOBIN)
	@go clean -cache -testcache

.PHONY: submodules
submodules: ## Initialize and update git submodules (required for reference library)
	@echo "[setup] Initializing git submodules..."
	@git submodule update --init --recursive

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "[deps] Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: test
test: ## Run all tests
	@echo "[test] Running all tests..."
	@$(foreach mod,$(MODULE_DIRS),(cd $(mod) && go test ./...) &&) true

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "[test] Running unit tests..."
	@go test ./pkg/... ./internal/... .

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "[test] Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[test] Coverage report generated: coverage.html"

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "[test] Running tests with verbose output..."
	@go test -v ./...

.PHONY: bench
bench: ## Run benchmarks
	@echo "[bench] Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: lint
lint: golangci-lint tidy-lint ## Run all linters

# Install golangci-lint with the required version if it is not already installed or version mismatch.
.PHONY: install-golangci-lint
install-golangci-lint:
	@mkdir -p $(GOBIN)
	@if [ "$(GOLANGCI_LINT_VERSION)" != "$(REQUIRED_GOLANGCI_LINT_VERSION)" ]; then \
		echo "[lint] Installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) (current: $(GOLANGCI_LINT_VERSION))"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
		echo "[lint] golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) installed successfully"; \
	else \
		echo "[lint] golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) already installed"; \
	fi

.PHONY: golangci-lint
golangci-lint: install-golangci-lint ## Run golangci-lint
	@echo "[lint] Running $(shell $(GOLANGCI_LINT_BINARY) version)"
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] golangci-lint: $(mod)" && \
		$(GOLANGCI_LINT_BINARY) run --timeout=10m --path-prefix $(mod)) &&) true

.PHONY: tidy-lint
tidy-lint: ## Check if go.mod and go.sum are tidy
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] mod tidy: $(mod)" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true

.PHONY: fmt
fmt: ## Format Go code
	@echo "[fmt] Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "[vet] Running go vet..."
	@go vet ./...

.PHONY: verify
verify: submodules deps fmt vet lint test ## Run all verification steps (format, vet, lint, test)
	@echo "[verify] All verification steps completed successfully"

.PHONY: dev
dev: deps fmt vet ## Quick development verification (deps, format, vet)
	@echo "[dev] Development verification completed successfully"

# Building
.PHONY: build
build: ## Build the CLI binary
	@echo "[build] Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./$(CLI_FOLDER)

.PHONY: build-cli
build-cli: build ## Build the CLI binary (alias for build)

.PHONY: install-cli
install-cli: build-cli ## Install CLI to system
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) CLI installed to /usr/local/bin/$(BINARY_NAME)"

# Release testing
.PHONY: snapshot
snapshot: ## Test release build locally
	@echo "[release] Testing release build..."
	@goreleaser release --snapshot --skip=publish --clean

.PHONY: release-test
release-test: snapshot ## Test release build without publishing (alias for snapshot)

.PHONY: release-snapshot
release-snapshot: snapshot ## Create a snapshot release (alias for snapshot)

.PHONY: install-goreleaser
install-goreleaser: ## Install GoReleaser
	@echo "[release] Installing GoReleaser..."
	@go install github.com/goreleaser/goreleaser@latest

# Tagging
.PHONY: tag
tag: ## Create and push a new tag (usage: make tag VERSION=v0.1.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required. Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "[release] Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "[release] Tag $(VERSION) created and pushed"

.PHONY: all
all: verify ## Run all verification steps (alias for verify)