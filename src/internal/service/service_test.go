package service

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/rokzabukovec/clip/internal/testutil"
)

func TestNewService(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	if svc == nil {
		t.Fatal("Expected service to be non-nil")
	}

	if svc.config != cfg {
		t.Error("Expected config to be set")
	}

	if svc.peerList == nil {
		t.Error("Expected peerList to be initialized")
	}

	if svc.discovery == nil {
		t.Error("Expected discovery service to be initialized")
	}

	if svc.handlers == nil {
		t.Error("Expected handlers to be initialized")
	}

	if svc.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
}

func TestService_Start(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	err := svc.Start()
	if err != nil {
		t.Fatalf("Expected Start() to succeed, got error: %v", err)
	}

	// Give some time for services to start
	time.Sleep(100 * time.Millisecond)

	// Test that the service is running by checking if we can get the full address
	addr := svc.GetFullAddress()
	if addr == "" {
		t.Error("Expected GetFullAddress() to return non-empty address")
	}

	expectedAddr := fmt.Sprintf("http://%s:%d", cfg.AdvertiseAddr, cfg.Port)
	if addr != expectedAddr {
		t.Errorf("Expected address to be '%s', got '%s'", expectedAddr, addr)
	}

	// Clean up
	svc.Stop()
}

func TestService_Stop(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	err := svc.Start()
	if err != nil {
		t.Fatalf("Expected Start() to succeed, got error: %v", err)
	}

	// Give some time for services to start
	time.Sleep(100 * time.Millisecond)

	// Stop the service
	svc.Stop()

	// Give some time for services to stop
	time.Sleep(100 * time.Millisecond)

	// Test that we can call Stop() multiple times without panicking
	svc.Stop()
}

func TestService_GetFullAddress(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	addr := svc.GetFullAddress()
	expectedAddr := fmt.Sprintf("http://%s:%d", cfg.AdvertiseAddr, cfg.Port)

	if addr != expectedAddr {
		t.Errorf("Expected address to be '%s', got '%s'", expectedAddr, addr)
	}
}

func TestService_GetHandlers(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	handlers := svc.GetHandlers()
	if handlers == nil {
		t.Error("Expected GetHandlers() to return non-nil handlers")
	}
}

func TestService_GetPeerList(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	peerList := svc.GetPeerList()
	if peerList == nil {
		t.Error("Expected GetPeerList() to return non-nil peer list")
	}

	// Test that we can add peers to the list
	initialCount := peerList.Count()
	if initialCount != 0 {
		t.Errorf("Expected initial peer count to be 0, got %d", initialCount)
	}
}

func TestService_HTTPEndpoints(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	err := svc.Start()
	if err != nil {
		t.Fatalf("Expected Start() to succeed, got error: %v", err)
	}
	defer svc.Stop()

	// Give some time for services to start
	time.Sleep(100 * time.Millisecond)

	// Start HTTP server
	handlers := svc.GetHandlers()
	mux := handlers.SetupRoutes()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	// Give some time for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)

	t.Run("GET /status", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/status")
		if err != nil {
			t.Fatalf("Failed to GET /status: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be 'application/json', got '%s'", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("GET /peers", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/peers")
		if err != nil {
			t.Fatalf("Failed to GET /peers: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be 'application/json', got '%s'", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("POST /join", func(t *testing.T) {
		// This test would require more setup to properly test the join endpoint
		// For now, we'll just test that the endpoint exists and returns method not allowed for GET
		resp, err := http.Get(baseURL + "/join")
		if err != nil {
			t.Fatalf("Failed to GET /join: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})

	t.Run("POST /heartbeat", func(t *testing.T) {
		// This test would require more setup to properly test the heartbeat endpoint
		// For now, we'll just test that the endpoint exists and returns method not allowed for GET
		resp, err := http.Get(baseURL + "/heartbeat")
		if err != nil {
			t.Fatalf("Failed to GET /heartbeat: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})

	t.Run("POST /gossip", func(t *testing.T) {
		// This test would require more setup to properly test the gossip endpoint
		// For now, we'll just test that the endpoint exists and returns method not allowed for GET
		resp, err := http.Get(baseURL + "/gossip")
		if err != nil {
			t.Fatalf("Failed to GET /gossip: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestService_AdvertiseAddressDetection(t *testing.T) {
	t.Run("with explicit advertise address", func(t *testing.T) {
		cfg := testutil.CreateTestConfig(t, "test-service")
		cfg.AdvertiseAddr = "192.168.1.100"

		svc := NewService(cfg)

		addr := svc.GetFullAddress()
		expectedAddr := "http://192.168.1.100:" + fmt.Sprintf("%d", cfg.Port)

		if addr != expectedAddr {
			t.Errorf("Expected address to be '%s', got '%s'", expectedAddr, addr)
		}
	})

	t.Run("with localhost advertise address", func(t *testing.T) {
		cfg := testutil.CreateTestConfig(t, "test-service")
		cfg.AdvertiseAddr = "localhost"

		svc := NewService(cfg)

		addr := svc.GetFullAddress()
		expectedAddr := "http://localhost:" + fmt.Sprintf("%d", cfg.Port)

		if addr != expectedAddr {
			t.Errorf("Expected address to be '%s', got '%s'", expectedAddr, addr)
		}
	})
}

func TestService_SeedNodes(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	cfg.SeedNodes = []string{"http://192.168.1.100:8080", "http://192.168.1.101:8080"}

	svc := NewService(cfg)

	// Test that service starts without error even with seed nodes
	// (the actual registration would fail in test environment, but that's expected)
	err := svc.Start()
	if err != nil {
		t.Fatalf("Expected Start() to succeed with seed nodes, got error: %v", err)
	}
	defer svc.Stop()
}

func TestService_ConcurrentOperations(t *testing.T) {
	cfg := testutil.CreateTestConfig(t, "test-service")
	svc := NewService(cfg)

	err := svc.Start()
	if err != nil {
		t.Fatalf("Expected Start() to succeed, got error: %v", err)
	}
	defer svc.Stop()

	// Test concurrent access to service methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test concurrent access to various methods
			svc.GetFullAddress()
			svc.GetHandlers()
			svc.GetPeerList()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
