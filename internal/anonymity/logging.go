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
// Deprecated: Use LogNetworkStatusWithMode for anon-only mode support.
func LogNetworkStatus() {
	LogNetworkStatusWithMode()
}
