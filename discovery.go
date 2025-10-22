package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	BroadcastPort     = 9999
	BroadcastInterval = 10 * time.Second
	DiscoveryMessage  = "CLIP_PEER_DISCOVERY"
)

type BroadcastMessage struct {
	MessageType string `json:"type"`
	ID          string `json:"id"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
}

// StartBroadcastListener listens for broadcast messages from other peers
func (s *Service) StartBroadcastListener() {
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
			case <-s.stopChan:
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

				s.handleBroadcast(buf[:n], remoteAddr)
			}
		}
	}()
}

func (s *Service) handleBroadcast(data []byte, remoteAddr *net.UDPAddr) {
	var msg BroadcastMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	if msg.ID == s.ID {
		return
	}

	if msg.MessageType != DiscoveryMessage {
		return
	}

	if _, exists := s.PeerList.Get(msg.ID); exists {
		return
	}

	log.Printf("Discovered new peer via broadcast: %s at %s", msg.ID, msg.Address)

	peer := &Peer{
		ID:      msg.ID,
		Address: msg.Address,
	}
	s.PeerList.Add(peer)

	thisPeer := &Peer{
		ID:      s.ID,
		Address: s.GetFullAddress(),
	}
	if err := s.sendJoinRequest(msg.Address, thisPeer); err != nil {
		log.Printf("Failed to join discovered peer %s: %v", msg.ID, err)
	} else {
		log.Printf("Successfully joined discovered peer: %s", msg.ID)
	}
}

func (s *Service) StartBroadcastAnnouncer() {

	broadcastAddr, err := s.getBroadcastAddress()
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
			s.sendBroadcast(broadcastAddr)
		case <-s.stopChan:
			return
		}
	}
}

func (s *Service) sendBroadcast(broadcastAddr string) {
	msg := BroadcastMessage{
		MessageType: DiscoveryMessage,
		ID:          s.ID,
		Address:     s.GetFullAddress(),
		Port:        s.Port,
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

func (s *Service) getBroadcastAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					// Calculate broadcast address
					broadcast := make(net.IP, len(ip4))
					for i := range ip4 {
						broadcast[i] = ip4[i] | ^ipnet.Mask[i]
					}
					return broadcast.String(), nil
				}
			}
		}
	}

	// Fallback to limited broadcast
	return "255.255.255.255", nil
}
