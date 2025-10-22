#!/bin/bash

# Script to test the cluster by querying all nodes

echo "Testing Cluster Connectivity..."
echo ""

# Function to query a node's status
query_node() {
    port=$1
    echo "==================================="
    echo "Node on port $port:"
    echo "==================================="
    curl -s http://localhost:$port/status | jq '.' 2>/dev/null || curl -s http://localhost:$port/status
    echo ""
}

# Wait a bit for nodes to discover each other
echo "Waiting for peer discovery..."
sleep 5
echo ""

# Query each node
query_node 8080
query_node 8081
query_node 8082

echo "==================================="
echo "Test complete!"
echo "==================================="
echo ""
echo "All nodes should show the same peers in their lists."
echo "You can also test manually with:"
echo "  curl http://localhost:8080/status"
echo "  curl http://localhost:8081/status"
echo "  curl http://localhost:8082/status"
