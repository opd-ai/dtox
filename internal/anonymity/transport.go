// Package anonymity provides utilities for handling anonymity network support
// in dtox, including Tor and I2P configuration and transport initialization.
package anonymity

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/opd-ai/toxcore/transport"
)

const (
	// Default timeout for transport availability checks
	defaultDialTimeout = 2 * time.Second
)

// TransportStatus represents the availability status of an anonymity transport.
type TransportStatus struct {
	Available   bool
	Address     string
	Error       error
	NetworkType string
}

// MultiTransportManager manages the multi-transport system for anonymity networks.
// It provides a unified interface for accessing Tor, I2P, and standard IP transports.
type MultiTransportManager struct {
	mt     *transport.MultiTransport
	mu     sync.RWMutex
	closed bool
}

// NewMultiTransportManager creates and initializes a new MultiTransportManager.
// The manager automatically registers IP, Tor, I2P, and Nym transports.
// Transport availability depends on the underlying services being running.
func NewMultiTransportManager() *MultiTransportManager {
	mt := transport.NewMultiTransport()
	return &MultiTransportManager{
		mt: mt,
	}
}

// GetMultiTransport returns the underlying MultiTransport instance.
// This allows direct access to the transport layer for advanced use cases.
func (m *MultiTransportManager) GetMultiTransport() *transport.MultiTransport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mt
}

// Close closes the multi-transport manager and releases all resources.
func (m *MultiTransportManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	if m.mt != nil {
		return m.mt.Close()
	}
	return nil
}

// GetSupportedNetworks returns a list of all network types supported by registered transports.
func (m *MultiTransportManager) GetSupportedNetworks() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mt == nil {
		return nil
	}
	return m.mt.GetSupportedNetworks()
}

// Dial establishes a connection to the given address using the appropriate transport.
// The address format determines which transport is used:
// - Standard IP addresses use IPTransport
// - .onion addresses use TorTransport
// - .i2p addresses use I2PTransport
func (m *MultiTransportManager) Dial(address string) (net.Conn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mt == nil {
		return nil, ErrTransportNotInitialized
	}
	return m.mt.Dial(address)
}

// DialPacket creates a packet connection to the given address using the appropriate transport.
// When Tor and I2P are both available, I2P is preferred for packet connections since Tor is TCP-only.
func (m *MultiTransportManager) DialPacket(address string) (net.PacketConn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mt == nil {
		return nil, ErrTransportNotInitialized
	}
	return m.mt.DialPacket(address)
}

// Listen creates a listener on the given address using the appropriate transport.
func (m *MultiTransportManager) Listen(address string) (net.Listener, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mt == nil {
		return nil, ErrTransportNotInitialized
	}
	return m.mt.Listen(address)
}

// CheckTorAvailability checks if Tor transport is available.
// Returns a TransportStatus with availability information.
func (m *MultiTransportManager) CheckTorAvailability() TransportStatus {
	torAddr := os.Getenv("TOR_CONTROL_ADDR")
	if torAddr == "" {
		torAddr = "127.0.0.1:9051"
	}

	status := TransportStatus{
		Address:     torAddr,
		NetworkType: "tor",
	}

	// Try to connect to the Tor control port
	conn, err := net.DialTimeout("tcp", torAddr, defaultDialTimeout)
	if err != nil {
		status.Available = false
		status.Error = err
		return status
	}
	conn.Close()

	status.Available = true
	return status
}

// CheckI2PAvailability checks if I2P transport is available.
// Returns a TransportStatus with availability information.
func (m *MultiTransportManager) CheckI2PAvailability() TransportStatus {
	i2pAddr := os.Getenv("I2P_SAM_ADDR")
	if i2pAddr == "" {
		i2pAddr = "127.0.0.1:7656"
	}

	status := TransportStatus{
		Address:     i2pAddr,
		NetworkType: "i2p",
	}

	// Try to connect to the I2P SAM bridge
	conn, err := net.DialTimeout("tcp", i2pAddr, defaultDialTimeout)
	if err != nil {
		status.Available = false
		status.Error = err
		return status
	}
	conn.Close()

	status.Available = true
	return status
}

// GetTransportStatuses returns the availability status of all anonymity transports.
func (m *MultiTransportManager) GetTransportStatuses() map[string]TransportStatus {
	statuses := make(map[string]TransportStatus)
	statuses["tor"] = m.CheckTorAvailability()
	statuses["i2p"] = m.CheckI2PAvailability()
	return statuses
}

// LogTransportStatus logs detailed status information about all transports.
// This is called at startup to inform users about available anonymity networks.
func (m *MultiTransportManager) LogTransportStatus() {
	log.Println(logHeader)

	// Check and log Tor status
	torStatus := m.CheckTorAvailability()
	if torStatus.Available {
		log.Printf("Tor: AVAILABLE (control: %s)", torStatus.Address)
		log.Println("  - .onion addresses will route through Tor")
	} else {
		log.Printf("Tor: NOT AVAILABLE (control: %s)", torStatus.Address)
		log.Printf("  - Error: %v", torStatus.Error)
		log.Println("  - Start Tor to enable .onion address support")
	}

	// Check and log I2P status
	i2pStatus := m.CheckI2PAvailability()
	if i2pStatus.Available {
		log.Printf("I2P: AVAILABLE (SAM bridge: %s)", i2pStatus.Address)
		log.Println("  - .i2p addresses will route through I2P")
	} else {
		log.Printf("I2P: NOT AVAILABLE (SAM bridge: %s)", i2pStatus.Address)
		log.Printf("  - Error: %v", i2pStatus.Error)
		log.Println("  - Start I2P router with SAM bridge to enable .i2p address support")
	}

	log.Println()
	log.Println("Transport routing (automatic based on address):")
	log.Println("  - Regular IP addresses -> Direct UDP/TCP")
	if torStatus.Available {
		log.Println("  - .onion addresses -> Tor network")
	}
	if i2pStatus.Available {
		log.Println("  - .i2p addresses -> I2P network")
	}

	// Log supported networks from the multi-transport
	if networks := m.GetSupportedNetworks(); len(networks) > 0 {
		log.Println()
		log.Printf("Supported network types: %v", networks)
	}

	log.Println(logSeparator)
}

// Global singleton instance for convenient access
var (
	globalManager     *MultiTransportManager
	globalManagerOnce sync.Once
	globalManagerMu   sync.Mutex
)

// GetGlobalManager returns the global MultiTransportManager instance.
// The manager is initialized lazily on first access.
func GetGlobalManager() *MultiTransportManager {
	globalManagerOnce.Do(func() {
		globalManager = NewMultiTransportManager()
	})
	return globalManager
}

// CloseGlobalManager closes the global manager if it was initialized.
// This should be called during application shutdown.
func CloseGlobalManager() error {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	if globalManager != nil {
		err := globalManager.Close()
		globalManager = nil
		// Reset the once so a new manager can be created if needed
		globalManagerOnce = sync.Once{}
		return err
	}
	return nil
}
