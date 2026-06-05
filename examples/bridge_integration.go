// Package main demonstrates how to integrate the Tor-over-Tox bridge into a Go Tox client.
//
// This example shows:
// 1. Initializing the multi-transport manager for anonymity network support
// 2. Creating and starting the Tor-over-Tox bridge
// 3. Querying bridge status
// 4. Gracefully shutting down the bridge
//
// Usage:
//   go run examples/bridge_integration.go
//
// The bridge will:
// - Listen on 127.0.0.1:19050 as a SOCKS5 proxy
// - Automatically route traffic through available Tox friends
// - Fall back to direct Tor if no Tox friends are available
// - Track connection metrics and bridge status
//
// External clients can connect to the SOCKS5 proxy:
//   curl --socks5 127.0.0.1:19050 https://example.onion
//   tor-browser --proxy-server=socks5://127.0.0.1:19050
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/dtox/internal/anonymity"
	"github.com/opd-ai/dtox/internal/bridge"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("=== Tor-over-Tox Bridge Integration Example ===\n")

	// Step 1: Initialize the multi-transport manager
	// This manages Tor, I2P, and other anonymity networks
	transportMgr := anonymity.NewMultiTransportManager()
	defer transportMgr.Close()

	log.Println("✓ Multi-transport manager initialized")

	// Log available transports
	transportMgr.LogTransportStatus()

	// Step 2: Get the underlying MultiTransport for the bridge
	// The bridge needs direct access to route traffic
	multiTransport := transportMgr.GetMultiTransport()
	if multiTransport == nil {
		log.Fatal("Failed to get MultiTransport")
	}

	// Step 3: Create and initialize the Tor-over-Tox bridge
	// enabled=true means the bridge starts immediately
	// To disable the bridge initially, pass enabled=false
	bridgeInstance, err := bridge.NewTOXBridge(multiTransport, true)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}
	defer bridgeInstance.Close()

	log.Println("✓ Tor-over-Tox bridge initialized and listening on 127.0.0.1:19050")

	// Step 4: Query bridge status
	status := bridgeInstance.Status()
	fmt.Printf("\nBridge Status:\n")
	fmt.Printf("  Enabled:           %v\n", status.Enabled)
	fmt.Printf("  Listening Address: %s\n", status.ListeningAddr)
	fmt.Printf("  Tor Available:     %v\n", status.TorAvailable)
	fmt.Printf("  Active Tox Friends: %d\n", status.ActiveToxFriends)
	fmt.Printf("  Total Connections: %d\n", status.TotalConnections)
	fmt.Printf("  Last Update:       %v\n", status.LastFailoverUpdate)

	// Step 5: Demonstrate bridge usage
	// In a real application, you would:
	// 1. Point your Tor client to use 127.0.0.1:19050 as a SOCKS5 proxy
	// 2. The bridge automatically handles failover between Tox friends and direct Tor
	// 3. Monitor bridge status periodically for debugging

	log.Println("\n✓ Bridge is ready to accept connections on 127.0.0.1:19050")
	log.Println("  Configure your client to use: socks5://127.0.0.1:19050")

	// Simulate the bridge running for 5 seconds while monitoring status
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)

		status := bridgeInstance.Status()
		log.Printf("[Status %d] Connections: %d, Active Friends: %d, Tor: %v",
			i+1, status.TotalConnections, status.ActiveToxFriends, status.TorAvailable)
	}

	log.Println("\n✓ Shutting down bridge...")
	if err := bridgeInstance.Close(); err != nil {
		log.Printf("Warning during bridge shutdown: %v", err)
	}

	log.Println("✓ Example completed successfully")
}
