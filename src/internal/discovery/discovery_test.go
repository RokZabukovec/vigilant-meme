package discovery

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestNewDiscoveryService(t *testing.T) {
	serviceID := "test-service"
	serviceAddr := "http://192.168.1.100:8080"
	servicePort := 8080
	broadcastPort := 9999
	onPeerFound := func(id, address string) {
		// Test callback
	}

	ds := NewDiscoveryService(serviceID, serviceAddr, servicePort, broadcastPort, onPeerFound)

	if ds.serviceID != serviceID {
		t.Errorf("Expected serviceID to be '%s', got '%s'", serviceID, ds.serviceID)
	}
	if ds.serviceAddr != serviceAddr {
		t.Errorf("Expected serviceAddr to be '%s', got '%s'", serviceAddr, ds.serviceAddr)
	}
	if ds.servicePort != servicePort {
		t.Errorf("Expected servicePort to be %d, got %d", servicePort, ds.servicePort)
	}
	if ds.broadcastPort != broadcastPort {
		t.Errorf("Expected broadcastPort to be %d, got %d", broadcastPort, ds.broadcastPort)
	}
	if ds.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
	if ds.onPeerFound == nil {
		t.Error("Expected onPeerFound callback to be set")
	}
}

func TestBroadcastMessage(t *testing.T) {
	msg := BroadcastMessage{
		MessageType: DiscoveryMessage,
		ID:          "test-peer",
		Address:     "http://192.168.1.100:8080",
		Port:        8080,
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal BroadcastMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled BroadcastMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal BroadcastMessage: %v", err)
	}

	if unmarshaled.MessageType != msg.MessageType {
		t.Errorf("Expected MessageType to be '%s', got '%s'", msg.MessageType, unmarshaled.MessageType)
	}
	if unmarshaled.ID != msg.ID {
		t.Errorf("Expected ID to be '%s', got '%s'", msg.ID, unmarshaled.ID)
	}
	if unmarshaled.Address != msg.Address {
		t.Errorf("Expected Address to be '%s', got '%s'", msg.Address, unmarshaled.Address)
	}
	if unmarshaled.Port != msg.Port {
		t.Errorf("Expected Port to be %d, got %d", msg.Port, unmarshaled.Port)
	}
}

func TestDiscoveryService_handleBroadcast(t *testing.T) {
	serviceID := "test-service"
	serviceAddr := "http://192.168.1.100:8080"
	servicePort := 8080
	broadcastPort := 9999
	var onPeerFoundCalled bool
	var foundPeerID, foundPeerAddr string
	onPeerFound := func(id, address string) {
		onPeerFoundCalled = true
		foundPeerID = id
		foundPeerAddr = address
	}

	ds := NewDiscoveryService(serviceID, serviceAddr, servicePort, broadcastPort, onPeerFound)

	t.Run("valid discovery message", func(t *testing.T) {
		onPeerFoundCalled = false
		foundPeerID = ""
		foundPeerAddr = ""

		msg := BroadcastMessage{
			MessageType: DiscoveryMessage,
			ID:          "other-peer",
			Address:     "http://192.168.1.101:8080",
			Port:        8080,
		}

		data, _ := json.Marshal(msg)
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 101), Port: 9999}

		ds.handleBroadcast(data, remoteAddr)

		if !onPeerFoundCalled {
			t.Error("Expected onPeerFound callback to be called")
		}
		if foundPeerID != "other-peer" {
			t.Errorf("Expected found peer ID to be 'other-peer', got '%s'", foundPeerID)
		}
		if foundPeerAddr != "http://192.168.1.101:8080" {
			t.Errorf("Expected found peer address to be 'http://192.168.1.101:8080', got '%s'", foundPeerAddr)
		}
	})

	t.Run("ignore own message", func(t *testing.T) {
		onPeerFoundCalled = false

		msg := BroadcastMessage{
			MessageType: DiscoveryMessage,
			ID:          serviceID, // Same as service ID
			Address:     "http://192.168.1.100:8080",
			Port:        8080,
		}

		data, _ := json.Marshal(msg)
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 100), Port: 9999}

		ds.handleBroadcast(data, remoteAddr)

		if onPeerFoundCalled {
			t.Error("Expected onPeerFound callback not to be called for own message")
		}
	})

	t.Run("ignore invalid message type", func(t *testing.T) {
		onPeerFoundCalled = false

		msg := BroadcastMessage{
			MessageType: "INVALID_MESSAGE",
			ID:          "other-peer",
			Address:     "http://192.168.1.101:8080",
			Port:        8080,
		}

		data, _ := json.Marshal(msg)
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 101), Port: 9999}

		ds.handleBroadcast(data, remoteAddr)

		if onPeerFoundCalled {
			t.Error("Expected onPeerFound callback not to be called for invalid message type")
		}
	})

	t.Run("ignore invalid JSON", func(t *testing.T) {
		onPeerFoundCalled = false

		invalidData := []byte("invalid json")
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 101), Port: 9999}

		ds.handleBroadcast(invalidData, remoteAddr)

		if onPeerFoundCalled {
			t.Error("Expected onPeerFound callback not to be called for invalid JSON")
		}
	})
}

func TestDiscoveryService_sendBroadcast(t *testing.T) {
	serviceID := "test-service"
	serviceAddr := "http://192.168.1.100:8080"
	servicePort := 8080
	broadcastPort := 9999
	ds := NewDiscoveryService(serviceID, serviceAddr, servicePort, broadcastPort, nil)

	// This test is limited because sendBroadcast requires actual network operations
	// We can test that it doesn't panic with a valid broadcast address
	broadcastAddr := "255.255.255.255"

	// This should not panic
	ds.sendBroadcast(broadcastAddr)
}

func TestDiscoveryService_Stop(t *testing.T) {
	serviceID := "test-service"
	serviceAddr := "http://192.168.1.100:8080"
	servicePort := 8080
	broadcastPort := 9999
	ds := NewDiscoveryService(serviceID, serviceAddr, servicePort, broadcastPort, nil)

	// Test that stop channel is closed
	select {
	case <-ds.stopChan:
		t.Error("Expected stopChan to be open before Stop()")
	default:
		// Expected - channel should be open
	}

	ds.Stop()

	// Test that stop channel is closed
	select {
	case <-ds.stopChan:
		// Expected - channel should be closed
	default:
		t.Error("Expected stopChan to be closed after Stop()")
	}
}

func TestConstants(t *testing.T) {
	if BroadcastPort != 9999 {
		t.Errorf("Expected BroadcastPort to be 9999, got %d", BroadcastPort)
	}
	if BroadcastInterval != 10*time.Second {
		t.Errorf("Expected BroadcastInterval to be 10s, got %v", BroadcastInterval)
	}
	if DiscoveryMessage != "CLIP_PEER_DISCOVERY" {
		t.Errorf("Expected DiscoveryMessage to be 'CLIP_PEER_DISCOVERY', got '%s'", DiscoveryMessage)
	}
}

func TestBroadcastMessage_EdgeCases(t *testing.T) {
	t.Run("empty message", func(t *testing.T) {
		msg := BroadcastMessage{}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal empty BroadcastMessage: %v", err)
		}

		var unmarshaled BroadcastMessage
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal empty BroadcastMessage: %v", err)
		}

		if unmarshaled.MessageType != "" {
			t.Errorf("Expected empty MessageType, got '%s'", unmarshaled.MessageType)
		}
		if unmarshaled.ID != "" {
			t.Errorf("Expected empty ID, got '%s'", unmarshaled.ID)
		}
		if unmarshaled.Address != "" {
			t.Errorf("Expected empty Address, got '%s'", unmarshaled.Address)
		}
		if unmarshaled.Port != 0 {
			t.Errorf("Expected Port to be 0, got %d", unmarshaled.Port)
		}
	})

	t.Run("special characters in ID and address", func(t *testing.T) {
		msg := BroadcastMessage{
			MessageType: DiscoveryMessage,
			ID:          "test-peer-123_456",
			Address:     "http://192.168.1.100:8080/path?query=value",
			Port:        8080,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal BroadcastMessage with special chars: %v", err)
		}

		var unmarshaled BroadcastMessage
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal BroadcastMessage with special chars: %v", err)
		}

		if unmarshaled.ID != msg.ID {
			t.Errorf("Expected ID to be '%s', got '%s'", msg.ID, unmarshaled.ID)
		}
		if unmarshaled.Address != msg.Address {
			t.Errorf("Expected Address to be '%s', got '%s'", msg.Address, unmarshaled.Address)
		}
	})
}
