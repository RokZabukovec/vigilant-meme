package network

import (
	"fmt"
	"net"
	"testing"
)

func TestGetOutboundIP(t *testing.T) {
	ip := GetOutboundIP()

	// The function should return a valid IP address or empty string
	if ip != "" {
		// If it returns an IP, it should be a valid IPv4 address
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			t.Errorf("GetOutboundIP() returned invalid IP: %s", ip)
		}
		if parsedIP.To4() == nil {
			t.Errorf("GetOutboundIP() returned IPv6 address, expected IPv4: %s", ip)
		}
	}
}

func TestGetAllLocalIPs(t *testing.T) {
	ips := GetAllLocalIPs()

	// Should return at least one IP (localhost might be filtered out)
	// But we can't guarantee the exact number since it depends on the system

	// All returned IPs should be valid IPv4 addresses
	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			t.Errorf("GetAllLocalIPs() returned invalid IP: %s", ip)
		}
		if parsedIP.To4() == nil {
			t.Errorf("GetAllLocalIPs() returned IPv6 address, expected IPv4: %s", ip)
		}
		if parsedIP.IsLoopback() {
			t.Errorf("GetAllLocalIPs() returned loopback address, should be filtered out: %s", ip)
		}
	}
}

func TestGetBroadcastAddress(t *testing.T) {
	tests := []struct {
		name     string
		ip       net.IP
		mask     net.IPMask
		expected string
	}{
		{
			name:     "192.168.1.0/24",
			ip:       net.IPv4(192, 168, 1, 0),
			mask:     net.IPv4Mask(255, 255, 255, 0),
			expected: "192.168.1.255",
		},
		{
			name:     "10.0.0.0/8",
			ip:       net.IPv4(10, 0, 0, 0),
			mask:     net.IPv4Mask(255, 0, 0, 0),
			expected: "10.255.255.255",
		},
		{
			name:     "172.16.0.0/12",
			ip:       net.IPv4(172, 16, 0, 0),
			mask:     net.IPv4Mask(255, 240, 0, 0),
			expected: "172.31.255.255",
		},
		{
			name:     "192.168.1.100/24",
			ip:       net.IPv4(192, 168, 1, 100),
			mask:     net.IPv4Mask(255, 255, 255, 0),
			expected: "192.168.1.255",
		},
		{
			name:     "IPv6 address",
			ip:       net.ParseIP("2001:db8::1"),
			mask:     net.IPMask(net.ParseIP("ffff:ffff::")),
			expected: "", // Should return nil for IPv6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBroadcastAddress(tt.ip, tt.mask)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected nil for IPv6, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected broadcast address, got nil")
			}

			if result.String() != tt.expected {
				t.Errorf("Expected broadcast address %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestFindBroadcastAddress(t *testing.T) {
	addr, err := FindBroadcastAddress()

	if err != nil {
		t.Fatalf("FindBroadcastAddress() failed: %v", err)
	}

	if addr == "" {
		t.Fatal("FindBroadcastAddress() returned empty address")
	}

	// Should be a valid IP address
	parsedIP := net.ParseIP(addr)
	if parsedIP == nil {
		t.Errorf("FindBroadcastAddress() returned invalid IP: %s", addr)
	}

	// Should be IPv4
	if parsedIP.To4() == nil {
		t.Errorf("FindBroadcastAddress() returned IPv6 address, expected IPv4: %s", addr)
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip      string
		isValid bool
	}{
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"2001:db8::1", true},
		{"invalid", false},
		{"192.168.1.256", false},
		{"192.168.1", false},
		{"", false},
		{"192.168.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := IsValidIP(tt.ip)
			if result != tt.isValid {
				t.Errorf("IsValidIP(%s) = %v, want %v", tt.ip, result, tt.isValid)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		port    int
		isValid bool
	}{
		{1, true},
		{8080, true},
		{65535, true},
		{0, false},
		{-1, false},
		{65536, false},
		{100000, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("port_%d", tt.port), func(t *testing.T) {
			result := IsValidPort(tt.port)
			if result != tt.isValid {
				t.Errorf("IsValidPort(%d) = %v, want %v", tt.port, result, tt.isValid)
			}
		})
	}
}
