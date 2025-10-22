package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	HeartbeatInterval = 5 * time.Second
	PeerTimeout       = 15 * time.Second
	GossipInterval    = 10 * time.Second
)

type Service struct {
	ID            string
	BindAddress   string
	AdvertiseAddr string
	Port          int
	PeerList      *PeerList
	SeedNodes     []string
	stopChan      chan struct{}
}

func NewService(id, bindAddress, advertiseAddr string, port int, seedNodes []string) *Service {
	return &Service{
		ID:            id,
		BindAddress:   bindAddress,
		AdvertiseAddr: advertiseAddr,
		Port:          port,
		PeerList:      NewPeerList(),
		SeedNodes:     seedNodes,
		stopChan:      make(chan struct{}),
	}
}

func (s *Service) Start() error {
	// Start broadcast discovery for automatic peer detection on LAN
	s.StartBroadcastListener()
	go s.StartBroadcastAnnouncer()

	// Register with seed nodes if provided (optional now with broadcast discovery)
	if len(s.SeedNodes) > 0 {
		if err := s.registerWithSeeds(); err != nil {
			log.Printf("Warning: Failed to register with seed nodes: %v", err)
		}
	} else {
		log.Printf("No seed nodes specified - relying on broadcast discovery")
	}

	go s.heartbeatLoop()
	go s.healthCheckLoop()
	go s.gossipLoop()

	log.Printf("Service %s started (binding: %s:%d, advertising: %s:%d)", s.ID, s.BindAddress, s.Port, s.AdvertiseAddr, s.Port)
	return nil
}

func (s *Service) Stop() {
	close(s.stopChan)
}

func (s *Service) GetFullAddress() string {
	return fmt.Sprintf("http://%s:%d", s.AdvertiseAddr, s.Port)
}

func (s *Service) registerWithSeeds() error {
	thisPeer := &Peer{
		ID:      s.ID,
		Address: s.GetFullAddress(),
	}

	for _, seed := range s.SeedNodes {
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

func (s *Service) sendJoinRequest(peerAddr string, peer *Peer) error {
	data, err := json.Marshal(peer)
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

	var peers []*Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return err
	}

	for _, p := range peers {
		if p.ID != s.ID {
			s.PeerList.Add(p)
		}
	}

	return nil
}

func (s *Service) heartbeatLoop() {
	ticker := time.NewTicker(HeartbeatInterval)
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

func (s *Service) sendHeartbeats() {
	peers := s.PeerList.GetAlive()

	heartbeat := map[string]string{
		"id":      s.ID,
		"address": s.GetFullAddress(),
	}

	for _, peer := range peers {
		go func(p *Peer) {
			data, _ := json.Marshal(heartbeat)
			resp, err := http.Post(p.Address+"/heartbeat", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Failed to send heartbeat to %s: %v", p.ID, err)
				return
			}
			defer resp.Body.Close()
		}(peer)
	}
}

func (s *Service) healthCheckLoop() {
	ticker := time.NewTicker(HeartbeatInterval)
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

func (s *Service) checkPeerHealth() {
	peers := s.PeerList.GetAll()
	now := time.Now()

	for _, peer := range peers {
		if now.Sub(peer.LastSeen) > PeerTimeout {
			if peer.IsAlive {
				log.Printf("Peer %s marked as dead (last seen: %v ago)", peer.ID, now.Sub(peer.LastSeen))
				s.PeerList.MarkDead(peer.ID)
			}
		}
	}
}

func (s *Service) gossipLoop() {
	ticker := time.NewTicker(GossipInterval)
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

func (s *Service) gossipWithPeers() {
	peers := s.PeerList.GetAlive()
	if len(peers) == 0 {
		return
	}

	myPeers := s.PeerList.GetAll()

	for _, peer := range peers {
		go func(p *Peer) {
			data, _ := json.Marshal(myPeers)
			resp, err := http.Post(p.Address+"/gossip", "application/json", bytes.NewBuffer(data))
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}(peer)

		break
	}
}
