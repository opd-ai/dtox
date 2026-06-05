package torbridge

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
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
		SOCKSAddr:    "127.0.0.1:0",
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
	if tb.GetSOCKSAddr() == "" || strings.HasSuffix(tb.GetSOCKSAddr(), ":0") {
		t.Errorf("Expected bound SOCKS address with assigned port, got %s", tb.GetSOCKSAddr())
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
		SOCKSAddr:    "127.0.0.1:0",
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
	conn, err := net.DialTimeout("tcp", tb.GetSOCKSAddr(), 1*time.Second)
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
		SOCKSAddr:    "127.0.0.1:0",
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

// TestCloseClearsSOCKSAddr verifies that closing SOCKS listener clears exposed address.
func TestCloseClearsSOCKSAddr(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:0",
		EnableBridge: false,
	}

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}

	if tb.GetSOCKSAddr() == "" {
		t.Fatal("Expected SOCKS address before close")
	}

	if err := tb.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if tb.GetSOCKSAddr() != "" {
		t.Errorf("Expected SOCKS address to be cleared after close, got %q", tb.GetSOCKSAddr())
	}
	if tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS to be disabled after close")
	}
}

// TestContextCancellation verifies that canceling the context properly
// signals shutdown to the TorBridge.
func TestContextCancellation(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:0",
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

// TestConfigNilDefaults verifies that passing nil config keeps defaults and
// still returns a usable bridge when at least one service can initialize.
func TestConfigNilDefaults(t *testing.T) {
	tb, err := New(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to create TorBridge from nil config: %v", err)
	}
	defer tb.Close()

	if tb.GetSOCKSAddr() == "" {
		t.Error("Expected SOCKS to initialize with nil config")
	}
	if !tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS to be enabled")
	}
	if tb.IsBridgeEnabled() {
		t.Error("Expected bridge to remain disabled without ToxInstance")
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

	tb, err := New(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create TorBridge: %v", err)
	}
	defer tb.Close()

	// Should use default address
	if tb.GetSOCKSAddr() == "" {
		t.Error("Expected SOCKS address to be set to default, got empty string")
	}

}

// TestServiceIndependence verifies that failure of bridge startup
// (due to missing ToxInstance) doesn't prevent SOCKS-only operation.
func TestServiceIndependence(t *testing.T) {
	config := &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    "127.0.0.1:0",
		EnableBridge: true,
		ToxInstance:  nil, // This will cause bridge to fail
	}

	tb, err := New(context.Background(), config)

	// Should succeed with SOCKS service even if bridge can't initialize
	if err != nil {
		t.Fatalf("Expected SOCKS-only success on partial startup, got error: %v", err)
	}

	if tb == nil {
		t.Fatal("Expected TorBridge instance")
	}
	defer tb.Close()

	if !tb.IsSOCKSEnabled() {
		t.Error("Expected SOCKS service to stay enabled")
	}
	if tb.IsBridgeEnabled() {
		t.Error("Expected bridge service to remain disabled")
	}
}

// TestConfigWithCustomPort verifies that custom SOCKS port configuration works.
func TestConfigWithCustomPort(t *testing.T) {
	customPort := "127.0.0.1:0"
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

	if tb.GetSOCKSAddr() == "" || strings.HasSuffix(tb.GetSOCKSAddr(), ":0") {
		t.Errorf("Expected SOCKS address to be bound to an assigned port, got %s", tb.GetSOCKSAddr())
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
