# Memory Bank Makefile

.PHONY: all build test test-unit test-integration test-all clean lint fmt vet deps help

# Variables
BINARY_NAME=memory-bank
BUILD_DIR=./cmd/$(BINARY_NAME)
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*')

# Default target
all: deps test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run all tests (unit tests only by default)
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -short -v ./...

# Run integration tests (requires external services)
test-integration:
	@echo "Running integration tests..."
	@echo "Make sure Ollama and ChromaDB are running:"
	@echo "  - Ollama: http://localhost:11434"
	@echo "  - ChromaDB: http://localhost:8000"
	go test -v ./internal/infra -run TestOllamaIntegration -timeout 5m
	go test -v ./internal/infra -run TestChromaDBIntegration -timeout 5m
	go test -v ./internal/infra -run TestFullIntegration -timeout 10m

# Run all tests including integration
test-all: test-unit test-integration

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -short -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the code
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all quality checks
check: fmt vet lint test-unit

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f memory_bank.db
	go clean -cache

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

# Start external services for integration testing
dev-services:
	@echo "Starting external services..."
	@echo "Starting Ollama (if not already running)..."
	@if ! curl -s http://localhost:11434/api/version > /dev/null 2>&1; then \
		echo "Please start Ollama manually: ollama serve"; \
	else \
		echo "Ollama is already running"; \
	fi
	@echo "Starting ChromaDB..."
	@if ! curl -s http://localhost:8000/api/v1/heartbeat > /dev/null 2>&1; then \
		echo "Starting ChromaDB in Docker..."; \
		docker run -d --name memory-bank-chromadb -p 8000:8000 chromadb/chroma || echo "ChromaDB already running or Docker not available"; \
	else \
		echo "ChromaDB is already running"; \
	fi

# Stop external services
dev-services-stop:
	@echo "Stopping external services..."
	@docker stop memory-bank-chromadb 2>/dev/null || true
	@docker rm memory-bank-chromadb 2>/dev/null || true

# Create sample configuration
config-sample:
	@echo "Creating sample configuration..."
	./$(BINARY_NAME) config init --force

# Database operations
db-migrate:
	@echo "Running database migrations..."
	./$(BINARY_NAME) migrate up

db-rollback:
	@echo "Rolling back last migration..."
	./$(BINARY_NAME) migrate down

db-status:
	@echo "Checking migration status..."
	./$(BINARY_NAME) migrate status

# Release build (with optimizations)
release:
	@echo "Building release version..."
	CGO_ENABLED=1 go build -ldflags="-w -s" -o $(BINARY_NAME) $(BUILD_DIR)

# Cross-platform builds
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $(BINARY_NAME)-linux-amd64 $(BUILD_DIR)

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o $(BINARY_NAME)-darwin-amd64 $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o $(BINARY_NAME)-darwin-arm64 $(BUILD_DIR)

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o $(BINARY_NAME)-windows-amd64.exe $(BUILD_DIR)

build-all: build-linux build-darwin build-windows

# Help
help:
	@echo "Memory Bank Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         Build the application"
	@echo "  release       Build optimized release version"
	@echo "  build-all     Build for all platforms"
	@echo ""
	@echo "Test Commands:"
	@echo "  test          Run unit tests (default)"
	@echo "  test-unit     Run unit tests only"
	@echo "  test-integration Run integration tests (requires services)"
	@echo "  test-all      Run all tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo ""
	@echo "Quality Commands:"
	@echo "  fmt           Format code"
	@echo "  vet           Run go vet"
	@echo "  lint          Run golangci-lint"
	@echo "  check         Run all quality checks"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev-setup     Set up development environment"
	@echo "  dev-services  Start external services for testing"
	@echo "  dev-services-stop Stop external services"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-migrate    Run database migrations"
	@echo "  db-rollback   Rollback last migration"
	@echo "  db-status     Check migration status"
	@echo ""
	@echo "Configuration Commands:"
	@echo "  config-sample Create sample configuration"
	@echo ""
	@echo "Utility Commands:"
	@echo "  deps          Install/update dependencies"
	@echo "  clean         Clean build artifacts"
	@echo "  help          Show this help message"