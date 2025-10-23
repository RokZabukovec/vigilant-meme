#!/bin/bash

# Start Cluster Script for Clip
# This script starts multiple Clip nodes for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_PATH="../build/clip"
NODES=3
BASE_PORT=8080
BASE_BROADCAST_PORT=9999

echo -e "${BLUE}Starting Clip Cluster with $NODES nodes${NC}"

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY_PATH${NC}"
    echo -e "${YELLOW}Please run 'make build' first${NC}"
    exit 1
fi

# Function to start a node
start_node() {
    local node_id=$1
    local port=$2
    local broadcast_port=$3
    
    echo -e "${GREEN}Starting node $node_id on port $port (broadcast: $broadcast_port)${NC}"
    
    # Start the node in background
    $BINARY_PATH \
        -id="node$node_id" \
        -port=$port \
        -log-level=info &
    
    local pid=$!
    echo $pid > "node${node_id}.pid"
    
    # Wait a bit for the node to start
    sleep 2
    
    # Check if the node is running
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}Node $node_id started successfully (PID: $pid)${NC}"
    else
        echo -e "${RED}Failed to start node $node_id${NC}"
        exit 1
    fi
}

# Function to stop all nodes
stop_nodes() {
    echo -e "${YELLOW}Stopping all nodes...${NC}"
    for i in $(seq 1 $NODES); do
        if [ -f "node${i}.pid" ]; then
            local pid=$(cat "node${i}.pid")
            if kill -0 $pid 2>/dev/null; then
                echo -e "${YELLOW}Stopping node $i (PID: $pid)${NC}"
                kill $pid
            fi
            rm -f "node${i}.pid"
        fi
    done
    echo -e "${GREEN}All nodes stopped${NC}"
}

# Function to show cluster status
show_status() {
    echo -e "${BLUE}Cluster Status:${NC}"
    for i in $(seq 1 $NODES); do
        local port=$((BASE_PORT + i - 1))
        echo -e "${YELLOW}Node $i:${NC} http://localhost:$port/status"
        curl -s "http://localhost:$port/status" | jq '.id, .address, .total_peers, .alive_peers' 2>/dev/null || echo "  Not responding"
        echo
    done
}

# Trap to cleanup on exit
trap stop_nodes EXIT INT TERM

# Start all nodes
for i in $(seq 1 $NODES); do
    local port=$((BASE_PORT + i - 1))
    local broadcast_port=$((BASE_BROADCAST_PORT + i - 1))
    start_node $i $port $broadcast_port
done

echo -e "${GREEN}All nodes started!${NC}"
echo -e "${BLUE}Cluster endpoints:${NC}"
for i in $(seq 1 $NODES); do
    local port=$((BASE_PORT + i - 1))
    echo -e "  Node $i: http://localhost:$port"
done

echo -e "${YELLOW}Press Ctrl+C to stop all nodes${NC}"

# Wait for user interrupt
while true; do
    sleep 5
    show_status
done