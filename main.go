package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	id := flag.String("id", "", "Unique identifier for this service instance (required)")
	address := flag.String("address", "0.0.0.0", "IP address to bind to (0.0.0.0 for all interfaces)")
	advertiseAddr := flag.String("advertise", "", "IP address to advertise to other peers (auto-detected if not specified)")
	port := flag.Int("port", 8080, "Port to listen on")
	seeds := flag.String("seeds", "", "Comma-separated list of seed node addresses (e.g., http://192.168.1.100:8080,http://192.168.1.101:8080)")

	flag.Parse()

	if *id == "" {
		fmt.Println("Error: -id flag is required")
		flag.Usage()
		os.Exit(1)
	}

	finalAdvertiseAddr := GetInstanceIP(advertiseAddr)

	var seedNodes []string
	if *seeds != "" {
		seedNodes = strings.Split(*seeds, ",")
		for i, seed := range seedNodes {
			seedNodes[i] = strings.TrimSpace(seed)
		}
	}

	service := NewService(*id, *address, finalAdvertiseAddr, *port, seedNodes)

	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	mux := service.SetupRoutes()
	bindAddr := fmt.Sprintf("%s:%d", *address, *port)
	server := &http.Server{
		Addr:    bindAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	fmt.Println("\n=== Service Started ===")
	fmt.Printf("ID:               %s\n", *id)
	fmt.Printf("Binding to:       %s:%d\n", *address, *port)
	fmt.Printf("Advertising as:   %s\n", service.GetFullAddress())
	fmt.Printf("Discovery:        Broadcast enabled (UDP port %d)\n", 9999)
	if len(seedNodes) > 0 {
		fmt.Printf("Seed nodes:       %v\n", seedNodes)
	} else {
		fmt.Println("Seed nodes:       None (auto-discovery via broadcast)")
	}

	localIPs := GetAllLocalIPs()
	if len(localIPs) > 0 {
		fmt.Println("\nDetected network IPs:")
		for _, ip := range localIPs {
			fmt.Printf("  - %s\n", ip)
		}
	}

	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /status    - View service status and peer list")
	fmt.Println("  GET  /peers     - List all known peers")
	fmt.Println("  POST /join      - Join the cluster")
	fmt.Println("  POST /heartbeat - Receive heartbeat")
	fmt.Println("  POST /gossip    - Receive peer gossip")
	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println("======================\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down service...")
	service.Stop()
	log.Println("Service stopped")
}

func GetInstanceIP(advertiseAddr *string) string {
	var finalAdvertiseAddr string
	if *advertiseAddr == "" {
		finalAdvertiseAddr = GetOutboundIP()
		if finalAdvertiseAddr == "" {
			fmt.Println("Warning: Could not auto-detect network IP. Using localhost.")
			fmt.Println("This will prevent other computers from connecting to this node.")
			fmt.Printf("Please specify an IP address manually using -advertise flag.\n\n")
			finalAdvertiseAddr = "localhost"
		} else {
			log.Printf("Auto-detected network IP: %s", finalAdvertiseAddr)
		}
	} else {
		finalAdvertiseAddr = *advertiseAddr

		if *advertiseAddr == "localhost" || *advertiseAddr == "127.0.0.1" {
			fmt.Println("\n⚠️  WARNING: You are using localhost as the advertise address!")
			fmt.Println("   This will only work for peers on the same machine.")
			fmt.Println("   For cross-machine communication, use your network IP address.\n")
		}
	}

	return finalAdvertiseAddr
}
