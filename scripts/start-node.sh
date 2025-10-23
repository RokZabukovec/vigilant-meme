#!/bin/bash

# Helper script to start a cluster node with proper configuration

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if ID is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <node-id> [port] [seed-address]"
    echo ""
    echo "Examples:"
    echo "  $0 node1                              # Start first node on default port 8080"
    echo "  $0 node2 8081 http://192.168.1.100:8080  # Start node on port 8081, connecting to seed"
    echo "  $0 node3 8080 http://192.168.1.100:8080  # Start node on port 8080, connecting to seed"
    exit 1
fi

NODE_ID=$1
PORT=${2:-8080}
SEED=${3:-}

# Build if needed
if [ ! -f "./Clip" ]; then
    echo -e "${YELLOW}Building application...${NC}"
    go build -o Clip
    if [ $? -ne 0 ]; then
        echo "Build failed!"
        exit 1
    fi
fi

# Build command
CMD="./Clip -id=$NODE_ID -port=$PORT"

if [ ! -z "$SEED" ]; then
    CMD="$CMD -seeds=$SEED"
fi

echo -e "${GREEN}Starting node: $NODE_ID on port $PORT${NC}"
if [ ! -z "$SEED" ]; then
    echo -e "${GREEN}Connecting to seed: $SEED${NC}"
fi
echo ""

# Execute
exec $CMD
