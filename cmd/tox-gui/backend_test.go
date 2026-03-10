package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/opd-ai/dtox/internal/anonymity"
)

// saveAndClearEnvVars saves the current values of Tor and I2P environment variables
// and clears them for testing. Returns a cleanup function that restores the original values.
func saveAndClearEnvVars() func() {
	origTor := os.Getenv("TOR_CONTROL_ADDR")
	origI2P := os.Getenv("I2P_SAM_ADDR")
	
	os.Unsetenv("TOR_CONTROL_ADDR")
	os.Unsetenv("I2P_SAM_ADDR")
	
	return func() {
		if origTor != "" {
			os.Setenv("TOR_CONTROL_ADDR", origTor)
		} else {
			os.Unsetenv("TOR_CONTROL_ADDR")
		}
		if origI2P != "" {
			os.Setenv("I2P_SAM_ADDR", origI2P)
		} else {
			os.Unsetenv("I2P_SAM_ADDR")
		}
	}
}

// TestLogAnonymityNetworkStatus verifies that the anonymity network logging
// function correctly displays Tor and I2P configuration.
func TestLogAnonymityNetworkStatus(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with default values (no environment variables set)
	t.Run("DefaultConfiguration", func(t *testing.T) {
		buf.Reset()
		cleanup := saveAndClearEnvVars()
		defer cleanup()
		
		anonymity.LogNetworkStatus()
		
		output := buf.String()
		
		// Verify output contains expected messages
		if !strings.Contains(output, "Anonymity Network Support") {
			t.Error("Expected 'Anonymity Network Support' in output")
		}
		
		if !strings.Contains(output, "Tor: CONFIGURED") {
			t.Error("Expected 'Tor: CONFIGURED' in output")
		}
		
		if !strings.Contains(output, "127.0.0.1:9051") {
			t.Error("Expected default Tor control address '127.0.0.1:9051' in output")
		}
		
		if !strings.Contains(output, "I2P: CONFIGURED") {
			t.Error("Expected 'I2P: CONFIGURED' in output")
		}
		
		if !strings.Contains(output, "127.0.0.1:7656") {
			t.Error("Expected default I2P SAM address '127.0.0.1:7656' in output")
		}
		
		if !strings.Contains(output, ".onion addresses") {
			t.Error("Expected mention of .onion addresses in output")
		}
		
		if !strings.Contains(output, ".i2p addresses") {
			t.Error("Expected mention of .i2p addresses in output")
		}
	})

	// Test with custom environment variables
	t.Run("CustomConfiguration", func(t *testing.T) {
		buf.Reset()
		cleanup := saveAndClearEnvVars()
		defer cleanup()
		
		// Set custom environment variables
		os.Setenv("TOR_CONTROL_ADDR", "192.168.1.10:9151")
		os.Setenv("I2P_SAM_ADDR", "192.168.1.20:7756")
		
		anonymity.LogNetworkStatus()
		
		output := buf.String()
		
		// Verify custom addresses are used
		if !strings.Contains(output, "192.168.1.10:9151") {
			t.Error("Expected custom Tor control address '192.168.1.10:9151' in output")
		}
		
		if !strings.Contains(output, "192.168.1.20:7756") {
			t.Error("Expected custom I2P SAM address '192.168.1.20:7756' in output")
		}
	})
}

