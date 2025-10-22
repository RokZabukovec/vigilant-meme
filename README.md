# Distributed Network Service

A peer-to-peer distributed network service written in Go where all instances have knowledge of each other's locations. Each instance automatically discovers and tracks other instances in the network using a gossip protocol and heartbeat mechanism.

## Features

- **⚡ Automatic Broadcast Discovery**: Instances automatically find each other on the same LAN via UDP broadcast - NO seed nodes required!
- **Automatic Peer Discovery**: Instances can also discover each other through optional seed nodes and gossip protocol
- **Cross-Machine Support**: Works seamlessly across different computers on the same network
- **Auto IP Detection**: Automatically detects and advertises the correct network IP address
- **Health Monitoring**: Continuous heartbeat mechanism to detect when peers go offline
- **Gossip Protocol**: Efficient peer information propagation across the network
- **REST API**: Simple HTTP API for querying cluster state and managing peers
- **No External Dependencies**: Self-contained implementation without requiring external service discovery tools

## How It Works

1. **Broadcast Discovery**: Instances periodically broadcast their presence on UDP port 9999 to discover peers on the same LAN
2. **Bootstrap** (Optional): New instances can also register with one or more seed nodes for faster discovery or cross-subnet connections
3. **Heartbeat**: Instances send periodic heartbeats to known peers to indicate they're alive
4. **Gossip**: Instances periodically exchange their peer lists to discover new peers
5. **Health Check**: Peers that haven't been seen within the timeout period are marked as dead

## Building

```bash
go build -o Clip
```

## Quick Start with Helper Script

For convenience, you can use the `start-node.sh` helper script:

```bash
# Start first node
./start-node.sh node1

# Start second node connecting to first node
./start-node.sh node2 8081 http://192.168.1.100:8080

# Start third node on different port
./start-node.sh node3 8082 http://192.168.1.100:8080
```

## Running

### Automatic Discovery (Recommended for Same Network)

**NEW!** Instances now automatically find each other on the same network via broadcast - no seed nodes needed!

```bash
# Start instances on the same network - they'll find each other automatically!

# Computer A:
./Clip -id=nodeA -port=8080

# Computer B:
./Clip -id=nodeB -port=8080

# Computer C:
./Clip -id=nodeC -port=8080

# That's it! They'll discover each other automatically via UDP broadcast.
```

**Test script for local testing:**
```bash
./test-broadcast.sh  # Starts 3 nodes that auto-discover each other
```

### Local Testing (Single Machine)

For testing on a single machine, the service will auto-detect your network IP:

```bash
# Start the first node (it will auto-detect your IP)
./Clip -id=node1 -port=8080

# Start additional nodes - NO SEEDS NEEDED (broadcast discovery)
./Clip -id=node2 -port=8081
./Clip -id=node3 -port=8082
```

### Cross-Machine Deployment (Multiple Computers)

With broadcast discovery, deployment is incredibly simple:

**On Computer A (e.g., 192.168.1.100):**
```bash
./Clip -id=nodeA -port=8080
# Will auto-detect as 192.168.1.100 and broadcast presence
```

**On Computer B (e.g., 192.168.1.101):**
```bash
./Clip -id=nodeB -port=8080
# Will auto-detect as 192.168.1.101 and discover nodeA via broadcast
```

**On Computer C (e.g., 192.168.1.102):**
```bash
./Clip -id=nodeC -port=8080
# Will auto-detect as 192.168.1.102 and discover both nodes via broadcast
```

**Optional: Using Seed Nodes** (for faster discovery or cross-subnet):
```bash
# If you want faster initial discovery or need cross-subnet communication
./Clip -id=nodeB -port=8080 -seeds=http://192.168.1.100:8080
./Clip -id=nodeC -port=8080 -seeds=http://192.168.1.100:8080
```

### Command Line Flags

- `-id`: Unique identifier for the service instance (required)
- `-address`: IP address to bind to (default: `0.0.0.0` - all interfaces)
- `-advertise`: IP address to advertise to other peers (auto-detected if not specified)
- `-port`: Port to listen on (default: 8080)
- `-seeds`: Comma-separated list of seed node addresses (OPTIONAL - broadcast discovery is automatic)

**Note:** 
- The service automatically uses UDP broadcast on port 9999 for peer discovery on the same LAN
- Seed nodes are now optional but can be used for faster discovery or cross-subnet communication
- By default, the service binds to `0.0.0.0` (all network interfaces) and auto-detects your network IP

## API Endpoints

### GET /status
Returns the status of this service instance including all known peers.

```bash
curl http://localhost:8080/status
```

Response:
```json
{
  "id": "node1",
  "address": "http://localhost:8080",
  "total_peers": 3,
  "alive_peers": 2,
  "peers": [
    {
      "id": "node1",
      "address": "http://localhost:8080",
      "last_seen": "2025-10-17T10:30:00Z",
      "is_alive": true
    },
    {
      "id": "node2",
      "address": "http://localhost:8081",
      "last_seen": "2025-10-17T10:29:55Z",
      "is_alive": true
    }
  ]
}
```

### GET /peers
Returns list of all known peers.

```bash
curl http://localhost:8080/peers
```

### POST /join
Used internally by nodes to join the cluster. Returns the current peer list.

### POST /heartbeat
Used internally by nodes to send heartbeat messages.

### POST /gossip
Used internally by nodes to exchange peer information.

## Testing the Cluster

After starting multiple instances, you can verify they know about each other:

```bash
# Check node 1's view of the cluster
curl http://localhost:8080/status | jq

# Check node 2's view of the cluster
curl http://localhost:8081/status | jq

# Check node 3's view of the cluster
curl http://localhost:8082/status | jq
```

All nodes should show the same peers in their lists (with slight timing differences in `last_seen` timestamps).

## Network IP Detection

When you start the service, it will display all detected network IPs:

```
=== Service Started ===
ID:               node1
Binding to:       0.0.0.0:8080
Advertising as:   http://192.168.1.100:8080
Seed nodes:       None (first node in cluster)

Detected network IPs:
  - 192.168.1.100
  - 10.0.0.5

Available endpoints:
  ...
```

This helps you know which IP address your node is advertising to other peers in the network.

## Configuration

You can adjust timing parameters in `service.go` and `discovery.go`:

**Service Parameters:**
- `HeartbeatInterval`: How often to send heartbeats (default: 5 seconds)
- `PeerTimeout`: How long before marking a peer as dead (default: 15 seconds)
- `GossipInterval`: How often to exchange peer information (default: 10 seconds)

**Broadcast Discovery Parameters:**
- `BroadcastPort`: UDP port for broadcast discovery (default: 9999)
- `BroadcastInterval`: How often to broadcast presence (default: 10 seconds)

## Troubleshooting

### Broadcast Discovery Not Working

If nodes on the same network aren't discovering each other automatically:

1. **Check Firewall for UDP Port 9999**: Broadcast discovery requires UDP port 9999 to be open
   ```bash
   # On Linux
   sudo ufw allow 9999/udp
   sudo ufw allow 8080/tcp
   
   # On Windows (PowerShell as Administrator)
   New-NetFirewallRule -DisplayName "Clip Discovery" -Direction Inbound -LocalPort 9999 -Protocol UDP -Action Allow
   New-NetFirewallRule -DisplayName "Clip HTTP" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
   ```

2. **Check Logs**: Look for "Broadcast discovery listener started" and "Discovered new peer via broadcast" messages

3. **Different Subnets**: Broadcast only works on the same subnet. If nodes are on different subnets, use seed nodes:
   ```bash
   ./Clip -id=node1 -seeds=http://192.168.1.100:8080
   ```

### Peers Can't Connect Across Network

If nodes on different computers can't see each other:

1. **Verify Network Connectivity**: Make sure computers can ping each other
   ```bash
   ping 192.168.1.100
   ```

2. **Check Advertised IP**: When the service starts, it shows "Advertising as: ...". Make sure this is the correct network IP that other computers can reach.

3. **Manual IP Override**: If auto-detection fails, manually specify the IP:
   ```bash
   ./Clip -id=node1 -advertise=192.168.1.100 -port=8080
   ```

### Service Says "Could not auto-detect network IP"

This means your computer doesn't have a network connection. You can:
- Connect to a network (WiFi or Ethernet)
- Manually specify the IP with `-advertise` flag
- Use `localhost` for local-only testing (nodes on other computers won't be able to connect)

## Architecture

```
┌─────────────┐     heartbeat/gossip     ┌─────────────┐
│   Node 1    │◄────────────────────────►│   Node 2    │
│ :8080       │                          │ :8081       │
└─────────────┘                          └─────────────┘
       ▲                                        ▲
       │                                        │
       │          heartbeat/gossip              │
       │                                        │
       └────────────────┬───────────────────────┘
                        │
                        ▼
                 ┌─────────────┐
                 │   Node 3    │
                 │ :8082       │
                 └─────────────┘
```

## License

MIT
