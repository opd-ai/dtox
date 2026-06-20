# dtox
Tox client with cross-platform GUI support

## Features

- Tox messaging protocol support
- **Cross-platform GUI**:
  - **Linux**: Native Wayland/X11 support using [wain](https://github.com/opd-ai/wain) (GPU rendering, static binary)
  - **Windows, macOS, Android, iOS**: [wayne](https://github.com/opd-ai/wayne) (Ebitengine-based)
- **Multiple anonymity network support (Tor and I2P) - enabled by default**
- **Embedded SOCKS endpoint listener** for local integrations
- **Tor-over-Tox bridge** for friend-accessible Tor routing (via [opd-ai/toxpt](https://github.com/opd-ai/toxpt))
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

## Tor Bridge Integration

dtox includes an embedded **dual-service Tor bridge** that enables both local Tor access and Tor routing for your Tox friends:

### Components

1. **SOCKS Proxy Service** (via [opd-ai/go-tor](https://github.com/opd-ai/go-tor))
   - Local Tor access on `127.0.0.1:19050`
   - For applications that require SOCKS5 proxy support
   - Always-on by default

2. **Tor-over-Tox Bridge Service** (via [opd-ai/toxpt](https://github.com/opd-ai/toxpt))
   - Allows connected Tox friends to route their Tor traffic through you
   - Friend-only access (not exposed externally)
   - Always-on by default

### Independent Operation

Both services run concurrently and independently:
- If SOCKS proxy fails, the bridge continues operating
- If bridge fails, SOCKS proxy continues operating
- Failure of one service does not affect the other

### Configuration

Both services are **enabled by default**. To disable either service (or customize the SOCKS port), developers can modify the `torBridgeConfig` (`torbridge.Config`) in `cmd/tox-gui/backend.go`:

```go
torBridgeConfig := torbridge.DefaultConfig()
torBridgeConfig.EnableSOCKS = true  // Local SOCKS proxy on 127.0.0.1:19050
torBridgeConfig.EnableBridge = true // Friend-accessible bridge
torBridgeConfig.ToxInstance = tox
tb, err := torbridge.New(context.Background(), torBridgeConfig)
```

### Integration in dtox

For dtox contributors, integrating the Tor bridge is straightforward:

```go
import "github.com/opd-ai/dtox/internal/torbridge" // internal: only importable from dtox module

// After creating your Tox instance:
config := torbridge.DefaultConfig()
config.ToxInstance = toxInstance

tb, err := torbridge.New(context.Background(), config)
if err != nil {
    log.Printf("Tor bridge unavailable: %v", err)
    // Application continues with Tox-only mode
}
defer tb.Close()

// Access services:
if tb != nil && tb.IsSOCKSEnabled() {
    log.Printf("SOCKS proxy: %s", tb.GetSOCKSAddr())
}
if tb != nil && tb.IsBridgeEnabled() {
    log.Println("Tor-over-Tox bridge active for friends")
}
```

### Minimal Integration Code

The integration requires less than 150 lines of code and maintains **zero breaking changes** to existing Tor or Tox APIs:
- Both services leverage existing implementations from [opd-ai/go-tor](https://github.com/opd-ai/go-tor) and [opd-ai/toxpt](https://github.com/opd-ai/toxpt)
- No Tor compilation or management required
- No transport duplication
- Comprehensive inline documentation

### Startup Logging

When dtox starts, you'll see status messages:
```
Tor bridge services initialized:
  - SOCKS proxy: 127.0.0.1:19050 (local Tor access)
  - Tor-over-Tox bridge: friend-accessible
```

If bridge initialization fails, dtox continues in Tox-only mode:
```
Failed to initialize Tor bridge services: [error details]
Continuing with Tox-only mode. To enable Tor bridge:
  - Ensure requested listener ports are available and permitted
  - Check torbridge configuration (SOCKSAddr, EnableBridge, ToxInstance)
  - Review torbridge initialization errors in startup logs
```



## Build Reference

See `Makefile` for build instructions.

Donate Monero(The only good cryptocurrency) to support development
==================================================================

 - `monero:43H3Uqnc9rfEsJjUXZYmam45MbtWmREFSANAWY5hijY4aht8cqYaT2BCNhfBhua5XwNdx9Tb6BEdt4tjUHJDwNW5H7mTiwe`

