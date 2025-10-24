package testutil

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/rokzabukovec/clip/internal/config"
	"github.com/rokzabukovec/clip/internal/peer"
)

// GetFreePort returns a free port for testing
func GetFreePort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

// CreateTestConfig creates a test configuration with a free port
func CreateTestConfig(t *testing.T, id string) *config.Config {
	t.Helper()

	port := GetFreePort(t)
	broadcastPort := GetFreePort(t)

	return &config.Config{
		ID:                id,
		BindAddress:       "127.0.0.1",
		AdvertiseAddr:     "127.0.0.1",
		Port:              port,
		BroadcastPort:     broadcastPort,
		BroadcastInterval: 100 * time.Millisecond, // Fast for testing
		HeartbeatInterval: 50 * time.Millisecond,  // Fast for testing
		PeerTimeout:       200 * time.Millisecond, // Fast for testing
		GossipInterval:    100 * time.Millisecond, // Fast for testing
		LogLevel:          "error",                // Reduce log noise in tests
		LogFormat:         "text",
	}
}

// CreateTestPeer creates a test peer
func CreateTestPeer(id, address string) *peer.Peer {
	return &peer.Peer{
		ID:      id,
		Address: address,
		IsAlive: true,
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-time.After(time.Until(deadline)):
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
	}
}

// WaitForPeerCount waits for the peer list to have the expected number of peers
func WaitForPeerCount(t *testing.T, peerList *peer.PeerList, expectedCount int, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, func() bool {
		return peerList.Count() == expectedCount
	}, timeout, fmt.Sprintf("peer count to be %d", expectedCount))
}

// WaitForAlivePeerCount waits for the peer list to have the expected number of alive peers
func WaitForAlivePeerCount(t *testing.T, peerList *peer.PeerList, expectedCount int, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, func() bool {
		return peerList.CountAlive() == expectedCount
	}, timeout, fmt.Sprintf("alive peer count to be %d", expectedCount))
}

// WaitForPeerExists waits for a peer to exist in the peer list
func WaitForPeerExists(t *testing.T, peerList *peer.PeerList, peerID string, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, func() bool {
		return peerList.Exists(peerID)
	}, timeout, fmt.Sprintf("peer %s to exist", peerID))
}

// WaitForPeerNotExists waits for a peer to not exist in the peer list
func WaitForPeerNotExists(t *testing.T, peerList *peer.PeerList, peerID string, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, func() bool {
		return !peerList.Exists(peerID)
	}, timeout, fmt.Sprintf("peer %s to not exist", peerID))
}
