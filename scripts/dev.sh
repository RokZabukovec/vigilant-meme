#!/bin/bash

# Development Script for Clip
# This script sets up the development environment and runs the service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up Clip development environment${NC}"

# Check if we're in the right directory
if [ ! -f "Makefile" ]; then
    echo -e "${RED}Error: Makefile not found. Please run this script from the project root.${NC}"
    exit 1
fi

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
make deps

# Format code
echo -e "${YELLOW}Formatting code...${NC}"
make fmt

# Run linter
echo -e "${YELLOW}Running linter...${NC}"
make vet

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
make test

# Build the application
echo -e "${YELLOW}Building application...${NC}"
make build

echo -e "${GREEN}Development environment ready!${NC}"
echo -e "${BLUE}Available commands:${NC}"
echo -e "  ${YELLOW}make run${NC}        - Run the application"
echo -e "  ${YELLOW}make dev${NC}        - Run with auto-reload (requires air)"
echo -e "  ${YELLOW}make test${NC}       - Run tests"
echo -e "  ${YELLOW}make build${NC}      - Build the application"
echo -e "  ${YELLOW}make clean${NC}      - Clean build artifacts"
echo -e "  ${YELLOW}./scripts/start-cluster.sh${NC} - Start a test cluster"