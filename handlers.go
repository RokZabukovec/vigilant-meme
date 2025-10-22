package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleJoin handles join requests from new peers
func (s *Service) HandleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newPeer Peer
	if err := json.NewDecoder(r.Body).Decode(&newPeer); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("New peer joining: %s at %s", newPeer.ID, newPeer.Address)

	// Add the new peer to our list
	s.PeerList.Add(&newPeer)

	// Notify this peer about ourselves
	thisPeer := &Peer{
		ID:      s.ID,
		Address: s.GetFullAddress(),
	}
	s.PeerList.Add(thisPeer)

	// Return our current peer list to the new peer
	peers := s.PeerList.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// HandleHeartbeat handles heartbeat messages from peers
func (s *Service) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var heartbeat map[string]string
	if err := json.NewDecoder(r.Body).Decode(&heartbeat); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	peerID := heartbeat["id"]
	peerAddress := heartbeat["address"]

	// Update or add the peer
	if _, exists := s.PeerList.Get(peerID); exists {
		s.PeerList.UpdateLastSeen(peerID)
	} else {
		// Add new peer discovered through heartbeat
		s.PeerList.Add(&Peer{
			ID:      peerID,
			Address: peerAddress,
		})
		log.Printf("Discovered new peer through heartbeat: %s at %s", peerID, peerAddress)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGossip handles gossip messages containing peer information
func (s *Service) HandleGossip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peers []*Peer
	if err := json.NewDecoder(r.Body).Decode(&peers); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Merge the received peer list with ours
	for _, peer := range peers {
		if peer.ID != s.ID {
			// Only add if we don't know about this peer or update if we do
			if existing, exists := s.PeerList.Get(peer.ID); exists {
				// Update only if the received info is newer
				if peer.LastSeen.After(existing.LastSeen) {
					s.PeerList.Add(peer)
				}
			} else {
				s.PeerList.Add(peer)
				log.Printf("Discovered new peer through gossip: %s at %s", peer.ID, peer.Address)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// HandlePeers returns the list of all known peers
func (s *Service) HandlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peers := s.PeerList.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// HandleStatus returns the status of this service instance
func (s *Service) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alivePeers := s.PeerList.GetAlive()
	allPeers := s.PeerList.GetAll()

	status := map[string]interface{}{
		"id":          s.ID,
		"address":     s.GetFullAddress(),
		"total_peers": len(allPeers),
		"alive_peers": len(alivePeers),
		"peers":       allPeers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// SetupRoutes sets up HTTP routes for the service
func (s *Service) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/join", s.HandleJoin)
	mux.HandleFunc("/heartbeat", s.HandleHeartbeat)
	mux.HandleFunc("/gossip", s.HandleGossip)
	mux.HandleFunc("/peers", s.HandlePeers)
	mux.HandleFunc("/status", s.HandleStatus)

	return mux
}
