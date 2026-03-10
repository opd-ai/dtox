package anonymity

import (
	"os"
	"sync"
	"testing"
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

func TestNewMultiTransportManager(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	manager := NewMultiTransportManager()
	if manager == nil {
		t.Fatal("NewMultiTransportManager returned nil")
	}
	defer manager.Close()

	// Verify the manager has supported networks
	networks := manager.GetSupportedNetworks()
	if len(networks) == 0 {
		t.Error("Expected at least one supported network")
	}

	// Check for expected network types
	hasIP := false
	hasTor := false
	hasI2P := false
	for _, network := range networks {
		switch network {
		case "tcp", "udp", "tcp4", "tcp6", "udp4", "udp6":
			hasIP = true
		case "tor":
			hasTor = true
		case "i2p":
			hasI2P = true
		}
	}

	if !hasIP {
		t.Error("Expected IP transport to be supported")
	}
	t.Logf("Supported networks: %v (IP: %v, Tor: %v, I2P: %v)", networks, hasIP, hasTor, hasI2P)
}

func TestMultiTransportManagerClose(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	manager := NewMultiTransportManager()
	if manager == nil {
		t.Fatal("NewMultiTransportManager returned nil")
	}

	// First close should succeed
	err := manager.Close()
	if err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	// Second close should also succeed (idempotent)
	err = manager.Close()
	if err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

func TestCheckTorAvailability(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	manager := NewMultiTransportManager()
	defer manager.Close()

	// Test with default address
	t.Run("DefaultAddress", func(t *testing.T) {
		status := manager.CheckTorAvailability()

		if status.Address != "127.0.0.1:9051" {
			t.Errorf("Expected default Tor address '127.0.0.1:9051', got '%s'", status.Address)
		}

		if status.NetworkType != "tor" {
			t.Errorf("Expected network type 'tor', got '%s'", status.NetworkType)
		}

		// Note: Availability depends on whether Tor is actually running
		t.Logf("Tor availability: %v (address: %s)", status.Available, status.Address)
		if !status.Available && status.Error != nil {
			t.Logf("Tor error: %v", status.Error)
		}
	})

	// Test with custom address
	t.Run("CustomAddress", func(t *testing.T) {
		os.Setenv("TOR_CONTROL_ADDR", "192.168.1.10:9151")

		status := manager.CheckTorAvailability()

		if status.Address != "192.168.1.10:9151" {
			t.Errorf("Expected custom Tor address '192.168.1.10:9151', got '%s'", status.Address)
		}

		// This should fail as the address is not reachable
		if status.Available {
			t.Error("Expected Tor to be unavailable with non-existent address")
		}
	})
}

func TestCheckI2PAvailability(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	manager := NewMultiTransportManager()
	defer manager.Close()

	// Test with default address
	t.Run("DefaultAddress", func(t *testing.T) {
		status := manager.CheckI2PAvailability()

		if status.Address != "127.0.0.1:7656" {
			t.Errorf("Expected default I2P address '127.0.0.1:7656', got '%s'", status.Address)
		}

		if status.NetworkType != "i2p" {
			t.Errorf("Expected network type 'i2p', got '%s'", status.NetworkType)
		}

		// Note: Availability depends on whether I2P is actually running
		t.Logf("I2P availability: %v (address: %s)", status.Available, status.Address)
		if !status.Available && status.Error != nil {
			t.Logf("I2P error: %v", status.Error)
		}
	})

	// Test with custom address
	t.Run("CustomAddress", func(t *testing.T) {
		os.Setenv("I2P_SAM_ADDR", "192.168.1.20:7756")

		status := manager.CheckI2PAvailability()

		if status.Address != "192.168.1.20:7756" {
			t.Errorf("Expected custom I2P address '192.168.1.20:7756', got '%s'", status.Address)
		}

		// This should fail as the address is not reachable
		if status.Available {
			t.Error("Expected I2P to be unavailable with non-existent address")
		}
	})
}

func TestGetTransportStatuses(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	manager := NewMultiTransportManager()
	defer manager.Close()

	statuses := manager.GetTransportStatuses()

	// Should have status for both Tor and I2P
	if _, ok := statuses["tor"]; !ok {
		t.Error("Expected status for 'tor' transport")
	}

	if _, ok := statuses["i2p"]; !ok {
		t.Error("Expected status for 'i2p' transport")
	}

	t.Logf("Transport statuses: Tor=%v, I2P=%v",
		statuses["tor"].Available, statuses["i2p"].Available)
}

func TestGlobalManager(t *testing.T) {
	cleanup := saveAndClearEnvVars()
	defer cleanup()

	// Reset global manager state
	globalManagerMu.Lock()
	globalManager = nil
	globalManagerOnce = sync.Once{}
	globalManagerMu.Unlock()

	// Get global manager
	manager1 := GetGlobalManager()
	if manager1 == nil {
		t.Fatal("GetGlobalManager returned nil")
	}

	// Get again - should return same instance
	manager2 := GetGlobalManager()
	if manager1 != manager2 {
		t.Error("GetGlobalManager should return the same instance")
	}

	// Close global manager
	err := CloseGlobalManager()
	if err != nil {
		t.Errorf("CloseGlobalManager failed: %v", err)
	}

	// After closing, a new call should create a new manager
	manager3 := GetGlobalManager()
	if manager3 == nil {
		t.Fatal("GetGlobalManager returned nil after close")
	}
	if manager3 == manager1 {
		t.Error("GetGlobalManager should return a new instance after close")
	}

	// Cleanup
	CloseGlobalManager()
}

func TestTransportNotInitializedError(t *testing.T) {
	manager := &MultiTransportManager{
		mt: nil,
	}

	// Dial should return error when transport is nil
	_, err := manager.Dial("127.0.0.1:8080")
	if err != ErrTransportNotInitialized {
		t.Errorf("Expected ErrTransportNotInitialized, got %v", err)
	}

	// DialPacket should return error when transport is nil
	_, err = manager.DialPacket("127.0.0.1:8080")
	if err != ErrTransportNotInitialized {
		t.Errorf("Expected ErrTransportNotInitialized, got %v", err)
	}

	// Listen should return error when transport is nil
	_, err = manager.Listen("127.0.0.1:8080")
	if err != ErrTransportNotInitialized {
		t.Errorf("Expected ErrTransportNotInitialized, got %v", err)
	}

	// GetSupportedNetworks should return nil when transport is nil
	networks := manager.GetSupportedNetworks()
	if networks != nil {
		t.Errorf("Expected nil networks, got %v", networks)
	}
}

func TestErrors(t *testing.T) {
	// Verify errors are properly defined
	if ErrTransportNotInitialized == nil {
		t.Error("ErrTransportNotInitialized should not be nil")
	}

	if ErrTransportClosed == nil {
		t.Error("ErrTransportClosed should not be nil")
	}

	if ErrTorNotAvailable == nil {
		t.Error("ErrTorNotAvailable should not be nil")
	}

	if ErrI2PNotAvailable == nil {
		t.Error("ErrI2PNotAvailable should not be nil")
	}

	// Verify error messages
	if ErrTransportNotInitialized.Error() != "transport not initialized" {
		t.Errorf("Unexpected error message: %s", ErrTransportNotInitialized.Error())
	}
}
