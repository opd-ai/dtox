# dtox
tox client using wain

## Features

- Tox messaging protocol support
- Modern GUI using wain
- **Multiple anonymity network support (Tor and I2P)**
- Secure peer-to-peer communication

## Anonymity Network Support

dtox supports multiple anonymity networks through the opd-ai/toxcore library:

### Tor Support

Tor support is built-in and automatically active. The Tox client can:
- Connect to bootstrap nodes via Tor
- Accept friend requests from .onion addresses
- Route traffic through the Tor network for enhanced privacy

**Configuration:**
- `TOR_CONTROL_ADDR`: Tor control port address (default: `127.0.0.1:9051`)
- `TOR_SOCKS_ADDR`: Tor SOCKS5 proxy address (default: `127.0.0.1:9050`)
- Ensure Tor is running on your system before starting dtox

**Example:**
```bash
# Start Tor service
tor &

# Run dtox with Tor support
TOR_CONTROL_ADDR=127.0.0.1:9051 ./tox-gui
```

### Anonymous-Only Mode (No Clearnet)

For maximum privacy, dtox supports an anonymous-only mode that routes **all** traffic through Tor:

```bash
# Run with anonymous-only mode (no clearnet connections)
./tox-gui --anon-only
```

**What this does:**
- Disables UDP connections (forces TCP-only mode)
- Routes all TCP connections through the Tor SOCKS5 proxy
- Prevents any direct clearnet connections that could leak your IP
- Bootstrap nodes are contacted through Tor

**Configuration with custom Tor SOCKS address:**
```bash
TOR_SOCKS_ADDR=127.0.0.1:9150 ./tox-gui --anon-only
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

# Run dtox with I2P support
I2P_SAM_ADDR=127.0.0.1:7656 ./tox-gui
```

### Using Both Networks Simultaneously

You can enable both Tor and I2P at the same time for maximum privacy and connectivity:

```bash
# Start both Tor and I2P
tor &
# (I2P router should be running)

# Run dtox with both networks
TOR_CONTROL_ADDR=127.0.0.1:9051 I2P_SAM_ADDR=127.0.0.1:7656 ./tox-gui
```

**Note:** When connecting to peers:
- Regular IP addresses will use direct UDP/TCP connections
- `.onion` addresses will automatically route through Tor
- `.i2p` addresses will automatically route through I2P
- The transport selection is automatic based on the address format

## Building

See `Makefile` for build instructions.
