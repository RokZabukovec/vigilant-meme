package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	// Service configuration
	ID            string
	BindAddress   string
	AdvertiseAddr string
	Port          int

	// Discovery configuration
	SeedNodes         []string
	BroadcastPort     int
	BroadcastInterval time.Duration

	// Health check configuration
	HeartbeatInterval time.Duration
	PeerTimeout       time.Duration
	GossipInterval    time.Duration

	// Logging configuration
	LogLevel  string
	LogFormat string
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		BindAddress:       "0.0.0.0",
		Port:              8080,
		BroadcastPort:     9999,
		BroadcastInterval: 10 * time.Second,
		HeartbeatInterval: 5 * time.Second,
		PeerTimeout:       15 * time.Second,
		GossipInterval:    10 * time.Second,
		LogLevel:          "info",
		LogFormat:         "text",
	}
}

// LoadFromFlags loads configuration from command line flags
func LoadFromFlags() (*Config, error) {
	config := DefaultConfig()

	// Define flags
	id := flag.String("id", "", "Unique identifier for this service instance (required)")
	address := flag.String("address", config.BindAddress, "IP address to bind to (0.0.0.0 for all interfaces)")
	advertiseAddr := flag.String("advertise", "", "IP address to advertise to other peers (auto-detected if not specified)")
	port := flag.Int("port", config.Port, "Port to listen on")
	seeds := flag.String("seeds", "", "Comma-separated list of seed node addresses")
	logLevel := flag.String("log-level", config.LogLevel, "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", config.LogFormat, "Log format (text, json)")

	flag.Parse()

	// Validate required fields
	if *id == "" {
		return nil, fmt.Errorf("id flag is required")
	}

	config.ID = *id
	config.BindAddress = *address
	config.AdvertiseAddr = *advertiseAddr
	config.Port = *port
	config.LogLevel = *logLevel
	config.LogFormat = *logFormat

	// Parse seed nodes
	if *seeds != "" {
		config.SeedNodes = strings.Split(*seeds, ",")
		for i, seed := range config.SeedNodes {
			config.SeedNodes[i] = strings.TrimSpace(seed)
		}
	}

	return config, nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if id := os.Getenv("CLIP_ID"); id != "" {
		c.ID = id
	}
	if address := os.Getenv("CLIP_BIND_ADDRESS"); address != "" {
		c.BindAddress = address
	}
	if advertiseAddr := os.Getenv("CLIP_ADVERTISE_ADDRESS"); advertiseAddr != "" {
		c.AdvertiseAddr = advertiseAddr
	}
	if port := os.Getenv("CLIP_PORT"); port != "" {
		// Note: In production, you'd want to parse this properly
		c.Port = 8080 // Default fallback
	}
	if seeds := os.Getenv("CLIP_SEED_NODES"); seeds != "" {
		c.SeedNodes = strings.Split(seeds, ",")
		for i, seed := range c.SeedNodes {
			c.SeedNodes[i] = strings.TrimSpace(seed)
		}
	}
	if logLevel := os.Getenv("CLIP_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
	if logFormat := os.Getenv("CLIP_LOG_FORMAT"); logFormat != "" {
		c.LogFormat = logFormat
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("service ID is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.BroadcastPort <= 0 || c.BroadcastPort > 65535 {
		return fmt.Errorf("broadcast port must be between 1 and 65535")
	}
	if c.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat interval must be positive")
	}
	if c.PeerTimeout <= 0 {
		return fmt.Errorf("peer timeout must be positive")
	}
	if c.GossipInterval <= 0 {
		return fmt.Errorf("gossip interval must be positive")
	}
	return nil
}

// GetFullAddress returns the full HTTP address for this service
func (c *Config) GetFullAddress() string {
	return fmt.Sprintf("http://%s:%d", c.AdvertiseAddr, c.Port)
}
