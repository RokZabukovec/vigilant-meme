package main

import (
	"sync"
	"time"
)

type Peer struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	IsAlive  bool      `json:"is_alive"`
}

type PeerList struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

func NewPeerList() *PeerList {
	return &PeerList{
		peers: make(map[string]*Peer),
	}
}

func (pl *PeerList) Add(peer *Peer) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	peer.LastSeen = time.Now().UTC()
	peer.IsAlive = true
	pl.peers[peer.ID] = peer
}

func (pl *PeerList) Remove(id string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	delete(pl.peers, id)
}

func (pl *PeerList) Get(id string) (*Peer, bool) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	peer, exists := pl.peers[id]
	return peer, exists
}

// GetAll returns all peers
func (pl *PeerList) GetAll() []*Peer {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	peers := make([]*Peer, 0, len(pl.peers))
	for _, peer := range pl.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetAlive returns only alive peers
func (pl *PeerList) GetAlive() []*Peer {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	peers := make([]*Peer, 0)
	for _, peer := range pl.peers {
		if peer.IsAlive {
			peers = append(peers, peer)
		}
	}
	return peers
}

// MarkDead marks a peer as dead
func (pl *PeerList) MarkDead(id string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if peer, exists := pl.peers[id]; exists {
		peer.IsAlive = false
	}
}

// UpdateLastSeen updates the last seen time for a peer
func (pl *PeerList) UpdateLastSeen(id string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if peer, exists := pl.peers[id]; exists {
		peer.LastSeen = time.Now()
		peer.IsAlive = true
	}
}
