package peer

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewPeerList(t *testing.T) {
	pl := NewPeerList()
	if pl == nil {
		t.Fatal("NewPeerList() returned nil")
	}
	if pl.peers == nil {
		t.Fatal("NewPeerList() peers map is nil")
	}
	if len(pl.peers) != 0 {
		t.Errorf("Expected empty peer list, got %d peers", len(pl.peers))
	}
}

func TestPeerList_Add(t *testing.T) {
	pl := NewPeerList()
	peer := &Peer{
		ID:      "test-peer",
		Address: "http://192.168.1.100:8080",
	}

	pl.Add(peer)

	if !pl.Exists("test-peer") {
		t.Error("Expected peer to exist after Add()")
	}

	retrieved, exists := pl.Get("test-peer")
	if !exists {
		t.Fatal("Expected to retrieve peer after Add()")
	}

	if retrieved.ID != "test-peer" {
		t.Errorf("Expected peer ID to be 'test-peer', got '%s'", retrieved.ID)
	}

	if retrieved.Address != "http://192.168.1.100:8080" {
		t.Errorf("Expected peer address to be 'http://192.168.1.100:8080', got '%s'", retrieved.Address)
	}

	if !retrieved.IsAlive {
		t.Error("Expected peer to be alive after Add()")
	}

	// Check that LastSeen is set to a recent time
	now := time.Now().UTC()
	if retrieved.LastSeen.After(now) || retrieved.LastSeen.Before(now.Add(-time.Second)) {
		t.Errorf("Expected LastSeen to be recent, got %v", retrieved.LastSeen)
	}
}

func TestPeerList_Remove(t *testing.T) {
	pl := NewPeerList()
	peer := &Peer{
		ID:      "test-peer",
		Address: "http://192.168.1.100:8080",
	}

	pl.Add(peer)
	if !pl.Exists("test-peer") {
		t.Fatal("Expected peer to exist before Remove()")
	}

	pl.Remove("test-peer")
	if pl.Exists("test-peer") {
		t.Error("Expected peer to not exist after Remove()")
	}

	// Removing non-existent peer should not panic
	pl.Remove("non-existent")
}

func TestPeerList_Get(t *testing.T) {
	pl := NewPeerList()
	peer := &Peer{
		ID:      "test-peer",
		Address: "http://192.168.1.100:8080",
	}

	pl.Add(peer)

	// Test getting existing peer
	retrieved, exists := pl.Get("test-peer")
	if !exists {
		t.Fatal("Expected to retrieve existing peer")
	}
	if retrieved.ID != "test-peer" {
		t.Errorf("Expected peer ID to be 'test-peer', got '%s'", retrieved.ID)
	}

	// Test getting non-existent peer
	_, exists = pl.Get("non-existent")
	if exists {
		t.Error("Expected non-existent peer to not exist")
	}
}

func TestPeerList_GetAll(t *testing.T) {
	pl := NewPeerList()

	// Test empty list
	peers := pl.GetAll()
	if len(peers) != 0 {
		t.Errorf("Expected empty list, got %d peers", len(peers))
	}

	// Add some peers
	peer1 := &Peer{ID: "peer1", Address: "http://192.168.1.100:8080"}
	peer2 := &Peer{ID: "peer2", Address: "http://192.168.1.101:8080"}
	peer3 := &Peer{ID: "peer3", Address: "http://192.168.1.102:8080"}

	pl.Add(peer1)
	pl.Add(peer2)
	pl.Add(peer3)

	peers = pl.GetAll()
	if len(peers) != 3 {
		t.Errorf("Expected 3 peers, got %d", len(peers))
	}

	// Check that all peers are present
	peerMap := make(map[string]bool)
	for _, p := range peers {
		peerMap[p.ID] = true
	}

	expectedPeers := []string{"peer1", "peer2", "peer3"}
	for _, expected := range expectedPeers {
		if !peerMap[expected] {
			t.Errorf("Expected peer '%s' to be in GetAll() result", expected)
		}
	}
}

func TestPeerList_GetAlive(t *testing.T) {
	pl := NewPeerList()

	// Add peers - Add() always sets IsAlive to true
	peer1 := &Peer{ID: "alive-peer", Address: "http://192.168.1.100:8080"}
	peer2 := &Peer{ID: "dead-peer", Address: "http://192.168.1.101:8080"}
	peer3 := &Peer{ID: "another-alive", Address: "http://192.168.1.102:8080"}

	pl.Add(peer1)
	pl.Add(peer2)
	pl.Add(peer3)

	// Mark one peer as dead
	pl.MarkDead("dead-peer")

	alivePeers := pl.GetAlive()
	if len(alivePeers) != 2 {
		t.Errorf("Expected 2 alive peers, got %d", len(alivePeers))
	}

	// Check that only alive peers are returned
	for _, p := range alivePeers {
		if !p.IsAlive {
			t.Errorf("Expected all returned peers to be alive, got dead peer: %s", p.ID)
		}
	}
}

func TestPeerList_MarkDead(t *testing.T) {
	pl := NewPeerList()
	peer := &Peer{
		ID:      "test-peer",
		Address: "http://192.168.1.100:8080",
		IsAlive: true,
	}

	pl.Add(peer)

	// Mark peer as dead
	pl.MarkDead("test-peer")

	retrieved, exists := pl.Get("test-peer")
	if !exists {
		t.Fatal("Expected peer to still exist after MarkDead()")
	}

	if retrieved.IsAlive {
		t.Error("Expected peer to be marked as dead")
	}

	// Marking non-existent peer as dead should not panic
	pl.MarkDead("non-existent")
}

func TestPeerList_UpdateLastSeen(t *testing.T) {
	pl := NewPeerList()
	peer := &Peer{
		ID:      "test-peer",
		Address: "http://192.168.1.100:8080",
		IsAlive: false,
	}

	pl.Add(peer)

	// Update last seen
	pl.UpdateLastSeen("test-peer")

	retrieved, exists := pl.Get("test-peer")
	if !exists {
		t.Fatal("Expected peer to exist after UpdateLastSeen()")
	}

	if !retrieved.IsAlive {
		t.Error("Expected peer to be alive after UpdateLastSeen()")
	}

	// Check that LastSeen was updated
	now := time.Now().UTC()
	if retrieved.LastSeen.After(now) || retrieved.LastSeen.Before(now.Add(-time.Second)) {
		t.Errorf("Expected LastSeen to be recent, got %v", retrieved.LastSeen)
	}

	// Updating non-existent peer should not panic
	pl.UpdateLastSeen("non-existent")
}

func TestPeerList_Count(t *testing.T) {
	pl := NewPeerList()

	// Test empty list
	if pl.Count() != 0 {
		t.Errorf("Expected count to be 0, got %d", pl.Count())
	}

	// Add peers
	peer1 := &Peer{ID: "peer1", Address: "http://192.168.1.100:8080"}
	peer2 := &Peer{ID: "peer2", Address: "http://192.168.1.101:8080"}

	pl.Add(peer1)
	if pl.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", pl.Count())
	}

	pl.Add(peer2)
	if pl.Count() != 2 {
		t.Errorf("Expected count to be 2, got %d", pl.Count())
	}

	// Remove a peer
	pl.Remove("peer1")
	if pl.Count() != 1 {
		t.Errorf("Expected count to be 1 after removal, got %d", pl.Count())
	}
}

func TestPeerList_CountAlive(t *testing.T) {
	pl := NewPeerList()

	// Test empty list
	if pl.CountAlive() != 0 {
		t.Errorf("Expected alive count to be 0, got %d", pl.CountAlive())
	}

	// Add peers - Add() always sets IsAlive to true
	peer1 := &Peer{ID: "alive-peer", Address: "http://192.168.1.100:8080"}
	peer2 := &Peer{ID: "dead-peer", Address: "http://192.168.1.101:8080"}
	peer3 := &Peer{ID: "another-alive", Address: "http://192.168.1.102:8080"}

	pl.Add(peer1)
	if pl.CountAlive() != 1 {
		t.Errorf("Expected alive count to be 1, got %d", pl.CountAlive())
	}

	pl.Add(peer2)
	if pl.CountAlive() != 2 {
		t.Errorf("Expected alive count to be 2, got %d", pl.CountAlive())
	}

	pl.Add(peer3)
	if pl.CountAlive() != 3 {
		t.Errorf("Expected alive count to be 3, got %d", pl.CountAlive())
	}

	// Mark one peer as dead
	pl.MarkDead("dead-peer")
	if pl.CountAlive() != 2 {
		t.Errorf("Expected alive count to be 2 after marking one dead, got %d", pl.CountAlive())
	}
}

func TestPeerList_Exists(t *testing.T) {
	pl := NewPeerList()

	// Test non-existent peer
	if pl.Exists("non-existent") {
		t.Error("Expected non-existent peer to not exist")
	}

	// Add peer
	peer := &Peer{ID: "test-peer", Address: "http://192.168.1.100:8080"}
	pl.Add(peer)

	if !pl.Exists("test-peer") {
		t.Error("Expected existing peer to exist")
	}

	// Remove peer
	pl.Remove("test-peer")
	if pl.Exists("test-peer") {
		t.Error("Expected removed peer to not exist")
	}
}

func TestPeerList_Concurrency(t *testing.T) {
	pl := NewPeerList()
	const numGoroutines = 100
	const numPeers = 10

	var wg sync.WaitGroup

	// Test concurrent adds
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numPeers; j++ {
				peer := &Peer{
					ID:      fmt.Sprintf("peer-%d-%d", id, j),
					Address: fmt.Sprintf("http://192.168.1.%d:8080", j),
				}
				pl.Add(peer)
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * numPeers
	if pl.Count() != expectedCount {
		t.Errorf("Expected %d peers after concurrent adds, got %d", expectedCount, pl.Count())
	}

	// Test concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numPeers; j++ {
				peerID := fmt.Sprintf("peer-%d-%d", id, j)

				// Read operations
				pl.Get(peerID)
				pl.Exists(peerID)
				pl.Count()
				pl.CountAlive()
				pl.GetAll()
				pl.GetAlive()

				// Write operations
				if j%2 == 0 {
					pl.UpdateLastSeen(peerID)
				} else {
					pl.MarkDead(peerID)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	finalCount := pl.Count()
	if finalCount != expectedCount {
		t.Errorf("Expected %d peers after concurrent operations, got %d", expectedCount, finalCount)
	}
}
