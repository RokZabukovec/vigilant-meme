package peer

import (
	"sync"
	"time"
)

// Peer represents a peer in the network
type Peer struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	IsAlive  bool      `json:"is_alive"`
}

// PeerList manages a thread-safe collection of peers
type PeerList struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

// NewPeerList creates a new peer list
func NewPeerList() *PeerList {
	return &PeerList{
		peers: make(map[string]*Peer),
	}
}

// Add adds a peer to the list
func (pl *PeerList) Add(peer *Peer) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	peer.LastSeen = time.Now().UTC()
	peer.IsAlive = true
	pl.peers[peer.ID] = peer
}

// Remove removes a peer from the list
func (pl *PeerList) Remove(id string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	delete(pl.peers, id)
}

// Get retrieves a peer by ID
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

// Count returns the total number of peers
func (pl *PeerList) Count() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return len(pl.peers)
}

// CountAlive returns the number of alive peers
func (pl *PeerList) CountAlive() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	count := 0
	for _, peer := range pl.peers {
		if peer.IsAlive {
			count++
		}
	}
	return count
}

// Exists checks if a peer exists
func (pl *PeerList) Exists(id string) bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	_, exists := pl.peers[id]
	return exists
}
