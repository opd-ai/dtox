# dtox – Tox Messenger GUI client
#
# Build a statically linked binary using musl and the Rust render library
# from wain.
#
# Prerequisites:
#   musl-gcc           sudo apt-get install musl-tools
#   Rust musl target   rustup target add x86_64-unknown-linux-musl
#
# Usage:
#   make              – build tox-gui
#   make clean        – remove build artifacts

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

.PHONY: all build rust stub clean

all: build

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
	rm -f $(OUTPUT)
	rm -rf $(BUILD_DIR) 2>/dev/null || chmod -R u+w $(BUILD_DIR) 2>/dev/null && rm -rf $(BUILD_DIR) 2>/dev/null; true
