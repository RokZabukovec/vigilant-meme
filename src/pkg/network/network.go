package network

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

// GetBroadcastAddress calculates the broadcast address for the given IP and mask
func GetBroadcastAddress(ip net.IP, mask net.IPMask) net.IP {
	if ip.To4() == nil {
		return nil
	}

	// Convert to IPv4 format
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}

	// Ensure mask length matches IPv4 length (4 bytes)
	if len(mask) != 4 {
		return nil
	}

	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = ip4[i] | ^mask[i]
	}
	return broadcast
}

// FindBroadcastAddress finds the broadcast address for the first available network interface
func FindBroadcastAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
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
					broadcast := GetBroadcastAddress(ip4, ipnet.Mask)
					return broadcast.String(), nil
				}
			}
		}
	}

	// Fallback to limited broadcast
	return "255.255.255.255", nil
}

// IsValidIP checks if the given string is a valid IP address
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidPort checks if the given port is valid
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}
