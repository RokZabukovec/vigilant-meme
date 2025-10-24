# Makefile for Clip - Distributed Network Service

# Variables
BINARY_NAME=clip
BUILD_DIR=build
SRC_DIR=src
CMD_DIR=$(SRC_DIR)/cmd/clip
INTERNAL_DIR=$(SRC_DIR)/internal
PKG_DIR=$(SRC_DIR)/pkg

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: all build clean test test-coverage test-all test-race test-bench test-pkg deps fmt vet lint run help install

# Default target
all: clean deps fmt vet test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/clip
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/clip
	cd $(SRC_DIR) && GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/clip
	cd $(SRC_DIR) && GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/clip
	cd $(SRC_DIR) && GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/clip
	@echo "Multi-platform build complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	cd $(SRC_DIR) && $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd $(SRC_DIR) && $(GOTEST) -v -coverprofile=coverage.out ./...
	cd $(SRC_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(SRC_DIR)/coverage.html"

# Run comprehensive tests with race detection and benchmarks
test-all:
	@echo "Running comprehensive tests..."
	cd $(SRC_DIR) && ./run_tests.sh

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	cd $(SRC_DIR) && $(GOTEST) -race ./...

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	cd $(SRC_DIR) && $(GOTEST) -bench=. ./...

# Run tests for specific package
test-pkg:
	@echo "Running tests for package: $(PKG)"
	@if [ -z "$(PKG)" ]; then echo "Usage: make test-pkg PKG=package_name"; exit 1; fi
	cd $(SRC_DIR) && $(GOTEST) -v ./$(PKG)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	cd $(SRC_DIR) && $(GOMOD) download
	cd $(SRC_DIR) && $(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	cd $(SRC_DIR) && $(GOCMD) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	cd $(SRC_DIR) && $(GOCMD) vet ./...

# Run golangci-lint (if available)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(SRC_DIR) && golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) -id=test-node -port=8080

# Run with specific parameters
run-local:
	@echo "Running local test..."
	./$(BUILD_DIR)/$(BINARY_NAME) -id=node1 -port=8080

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete"

# Development mode - run with auto-reload (requires air)
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		cd $(SRC_DIR) && air; \
	else \
		echo "air not found, please install it: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

# Create release package
release: clean deps fmt vet test build-all
	@echo "Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-* $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@cp -r scripts $(BUILD_DIR)/release/ 2>/dev/null || true
	@cp -r configs $(BUILD_DIR)/release/ 2>/dev/null || true
	@echo "Release package created in $(BUILD_DIR)/release/"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t clip:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9999:9999/udp --name clip-container clip:latest

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, deps, fmt, vet, test, and build"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-all     - Run comprehensive tests with race detection and benchmarks"
	@echo "  test-race    - Run tests with race detection"
	@echo "  test-bench   - Run benchmark tests"
	@echo "  test-pkg     - Run tests for specific package (PKG=package_name)"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  run          - Build and run the application"
	@echo "  run-local    - Run with local test parameters"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  dev          - Run in development mode with auto-reload"
	@echo "  release      - Create release package"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help         - Show this help message"