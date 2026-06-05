// Package bridge provides a Tor-over-Tox bridge integration layer for Go Tox clients.
//
// # Overview
//
// This package enables automatic routing of Tor traffic through peer-to-peer Tox friend bridges
// with graceful fallback to direct Tor connectivity. The bridge operates as a SOCKS5 proxy
// that transparently routes connections through available Tox bridges with intelligent failover.
//
// # Architecture
//
// The bridge consists of three main components:
//
// 1. Bridge (bridge.go): Main lifecycle and SOCKS5 server management
//    - Listens on 127.0.0.1:19050 for SOCKS5 connections
//    - Manages graceful startup and shutdown
//    - Tracks connection metrics
//
// 2. SOCKS5 Handler (socks5.go): RFC 1928 protocol implementation
//    - Parses SOCKS5 greeting and request protocol
//    - Supports CONNECT command for TCP connections
//    - Routes connections through the multi-transport system
//    - Implements bidirectional connection relay
//
// 3. Failover State Machine (failover.go): Automatic routing decision logic
//    - Monitors Tox friend bridge availability
//    - Transitions between primary (Tox friends) and fallback (direct Tor) routes
//    - Thread-safe state management
//    - Automatic detection and switching
//
// # Integration Pattern
//
// Go Tox clients integrate the bridge with minimal code:
//
//	// 1. Initialize multi-transport manager (already in use for Tor/I2P support)
//	transportMgr := anonymity.NewMultiTransportManager()
//	multiTransport := transportMgr.GetMultiTransport()
//
//	// 2. Create and start the bridge (enabled by default)
//	bridge, err := NewTOXBridge(multiTransport, true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer bridge.Close()
//
//	// 3. Route client traffic through 127.0.0.1:19050 as SOCKS5 proxy
//	// Bridge automatically handles failover between Tox friends and Tor
//
// # Failover State Machine
//
// The bridge maintains a state machine that manages routing decisions:
//
// - RouteToxFriends: When Tox friend bridges are available (primary route)
// - RouteDirect: When no Tox friends are available (fallback to direct Tor)
//
// The state machine transitions automatically based on:
// - Number of active Tox friend bridges
// - Tor availability status
// - Real-time monitoring with 10-second check intervals
//
// State transitions are logged for debugging and operational visibility.
//
// # SOCKS5 Protocol Support
//
// The bridge implements RFC 1928 SOCKS5 protocol:
//
// - Version negotiation: Responds with SOCKS5AuthNoAuth (no authentication required)
// - Request handling: Supports CONNECT command for TCP connections
// - Address types: IPv4, IPv6, and domain names
// - Connection relay: Bidirectional streaming with proper error handling
//
// Unsupported commands (BIND, UDP_ASSOCIATE) return appropriate SOCKS5 error replies.
//
// # Status Monitoring
//
// The bridge exposes operational status via the Status() method:
//
//	status := bridge.Status()
//	// Returns: Enabled, ListeningAddr, ActiveToxFriends, TorAvailable,
//	//          LastFailoverUpdate, TotalConnections
//
// This allows applications to monitor bridge health and debug routing decisions.
//
// # Zero Breaking Changes
//
// The bridge integration is:
// - Non-invasive to existing dtox functionality
// - Optional feature (disabled by passing enabled=false)
// - No changes to toxcore or transport APIs
// - Backward compatible with existing code
//
// # Example
//
// See examples/bridge_integration.go for a complete working example demonstrating:
// - Bridge initialization
// - Status monitoring
// - Graceful shutdown
// - Integration patterns for production applications
//
// # Testing
//
// Comprehensive test suite covers:
// - Bridge initialization and lifecycle
// - SOCKS5 protocol handling
// - Failover state machine transitions
// - Concurrency and thread safety
// - Connection tracking and metrics
//
// Run tests with:
//
//	go test ./internal/bridge/... -v
package bridge
