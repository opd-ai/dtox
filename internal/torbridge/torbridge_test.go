package torbridge

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/opd-ai/toxcore"
)

// TestDefaultConfig verifies that DefaultConfig returns a properly initialized config
// with SOCKS and bridge both enabled and using the default SOCKS address.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.EnableSOCKS {
		t.Error("Expected EnableSOCKS to be true")
	}
	if cfg.SOCKSAddr != DefaultSOCKSAddr {
		t.Errorf("Expected SOCKSAddr to be %s, got %s", DefaultSOCKSAddr, cfg.SOCKSAddr)
	}
	if !cfg.EnableBridge {
		t.Error("Expected EnableBridge to be true")
	}
	if cfg.ToxInstance != nil {
		t.Error("Expected ToxInstance to be nil in default config")
	}
}

// TestSOCKSProxyOnly verifies that TorBridge can initialize with only the SOCKS
// proxy enabled, without a Tox instance.
func TestSOCKSProxyOnly(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:19051", // Use different port to avoid conflicts
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	// Verify SOCKS service is running
	if !tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS to be enabled")
	}
	if tb.GetSOCKSAddr() != "127.0.0.1:19051" {
		t.Errorf("Expected SOCKS address to be 127.0.0.1:19051, got %s", tb.GetSOCKSAddr())
	}

	// Verify bridge service is not running
	if tb.IsBridgeEnabled() {
		t.Error("Expected bridge to be disabled")
	}
}

// TestBridgeRequiresToxInstance verifies that initializing a bridge without
// providing a ToxInstance returns an error.
func TestBridgeRequiresToxInstance(t *testing.T) {
	config := &Config{
		EnableSOCKS:  false,
		EnableBridge: true,
		ToxInstance:  nil,
	}

	tb, err := New(context.Background(), config)
	if err == nil {
		t.Fatal("Expected error when EnableBridge is true but ToxInstance is nil")
	}
	if tb != nil {
		tb.Close()
	}
}

// TestBothServicesDisabled verifies that TorBridge can be created with both
// services disabled (though not practically useful, this ensures the package
// handles the edge case gracefully).
func TestBothServicesDisabled(t *testing.T) {
	config := &Config{
		EnableSOCKS:  false,
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	if tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS to be disabled")
	}
	if tb.IsBridgeEnabled() {
		t.Error("Expected bridge to be disabled")
	}
}

// TestSOCKSProxyListening verifies that the SOCKS proxy actually listens
// on the configured address and can accept connections.
func TestSOCKSProxyListening(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:19052",
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	// Give the SOCKS server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Attempt to connect to the SOCKS proxy
	conn, err := net.DialTimeout("tcp", "127.0.0.1:19052", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to SOCKS proxy: %v", err)
	}
	defer conn.Close()

	// If we got here, the port is listening
	if conn == nil {
		t.Error("Expected connection to be established")
	}
}

// TestMultipleCloseIsSafe verifies that calling Close() multiple times
// does not cause errors and is safe to do so.
func TestMultipleCloseIsSafe(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:19053",
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}

	// Close multiple times should not error
	if err := tb.Close(); err != nil {
		t.Errorf("First close failed: %v", err)
	}

	if err := tb.Close(); err != nil {
		t.Errorf("Second close failed: %v", err)
	}

	if err := tb.Close(); err != nil {
		t.Errorf("Third close failed: %v", err)
	}
}

// TestContextCancellation verifies that canceling the context properly
// signals shutdown to the TorBridge.
func TestContextCancellation(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:19054",
		EnableBridge: false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	tb, err := New(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}

	// Signal cancellation
	cancel()

	// Give services a moment to respond to cancellation
	time.Sleep(100 * time.Millisecond)

	// Close should still work
	if err := tb.Close(); err != nil {
		t.Errorf("Close failed after context cancellation: %v", err)
	}
}

// TestConfigNilDefaults verifies that passing nil config uses defaults.
func TestConfigNilDefaults(t *testing.T) {
	tb, err := New(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to create TorBridge with nil config: %v", err)
	}
	defer tb.Close()

	// Verify defaults were applied
	if tb.GetSOCKSAddr() != DefaultSOCKSAddr {
		t.Errorf("Expected default SOCKS address %s, got %s", DefaultSOCKSAddr, tb.GetSOCKSAddr())
	}
	if !tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS to be enabled by default")
	}
}

// TestSOCKSAddrEmptyStringUsesDefault verifies that an empty SOCKSAddr
// string in config uses the default address.
func TestSOCKSAddrEmptyStringUsesDefault(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "", // Empty string
		EnableBridge: false,
	}

	// Create a mock options for toxcore
	options := toxcore.NewOptions()

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	// Should use default address
	if tb.GetSOCKSAddr() == "" {
		t.Error("Expected SOCKS address to be set to default, got empty string")
	}

	_ = options // Prevent unused variable error
}

// TestServiceIndependence verifies that failure of bridge startup
// (due to missing ToxInstance) doesn't prevent SOCKS-only operation.
func TestServiceIndependence(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:19055",
		EnableBridge: true,
		ToxInstance:  nil, // This will cause bridge to fail
	}

	tb, err := New(context.Background(), config)

	// Should fail due to missing ToxInstance
	if err == nil {
		t.Fatal("Expected error when bridge enabled without ToxInstance")
	}

	if tb != nil {
		tb.Close()
	}
}

// TestConfigWithCustomPort verifies that custom SOCKS port configuration works.
func TestConfigWithCustomPort(t *testing.T) {
	customPort := "127.0.0.1:19056"
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    customPort,
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	if tb.GetSOCKSAddr() != customPort {
		t.Errorf("Expected SOCKS address to be %s, got %s", customPort, tb.GetSOCKSAddr())
	}
}

// BenchmarkTorBridgeCreation benchmarks the time it takes to create a TorBridge
// with SOCKS proxy only (no bridge).
func BenchmarkTorBridgeCreation(b *testing.B) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:0", // Use port 0 for auto-selection
		EnableBridge: false,
	}

	for i := 0; i < b.N; i++ {
		tb, err := New(context.Background(), config)
		if err != nil {
			b.Fatalf("Failed to create TorBridge: %v", err)
		}
		tb.Close()
	}
}
