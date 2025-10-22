package main

import (
	"log"
	"net"
)

// GetOutboundIP gets the preferred outbound IP address of this machine
// This is useful for advertising the correct IP to other peers
func GetOutboundIP() string {
	// Try to connect to a public DNS server to determine our outbound IP
	// This doesn't actually establish a connection, just determines routing
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Warning: Could not detect network IP: %v", err)
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// GetAllLocalIPs returns all local IP addresses of this machine
func GetAllLocalIPs() []string {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Warning: Could not get network interfaces: %v", err)
		return ips
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips
}
