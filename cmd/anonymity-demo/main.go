package main

import (
	"log"
	"os"
)

// Simple demonstration program showing Tor and I2P support logging
func main() {
	log.SetFlags(log.Ltime)
	log.Println("dtox - Tox Client with Anonymity Network Support")
	log.Println("================================================")
	
	// This function is from the main tox-gui package
	logAnonymityNetworkStatus()
	
	log.Println()
	log.Println("The toxcore library automatically routes traffic based on address:")
	log.Println("  - Regular IP addresses -> Direct UDP/TCP")
	log.Println("  - .onion addresses -> Tor network")
	log.Println("  - .i2p addresses -> I2P network")
	log.Println()
	log.Println("To use with custom addresses, set environment variables:")
	log.Println("  TOR_CONTROL_ADDR=<address:port>")
	log.Println("  I2P_SAM_ADDR=<address:port>")
}

// logAnonymityNetworkStatus logs the availability of anonymity networks (Tor and I2P).
// This is a copy of the function from backend.go for demonstration purposes.
func logAnonymityNetworkStatus() {
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
