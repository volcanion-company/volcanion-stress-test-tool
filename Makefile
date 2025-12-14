# Makefile for Volcanion Stress Test Tool

# Variables
BINARY_NAME=volcanion-stress-test
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOMOD=$(GOCMD) mod

# Directories
BUILD_DIR=./dist
CMD_DIR=./cmd/server
WEB_DIR=./web

.PHONY: all build clean test lint fmt help docker docker-build docker-run

# Default target
all: lint test build

## Build targets

build: ## Build the Go binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

build-all: build-linux build-darwin build-windows ## Build for all platforms

## Frontend targets

frontend-install: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	cd $(WEB_DIR) && npm ci

frontend-build: ## Build frontend
	@echo "Building frontend..."
	cd $(WEB_DIR) && npm run build

frontend-dev: ## Run frontend development server
	@echo "Starting frontend dev server..."
	cd $(WEB_DIR) && npm run dev

## Test targets

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Lint and format targets

lint: ## Run linters
	@echo "Running linters..."
	$(GOVET) ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout 5m; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

fmt-check: ## Check code formatting
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

## Dependency targets

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

## Docker targets

docker-build: ## Build backend Docker image
	@echo "Building backend Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest \
		.

docker-build-web: ## Build frontend Docker image
	@echo "Building frontend Docker image..."
	docker build \
		-t $(BINARY_NAME)-web:$(VERSION) \
		-t $(BINARY_NAME)-web:latest \
		$(WEB_DIR)

docker-build-all: docker-build docker-build-web ## Build all Docker images
	@echo "All Docker images built successfully"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -d \
		-p 8080:8080 \
		--name $(BINARY_NAME) \
		-e JWT_SECRET=change-me-in-production \
		$(BINARY_NAME):latest

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

docker-compose-build: ## Build all images with docker-compose
	@echo "Building all images..."
	docker-compose build

docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services..."
	docker-compose up -d

docker-compose-up-build: ## Build and start all services
	@echo "Building and starting services..."
	docker-compose up -d --build

docker-compose-down: ## Stop all services with docker-compose
	@echo "Stopping services..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

## Development targets

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

## Utility targets

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f benchmark.txt

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

security: ## Run security checks
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping..."; \
	fi
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed, skipping..."; \
	fi

## Help

help: ## Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
