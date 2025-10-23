package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/rokzabukovec/clip/pkg/network"
)

const (
	BroadcastPort     = 9999
	BroadcastInterval = 10 * time.Second
	DiscoveryMessage  = "CLIP_PEER_DISCOVERY"
)

// BroadcastMessage represents a message sent via UDP broadcast
type BroadcastMessage struct {
	MessageType string `json:"type"`
	ID          string `json:"id"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
}

// DiscoveryService handles peer discovery via UDP broadcast
type DiscoveryService struct {
	serviceID   string
	serviceAddr string
	servicePort int
	stopChan    chan struct{}
	onPeerFound func(id, address string)
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(serviceID, serviceAddr string, servicePort int, onPeerFound func(id, address string)) *DiscoveryService {
	return &DiscoveryService{
		serviceID:   serviceID,
		serviceAddr: serviceAddr,
		servicePort: servicePort,
		stopChan:    make(chan struct{}),
		onPeerFound: onPeerFound,
	}
}

// StartBroadcastListener starts listening for broadcast messages from other peers
func (ds *DiscoveryService) StartBroadcastListener() {
	addr := net.UDPAddr{
		Port: BroadcastPort,
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Warning: Could not start broadcast listener: %v", err)
		log.Printf("Automatic peer discovery will not work. Use -seeds flag instead.")
		return
	}

	log.Printf("Broadcast discovery listener started on port %d", BroadcastPort)

	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)

		for {
			select {
			case <-ds.stopChan:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				n, remoteAddr, err := conn.ReadFromUDP(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					log.Printf("Error reading broadcast: %v", err)
					continue
				}

				ds.handleBroadcast(buf[:n], remoteAddr)
			}
		}
	}()
}

// StartBroadcastAnnouncer starts announcing this service's presence via broadcast
func (ds *DiscoveryService) StartBroadcastAnnouncer() {
	broadcastAddr, err := network.FindBroadcastAddress()
	if err != nil {
		log.Printf("Warning: Could not determine broadcast address: %v", err)
		log.Printf("Automatic peer discovery announcements will not work. Use -seeds flag instead.")
		return
	}

	log.Printf("Broadcasting presence to %s every %v", broadcastAddr, BroadcastInterval)

	ticker := time.NewTicker(BroadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ds.sendBroadcast(broadcastAddr)
		case <-ds.stopChan:
			return
		}
	}
}

// Stop stops the discovery service
func (ds *DiscoveryService) Stop() {
	close(ds.stopChan)
}

// handleBroadcast processes incoming broadcast messages
func (ds *DiscoveryService) handleBroadcast(data []byte, remoteAddr *net.UDPAddr) {
	var msg BroadcastMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	if msg.ID == ds.serviceID {
		return
	}

	if msg.MessageType != DiscoveryMessage {
		return
	}

	log.Printf("Discovered new peer via broadcast: %s at %s", msg.ID, msg.Address)

	if ds.onPeerFound != nil {
		ds.onPeerFound(msg.ID, msg.Address)
	}
}

// sendBroadcast sends a broadcast message announcing this service's presence
func (ds *DiscoveryService) sendBroadcast(broadcastAddr string) {
	msg := BroadcastMessage{
		MessageType: DiscoveryMessage,
		ID:          ds.serviceID,
		Address:     ds.serviceAddr,
		Port:        ds.servicePort,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastAddr, BroadcastPort))
	if err != nil {
		log.Printf("Error resolving broadcast address: %v", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("Error creating UDP connection: %v", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error sending broadcast: %v", err)
	}
}
