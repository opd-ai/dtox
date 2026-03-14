# Tox Messenger GUI — Cross-Platform Client

A Tox Messenger GUI client with cross-platform support:

- **Linux**: Built with [wain](https://github.com/opd-ai/wain) (Go GUI toolkit, Wayland/X11, GPU + software rendering). Produces a **statically linked** binary.
- **Windows, macOS, Android, iOS**: Built with [wayne](https://github.com/opd-ai/wayne) (Ebitengine-based cross-platform GUI).

Both backends use [toxcore](https://github.com/opd-ai/toxcore) (pure Go Tox protocol with Noise-IK encryption).

## Features

- End-to-end encrypted messaging via the Tox protocol
- Pure Go Tox protocol implementation (no C toxcore dependency)
- Dark theme with Tox-branded green accents
- Friend list with online/offline indicators
- Real-time message send/receive
- Friend request sending and accepting
- Tox ID display for sharing
- Automatic DHT bootstrap on startup
- Graceful shutdown on window close or SIGTERM

## Build

### Linux (Static Binary)

#### Prerequisites

| Tool | Install |
|------|---------|
| Go 1.24+ | <https://go.dev/dl/> |
| Rust + Cargo | <https://rustup.rs/> |
| musl-gcc | `sudo apt-get install musl-tools` (Debian/Ubuntu) |
| musl Rust target | `rustup target add x86_64-unknown-linux-musl` |

#### Quick build

```sh
make
```

This will:
1. Build the Rust render-sys static library (GPU rendering backend)
2. Compile the dl_find_object compatibility stub
3. Build the Go binary with static linking via musl

#### Manual build

```sh
# 1. Build the Rust rendering library
WAIN_DIR=$(go env GOMODCACHE)/github.com/opd-ai/wain@<version>
cp -r "$WAIN_DIR/render-sys" /tmp/render-sys-build
cd /tmp/render-sys-build
cargo build --release --target x86_64-unknown-linux-musl

# 2. Compile the stub object
musl-gcc -c -o /tmp/dl_find_object_stub.o "$WAIN_DIR/internal/render/dl_find_object_stub.c"

# 3. Build the Go binary
cd <repo-root>
CC=musl-gcc CGO_ENABLED=1 \
  CGO_LDFLAGS="/tmp/render-sys-build/target/x86_64-unknown-linux-musl/release/librender_sys.a /tmp/dl_find_object_stub.o -ldl -lm -lpthread" \
  CGO_LDFLAGS_ALLOW=".*" \
  go build -ldflags "-extldflags '-static' -s -w" -tags netgo -o tox-gui ./cmd/tox-gui/
```

#### Verify static linkage

```sh
file tox-gui
# → ELF 64-bit LSB executable, x86-64, statically linked, stripped
ldd tox-gui
# → not a dynamic executable
```

### Windows

```sh
make windows
```

Or manually:
```sh
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o tox-gui.exe ./cmd/tox-gui/
```

### macOS

Build natively on macOS:
```sh
go build -o tox-gui ./cmd/tox-gui/
```

Or use the Makefile target (may require a cross-compiler toolchain):
```sh
make darwin
```

## Runtime Requirements

### Linux
- Wayland compositor **or** X11 server
- Network access (UDP port 33445 for DHT bootstrap)

### Windows / macOS
- Network access (UDP port 33445 for DHT bootstrap)
- OpenGL or Metal support (for Ebitengine rendering)

## Architecture

```
cmd/tox-gui/
├── main.go        Entry point: app lifecycle orchestration
├── ui.go          Widget tree construction, layout, event wiring
├── backend.go     Tox instance wrapper, callback bridge
├── state.go       Thread-safe shared application state
├── theme.go       Tox-branded dark theme colors and styles
├── bootstrap.go   DHT bootstrap node list and connection logic
└── README.md      This file

internal/ui/
├── ui_linux.go    Linux platform: re-exports wain types
└── ui_other.go    Windows/macOS/Android/iOS: wrapper types for wayne
```

### UI Layout

```
Window (900×650)
└── root: Column (100% × 100%)
    ├── header: Row (100% × 8%)          — App title + connection status
    ├── body: Row (100% × 84%)           — Sidebar + chat area
    │   ├── sidebar: Column (25% × 100%) — Friend list
    │   └── chatArea: Column (75% × 100%)
    │       ├── chatHeader: Label         — Selected friend name
    │       ├── messageScroll: ScrollView — Messages
    │       └── inputRow: Row             — Input + send button
    └── footer: Row (100% × 8%)          — Add friend + Tox ID display
```

## Security

- All messages are end-to-end encrypted via Noise-IK (toxcore)
- Private keys are never logged
- Message content is never logged; only metadata (friend ID, timestamp) appears in logs
- Tox address is validated (76 hex characters) before friend requests
- Static binary eliminates dynamic library injection attacks

## Known Limitations

### Linux (wain)
- The wain widget rendering pipeline is under active development; some visual
  features (e.g., text placeholder rendering, scroll indicators) may not be
  fully rendered yet.
- The wain library requires CGO for its GPU rendering backend (Rust FFI via
  musl). The Go application code itself contains zero `import "C"` statements.

### Windows / macOS (wayne)
- Wayne uses Ebitengine which requires OpenGL or Metal support.
- Some visual features may differ slightly from the Linux version due to
  different rendering backends.
- Window close callbacks are not natively supported by Ebitengine; graceful
  shutdown relies on signal handling.

### All Platforms
- Message history is in-memory only; it is lost on restart.
  Use `tox.Save()` / `NewFromSavedata()` to persist Tox state (friend list,
  keys) across sessions in a future update.

## License

See the repository root [LICENSE](../../LICENSE) file.
