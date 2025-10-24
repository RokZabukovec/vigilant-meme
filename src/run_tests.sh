#!/bin/bash

# Test runner script for Clip
# This script runs all tests with proper coverage reporting

set -e

echo "🧪 Running Clip Tests"
echo "===================="

# Change to the src directory
cd "$(dirname "$0")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Get Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_status "Using Go version $GO_VERSION"

# Run tests with verbose output
echo ""
echo "Running unit tests..."
echo "-------------------"

# Run tests for each package
PACKAGES=(
    "./internal/config"
    "./internal/peer"
    "./internal/handlers"
    "./internal/discovery"
    "./internal/service"
    "./pkg/network"
    "./cmd/clip"
)

FAILED_PACKAGES=()

for pkg in "${PACKAGES[@]}"; do
    echo ""
    echo "Testing $pkg..."
    if go test -v "$pkg"; then
        print_status "Tests passed for $pkg"
    else
        print_error "Tests failed for $pkg"
        FAILED_PACKAGES+=("$pkg")
    fi
done

# Run all tests together for coverage
echo ""
echo "Running all tests with coverage..."
echo "--------------------------------"

if go test -v -coverprofile=coverage.out ./...; then
    print_status "All tests passed"
else
    print_error "Some tests failed"
    if [ ${#FAILED_PACKAGES[@]} -gt 0 ]; then
        echo "Failed packages:"
        for pkg in "${FAILED_PACKAGES[@]}"; do
            echo "  - $pkg"
        done
    fi
fi

# Generate coverage report
echo ""
echo "Generating coverage report..."
echo "----------------------------"

if [ -f coverage.out ]; then
    # Show coverage summary
    go tool cover -func=coverage.out | tail -1
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_status "HTML coverage report generated: coverage.html"
    
    # Show coverage by package
    echo ""
    echo "Coverage by package:"
    go tool cover -func=coverage.out | grep -E "^(github.com/rokzabukovec/clip/)" | awk '{print $1 " " $3}' | sort
else
    print_warning "No coverage file generated"
fi

# Run race detection tests
echo ""
echo "Running race detection tests..."
echo "-----------------------------"

if go test -race ./...; then
    print_status "No race conditions detected"
else
    print_error "Race conditions detected"
fi

# Run benchmark tests
echo ""
echo "Running benchmark tests..."
echo "------------------------"

if go test -bench=. ./...; then
    print_status "Benchmark tests completed"
else
    print_warning "Some benchmark tests failed or were skipped"
fi

# Clean up
echo ""
echo "Cleaning up..."
rm -f coverage.out

echo ""
echo "Test Summary"
echo "============"
if [ ${#FAILED_PACKAGES[@]} -eq 0 ]; then
    print_status "All tests passed! 🎉"
    exit 0
else
    print_error "Some tests failed. Please check the output above."
    exit 1
fi