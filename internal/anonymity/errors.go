// Package anonymity provides utilities for handling anonymity network support
// in dtox, including Tor and I2P configuration and transport initialization.
package anonymity

import "errors"

// Errors for the anonymity package.
var (
	// ErrTransportNotInitialized is returned when attempting to use a transport
	// that has not been initialized.
	ErrTransportNotInitialized = errors.New("transport not initialized")

	// ErrTransportClosed is returned when attempting to use a transport
	// that has been closed.
	ErrTransportClosed = errors.New("transport closed")

	// ErrTorNotAvailable is returned when Tor is not available.
	ErrTorNotAvailable = errors.New("tor service not available")

	// ErrI2PNotAvailable is returned when I2P is not available.
	ErrI2PNotAvailable = errors.New("i2p service not available")
)
