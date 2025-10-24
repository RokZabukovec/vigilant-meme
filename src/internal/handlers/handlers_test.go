package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rokzabukovec/clip/internal/peer"
)

func TestNewHandler(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	onPeerJoin := func(p *peer.Peer) {
		// Test callback
	}

	h := NewHandler(peerList, serviceID, onPeerJoin)

	if h.peerList != peerList {
		t.Error("Expected peerList to be set")
	}
	if h.serviceID != serviceID {
		t.Errorf("Expected serviceID to be '%s', got '%s'", serviceID, h.serviceID)
	}
	if h.onPeerJoin == nil {
		t.Error("Expected onPeerJoin callback to be set")
	}
}

func TestHandler_HandleJoin(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	var onPeerJoinCalled bool
	var joinedPeer *peer.Peer
	onPeerJoin := func(p *peer.Peer) {
		onPeerJoinCalled = true
		joinedPeer = p
	}

	h := NewHandler(peerList, serviceID, onPeerJoin)

	t.Run("valid join request", func(t *testing.T) {
		onPeerJoinCalled = false
		joinedPeer = nil

		newPeer := peer.Peer{
			ID:      "new-peer",
			Address: "http://192.168.1.100:8080",
		}

		jsonData, _ := json.Marshal(newPeer)
		req := httptest.NewRequest("POST", "/join", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleJoin(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if !onPeerJoinCalled {
			t.Error("Expected onPeerJoin callback to be called")
		}

		if joinedPeer == nil || joinedPeer.ID != "new-peer" {
			t.Error("Expected joined peer to be set correctly")
		}

		// Check that peer was added to the list
		if !peerList.Exists("new-peer") {
			t.Error("Expected peer to be added to peer list")
		}

		// Check response contains peer list
		var responsePeers []*peer.Peer
		if err := json.NewDecoder(w.Body).Decode(&responsePeers); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(responsePeers) != 1 {
			t.Errorf("Expected 1 peer in response, got %d", len(responsePeers))
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/join", nil)
		w := httptest.NewRecorder()

		h.HandleJoin(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/join", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleJoin(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestHandler_HandleHeartbeat(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	h := NewHandler(peerList, serviceID, nil)

	t.Run("valid heartbeat for existing peer", func(t *testing.T) {
		// Add a peer first
		existingPeer := &peer.Peer{
			ID:      "existing-peer",
			Address: "http://192.168.1.100:8080",
		}
		peerList.Add(existingPeer)

		heartbeat := map[string]string{
			"id":      "existing-peer",
			"address": "http://192.168.1.100:8080",
		}

		jsonData, _ := json.Marshal(heartbeat)
		req := httptest.NewRequest("POST", "/heartbeat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleHeartbeat(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that peer's last seen was updated
		p, exists := peerList.Get("existing-peer")
		if !exists {
			t.Fatal("Expected peer to exist")
		}
		if !p.IsAlive {
			t.Error("Expected peer to be alive after heartbeat")
		}
	})

	t.Run("valid heartbeat for new peer", func(t *testing.T) {
		heartbeat := map[string]string{
			"id":      "new-peer",
			"address": "http://192.168.1.101:8080",
		}

		jsonData, _ := json.Marshal(heartbeat)
		req := httptest.NewRequest("POST", "/heartbeat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleHeartbeat(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that new peer was added
		if !peerList.Exists("new-peer") {
			t.Error("Expected new peer to be added")
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/heartbeat", nil)
		w := httptest.NewRecorder()

		h.HandleHeartbeat(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/heartbeat", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleHeartbeat(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestHandler_HandleGossip(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	h := NewHandler(peerList, serviceID, nil)

	t.Run("valid gossip with new peers", func(t *testing.T) {
		peers := []*peer.Peer{
			{ID: "peer1", Address: "http://192.168.1.100:8080"},
			{ID: "peer2", Address: "http://192.168.1.101:8080"},
		}

		jsonData, _ := json.Marshal(peers)
		req := httptest.NewRequest("POST", "/gossip", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleGossip(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that peers were added
		if !peerList.Exists("peer1") {
			t.Error("Expected peer1 to be added")
		}
		if !peerList.Exists("peer2") {
			t.Error("Expected peer2 to be added")
		}
	})

	t.Run("gossip with self should be ignored", func(t *testing.T) {
		peers := []*peer.Peer{
			{ID: serviceID, Address: "http://192.168.1.100:8080"},
			{ID: "other-peer", Address: "http://192.168.1.101:8080"},
		}

		jsonData, _ := json.Marshal(peers)
		req := httptest.NewRequest("POST", "/gossip", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleGossip(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that self was not added
		if peerList.Exists(serviceID) {
			t.Error("Expected self to not be added")
		}
		// Check that other peer was added
		if !peerList.Exists("other-peer") {
			t.Error("Expected other peer to be added")
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/gossip", nil)
		w := httptest.NewRecorder()

		h.HandleGossip(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/gossip", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleGossip(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestHandler_HandlePeers(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	h := NewHandler(peerList, serviceID, nil)

	// Add some peers
	peer1 := &peer.Peer{ID: "peer1", Address: "http://192.168.1.100:8080"}
	peer2 := &peer.Peer{ID: "peer2", Address: "http://192.168.1.101:8080"}
	peerList.Add(peer1)
	peerList.Add(peer2)

	t.Run("valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/peers", nil)
		w := httptest.NewRecorder()

		h.HandlePeers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be 'application/json', got '%s'", w.Header().Get("Content-Type"))
		}

		var responsePeers []*peer.Peer
		if err := json.NewDecoder(w.Body).Decode(&responsePeers); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(responsePeers) != 2 {
			t.Errorf("Expected 2 peers in response, got %d", len(responsePeers))
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/peers", nil)
		w := httptest.NewRecorder()

		h.HandlePeers(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandler_HandleStatus(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		peerList := peer.NewPeerList()
		serviceID := "test-service"
		h := NewHandler(peerList, serviceID, nil)

		// Add some peers with different states
		peer1 := &peer.Peer{ID: "alive-peer", Address: "http://192.168.1.100:8080"}
		peer2 := &peer.Peer{ID: "dead-peer", Address: "http://192.168.1.101:8080"}
		peerList.Add(peer1)
		peerList.Add(peer2)

		// Mark one peer as dead
		peerList.MarkDead("dead-peer")

		req := httptest.NewRequest("GET", "/status", nil)
		w := httptest.NewRecorder()

		h.HandleStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be 'application/json', got '%s'", w.Header().Get("Content-Type"))
		}

		var status map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if status["id"] != serviceID {
			t.Errorf("Expected service ID to be '%s', got '%v'", serviceID, status["id"])
		}

		if status["total_peers"] != float64(2) {
			t.Errorf("Expected total_peers to be 2, got %v", status["total_peers"])
		}

		if status["alive_peers"] != float64(1) {
			t.Errorf("Expected alive_peers to be 1, got %v", status["alive_peers"])
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		peerList := peer.NewPeerList()
		serviceID := "test-service"
		h := NewHandler(peerList, serviceID, nil)

		req := httptest.NewRequest("POST", "/status", nil)
		w := httptest.NewRecorder()

		h.HandleStatus(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandler_SetupRoutes(t *testing.T) {
	peerList := peer.NewPeerList()
	serviceID := "test-service"
	h := NewHandler(peerList, serviceID, nil)

	mux := h.SetupRoutes()

	if mux == nil {
		t.Fatal("Expected mux to be non-nil")
	}

	// Test that all routes are registered by making requests
	routes := []string{"/join", "/heartbeat", "/gossip", "/peers", "/status"}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// Should not get 404 (route not found)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s not found", route)
		}
	}
}
