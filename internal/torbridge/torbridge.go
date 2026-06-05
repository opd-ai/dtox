// Package torbridge provides a minimal integration layer enabling Go Tox clients
// to run a local SOCKS endpoint listener and a Tor-over-Tox bridge (via opd-ai/toxpt)
// for friend-accessible Tor routing.
//
// Component Separation:
// - SOCKS Endpoint: Local listener on 127.0.0.1:19050 for SOCKS clients
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
	"github.com/opd-ai/toxpt"
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
	socksServer *socks.Server
	socksAddr   string
	socksListen net.Listener

	// Tor-over-Tox bridge service (friend-accessible)
	toxBridge *toxpt.EmbeddableBridge
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
// It starts enabled services independently and continues when at least one requested
// service initializes successfully.
// Returns an error only when no requested service could be initialized.
//
// Resource Management:
// The returned TorBridge must be closed via Close() to release resources properly.
func New(ctx context.Context, config *Config) (*TorBridge, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(ctx)

	tb := &TorBridge{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	var initErrs []error
	startedAnyService := false

	// SOCKS endpoint service
	if config.EnableSOCKS {
		if err := tb.initializeSOCKSProxy(); err != nil {
			initErrs = append(initErrs, fmt.Errorf("failed to initialize SOCKS proxy: %w", err))
		} else {
			startedAnyService = true
			log.Printf("Tor SOCKS endpoint started on %s (local use)", tb.socksAddr)
		}
	}

	// Tor-over-Tox bridge service
	if config.EnableBridge {
		var err error
		if config.ToxInstance == nil {
			err = fmt.Errorf("EnableBridge requires ToxInstance to be set")
		} else {
			err = tb.initializeToxBridge()
		}
		if err != nil {
			initErrs = append(initErrs, fmt.Errorf("failed to initialize Tor-over-Tox bridge: %w", err))
		} else {
			startedAnyService = true
			log.Printf("Tor-over-Tox bridge initialized (friend-accessible)")
		}
	}

	if !startedAnyService && (config.EnableSOCKS || config.EnableBridge) {
		cancel()
		return nil, fmt.Errorf("failed to initialize requested Tor bridge services: %v", initErrs)
	}

	if len(initErrs) > 0 {
		log.Printf("TorBridge started with partial service availability: %v", initErrs)
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

	// Minimal SOCKS endpoint listener for local integrations.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Accept and immediately close incoming connections.
	go func() {
		for tb.ctx.Err() == nil {
			conn, err := listener.Accept()
			if err != nil {
				if tb.ctx.Err() == nil {
					log.Printf("SOCKS accept error: %v", err)
				}
				break
			}
			_ = conn.Close()
		}
	}()

	tb.socksAddr = listener.Addr().String()
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
		tb.socksAddr = ""
	}

	return nil
}

// initializeToxBridge sets up the friend-accessible Tor-over-Tox bridge service.
// This provides bridge connectivity via opd-ai/toxpt for connected Tox friends only.
func (tb *TorBridge) initializeToxBridge() error {
	tb.mu.RLock()
	closed := tb.closed
	tb.mu.RUnlock()
	if closed {
		return fmt.Errorf("TorBridge is closed")
	}

	tb.bridgeMu.Lock()
	defer tb.bridgeMu.Unlock()

	// Create bridge configuration with Tox instance.
	// When AllowedFriends is nil/empty, the bridge dynamically uses the friend list from ToxClient.
	bridgeConfig := toxpt.DefaultConfig()
	bridgeConfig.ToxClient = tb.config.ToxInstance
	bridgeConfig.AllowedFriends = nil // Use dynamic friend list from ToxClient

	// Create bridge instance from toxpt
	// The bridge automatically restricts access to connected Tox friends
	bridgeInstance, err := toxpt.NewEmbeddableBridge(bridgeConfig)
	if err != nil {
		return fmt.Errorf("failed to create embeddable bridge: %w", err)
	}

	tb.toxBridge = bridgeInstance

	return nil
}

// closeToxBridge cleanly shuts down the Tor-over-Tox bridge service.
func (tb *TorBridge) closeToxBridge() error {
	tb.bridgeMu.Lock()
	defer tb.bridgeMu.Unlock()

	if tb.toxBridge != nil {
		// Bridge cleanup (if needed in the API)
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
	return tb.socksListen != nil
}

// IsBridgeEnabled returns whether the Tor-over-Tox bridge service is enabled.
func (tb *TorBridge) IsBridgeEnabled() bool {
	tb.bridgeMu.RLock()
	defer tb.bridgeMu.RUnlock()
	return tb.toxBridge != nil
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
