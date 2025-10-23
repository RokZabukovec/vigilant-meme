#!/bin/bash

# Script to test automatic broadcast discovery - no seed nodes needed!

echo "Building the service..."
go build -o Clip

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting nodes with automatic broadcast discovery..."
echo ""

# Start node 1 - NO SEEDS NEEDED!
echo "Starting Node 1 on port 8080 (no seeds - will auto-discover)..."
./Clip -id=node1 -port=8080 &
NODE1_PID=$!
sleep 2

# Start node 2 - NO SEEDS NEEDED!
echo "Starting Node 2 on port 8081 (no seeds - will auto-discover)..."
./Clip -id=node2 -port=8081 &
NODE2_PID=$!
sleep 2

# Start node 3 - NO SEEDS NEEDED!
echo "Starting Node 3 on port 8082 (no seeds - will auto-discover)..."
./Clip -id=node3 -port=8082 &
NODE3_PID=$!
sleep 2

echo ""
echo "==================================="
echo "Cluster is running with AUTO-DISCOVERY!"
echo "==================================="
echo "Node 1: http://localhost:8080/status"
echo "Node 2: http://localhost:8081/status"
echo "Node 3: http://localhost:8082/status"
echo ""
echo "⚡ All nodes will automatically find each other via UDP broadcast!"
echo "   No need to specify seed nodes when on the same network."
echo ""
echo "Press Ctrl+C to stop all nodes"
echo "==================================="
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping all nodes..."
    kill $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    wait $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    echo "All nodes stopped."
    exit 0
}

# Trap SIGINT (Ctrl+C) and SIGTERM
trap cleanup SIGINT SIGTERM

# Wait for all background processes
wait
