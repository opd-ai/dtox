# dtox – Tox Messenger GUI client
#
# Cross-platform build targets:
#   make              – build tox-gui for Linux (static binary using wain/Wayland/X11)
#   make windows      – build tox-gui.exe for Windows (using wayne/Ebitengine)
#   make darwin       – build tox-gui-darwin for macOS (using wayne/Ebitengine)
#   make clean        – remove build artifacts
#
# Prerequisites for Linux static build:
#   musl-gcc           sudo apt-get install musl-tools
#   Rust musl target   rustup target add x86_64-unknown-linux-musl
#
# Prerequisites for Windows/macOS cross-compilation:
#   Go cross-compilation support (built-in)
#   CGO cross-compiler toolchain (for CGO-enabled builds)

SHELL := /bin/bash

CC := musl-gcc

# Architecture detection
RUST_HOST        := $(shell rustc -vV 2>/dev/null | awk '/^host:/{print $$2}')
HOST_ARCH        := $(firstword $(subst -, ,$(RUST_HOST)))
RUST_MUSL_TARGET := $(HOST_ARCH)-unknown-linux-musl

# Resolve the wain module directory from the Go module cache.
WAIN_VERSION     := $(shell go list -m -f '{{.Version}}' github.com/opd-ai/wain 2>/dev/null)
WAIN_DIR         := $(shell go env GOMODCACHE)/github.com/opd-ai/wain@$(WAIN_VERSION)

# Build directories (outside the source tree)
BUILD_DIR        := /tmp/dtox-build
RUST_SRC         := $(BUILD_DIR)/render-sys
RUST_LIB         := $(BUILD_DIR)/target/$(RUST_MUSL_TARGET)/release/librender_sys.a
DL_STUB_OBJ      := $(BUILD_DIR)/dl_find_object_stub.o

OUTPUT           := tox-gui
OUTPUT_WINDOWS   := tox-gui.exe
OUTPUT_DARWIN    := tox-gui-darwin

.PHONY: all build rust stub clean windows darwin

all: build

# Linux static build (using wain with Wayland/X11 + GPU rendering)
build: rust stub
	CC=$(CC) CGO_ENABLED=1 \
	  CGO_LDFLAGS="$(RUST_LIB) $(DL_STUB_OBJ) -ldl -lm -lpthread" \
	  CGO_LDFLAGS_ALLOW=".*" \
	  go build \
	    -ldflags "-extldflags '-static' -s -w" \
	    -tags netgo \
	    -o $(OUTPUT) ./cmd/tox-gui/
	@echo ""
	@echo "✓ Built $(OUTPUT)"
	@file $(OUTPUT) | grep -q "statically linked" && echo "✓ Statically linked" || echo "⚠ Not statically linked"

# Windows build (using wayne with Ebitengine)
# Note: Cross-compilation from Linux requires CGO_ENABLED=0 or a Windows cross-compiler
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	  go build \
	    -ldflags "-s -w" \
	    -o $(OUTPUT_WINDOWS) ./cmd/tox-gui/
	@echo ""
	@echo "✓ Built $(OUTPUT_WINDOWS) for Windows/amd64"

# macOS build (using wayne with Ebitengine)
# Note: Full CGO builds require macOS or a macOS cross-compiler toolchain
# This target produces a non-CGO build suitable for testing; for production,
# build natively on macOS with: go build -o tox-gui ./cmd/tox-gui/
darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
	  go build \
	    -ldflags "-s -w" \
	    -o $(OUTPUT_DARWIN) ./cmd/tox-gui/ 2>/dev/null || \
	  echo "⚠ macOS build requires native compilation or a cross-compiler toolchain"
	@if [ -f $(OUTPUT_DARWIN) ]; then echo "✓ Built $(OUTPUT_DARWIN) for macOS/amd64"; fi

rust: $(RUST_LIB)

$(RUST_LIB):
	@mkdir -p $(BUILD_DIR)
	@if [ ! -d "$(RUST_SRC)" ]; then cp -r "$(WAIN_DIR)/render-sys" "$(RUST_SRC)" && chmod -R u+w "$(RUST_SRC)"; fi
	CARGO_TARGET_DIR=$(BUILD_DIR)/target cargo build \
	  --release \
	  --target $(RUST_MUSL_TARGET) \
	  --manifest-path $(RUST_SRC)/Cargo.toml

stub: $(DL_STUB_OBJ)

$(DL_STUB_OBJ):
	@mkdir -p $(BUILD_DIR)
	$(CC) -c -o $(DL_STUB_OBJ) "$(WAIN_DIR)/internal/render/dl_find_object_stub.c"

clean:
	rm -f $(OUTPUT) $(OUTPUT_WINDOWS) $(OUTPUT_DARWIN)
	rm -rf $(BUILD_DIR) 2>/dev/null || chmod -R u+w $(BUILD_DIR) 2>/dev/null && rm -rf $(BUILD_DIR) 2>/dev/null; true
