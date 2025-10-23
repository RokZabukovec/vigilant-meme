package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rokzabukovec/clip/internal/peer"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	peerList   *peer.PeerList
	serviceID  string
	onPeerJoin func(peer *peer.Peer)
}

// NewHandler creates a new handler instance
func NewHandler(peerList *peer.PeerList, serviceID string, onPeerJoin func(peer *peer.Peer)) *Handler {
	return &Handler{
		peerList:   peerList,
		serviceID:  serviceID,
		onPeerJoin: onPeerJoin,
	}
}

// HandleJoin handles join requests from new peers
func (h *Handler) HandleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newPeer peer.Peer
	if err := json.NewDecoder(r.Body).Decode(&newPeer); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("New peer joining: %s at %s", newPeer.ID, newPeer.Address)

	// Add the new peer to our list
	h.peerList.Add(&newPeer)

	// Notify about the new peer
	if h.onPeerJoin != nil {
		h.onPeerJoin(&newPeer)
	}

	// Return our current peer list to the new peer
	peers := h.peerList.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// HandleHeartbeat handles heartbeat messages from peers
func (h *Handler) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
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
	if _, exists := h.peerList.Get(peerID); exists {
		h.peerList.UpdateLastSeen(peerID)
	} else {
		// Add new peer discovered through heartbeat
		h.peerList.Add(&peer.Peer{
			ID:      peerID,
			Address: peerAddress,
		})
		log.Printf("Discovered new peer through heartbeat: %s at %s", peerID, peerAddress)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGossip handles gossip messages containing peer information
func (h *Handler) HandleGossip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peers []*peer.Peer
	if err := json.NewDecoder(r.Body).Decode(&peers); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Merge the received peer list with ours
	for _, peer := range peers {
		if peer.ID != h.serviceID {
			// Only add if we don't know about this peer or update if we do
			if existing, exists := h.peerList.Get(peer.ID); exists {
				// Update only if the received info is newer
				if peer.LastSeen.After(existing.LastSeen) {
					h.peerList.Add(peer)
				}
			} else {
				h.peerList.Add(peer)
				log.Printf("Discovered new peer through gossip: %s at %s", peer.ID, peer.Address)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// HandlePeers returns the list of all known peers
func (h *Handler) HandlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peers := h.peerList.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// HandleStatus returns the status of this service instance
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alivePeers := h.peerList.GetAlive()
	allPeers := h.peerList.GetAll()

	status := map[string]interface{}{
		"id":          h.serviceID,
		"address":     "", // This will be set by the service
		"total_peers": len(allPeers),
		"alive_peers": len(alivePeers),
		"peers":       allPeers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// SetupRoutes sets up HTTP routes for the service
func (h *Handler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/join", h.HandleJoin)
	mux.HandleFunc("/heartbeat", h.HandleHeartbeat)
	mux.HandleFunc("/gossip", h.HandleGossip)
	mux.HandleFunc("/peers", h.HandlePeers)
	mux.HandleFunc("/status", h.HandleStatus)

	return mux
}
