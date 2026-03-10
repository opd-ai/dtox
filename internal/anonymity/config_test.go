package anonymity

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

// saveAndClearEnvVars saves the current values of Tor and I2P environment variables
// and clears them for testing. Returns a cleanup function that restores the original values.
func saveAndClearEnvVars() func() {
	origTorControl := os.Getenv("TOR_CONTROL_ADDR")
	origTorSOCKS := os.Getenv("TOR_SOCKS_ADDR")
	origI2P := os.Getenv("I2P_SAM_ADDR")

	os.Unsetenv("TOR_CONTROL_ADDR")
	os.Unsetenv("TOR_SOCKS_ADDR")
	os.Unsetenv("I2P_SAM_ADDR")

	return func() {
		if origTorControl != "" {
			os.Setenv("TOR_CONTROL_ADDR", origTorControl)
		} else {
			os.Unsetenv("TOR_CONTROL_ADDR")
		}
		if origTorSOCKS != "" {
			os.Setenv("TOR_SOCKS_ADDR", origTorSOCKS)
		} else {
			os.Unsetenv("TOR_SOCKS_ADDR")
		}
		if origI2P != "" {
			os.Setenv("I2P_SAM_ADDR", origI2P)
		} else {
			os.Unsetenv("I2P_SAM_ADDR")
		}
	}
}

// TestAnonOnlyMode verifies that the anon-only mode can be set and retrieved.
func TestAnonOnlyMode(t *testing.T) {
	// Save and reset anon-only mode after test
	origAnonOnly := IsAnonOnly()
	defer SetAnonOnly(origAnonOnly)

	t.Run("DefaultIsFalse", func(t *testing.T) {
		SetAnonOnly(false)
		if IsAnonOnly() {
			t.Error("Expected anon-only mode to be false by default")
		}
	})

	t.Run("SetToTrue", func(t *testing.T) {
		SetAnonOnly(true)
		if !IsAnonOnly() {
			t.Error("Expected anon-only mode to be true after setting")
		}
	})

	t.Run("SetToFalse", func(t *testing.T) {
		SetAnonOnly(true)
		SetAnonOnly(false)
		if IsAnonOnly() {
			t.Error("Expected anon-only mode to be false after resetting")
		}
	})
}

// TestGetTorControlAddr verifies the Tor control address configuration.
func TestGetTorControlAddr(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	t.Run("DefaultAddress", func(t *testing.T) {
		addr := GetTorControlAddr()
		if addr != "127.0.0.1:9051" {
			t.Errorf("Expected default Tor control address '127.0.0.1:9051', got '%s'", addr)
		}
	})

	t.Run("CustomAddress", func(t *testing.T) {
		os.Setenv("TOR_CONTROL_ADDR", "192.168.1.1:9151")
		addr := GetTorControlAddr()
		if addr != "192.168.1.1:9151" {
			t.Errorf("Expected custom Tor control address '192.168.1.1:9151', got '%s'", addr)
		}
	})
}

// TestGetTorSOCKSAddr verifies the Tor SOCKS proxy address configuration.
func TestGetTorSOCKSAddr(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	t.Run("DefaultAddress", func(t *testing.T) {
		addr := GetTorSOCKSAddr()
		if addr != "127.0.0.1:9050" {
			t.Errorf("Expected default Tor SOCKS address '127.0.0.1:9050', got '%s'", addr)
		}
	})

	t.Run("CustomAddress", func(t *testing.T) {
		os.Setenv("TOR_SOCKS_ADDR", "192.168.1.1:9150")
		addr := GetTorSOCKSAddr()
		if addr != "192.168.1.1:9150" {
			t.Errorf("Expected custom Tor SOCKS address '192.168.1.1:9150', got '%s'", addr)
		}
	})
}

// TestGetI2PSAMAddr verifies the I2P SAM bridge address configuration.
func TestGetI2PSAMAddr(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	t.Run("DefaultAddress", func(t *testing.T) {
		addr := GetI2PSAMAddr()
		if addr != "127.0.0.1:7656" {
			t.Errorf("Expected default I2P SAM address '127.0.0.1:7656', got '%s'", addr)
		}
	})

	t.Run("CustomAddress", func(t *testing.T) {
		os.Setenv("I2P_SAM_ADDR", "192.168.1.1:7756")
		addr := GetI2PSAMAddr()
		if addr != "192.168.1.1:7756" {
			t.Errorf("Expected custom I2P SAM address '192.168.1.1:7756', got '%s'", addr)
		}
	})
}

// TestParseHostPort verifies the host:port parsing function.
func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		defaultPort uint16
		wantHost    string
		wantPort    uint16
	}{
		{"IPv4WithPort", "127.0.0.1:9050", 8080, "127.0.0.1", 9050},
		{"HostnameWithPort", "localhost:9150", 8080, "localhost", 9150},
		{"IPv4NoPort", "127.0.0.1", 9050, "127.0.0.1", 9050},
		{"HostnameNoPort", "localhost", 9050, "localhost", 9050},
		{"EmptyString", "", 9050, "", 9050},
		{"PortOnly", ":9050", 8080, "", 9050},
		{"TrailingColon", "localhost:", 9050, "localhost", 9050},
		{"InvalidPort", "localhost:abc", 9050, "localhost:abc", 9050},
		{"PortTooLarge", "localhost:99999", 9050, "localhost:99999", 9050},
		{"PortZero", "localhost:0", 9050, "localhost:0", 9050},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := ParseHostPort(tt.addr, tt.defaultPort)
			if host != tt.wantHost {
				t.Errorf("ParseHostPort(%q, %d) host = %q, want %q", tt.addr, tt.defaultPort, host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("ParseHostPort(%q, %d) port = %d, want %d", tt.addr, tt.defaultPort, port, tt.wantPort)
			}
		})
	}
}

// TestLogNetworkStatusWithMode verifies the logging function shows anon-only mode correctly.
func TestLogNetworkStatusWithMode(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Save and reset anon-only mode after test
	origAnonOnly := IsAnonOnly()
	defer SetAnonOnly(origAnonOnly)

	cleanup := saveAndClearEnvVars()
	defer cleanup()

	t.Run("NormalMode", func(t *testing.T) {
		buf.Reset()
		SetAnonOnly(false)
		LogNetworkStatusWithMode()

		output := buf.String()

		if strings.Contains(output, "ANONYMOUS-ONLY MODE") {
			t.Error("Should not show ANONYMOUS-ONLY MODE when disabled")
		}

		if !strings.Contains(output, "Regular IP addresses -> Direct UDP/TCP") {
			t.Error("Expected normal routing info when anon-only is disabled")
		}
	})

	t.Run("AnonOnlyMode", func(t *testing.T) {
		buf.Reset()
		SetAnonOnly(true)
		LogNetworkStatusWithMode()

		output := buf.String()

		if !strings.Contains(output, "ANONYMOUS-ONLY MODE ENABLED") {
			t.Error("Expected 'ANONYMOUS-ONLY MODE ENABLED' when anon-only is true")
		}

		if !strings.Contains(output, "All traffic will route through Tor") {
			t.Error("Expected Tor routing info when anon-only is enabled")
		}

		if !strings.Contains(output, "Direct IP connections -> BLOCKED") {
			t.Error("Expected blocked direct connections info when anon-only is enabled")
		}
	})
}
