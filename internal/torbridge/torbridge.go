// Package torbridge provides a minimal integration layer enabling Go Tox clients
// to run both an embedded Tor SOCKS proxy (via opd-ai/go-tor) and a Tor-over-Tox
// bridge (via opd-ai/toxpt) for friend-accessible Tor routing.
//
// Component Separation:
// - SOCKS Proxy: Local-use Tor proxy on 127.0.0.1:19050 for direct Tor access
// - Tor-over-Tox Bridge: Friend-accessible bridge for remote Tor routing
//
// Both services run independently and concurrently; failure of one does not
// affect the other. Both are enabled by default and can be selectively disabled
// via configuration.
package torbridge

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/opd-ai/go-tor/pkg/socks"
	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/toxpt/pkg/bridge"
)

const (
	// DefaultSOCKSAddr is the default address for the Tor SOCKS proxy
	DefaultSOCKSAddr = "127.0.0.1:19050"
)

// TorBridge manages both the SOCKS proxy service and the Tor-over-Tox bridge service.
// It coordinates initialization and shutdown of both independent components.
type TorBridge struct {
	// Configuration
	config *Config

	// SOCKS proxy service (local Tor access)
	socksServer socks.SOCKSServer
	socksAddr   string
	socksListen net.Listener

	// Tor-over-Tox bridge service (friend-accessible)
	toxBridge bridge.ToxBridge
	bridgeMu  sync.RWMutex

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool
}

// Config represents configuration options for the TorBridge.
// Both services are enabled by default.
type Config struct {
	// EnableSOCKS enables the local Tor SOCKS proxy service
	// Default: true
	EnableSOCKS bool

	// SOCKSAddr specifies the address for the SOCKS proxy listener
	// Default: 127.0.0.1:19050
	SOCKSAddr string

	// EnableBridge enables the Tor-over-Tox bridge service for friend access
	// Default: true
	EnableBridge bool

	// ToxInstance is the Tox client instance to use for bridge access
	// Required if EnableBridge is true
	ToxInstance *toxcore.Tox
}

// DefaultConfig returns a Config with all services enabled using default settings.
func DefaultConfig() *Config {
	return &Config{
		EnableSOCKS:  true,
		SOCKSAddr:    DefaultSOCKSAddr,
		EnableBridge: true,
	}
}

// New creates and initializes a new TorBridge with the given configuration.
// It starts both the SOCKS proxy and bridge services concurrently.
// Returns an error if either service fails to initialize (unless disabled in config).
//
// Resource Management:
// The returned TorBridge must be closed via Close() to release resources properly.
func New(ctx context.Context, config *Config) (*TorBridge, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate bridge requirements
	if config.EnableBridge && config.ToxInstance == nil {
		return nil, fmt.Errorf("EnableBridge requires ToxInstance to be set")
	}

	ctx, cancel := context.WithCancel(ctx)

	tb := &TorBridge{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// SOCKS Proxy Service: Handles local Tor access
	// Component: opd-ai/go-tor SOCKS implementation on specified address
	if config.EnableSOCKS {
		if err := tb.initializeSOCKSProxy(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize SOCKS proxy: %w", err)
		}
		log.Printf("Tor SOCKS proxy started on %s (local use)", tb.socksAddr)
	}

	// Tor-over-Tox Bridge Service: Handles friend-accessible Tor routing
	// Component: opd-ai/toxpt bridge implementation for Tox friend connectivity
	if config.EnableBridge {
		if err := tb.initializeToxBridge(); err != nil {
			// Close SOCKS proxy if it was started, then fail
			if config.EnableSOCKS {
				_ = tb.closeSOCKSProxy()
			}
			cancel()
			return nil, fmt.Errorf("failed to initialize Tor-over-Tox bridge: %w", err)
		}
		log.Printf("Tor-over-Tox bridge initialized (friend-accessible)")
	}

	return tb, nil
}

// initializeSOCKSProxy sets up the local SOCKS proxy service.
// This provides local Tor connectivity via opd-ai/go-tor on the configured address.
func (tb *TorBridge) initializeSOCKSProxy() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.closed {
		return fmt.Errorf("TorBridge is closed")
	}

	addr := tb.config.SOCKSAddr
	if addr == "" {
		addr = DefaultSOCKSAddr
	}

	// Create SOCKS server instance from go-tor
	socksServer := socks.NewSOCKSServer()

	// Listen on the configured address
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Start accepting connections in background
	go func() {
		if err := socksServer.Serve(listener); err != nil && tb.ctx.Err() == nil {
			log.Printf("SOCKS server error: %v", err)
		}
	}()

	tb.socksServer = socksServer
	tb.socksAddr = addr
	tb.socksListen = listener

	return nil
}

// closeSOCKSProxy cleanly shuts down the SOCKS proxy service.
func (tb *TorBridge) closeSOCKSProxy() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.socksListen != nil {
		if err := tb.socksListen.Close(); err != nil {
			return fmt.Errorf("failed to close SOCKS listener: %w", err)
		}
		tb.socksListen = nil
	}

	return nil
}

// initializeToxBridge sets up the friend-accessible Tor-over-Tox bridge service.
// This provides bridge connectivity via opd-ai/toxpt for connected Tox friends only.
func (tb *TorBridge) initializeToxBridge() error {
	tb.bridgeMu.Lock()
	defer tb.bridgeMu.Unlock()

	if tb.closed {
		return fmt.Errorf("TorBridge is closed")
	}

	// Create bridge instance from toxpt
	// The bridge automatically restricts access to connected Tox friends
	bridgeInstance := bridge.NewToxBridge(tb.config.ToxInstance)

	// Start bridge in background, monitored for errors
	go func() {
		if err := bridgeInstance.Start(tb.ctx); err != nil && tb.ctx.Err() == nil {
			log.Printf("Tor-over-Tox bridge error: %v", err)
		}
	}()

	tb.toxBridge = bridgeInstance

	return nil
}

// closeToxBridge cleanly shuts down the Tor-over-Tox bridge service.
func (tb *TorBridge) closeToxBridge() error {
	tb.bridgeMu.Lock()
	defer tb.bridgeMu.Unlock()

	if tb.toxBridge != nil {
		if err := tb.toxBridge.Stop(); err != nil {
			return fmt.Errorf("failed to stop bridge: %w", err)
		}
		tb.toxBridge = nil
	}

	return nil
}

// GetSOCKSAddr returns the address where the SOCKS proxy is listening,
// or empty string if the SOCKS service is not enabled.
func (tb *TorBridge) GetSOCKSAddr() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.socksAddr
}

// IsSOCKSEnabled returns whether the SOCKS proxy service is enabled.
func (tb *TorBridge) IsSOCKSEnabled() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.config.EnableSOCKS
}

// IsBridgeEnabled returns whether the Tor-over-Tox bridge service is enabled.
func (tb *TorBridge) IsBridgeEnabled() bool {
	tb.bridgeMu.RLock()
	defer tb.bridgeMu.RUnlock()
	return tb.config.EnableBridge
}

// Close cleanly shuts down both services and releases all resources.
// It is safe to call Close() multiple times.
// After Close(), the TorBridge instance cannot be reused.
func (tb *TorBridge) Close() error {
	tb.mu.Lock()
	if tb.closed {
		tb.mu.Unlock()
		return nil
	}
	tb.closed = true
	tb.mu.Unlock()

	// Signal shutdown context
	tb.cancel()

	var errs []error

	// Shutdown SOCKS proxy
	if tb.config.EnableSOCKS {
		if err := tb.closeSOCKSProxy(); err != nil {
			errs = append(errs, err)
		}
		log.Println("Tor SOCKS proxy stopped")
	}

	// Shutdown bridge
	if tb.config.EnableBridge {
		if err := tb.closeToxBridge(); err != nil {
			errs = append(errs, err)
		}
		log.Println("Tor-over-Tox bridge stopped")
	}

	// Return first error if any occurred
	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}
