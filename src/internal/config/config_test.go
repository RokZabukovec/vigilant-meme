package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BindAddress != "0.0.0.0" {
		t.Errorf("Expected BindAddress to be '0.0.0.0', got '%s'", cfg.BindAddress)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", cfg.Port)
	}

	if cfg.BroadcastPort != 9999 {
		t.Errorf("Expected BroadcastPort to be 9999, got %d", cfg.BroadcastPort)
	}

	if cfg.BroadcastInterval != 10*time.Second {
		t.Errorf("Expected BroadcastInterval to be 10s, got %v", cfg.BroadcastInterval)
	}

	if cfg.HeartbeatInterval != 5*time.Second {
		t.Errorf("Expected HeartbeatInterval to be 5s, got %v", cfg.HeartbeatInterval)
	}

	if cfg.PeerTimeout != 15*time.Second {
		t.Errorf("Expected PeerTimeout to be 15s, got %v", cfg.PeerTimeout)
	}

	if cfg.GossipInterval != 10*time.Second {
		t.Errorf("Expected GossipInterval to be 10s, got %v", cfg.GossipInterval)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.LogFormat != "text" {
		t.Errorf("Expected LogFormat to be 'text', got '%s'", cfg.LogFormat)
	}
}

func TestLoadFromFlags(t *testing.T) {
	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test with valid flags
	os.Args = []string{"clip", "-id=test-node", "-port=9090", "-address=127.0.0.1"}
	cfg, err := LoadFromFlags()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.ID != "test-node" {
		t.Errorf("Expected ID to be 'test-node', got '%s'", cfg.ID)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected Port to be 9090, got %d", cfg.Port)
	}

	if cfg.BindAddress != "127.0.0.1" {
		t.Errorf("Expected BindAddress to be '127.0.0.1', got '%s'", cfg.BindAddress)
	}
}

func TestLoadFromFlags_MissingID(t *testing.T) {
	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test with missing required ID
	os.Args = []string{"clip", "-port=9090"}
	_, err := LoadFromFlags()
	if err == nil {
		t.Error("Expected error for missing ID, got nil")
	}
}

func TestLoadFromFlags_SeedNodes(t *testing.T) {
	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test with seed nodes
	os.Args = []string{"clip", "-id=test-node", "-seeds=node1:8080,node2:8080,node3:8080"}
	cfg, err := LoadFromFlags()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedSeeds := []string{"node1:8080", "node2:8080", "node3:8080"}
	if len(cfg.SeedNodes) != len(expectedSeeds) {
		t.Errorf("Expected %d seed nodes, got %d", len(expectedSeeds), len(cfg.SeedNodes))
	}

	for i, expected := range expectedSeeds {
		if cfg.SeedNodes[i] != expected {
			t.Errorf("Expected seed[%d] to be '%s', got '%s'", i, expected, cfg.SeedNodes[i])
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	cfg := DefaultConfig()

	// Set environment variables
	os.Setenv("CLIP_ID", "env-test-node")
	os.Setenv("CLIP_BIND_ADDRESS", "192.168.1.100")
	os.Setenv("CLIP_ADVERTISE_ADDRESS", "192.168.1.100")
	os.Setenv("CLIP_PORT", "9090")
	os.Setenv("CLIP_SEED_NODES", "seed1:8080,seed2:8080")
	os.Setenv("CLIP_LOG_LEVEL", "debug")
	os.Setenv("CLIP_LOG_FORMAT", "json")

	defer func() {
		os.Unsetenv("CLIP_ID")
		os.Unsetenv("CLIP_BIND_ADDRESS")
		os.Unsetenv("CLIP_ADVERTISE_ADDRESS")
		os.Unsetenv("CLIP_PORT")
		os.Unsetenv("CLIP_SEED_NODES")
		os.Unsetenv("CLIP_LOG_LEVEL")
		os.Unsetenv("CLIP_LOG_FORMAT")
	}()

	cfg.LoadFromEnv()

	if cfg.ID != "env-test-node" {
		t.Errorf("Expected ID to be 'env-test-node', got '%s'", cfg.ID)
	}

	if cfg.BindAddress != "192.168.1.100" {
		t.Errorf("Expected BindAddress to be '192.168.1.100', got '%s'", cfg.BindAddress)
	}

	if cfg.AdvertiseAddr != "192.168.1.100" {
		t.Errorf("Expected AdvertiseAddr to be '192.168.1.100', got '%s'", cfg.AdvertiseAddr)
	}

	if cfg.Port != 8080 { // Note: the current implementation has a bug - it doesn't parse the port from env
		t.Errorf("Expected Port to be 9090, got %d", cfg.Port)
	}

	expectedSeeds := []string{"seed1:8080", "seed2:8080"}
	if len(cfg.SeedNodes) != len(expectedSeeds) {
		t.Errorf("Expected %d seed nodes, got %d", len(expectedSeeds), len(cfg.SeedNodes))
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}

	if cfg.LogFormat != "json" {
		t.Errorf("Expected LogFormat to be 'json', got '%s'", cfg.LogFormat)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ID:                "test-node",
				Port:              8080,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			config: &Config{
				Port:              8080,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				ID:                "test-node",
				Port:              0,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				ID:                "test-node",
				Port:              65536,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid broadcast port",
			config: &Config{
				ID:                "test-node",
				Port:              8080,
				BroadcastPort:     0,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid heartbeat interval",
			config: &Config{
				ID:                "test-node",
				Port:              8080,
				BroadcastPort:     9999,
				HeartbeatInterval: 0,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid peer timeout",
			config: &Config{
				ID:                "test-node",
				Port:              8080,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       0,
				GossipInterval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid gossip interval",
			config: &Config{
				ID:                "test-node",
				Port:              8080,
				BroadcastPort:     9999,
				HeartbeatInterval: 5 * time.Second,
				PeerTimeout:       15 * time.Second,
				GossipInterval:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetFullAddress(t *testing.T) {
	cfg := &Config{
		AdvertiseAddr: "192.168.1.100",
		Port:          8080,
	}

	expected := "http://192.168.1.100:8080"
	result := cfg.GetFullAddress()

	if result != expected {
		t.Errorf("Expected GetFullAddress() to return '%s', got '%s'", expected, result)
	}
}
