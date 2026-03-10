// Package anonymity provides utilities for handling anonymity network support
// in dtox, including Tor and I2P configuration and logging.
package anonymity

import (
	"log"
	"os"
)

const (
	// Log formatting constants
	logSeparator = "================================"
	logHeader    = "=== Anonymity Network Support ==="
)

// LogNetworkStatus logs the configuration of anonymity networks (Tor and I2P).
// This helps users understand which privacy networks are configured.
// Note: This shows configuration only - the actual services must be running separately.
func LogNetworkStatus() {
	log.Println(logHeader)
	
	// Check Tor configuration
	torAddr := os.Getenv("TOR_CONTROL_ADDR")
	if torAddr == "" {
		torAddr = "127.0.0.1:9051" // default
	}
	log.Printf("Tor: CONFIGURED (control: %s)", torAddr)
	log.Println("  - .onion addresses will route through Tor")
	log.Println("  - Requires Tor service running on control port")
	
	// Check I2P configuration
	i2pAddr := os.Getenv("I2P_SAM_ADDR")
	if i2pAddr == "" {
		i2pAddr = "127.0.0.1:7656" // default
	}
	log.Printf("I2P: CONFIGURED (SAM bridge: %s)", i2pAddr)
	log.Println("  - .i2p addresses will route through I2P")
	log.Println("  - Requires I2P router with SAM bridge enabled")
	
	log.Println()
	log.Println("Transport routing (automatic based on address):")
	log.Println("  - Regular IP addresses -> Direct UDP/TCP")
	log.Println("  - .onion addresses -> Tor network")
	log.Println("  - .i2p addresses -> I2P network")
	
	log.Println()
	log.Println("Note: Start Tor/I2P services before connecting to anonymity addresses")
	log.Println(logSeparator)
}
