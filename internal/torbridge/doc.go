package torbridge

// ExampleUsage demonstrates how a Go Tox client initializes both the SOCKS proxy
// and Tor-over-Tox bridge services with a single function call.
//
// This is a minimal example showing the typical integration pattern:
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/opd-ai/toxcore"
//	    "github.com/opd-ai/dtox/internal/torbridge"
//	)
//
//	// 1. Initialize Tox instance (or get existing instance)
//	tox, err := toxcore.New(options)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 2. Create TorBridge config (using defaults or custom settings)
//	config := &torbridge.Config{
//	    EnableSOCKS:  true,
//	    SOCKSAddr:    torbridge.DefaultSOCKSAddr,
//	    EnableBridge: true,
//	    ToxInstance:  tox,
//	}
//
//	// 3. Initialize TorBridge - starts both services
//	tb, err := torbridge.New(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 4. Access services
//	log.Printf("SOCKS proxy at: %s", tb.GetSOCKSAddr())
//	log.Printf("Bridge accessible to Tox friends: %v", tb.IsBridgeEnabled())
//
//	// 5. On shutdown
//	tb.Close()
//
// The pattern above is less than 20 lines and provides:
// - Local Tor access via SOCKS proxy on 127.0.0.1:19050
// - Remote Tor access for Tox friends via bridge
// - Both services running independently
func ExampleUsage() {
	// This is a documentation example, not runnable code
}

// ExampleSOCKSOnlyMode demonstrates running only the SOCKS proxy without the bridge.
// Useful for clients that only need local Tor access.
//
//	config := &torbridge.Config{
//	    EnableSOCKS:  true,
//	    SOCKSAddr:    torbridge.DefaultSOCKSAddr,
//	    EnableBridge: false, // Disabled
//	}
//
//	tb, err := torbridge.New(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tb.Close()
//
func ExampleSOCKSOnlyMode() {
	// This is a documentation example, not runnable code
}

// ExampleCustomSOCKSPort demonstrates running the SOCKS proxy on a custom port.
//
//	config := &torbridge.Config{
//	    EnableSOCKS: true,
//	    SOCKSAddr:   "127.0.0.1:9050", // Custom port
//	    EnableBridge: true,
//	    ToxInstance: tox,
//	}
//
//	tb, err := torbridge.New(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tb.Close()
//
func ExampleCustomSOCKSPort() {
	// This is a documentation example, not runnable code
}

// ExampleGracefulShutdown demonstrates proper cleanup of TorBridge resources.
// It's safe to call Close() multiple times.
//
//	tb, err := torbridge.New(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// On shutdown (e.g., in a signal handler or defer block)
//	if err := tb.Close(); err != nil {
//	    log.Printf("Error closing TorBridge: %v", err)
//	}
//	// Services are now stopped and resources released
//
func ExampleGracefulShutdown() {
	// This is a documentation example, not runnable code
}

// ExampleChecking verifies which services are currently enabled.
//
//	if tb.IsSOCKSEnabled() {
//	    log.Printf("SOCKS proxy available at: %s", tb.GetSOCKSAddr())
//	}
//
//	if tb.IsBridgeEnabled() {
//	    log.Println("Tor-over-Tox bridge active (accessible to friends)")
//	}
//
func ExampleChecking() {
	// This is a documentation example, not runnable code
}

// IntegrationGuide provides documentation on integrating TorBridge into a
// Go Tox client application.
//
// STEP 1: Add to imports
//	import "github.com/opd-ai/dtox/internal/torbridge"
//
// STEP 2: Initialize in main application flow
// After creating your Tox instance, create TorBridge config:
//	config := torbridge.DefaultConfig()
//	config.ToxInstance = toxInstance
//
// STEP 3: Create TorBridge
//	tb, err := torbridge.New(context.Background(), config)
//	if err != nil {
//	    log.Printf("Failed to initialize Tor services: %v", err)
//	    // Application can still run with just Tox
//	}
//
// STEP 4: Shutdown on application exit
//	defer func() {
//	    if tb != nil {
//	        if err := tb.Close(); err != nil {
//	            log.Printf("Error closing TorBridge: %v", err)
//	        }
//	    }
//	}()
//
// STEP 5: Use services
// - SOCKS Proxy: For local Tor access, configure applications to use SOCKS at GetSOCKSAddr()
// - Bridge: Automatically advertised to Tox friends; they can route Tor traffic through you
//
// ADVANTAGES:
// - Minimal integration code (< 20 lines in typical client)
// - Zero breaking changes to existing Tox API
// - Independent services: bridge failure doesn't affect SOCKS
// - Configurable: Enable/disable services as needed
// - Comprehensive logging for service status
//
// TYPICAL OUTPUT:
//	Tor SOCKS proxy started on 127.0.0.1:19050 (local use)
//	Tor-over-Tox bridge initialized (friend-accessible)
//
const IntegrationGuide = `
TorBridge Integration Guide
===========================

TorBridge provides a minimal integration layer for Go Tox clients to enable:
1. Local Tor SOCKS proxy (opd-ai/go-tor on 127.0.0.1:19050)
2. Friend-accessible Tor-over-Tox bridge (opd-ai/toxpt)

QUICK START:

  import "github.com/opd-ai/dtox/internal/torbridge"

  // After creating Tox instance...
  config := torbridge.DefaultConfig()
  config.ToxInstance = toxInstance

  tb, err := torbridge.New(context.Background(), config)
  if err != nil {
    log.Printf("Tor services failed: %v", err)
  }
  defer tb.Close()

CONFIGURATION OPTIONS:

  config := &torbridge.Config{
    EnableSOCKS:  true,                         // Enable local SOCKS proxy
    SOCKSAddr:    torbridge.DefaultSOCKSAddr,  // Listen address (127.0.0.1:19050)
    EnableBridge: true,                         // Enable friend-accessible bridge
    ToxInstance:  toxInstance,                  // Tox client instance
  }

SERVICE MODES:

  1. SOCKS Only: Set EnableBridge=false (no ToxInstance required)
  2. Bridge Only: Set EnableSOCKS=false (not typical but possible)
  3. Both: Set both to true (recommended, default)

INDEPENDENT OPERATION:

  Both services run concurrently and independently:
  - If SOCKS fails, bridge can still run
  - If bridge fails, SOCKS can still run
  - Shutdown of one doesn't affect the other

RESOURCE CLEANUP:

  Always call Close() on shutdown:
    tb.Close()  // Safe to call multiple times

The integration requires <150 lines of code and maintains zero breaking
changes to the Tox API or existing libraries.
`
