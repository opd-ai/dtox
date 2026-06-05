# dtox
Tox client with cross-platform GUI support

## Features

- Tox messaging protocol support
- **Cross-platform GUI**:
  - **Linux**: Native Wayland/X11 support using [wain](https://github.com/opd-ai/wain) (GPU rendering, static binary)
  - **Windows, macOS, Android, iOS**: [wayne](https://github.com/opd-ai/wayne) (Ebitengine-based)
- **Multiple anonymity network support (Tor and I2P) - enabled by default**
- Secure peer-to-peer communication
- Automatic transport selection based on address format

## Building

### Linux (Static Binary)

```sh
make
```

This produces a fully static binary using wain's Wayland/X11 + GPU rendering backend.

### Windows

```sh
make windows
```

Or cross-compile from any platform:
```sh
GOOS=windows GOARCH=amd64 go build -o tox-gui.exe ./cmd/tox-gui/
```

### macOS

Build natively on macOS:
```sh
go build -o tox-gui ./cmd/tox-gui/
```

Or use the Makefile target (may require cross-compiler):
```sh
make darwin
```

See `Makefile` for all build targets and options.

## Anonymity Network Support

dtox supports multiple anonymity networks through the `opd-ai/toxcore` library's multi-transport system. **Tor and I2P support is enabled by default** - when the services are running on your system, dtox will automatically use them for `.onion` and `.i2p` addresses.

### How It Works

The application initializes a multi-transport manager at startup that:
1. Registers IP, Tor, I2P, and Nym transports automatically
2. Checks for service availability and logs status
3. Routes connections through the appropriate network based on address format

### Startup Status

When you start dtox, it will display the availability status of each anonymity network:
- **AVAILABLE**: The service is running and dtox can use it
- **NOT AVAILABLE**: The service is not running (you can still use dtox with regular IP addresses)

### Tor Support

Tor support is built-in and automatically active when Tor is running. The Tox client can:
- Connect to bootstrap nodes via Tor
- Accept friend requests from .onion addresses
- Route traffic through the Tor network for enhanced privacy

**Configuration:**
- `TOR_CONTROL_ADDR`: Tor control port address (default: `127.0.0.1:9051`)
- Ensure Tor is running on your system before starting dtox

**Example:**
```bash
# Start Tor service
tor &

# Run dtox with Tor support (no extra configuration needed if using defaults)
./tox-gui

# Or with custom Tor control address
TOR_CONTROL_ADDR=127.0.0.1:9051 ./tox-gui
```

### I2P Support

I2P (Invisible Internet Project) support is also built-in. The client can:
- Connect to I2P peers via .i2p addresses
- Use the SAM bridge protocol for I2P connectivity
- Maintain anonymous peer-to-peer connections

**Configuration:**
- `I2P_SAM_ADDR`: I2P SAM bridge address (default: `127.0.0.1:7656`)
- Ensure I2P router is running with SAM bridge enabled

**Example:**
```bash
# Ensure I2P router is running with SAM bridge on port 7656

# Run dtox with I2P support (no extra configuration needed if using defaults)
./tox-gui

# Or with custom I2P SAM address
I2P_SAM_ADDR=127.0.0.1:7656 ./tox-gui
```

### Using Both Networks Simultaneously

You can enable both Tor and I2P at the same time for maximum privacy and connectivity:

```bash
# Start both Tor and I2P
tor &
# (I2P router should be running)

# Run dtox with both networks (automatic detection)
./tox-gui

# Or with explicit configuration
TOR_CONTROL_ADDR=127.0.0.1:9051 I2P_SAM_ADDR=127.0.0.1:7656 ./tox-gui
```

### Automatic Transport Routing

When connecting to peers, dtox automatically selects the appropriate transport:
- **Regular IP addresses** → Direct UDP/TCP connections
- **`.onion` addresses** → Tor network
- **`.i2p` addresses** → I2P network
- **`.nym` addresses** → Nym mixnet (if available)

The transport selection is completely automatic based on the address format - no manual configuration required.

### Supported Networks

The multi-transport system supports the following network types:
- **IP**: tcp, udp, tcp4, tcp6, udp4, udp6
- **Tor**: .onion hidden services
- **I2P**: .i2p destinations via SAM bridge
- **Nym**: .nym mixnet addresses (experimental)

## Tor-over-Tox Bridge Integration

dtox provides a **Tor-over-Tox bridge** module that enables seamless integration with peer-to-peer Tox bridges for enhanced privacy and resilience. This feature is fully implemented and ready for use by Go Tox client developers.

### Bridge Overview

The bridge provides:
- **SOCKS5 Proxy**: Listen on 127.0.0.1:19050 for transparent traffic routing
- **Failover State Machine**: Chooses between direct routing and (future) Tox-friend routes
- **Status & Metrics**: Exposes connection counts and failover state
- **Simple Integration**: Single-function initialization for Go clients


### How It Works

The bridge implements a failover state machine:

1. **Primary Route**: When Tox friend bridges are available, traffic is routed through them
2. **Fallback Route**: If no Tox friends are available, traffic falls back to direct Tor
3. **Automatic Detection**: Bridge monitors Tox friend availability and switches routes automatically
4. **Transparent**: No client-side logic needed - the bridge handles routing automatically

### Integration for Go Clients

Minimal integration required - just 3 steps:

```go
import (
    "github.com/opd-ai/dtox/internal/anonymity"
    "github.com/opd-ai/dtox/internal/bridge"
)

func initBridge() (*bridge.TOXBridge, error) {
    // 1. Initialize multi-transport manager
    transportMgr := anonymity.NewMultiTransportManager()
    multiTransport := transportMgr.GetMultiTransport()
    
    // 2. Create bridge (enabled by default)
    return bridge.NewTOXBridge(multiTransport, true)
}

func main() {
    b, err := initBridge()
    if err != nil {
        log.Fatal(err)
    }
    defer b.Close()
    
    // 3. Configure your client to use SOCKS5 proxy at 127.0.0.1:19050
    // Bridge handles everything else automatically
    
    // Query bridge status if needed
    status := b.Status()
    log.Printf("Bridge running: %v, Tox friends: %d", 
        status.Enabled, status.ActiveToxFriends)
}
```

### Bridge API

**Initialization:**
```go
bridge, err := bridge.NewTOXBridge(multiTransport, enabled bool)
```

**Status Query:**
```go
status := bridge.Status()
// Returns: Enabled, ListeningAddr, ActiveToxFriends, TorAvailable, 
//          LastFailoverUpdate, TotalConnections
```

**Shutdown:**
```go
err := bridge.Close() // Graceful shutdown, safe to call multiple times
```

### Configuration

The bridge uses environment variables inherited from the multi-transport system:
- `TOR_CONTROL_ADDR`: Tor control port (default: `127.0.0.1:9051`)
- `I2P_SAM_ADDR`: I2P SAM bridge (default: `127.0.0.1:7656`)

### Example Usage

See `examples/bridge_integration.go` for a complete working example:

```bash
go run examples/bridge_integration.go
```

The example demonstrates:
- Bridge initialization
- Status monitoring
- Graceful shutdown
- Integration patterns for production applications

### SOCKS5 Compatibility

The bridge implements full SOCKS5 protocol support (RFC 1928):
- ✅ CONNECT command (TCP connections)
- ✅ IPv4, IPv6, and domain address types
- ✅ No authentication (SOCKS5AuthNoAuth)
- ✅ Connection relay and bidirectional streaming

External SOCKS5 clients can connect:
```bash
# curl with SOCKS5 proxy
curl --socks5 127.0.0.1:19050 https://example.onion

# Tor Browser with SOCKS5 proxy
# Configure Proxy Address: 127.0.0.1, Port: 19050, Type: SOCKS5
```

### Failover State Machine

The bridge maintains an internal state machine that tracks:
- Current routing mode (Tox friends vs. direct Tor)
- Number of active Tox friend bridges
- Tor availability status
- Last state transition time

State transitions are logged for debugging:
```
[Bridge] Failover state transition: RouteDirect → RouteToxFriends (friends=2, tor=true)
```

### Testing

Comprehensive test suite included:

```bash
go test ./internal/bridge/... -v
```

Tests cover:
- Bridge initialization and lifecycle
- SOCKS5 protocol handling
- Failover state machine transitions
- Concurrency and thread safety
- Connection tracking and metrics

### Zero Breaking Changes

The bridge integration:
- ✅ Non-invasive to existing dtox functionality
- ✅ Optional feature (disabled by passing `enabled=false`)
- ✅ No changes to toxcore or transport APIs
- ✅ Backward compatible with existing code

## Building

See `Makefile` for build instructions.
