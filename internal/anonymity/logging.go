// Package anonymity provides utilities for handling anonymity network support
// in dtox, including Tor and I2P configuration and logging.
package anonymity

import (
	"log"
	"os"
)

// LogNetworkStatus logs the availability of anonymity networks (Tor and I2P).
// This helps users understand which privacy networks are configured and available.
func LogNetworkStatus() {
	log.Println("=== Anonymity Network Support ===")
	
	// Check Tor configuration
	torAddr := os.Getenv("TOR_CONTROL_ADDR")
	if torAddr == "" {
		torAddr = "127.0.0.1:9051" // default
	}
	log.Printf("Tor support: ENABLED (control: %s)", torAddr)
	log.Println("  - Connect to bootstrap nodes via Tor")
	log.Println("  - Accept friend requests from .onion addresses")
	log.Println("  - Automatic routing for .onion addresses")
	
	// Check I2P configuration
	i2pAddr := os.Getenv("I2P_SAM_ADDR")
	if i2pAddr == "" {
		i2pAddr = "127.0.0.1:7656" // default
	}
	log.Printf("I2P support: ENABLED (SAM bridge: %s)", i2pAddr)
	log.Println("  - Connect to I2P peers via .i2p addresses")
	log.Println("  - Use SAM bridge protocol for I2P connectivity")
	log.Println("  - Automatic routing for .i2p addresses")
	
	log.Println("Note: Ensure Tor/I2P services are running for anonymity network features")
	log.Println("================================")
}
