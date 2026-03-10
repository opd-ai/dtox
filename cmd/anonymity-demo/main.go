package main

import (
	"log"

	"github.com/opd-ai/dtox/internal/anonymity"
)

// Simple demonstration program showing Tor and I2P support logging
func main() {
	log.SetFlags(log.Ltime)
	log.Println("dtox - Tox Client with Anonymity Network Support")
	log.Println("================================================")
	
	// This function is from the shared internal/anonymity package
	anonymity.LogNetworkStatus()
	
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
