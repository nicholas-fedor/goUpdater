# goUpdater Makefile
# Comprehensive build, test, and deployment targets for goUpdater

# Variables
BINARY_NAME=goUpdater
GO=go
DOCKER=docker
GORELEASER=goreleaser
GOLANGCI_LINT=golangci-lint-v2

# Default target
help: ## Show this help message
	@echo "goUpdater Makefile"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# Development Targets
# =============================================================================

.PHONY: build test lint run setup install

build: ## Build the application binary
	$(GO) build -ldflags "-X github.com/nicholas-fedor/goUpdater/internal/version.goVersion=$$(go version | cut -d' ' -f3) -X github.com/nicholas-fedor/goUpdater/internal/version.platform=$$(go env GOOS)/$$(go env GOARCH)" -o bin/$(BINARY_NAME) .

test: ## Run all tests
	$(GO) test -v -coverprofile coverage.out -covermode atomic ./...

lint: ## Run linter and fix issues
	$(GOLANGCI_LINT) run --fix --config build/golangci-lint/golangci-lint.yaml ./...

run: ## Run the application
	$(GO) run .

setup: ## Set up development environment (Go modules and docs dependencies)
	$(GO) mod tidy
	cd docs && npm install

install: ## Install the application
	$(GO) install .

# =============================================================================
# Dependency Management
# =============================================================================

.PHONY: mod-tidy mod-download

mod-tidy: ## Tidy and clean up Go modules
	$(GO) mod tidy

mod-download: ## Download Go module dependencies
	$(GO) mod download

# =============================================================================
# Documentation Targets
# =============================================================================

.PHONY: docs docs-build docs-serve docs-start docs-stop docs-deploy docs-clean docs-api

docs: docs-build docs-serve ## Build and serve documentation site for local development

docs-build: ## Build Docusaurus documentation site
	cd docs && npm run build

docs-serve: ## Serve Docusaurus documentation site locally
	cd docs && npm run serve

docs-start: ## Start Docusaurus development server
	cd docs && npm start

docs-stop: ## Stop Docusaurus development server
	pkill -f docusaurus

docs-deploy: ## Deploy Docusaurus documentation site
	cd docs && npm run deploy

docs-clean: ## Clean documentation artifacts
	cd docs && npm run clear

# =============================================================================
# Release Targets
# =============================================================================

.PHONY: release

release: ## Create a new release using GoReleaser
	$(GORELEASER) release --clean --config build/goreleaser/goreleaser.yaml

# =============================================================================
# Utility Targets
# =============================================================================

.PHONY: clean changelog

clean: ## Clean build artifacts
	rm -rf bin/

changelog: ## Generate changelog using Git Cliff
	git-cliff --config build/git-cliff/cliff.toml -o CHANGELOG.md
