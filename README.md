# Clip - Distributed Network Service

A professional, production-ready peer-to-peer distributed network service written in Go. All instances automatically discover and track each other's locations using a gossip protocol and heartbeat mechanism.

## 🚀 Features

- **⚡ Automatic Broadcast Discovery**: Instances automatically find each other on the same LAN via UDP broadcast - NO seed nodes required!
- **🔍 Automatic Peer Discovery**: Instances can also discover each other through optional seed nodes and gossip protocol
- **🌐 Cross-Machine Support**: Works seamlessly across different computers on the same network
- **🎯 Auto IP Detection**: Automatically detects and advertises the correct network IP address
- **💓 Health Monitoring**: Continuous heartbeat mechanism to detect when peers go offline
- **📡 Gossip Protocol**: Efficient peer information propagation across the network
- **🔌 REST API**: Simple HTTP API for querying cluster state and managing peers
- **📦 No External Dependencies**: Self-contained implementation without requiring external service discovery tools
- **🐳 Docker Support**: Ready-to-use Docker containers and Docker Compose setup
- **🔧 Professional Structure**: Clean, maintainable codebase with proper separation of concerns

## 📁 Project Structure

```
Clip/
├── src/                          # Source code
│   ├── cmd/clip/                 # Main application entry point
│   │   └── main.go
│   ├── internal/                 # Private application code
│   │   ├── config/              # Configuration management
│   │   ├── discovery/           # Peer discovery logic
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── logger/              # Structured logging
│   │   ├── peer/                # Peer management
│   │   └── service/             # Main service logic
│   └── pkg/                     # Public library code
│       ├── network/             # Network utilities
│       └── utils/               # General utilities
├── scripts/                     # Utility scripts
├── configs/                     # Configuration files
├── docs/                        # Documentation
├── build/                       # Build artifacts (generated)
├── Makefile                     # Build automation
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Multi-container setup
└── README.md                    # This file
```

## 🛠️ Quick Start

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)
- Docker (optional, for containerized deployment)

### Development Setup

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd Clip
   ./scripts/dev.sh
   ```

2. **Build the application**:
   ```bash
   make build
   ```

3. **Run a single node**:
   ```bash
   make run
   ```

4. **Start a test cluster**:
   ```bash
   ./scripts/start-cluster.sh
   ```

### Using Docker

1. **Build and run with Docker Compose**:
   ```bash
   docker-compose up --build
   ```

2. **Or build and run a single container**:
   ```bash
   make docker-build
   make docker-run
   ```

## 🏗️ Building

### Using Make (Recommended)

```bash
# Build for current platform
make build

# Build for multiple platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage

# Clean build artifacts
make clean

# Development mode with auto-reload
make dev
```

### Manual Build

```bash
cd src
go build -o ../build/clip ./cmd/clip
```

## 🚀 Running

### Automatic Discovery (Recommended)

**Instances automatically find each other on the same network via broadcast - no seed nodes needed!**

```bash
# Computer A:
./build/clip -id=nodeA -port=8080

# Computer B:
./build/clip -id=nodeB -port=8080

# Computer C:
./build/clip -id=nodeC -port=8080

# That's it! They'll discover each other automatically via UDP broadcast.
```

### Local Testing

```bash
# Start the first node (it will auto-detect your IP)
./build/clip -id=node1 -port=8080

# Start additional nodes - NO SEEDS NEEDED (broadcast discovery)
./build/clip -id=node2 -port=8081
./build/clip -id=node3 -port=8082
```

### Using Seed Nodes (Optional)

For faster initial discovery or cross-subnet communication:

```bash
./build/clip -id=nodeB -port=8080 -seeds=http://192.168.1.100:8080
./build/clip -id=nodeC -port=8080 -seeds=http://192.168.1.100:8080
```

## ⚙️ Configuration

### Command Line Flags

- `-id`: Unique identifier for the service instance (required)
- `-address`: IP address to bind to (default: `0.0.0.0` - all interfaces)
- `-advertise`: IP address to advertise to other peers (auto-detected if not specified)
- `-port`: Port to listen on (default: 8080)
- `-seeds`: Comma-separated list of seed node addresses (optional)
- `-log-level`: Log level (debug, info, warn, error) (default: info)
- `-log-format`: Log format (text, json) (default: text)

### Environment Variables

- `CLIP_ID`: Service identifier
- `CLIP_BIND_ADDRESS`: Bind address
- `CLIP_ADVERTISE_ADDRESS`: Advertise address
- `CLIP_PORT`: Service port
- `CLIP_SEED_NODES`: Comma-separated seed nodes
- `CLIP_LOG_LEVEL`: Log level
- `CLIP_LOG_FORMAT`: Log format

### Configuration File

See `configs/config.yaml` for a complete configuration example.

## 📡 API Endpoints

### GET /status
Returns the status of this service instance including all known peers.

```bash
curl http://localhost:8080/status
```

Response:
```json
{
  "id": "node1",
  "address": "http://192.168.1.100:8080",
  "total_peers": 3,
  "alive_peers": 2,
  "peers": [
    {
      "id": "node1",
      "address": "http://192.168.1.100:8080",
      "last_seen": "2025-01-17T10:30:00Z",
      "is_alive": true
    },
    {
      "id": "node2",
      "address": "http://192.168.1.101:8080",
      "last_seen": "2025-01-17T10:29:55Z",
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

## 🧪 Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
# Start a test cluster
./scripts/start-cluster.sh

# In another terminal, test the cluster
curl http://localhost:8080/status | jq
curl http://localhost:8081/status | jq
curl http://localhost:8082/status | jq
```

### Test Coverage
```bash
make test-coverage
```

## 🐳 Docker Deployment

### Single Container
```bash
# Build image
make docker-build

# Run container
make docker-run
```

### Multi-Container Cluster
```bash
# Start 3-node cluster
docker-compose up --build

# Scale to more nodes
docker-compose up --scale clip-node1=1 --scale clip-node2=1 --scale clip-node3=1
```

## 🔧 Development

### Code Structure

- **`cmd/clip/`**: Main application entry point
- **`internal/config/`**: Configuration management with flag and environment support
- **`internal/service/`**: Core service logic and orchestration
- **`internal/peer/`**: Peer management and thread-safe operations
- **`internal/discovery/`**: UDP broadcast discovery mechanism
- **`internal/handlers/`**: HTTP request handlers
- **`internal/logger/`**: Structured logging with multiple output formats
- **`pkg/network/`**: Network utilities and IP detection
- **`pkg/utils/`**: General utility functions

### Adding New Features

1. **New HTTP endpoints**: Add handlers in `internal/handlers/`
2. **New configuration options**: Extend `internal/config/`
3. **New discovery mechanisms**: Add to `internal/discovery/`
4. **New peer operations**: Extend `internal/peer/`

### Code Quality

The project includes:
- **Structured logging** with configurable levels and formats
- **Comprehensive error handling** with proper error propagation
- **Thread-safe operations** for concurrent access
- **Configuration validation** with clear error messages
- **Clean separation of concerns** with well-defined interfaces

## 🚨 Troubleshooting

### Broadcast Discovery Not Working

1. **Check Firewall**: Ensure UDP port 9999 is open
   ```bash
   # Linux
   sudo ufw allow 9999/udp
   sudo ufw allow 8080/tcp
   
   # Windows (PowerShell as Administrator)
   New-NetFirewallRule -DisplayName "Clip Discovery" -Direction Inbound -LocalPort 9999 -Protocol UDP -Action Allow
   New-NetFirewallRule -DisplayName "Clip HTTP" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
   ```

2. **Check Logs**: Look for "Broadcast discovery listener started" messages

3. **Different Subnets**: Use seed nodes for cross-subnet communication

### Peers Can't Connect

1. **Verify Network Connectivity**: `ping <peer-ip>`
2. **Check Advertise IP**: Ensure the service advertises the correct network IP
3. **Manual IP Override**: Use `-advertise` flag to specify IP manually

## 📊 Monitoring

### Health Checks

The service provides built-in health monitoring:
- **Heartbeat mechanism**: Detects dead peers automatically
- **Health check endpoint**: `/status` shows cluster health
- **Structured logging**: Easy integration with log aggregation systems

### Metrics

Monitor these key metrics:
- Total peers in cluster
- Alive peers count
- Last seen timestamps
- Discovery success rate

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Format code: `make fmt`
7. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

- Built with Go's excellent standard library
- Inspired by distributed systems principles
- Designed for simplicity and reliability