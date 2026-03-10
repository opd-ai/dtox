// Package anonymity provides utilities for handling anonymity network support
// in dtox, including Tor and I2P configuration and logging.
package anonymity

import (
	"log"
	"os"
	"sync"
)

// Config holds the anonymity network configuration settings.
type Config struct {
	mu       sync.RWMutex
	anonOnly bool
}

// globalConfig is the package-level configuration instance.
var globalConfig = &Config{}

// SetAnonOnly sets the anonymous-only mode flag. When enabled, all network
// traffic must route through Tor, preventing any clearnet connections.
func SetAnonOnly(enabled bool) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.anonOnly = enabled
}

// IsAnonOnly returns whether anonymous-only mode is enabled.
func IsAnonOnly() bool {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.anonOnly
}

// GetTorControlAddr returns the configured Tor control address.
func GetTorControlAddr() string {
	addr := os.Getenv("TOR_CONTROL_ADDR")
	if addr == "" {
		return "127.0.0.1:9051"
	}
	return addr
}

// GetI2PSAMAddr returns the configured I2P SAM bridge address.
func GetI2PSAMAddr() string {
	addr := os.Getenv("I2P_SAM_ADDR")
	if addr == "" {
		return "127.0.0.1:7656"
	}
	return addr
}

// GetTorSOCKSAddr returns the configured Tor SOCKS proxy address.
// This is the SOCKS5 proxy that Tor exposes for application use.
// Default is 127.0.0.1:9050 (standard Tor SOCKS port).
func GetTorSOCKSAddr() string {
	addr := os.Getenv("TOR_SOCKS_ADDR")
	if addr == "" {
		return "127.0.0.1:9050"
	}
	return addr
}

// ParseHostPort splits a host:port string into separate components.
// If parsing fails, returns the original string as host and the default port.
func ParseHostPort(addr string, defaultPort uint16) (string, uint16) {
	// Find the last colon (handles IPv6 addresses in brackets)
	lastColon := -1
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon == -1 {
		return addr, defaultPort
	}

	host := addr[:lastColon]
	portStr := addr[lastColon+1:]

	// Handle empty port string (trailing colon with no port)
	if portStr == "" {
		return host, defaultPort
	}

	// Parse port
	var port uint64
	for _, c := range portStr {
		if c < '0' || c > '9' {
			return addr, defaultPort
		}
		port = port*10 + uint64(c-'0')
	}

	if port > 65535 || port == 0 {
		return addr, defaultPort
	}

	return host, uint16(port)
}

// LogNetworkStatusWithMode logs the configuration of anonymity networks,
// including whether anon-only mode is enabled.
func LogNetworkStatusWithMode() {
	log.Println(logHeader)

	// Log anon-only mode status
	if IsAnonOnly() {
		log.Println("*** ANONYMOUS-ONLY MODE ENABLED ***")
		log.Println("  - All traffic will route through Tor")
		log.Println("  - Direct clearnet connections disabled")
		log.Println()
	}

	// Check Tor configuration
	torAddr := GetTorControlAddr()
	log.Printf("Tor: CONFIGURED (control: %s)", torAddr)
	log.Println("  - .onion addresses will route through Tor")
	log.Println("  - Requires Tor service running on control port")

	// Check I2P configuration
	i2pAddr := GetI2PSAMAddr()
	log.Printf("I2P: CONFIGURED (SAM bridge: %s)", i2pAddr)
	log.Println("  - .i2p addresses will route through I2P")
	log.Println("  - Requires I2P router with SAM bridge enabled")

	log.Println()
	if IsAnonOnly() {
		log.Println("Transport routing (anon-only mode):")
		log.Println("  - All bootstrap via Tor SOCKS proxy")
		log.Println("  - .onion addresses -> Tor network")
		log.Println("  - .i2p addresses -> I2P network")
		log.Println("  - Direct IP connections -> BLOCKED")
	} else {
		log.Println("Transport routing (automatic based on address):")
		log.Println("  - Regular IP addresses -> Direct UDP/TCP")
		log.Println("  - .onion addresses -> Tor network")
		log.Println("  - .i2p addresses -> I2P network")
	}

	log.Println()
	log.Println("Note: Start Tor/I2P services before connecting to anonymity addresses")
	log.Println(logSeparator)
}
