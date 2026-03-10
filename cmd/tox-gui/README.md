# tox-gui — Pure-Go Tox Messenger GUI Client

A statically-compiled, single-binary Tox messenger GUI for Linux built on:

| Dependency | Role |
|---|---|
| [`github.com/opd-ai/toxcore`](https://github.com/opd-ai/toxcore) | Tox protocol (DHT, friend management, messaging) |
| [`github.com/opd-ai/wain`](https://github.com/opd-ai/wain) | Pure-Go GUI toolkit (Wayland / X11, software renderer) |

## Features

- **Identity display** – your full Tox ID appears in the footer; share it with friends.
- **Friend list** – live sidebar showing each friend with an online / offline indicator.
- **Add friend** – paste a friend's Tox ID into the footer field and click *Add Friend*.
- **Auto-accept** – incoming friend requests are automatically accepted.
- **Chat view** – scrollable message history distinguishes sent vs received messages.
- **Message input** – type a message and press *Send* (or click the button).
- **Connection status** – header label tracks Tox network connection state.
- **Bootstrap on startup** – connects to well-known Tox bootstrap nodes automatically.
- **Thread-safe UI updates** – all Tox callbacks route through `app.Notify()`.

## Requirements

| Tool | Version |
|---|---|
| Go | ≥ 1.24 |
| Rust (cargo) | stable |
| musl-gcc | any (Ubuntu: `sudo apt-get install musl-tools`) |
| Linux display | Wayland *or* X11 |

## Building

```bash
# 1. Add the musl Rust target (once)
rustup target add x86_64-unknown-linux-musl

# 2. Clone wain and build its Rust static library
git clone https://github.com/opd-ai/wain.git /tmp/wain-src
cargo build --release \
  --manifest-path /tmp/wain-src/render-sys/Cargo.toml \
  --target x86_64-unknown-linux-musl

# 3. Build the dl_find_object stub required by musl + GCC 14+
musl-gcc -c /tmp/wain-src/internal/render/dl_find_object_stub.c \
  -o /tmp/dl_find_object_stub.o

# 4. Build tox-gui
RUST_LIB=/tmp/wain-src/render-sys/target/x86_64-unknown-linux-musl/release/librender_sys.a
CC=musl-gcc CGO_ENABLED=1 \
  CGO_LDFLAGS="${RUST_LIB} /tmp/dl_find_object_stub.o -ldl -lm -lpthread" \
  CGO_LDFLAGS_ALLOW=".*" \
  go build -ldflags="-s -w -extldflags=-static" -tags netgo \
  -o tox-gui ./cmd/tox-gui/
```

Or simply use the Makefile target (see `Makefile`):

```bash
make tox-gui
```

## Running

```bash
./tox-gui
```

`tox-gui` auto-detects whether Wayland or X11 is available.  If both are
absent it will exit with an error.

## Architecture

```
cmd/tox-gui/
├── main.go          # App lifecycle, window creation, signal handling
├── ui.go            # Widget tree construction + wain.Widget bridge
├── tox_backend.go   # Tox instance, callbacks, event loop goroutine
└── README.md        # This file
```

### Widget layout

```
╔══════════════════════════════════════════════════════════╗
║ Header Row:  [Tox Messenger]            [Connected(UDP)] ║
╠══════════╦═══════════════════════════════════════════════╣
║ Sidebar  ║  Chat Area                                    ║
║ Column   ║  ┌──────────────────────────────────────┐     ║
║          ║  │ ScrollView: message history           │     ║
║ ● Alice  ║  │  "Alice: hello"                       │     ║
║ ○ Bob    ║  │  "Me: hi there!"                      │     ║
║          ║  └──────────────────────────────────────┘     ║
║          ║  [ TextInput (80%)         ] [ Send (20%) ]   ║
╠══════════╩═══════════════════════════════════════════════╣
║ Footer: [Your Tox ID…] [Paste friend ID…] [Add Friend]  ║
╚══════════════════════════════════════════════════════════╝
```

### Thread safety

All Tox event callbacks (friend requests, messages, connection-status changes)
run on the goroutine driving `tox.Iterate()`.  Every update that touches a
wain widget is wrapped in `app.Notify(func() { … })`, which queues the closure
onto the wain event-loop goroutine so it executes safely.

## Security notes

- Private keys and raw message content are never logged.
- Tox provides end-to-end encryption for all messages via NaCl/libsodium-compatible
  routines implemented in [`github.com/opd-ai/toxcore/crypto`](https://github.com/opd-ai/toxcore).
- Static compilation (`-ldflags=-extldflags=-static`) eliminates dynamic-library
  substitution attacks.
