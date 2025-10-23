package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rokzabukovec/clip/internal/config"
	"github.com/rokzabukovec/clip/internal/discovery"
	"github.com/rokzabukovec/clip/internal/handlers"
	"github.com/rokzabukovec/clip/internal/peer"
	"github.com/rokzabukovec/clip/pkg/network"
)

// Service represents the main service instance
type Service struct {
	config        *config.Config
	peerList      *peer.PeerList
	discovery     *discovery.DiscoveryService
	handlers      *handlers.Handler
	stopChan      chan struct{}
	advertiseAddr string
}

// NewService creates a new service instance
func NewService(cfg *config.Config) *Service {
	peerList := peer.NewPeerList()

	// Determine advertise address
	advertiseAddr := cfg.AdvertiseAddr
	if advertiseAddr == "" {
		advertiseAddr = network.GetOutboundIP()
		if advertiseAddr == "" {
			log.Println("Warning: Could not auto-detect network IP. Using localhost.")
			log.Println("This will prevent other computers from connecting to this node.")
			log.Printf("Please specify an IP address manually using -advertise flag.\n")
			advertiseAddr = "localhost"
		} else {
			log.Printf("Auto-detected network IP: %s", advertiseAddr)
		}
	} else {
		if advertiseAddr == "localhost" || advertiseAddr == "127.0.0.1" {
			fmt.Println("\n⚠️  WARNING: You are using localhost as the advertise address!")
			fmt.Println("   This will only work for peers on the same machine.")
			fmt.Println("   For cross-machine communication, use your network IP address.")
		}
	}

	serviceAddr := fmt.Sprintf("http://%s:%d", advertiseAddr, cfg.Port)

	discoveryService := discovery.NewDiscoveryService(
		cfg.ID,
		serviceAddr,
		cfg.Port,
		func(id, address string) {
			// Callback when a peer is discovered via broadcast
			p := &peer.Peer{
				ID:      id,
				Address: address,
			}
			peerList.Add(p)
		},
	)

	handler := handlers.NewHandler(peerList, cfg.ID, func(p *peer.Peer) {
		// Callback when a peer joins
		log.Printf("Peer joined: %s at %s", p.ID, p.Address)
	})

	return &Service{
		config:        cfg,
		peerList:      peerList,
		discovery:     discoveryService,
		handlers:      handler,
		stopChan:      make(chan struct{}),
		advertiseAddr: advertiseAddr,
	}
}

// Start starts the service
func (s *Service) Start() error {
	// Start broadcast discovery for automatic peer detection on LAN
	s.discovery.StartBroadcastListener()
	go s.discovery.StartBroadcastAnnouncer()

	// Register with seed nodes if provided
	if len(s.config.SeedNodes) > 0 {
		if err := s.registerWithSeeds(); err != nil {
			log.Printf("Warning: Failed to register with seed nodes: %v", err)
		}
	} else {
		log.Printf("No seed nodes specified - relying on broadcast discovery")
	}

	go s.heartbeatLoop()
	go s.healthCheckLoop()
	go s.gossipLoop()

	log.Printf("Service %s started (binding: %s:%d, advertising: %s:%d)",
		s.config.ID, s.config.BindAddress, s.config.Port, s.advertiseAddr, s.config.Port)
	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	close(s.stopChan)
	s.discovery.Stop()
}

// GetFullAddress returns the full HTTP address for this service
func (s *Service) GetFullAddress() string {
	return fmt.Sprintf("http://%s:%d", s.advertiseAddr, s.config.Port)
}

// GetHandlers returns the HTTP handlers
func (s *Service) GetHandlers() *handlers.Handler {
	return s.handlers
}

// GetPeerList returns the peer list
func (s *Service) GetPeerList() *peer.PeerList {
	return s.peerList
}

// registerWithSeeds registers this service with seed nodes
func (s *Service) registerWithSeeds() error {
	thisPeer := &peer.Peer{
		ID:      s.config.ID,
		Address: s.GetFullAddress(),
	}

	for _, seed := range s.config.SeedNodes {
		if seed == s.GetFullAddress() {
			continue
		}

		if err := s.sendJoinRequest(seed, thisPeer); err != nil {
			log.Printf("Failed to register with seed %s: %v", seed, err)
			continue
		}
		log.Printf("Successfully registered with seed node: %s", seed)
	}

	return nil
}

// sendJoinRequest sends a join request to a peer
func (s *Service) sendJoinRequest(peerAddr string, p *peer.Peer) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := http.Post(peerAddr+"/join", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("join request failed with status: %d", resp.StatusCode)
	}

	var peers []*peer.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return err
	}

	for _, peer := range peers {
		if peer.ID != s.config.ID {
			s.peerList.Add(peer)
		}
	}

	return nil
}

// heartbeatLoop sends periodic heartbeats to known peers
func (s *Service) heartbeatLoop() {
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendHeartbeats()
		case <-s.stopChan:
			return
		}
	}
}

// sendHeartbeats sends heartbeats to all alive peers
func (s *Service) sendHeartbeats() {
	peers := s.peerList.GetAlive()

	heartbeat := map[string]string{
		"id":      s.config.ID,
		"address": s.GetFullAddress(),
	}

	for _, p := range peers {
		go func(peer *peer.Peer) {
			data, _ := json.Marshal(heartbeat)
			resp, err := http.Post(peer.Address+"/heartbeat", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Failed to send heartbeat to %s: %v", peer.ID, err)
				return
			}
			defer resp.Body.Close()
		}(p)
	}
}

// healthCheckLoop periodically checks peer health
func (s *Service) healthCheckLoop() {
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkPeerHealth()
		case <-s.stopChan:
			return
		}
	}
}

// checkPeerHealth marks peers as dead if they haven't been seen recently
func (s *Service) checkPeerHealth() {
	peers := s.peerList.GetAll()
	now := time.Now()

	for _, peer := range peers {
		if now.Sub(peer.LastSeen) > s.config.PeerTimeout {
			if peer.IsAlive {
				log.Printf("Peer %s marked as dead (last seen: %v ago)", peer.ID, now.Sub(peer.LastSeen))
				s.peerList.MarkDead(peer.ID)
			}
		}
	}
}

// gossipLoop periodically exchanges peer information with other peers
func (s *Service) gossipLoop() {
	ticker := time.NewTicker(s.config.GossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.gossipWithPeers()
		case <-s.stopChan:
			return
		}
	}
}

// gossipWithPeers exchanges peer information with other peers
func (s *Service) gossipWithPeers() {
	peers := s.peerList.GetAlive()
	if len(peers) == 0 {
		return
	}

	myPeers := s.peerList.GetAll()

	for _, p := range peers {
		go func(peer *peer.Peer) {
			data, _ := json.Marshal(myPeers)
			resp, err := http.Post(peer.Address+"/gossip", "application/json", bytes.NewBuffer(data))
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}(p)

		break
	}
}
