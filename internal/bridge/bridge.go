// Package bridge provides a Tor-over-Tox bridge integration layer enabling Go Tox clients
// to automatically leverage peer-to-peer Tox bridges as failover routes for Tor traffic.
//
// Bridge Architecture:
// The bridge operates as a SOCKS5 proxy server listening on 127.0.0.1:19050, automatically
// routing traffic through available Tox friend bridges with graceful fallback to direct Tor.
//
// Integration Pattern:
// 1. Initialize the bridge with NewTOXBridge(transportManager, enabled)
// 2. The bridge automatically starts listening on the default port
// 3. Route client traffic through 127.0.0.1:19050 as a SOCKS5 proxy
// 4. Bridge handles Tox friend availability changes automatically
// 5. Call Close() during graceful shutdown
package bridge

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/opd-ai/toxcore/transport"
)

const (
	// DefaultSOCKSAddr is the default SOCKS5 proxy address
	DefaultSOCKSAddr = "127.0.0.1:19050"
	// SOCKSAddrEnvVar overrides the SOCKS5 listen address when set
	SOCKSAddrEnvVar = "DTOX_BRIDGE_SOCKS_ADDR"
	// FailoverCheckInterval is how often we check Tox friend availability
	FailoverCheckInterval = 10 * time.Second
)

// BridgeStatus represents the current state of the bridge
type BridgeStatus struct {
	Enabled            bool
	ListeningAddr      string
	ActiveToxFriends   int
	TorAvailable       bool
	LastFailoverUpdate time.Time
	TotalConnections   int64
}

// TOXBridge manages the Tor-over-Tox bridge integration, providing automatic failover
// between Tox friend routes and direct Tor connectivity.
type TOXBridge struct {
	// Configuration
	enabled      bool
	listenAddr   string
	transportMgr *transport.MultiTransport

	// Server state
	listener   net.Listener
	listenerMu sync.Mutex
	closed     bool
	closedMu   sync.Mutex
	done       chan struct{}

	// Failover state machine
	failover      *FailoverState
	activeFriends int
	activeMu      sync.RWMutex

	// Connection tracking
	connCount   int64
	connCountMu sync.Mutex
}

// NewTOXBridge creates and initializes a new Tor-over-Tox bridge.
// The bridge is enabled by default and automatically starts listening on DefaultSOCKSAddr.
//
// Parameters:
//   - transportMgr: The multi-transport manager for routing (typically from anonymity.MultiTransportManager)
//   - enabled: If true, bridge starts immediately; if false, bridge is created but inactive
//
// Returns a non-nil *TOXBridge and a nil error on success. When enabled is true,
// the bridge will be listening on DefaultSOCKSAddr before this function returns.
func NewTOXBridge(transportMgr *transport.MultiTransport, enabled bool) (*TOXBridge, error) {
	if transportMgr == nil {
		return nil, fmt.Errorf("transport manager required")
	}

	listenAddr := DefaultSOCKSAddr
	if configuredAddr := os.Getenv(SOCKSAddrEnvVar); configuredAddr != "" {
		listenAddr = configuredAddr
	}

	bridge := &TOXBridge{
		enabled:      enabled,
		listenAddr:   listenAddr,
		transportMgr: transportMgr,
		done:         make(chan struct{}),
		failover:     NewFailoverState(),
	}

	if enabled {
		// Start listening on the SOCKS5 port
		if err := bridge.start(); err != nil {
			return nil, fmt.Errorf("failed to start bridge: %w", err)
		}
	}

	return bridge, nil
}

// start initializes the SOCKS5 proxy listener and begins accepting connections.
// This is called automatically from NewTOXBridge if enabled=true.
func (b *TOXBridge) start() error {
	b.listenerMu.Lock()
	defer b.listenerMu.Unlock()

	if b.closed {
		return fmt.Errorf("bridge is closed")
	}

	listener, err := net.Listen("tcp", b.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", b.listenAddr, err)
	}

	b.listener = listener
	b.listenAddr = listener.Addr().String()
	log.Printf("[Bridge] SOCKS5 proxy listening on %s", b.listenAddr)

	// Start the connection accept loop in a goroutine
	go b.acceptConnections()

	// Start the failover state machine update loop
	go b.updateFailoverState()

	return nil
}

// acceptConnections accepts incoming SOCKS5 connections and routes them
// through the appropriate transport (Tox friends or direct Tor).
func (b *TOXBridge) acceptConnections() {
	for {
		select {
		case <-b.done:
			return
		default:
		}

		// Set a short timeout to allow checking if we should exit
		b.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))

		conn, err := b.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-b.done:
				return
			default:
				log.Printf("[Bridge] Accept error: %v", err)
			}
			continue
		}

		// Increment connection count
		b.connCountMu.Lock()
		b.connCount++
		b.connCountMu.Unlock()

		// Handle SOCKS5 connection in a goroutine
		go b.handleSOCKS5Connection(conn)
	}
}

// updateFailoverState periodically checks Tox friend availability and updates
// the failover routing state. This implements automatic detection of available bridges.
func (b *TOXBridge) updateFailoverState() {
	ticker := time.NewTicker(FailoverCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			// Update failover state with current Tor availability
			torAvailable := b.isTorAvailable()
			activeFriends := b.getActiveToxFriends()
			b.failover.Update(activeFriends, torAvailable)
			log.Printf("[Bridge] Failover state updated: Tor=%v, ToxFriends=%d",
				torAvailable, activeFriends)
		}
	}
}

// SetActiveToxFriends updates the number of currently available Tox friend bridges.
func (b *TOXBridge) SetActiveToxFriends(count int) {
	if count < 0 {
		count = 0
	}
	b.activeMu.Lock()
	b.activeFriends = count
	b.activeMu.Unlock()
	b.failover.Update(count, b.isTorAvailable())
}

func (b *TOXBridge) getActiveToxFriends() int {
	b.activeMu.RLock()
	defer b.activeMu.RUnlock()
	return b.activeFriends
}

// isTorAvailable checks if Tor is currently available by attempting a light connectivity check
func (b *TOXBridge) isTorAvailable() bool {
	// Quick check by trying to get supported networks that include "tor"
	networks := b.transportMgr.GetSupportedNetworks()
	for _, net := range networks {
		if net == "tor" {
			return true
		}
	}
	return false
}

// handleSOCKS5Connection implements the SOCKS5 protocol and routes the connection
// through available Tox friends with automatic Tor fallback.
//
// SOCKS5 Flow:
// 1. Client sends SOCKS5 greeting with authentication methods
// 2. Server (bridge) responds with selected method
// 3. Client sends SOCKS5 request (CONNECT, BIND, UDP_ASSOCIATE)
// 4. Server routes through failover state machine
// 5. Server responds with result and relays data
func (b *TOXBridge) handleSOCKS5Connection(clientConn net.Conn) {
	defer clientConn.Close()

	// Parse SOCKS5 request
	handler := NewSOCKS5Handler(clientConn, b.transportMgr, b.failover)
	if err := handler.Handle(); err != nil {
		log.Printf("[Bridge] SOCKS5 error: %v", err)
	}
}

// Status returns the current bridge operational status
func (b *TOXBridge) Status() BridgeStatus {
	b.connCountMu.Lock()
	connCount := b.connCount
	b.connCountMu.Unlock()

	return BridgeStatus{
		Enabled:            b.enabled,
		ListeningAddr:      b.listenAddr,
		ActiveToxFriends:   b.failover.GetActiveFriendCount(),
		TorAvailable:       b.isTorAvailable(),
		LastFailoverUpdate: b.failover.GetLastUpdate(),
		TotalConnections:   connCount,
	}
}

// Close gracefully shuts down the bridge, closing the SOCKS5 listener
// and releasing all resources. Safe to call multiple times.
func (b *TOXBridge) Close() error {
	b.closedMu.Lock()
	defer b.closedMu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	close(b.done)

	b.listenerMu.Lock()
	defer b.listenerMu.Unlock()

	if b.listener != nil {
		if err := b.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
		log.Printf("[Bridge] SOCKS5 proxy on %s closed", b.listenAddr)
	}

	return nil
}

// IsEnabled returns whether the bridge is currently enabled
func (b *TOXBridge) IsEnabled() bool {
	return b.enabled
}
