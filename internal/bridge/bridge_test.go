package bridge

import (
	"net"
	"testing"
	"time"

	"github.com/opd-ai/dtox/internal/anonymity"
)

// TestNewTOXBridge tests bridge initialization with enabled and disabled states
func TestNewTOXBridge(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()
	if mt == nil {
		t.Fatal("MultiTransport is nil")
	}

	// Test initialization with bridge enabled
	bridge, err := NewTOXBridge(mt, true)
	if err != nil {
		t.Fatalf("Failed to create enabled bridge: %v", err)
	}
	defer bridge.Close()

	if !bridge.IsEnabled() {
		t.Error("Bridge should be enabled")
	}

	// Verify the bridge is listening
	status := bridge.Status()
	if !status.Enabled {
		t.Error("Status should show bridge enabled")
	}
	if status.ListeningAddr != DefaultSOCKSAddr {
		t.Errorf("Expected listening address %s, got %s", DefaultSOCKSAddr, status.ListeningAddr)
	}
}

// TestBridgeDisabled tests bridge initialization in disabled state
func TestBridgeDisabled(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()

	// Test initialization with bridge disabled
	bridge, err := NewTOXBridge(mt, false)
	if err != nil {
		t.Fatalf("Failed to create disabled bridge: %v", err)
	}
	defer bridge.Close()

	if bridge.IsEnabled() {
		t.Error("Bridge should be disabled")
	}

	status := bridge.Status()
	if status.Enabled {
		t.Error("Status should show bridge disabled")
	}
}

// TestBridgeClose tests graceful bridge shutdown
func TestBridgeClose(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()

	bridge, err := NewTOXBridge(mt, true)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Close the bridge
	if err := bridge.Close(); err != nil {
		t.Errorf("Failed to close bridge: %v", err)
	}

	// Closing again should be safe (idempotent)
	if err := bridge.Close(); err != nil {
		t.Errorf("Second close should be safe: %v", err)
	}
}

// TestBridgeNilTransport tests that bridge rejects nil transport manager
func TestBridgeNilTransport(t *testing.T) {
	bridge, err := NewTOXBridge(nil, true)
	if err == nil {
		t.Error("Bridge should reject nil transport manager")
	}
	if bridge != nil {
		bridge.Close()
	}
}

// TestBridgeStatus tests the status reporting interface
func TestBridgeStatus(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()

	bridge, err := NewTOXBridge(mt, true)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}
	defer bridge.Close()

	status := bridge.Status()

	// Verify status fields
	if !status.Enabled {
		t.Error("Status should indicate bridge is enabled")
	}
	if status.ListeningAddr != DefaultSOCKSAddr {
		t.Errorf("Status listening address mismatch")
	}
	if status.TotalConnections != 0 {
		t.Errorf("Initial connection count should be 0, got %d", status.TotalConnections)
	}

	// Status should include Tor availability check
	t.Logf("Tor Available: %v", status.TorAvailable)
}

// TestFailoverStateTransitions tests the failover state machine
func TestFailoverStateTransitions(t *testing.T) {
	fs := NewFailoverState()

	// Initially in direct mode
	if !fs.ShouldUseDirect() {
		t.Error("Should start in direct mode")
	}

	// Transition: Tox friends become available
	fs.Update(2, true)
	if !fs.ShouldUseToxFriends() {
		t.Error("Should transition to Tox friends mode when friends available")
	}

	if fs.GetActiveFriendCount() != 2 {
		t.Errorf("Expected 2 active friends, got %d", fs.GetActiveFriendCount())
	}

	// Transition: All Tox friends offline, fallback to direct
	fs.Update(0, true)
	if !fs.ShouldUseDirect() {
		t.Error("Should transition to direct mode when no friends available")
	}

	if fs.GetActiveFriendCount() != 0 {
		t.Errorf("Expected 0 active friends, got %d", fs.GetActiveFriendCount())
	}

	// Verify Tor availability tracking
	if !fs.IsTorAvailable() {
		t.Error("Tor should be available")
	}

	fs.Update(0, false)
	if fs.IsTorAvailable() {
		t.Error("Tor should be unavailable after update")
	}
}

// TestFailoverStateConcurrency tests thread safety of failover state machine
func TestFailoverStateConcurrency(t *testing.T) {
	fs := NewFailoverState()

	// Simulate concurrent updates from different goroutines
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				friends := (idx + j) % 5
				fs.Update(friends, (j%2) == 0)

				// Concurrent reads should not panic
				_ = fs.GetCurrentState()
				_ = fs.GetActiveFriendCount()
				_ = fs.IsTorAvailable()
				_ = fs.ShouldUseToxFriends()
				_ = fs.ShouldUseDirect()
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state is valid
	if !fs.ShouldUseDirect() && !fs.ShouldUseToxFriends() {
		t.Error("State should be either direct or Tox friends")
	}
}

// TestBridgeConnectionTracking tests that connection count is tracked
func TestBridgeConnectionTracking(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()

	bridge, err := NewTOXBridge(mt, true)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}
	defer bridge.Close()

	// Initial count should be 0
	status := bridge.Status()
	if status.TotalConnections != 0 {
		t.Errorf("Initial connection count should be 0, got %d", status.TotalConnections)
	}
}

// TestBridgeListenerBinding tests that the bridge binds to the correct address
func TestBridgeListenerBinding(t *testing.T) {
	tm := anonymity.NewMultiTransportManager()
	defer tm.Close()

	mt := tm.GetMultiTransport()

	bridge, err := NewTOXBridge(mt, true)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}
	defer bridge.Close()

	// Wait a moment for the listener to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect to the bridge port
	conn, err := net.Dial("tcp", DefaultSOCKSAddr)
	if err != nil {
		t.Fatalf("Failed to connect to bridge at %s: %v", DefaultSOCKSAddr, err)
	}
	defer conn.Close()

	// Send a SOCKS5 greeting (should work even if it fails later in protocol)
	greeting := []byte{0x05, 0x01, 0x00} // SOCKS5, 1 method, no auth
	if _, err := conn.Write(greeting); err != nil {
		t.Errorf("Failed to write SOCKS5 greeting: %v", err)
	}

	// Should receive a response
	response := make([]byte, 2)
	if _, err := conn.Read(response); err != nil {
		t.Errorf("Failed to read SOCKS5 response: %v", err)
	}

	// Verify SOCKS5 response format
	if response[0] != 0x05 {
		t.Errorf("Expected SOCKS5 version 0x05, got 0x%02x", response[0])
	}
}

// BenchmarkFailoverStateUpdate benchmarks the failover state machine update speed
func BenchmarkFailoverStateUpdate(b *testing.B) {
	fs := NewFailoverState()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs.Update((i%5)+1, (i%2) == 0)
	}
}

// BenchmarkFailoverStateQueries benchmarks failover state machine query speed
func BenchmarkFailoverStateQueries(b *testing.B) {
	fs := NewFailoverState()
	fs.Update(3, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fs.GetCurrentState()
		_ = fs.GetActiveFriendCount()
		_ = fs.IsTorAvailable()
		_ = fs.ShouldUseToxFriends()
		_ = fs.ShouldUseDirect()
	}
}
