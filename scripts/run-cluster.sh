#!/bin/bash

# Script to run a local cluster of 3 nodes for testing

echo "Building the service..."
go build -o cluster-service

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting cluster nodes..."
echo ""

# Start node 1 (seed node)
echo "Starting Node 1 on port 8080..."
./cluster-service -id=node1 -address=localhost -port=8080 &
NODE1_PID=$!
sleep 2

# Start node 2
echo "Starting Node 2 on port 8081..."
./cluster-service -id=node2 -address=localhost -port=8081 -seeds=http://localhost:8080 &
NODE2_PID=$!
sleep 2

# Start node 3
echo "Starting Node 3 on port 8082..."
./cluster-service -id=node3 -address=localhost -port=8082 -seeds=http://localhost:8080 &
NODE3_PID=$!
sleep 2

echo ""
echo "==================================="
echo "Cluster is running!"
echo "==================================="
echo "Node 1: http://localhost:8080/status"
echo "Node 2: http://localhost:8081/status"
echo "Node 3: http://localhost:8082/status"
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
