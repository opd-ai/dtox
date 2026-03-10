# Makefile for tox-gui
#
# Prerequisites
#   - Rust (stable) + musl target:  rustup target add x86_64-unknown-linux-musl
#   - musl-gcc:                      sudo apt-get install musl-tools
#   - Go ≥ 1.24

SHELL        := /bin/bash
CC           := musl-gcc
RUST_HOST    := $(shell rustc -vV 2>/dev/null | awk '/^host:/{print $$2}')
HOST_ARCH    := $(firstword $(subst -, ,$(RUST_HOST)))
MUSL_TARGET  := $(HOST_ARCH)-unknown-linux-musl

WAIN_DIR     ?= /tmp/wain-src
RUST_LIB     := $(WAIN_DIR)/render-sys/target/$(MUSL_TARGET)/release/librender_sys.a
DL_STUB_SRC  := $(WAIN_DIR)/internal/render/dl_find_object_stub.c
DL_STUB_OBJ  := /tmp/dl_find_object_stub.o

OUT          := tox-gui
PKG          := ./cmd/tox-gui/

.PHONY: all tox-gui clean deps rust-lib dl-stub

all: tox-gui

## ── Dependencies ─────────────────────────────────────────────────────────────

deps:
	@command -v musl-gcc >/dev/null 2>&1 || \
	  { echo "ERROR: musl-gcc not found – run: sudo apt-get install musl-tools"; exit 1; }
	@command -v cargo >/dev/null 2>&1 || \
	  { echo "ERROR: cargo not found – install Rust from https://rustup.rs/"; exit 1; }
	@rustup target list --installed | grep -q "$(MUSL_TARGET)" || \
	  rustup target add $(MUSL_TARGET)

## ── Rust rendering library ───────────────────────────────────────────────────

$(WAIN_DIR):
	git clone --depth=1 https://github.com/opd-ai/wain.git $(WAIN_DIR)

rust-lib: deps $(WAIN_DIR) $(RUST_LIB)

$(RUST_LIB): $(WAIN_DIR)
	cargo build --release \
	  --manifest-path $(WAIN_DIR)/render-sys/Cargo.toml \
	  --target $(MUSL_TARGET)

## ── dl_find_object stub (needed for GCC 14 + musl) ──────────────────────────

dl-stub: $(DL_STUB_OBJ)

$(DL_STUB_OBJ): $(DL_STUB_SRC)
	$(CC) -c $< -o $@

## ── Go binary ────────────────────────────────────────────────────────────────

tox-gui: rust-lib dl-stub
	CC=$(CC) CGO_ENABLED=1 \
	  CGO_LDFLAGS="$(RUST_LIB) $(DL_STUB_OBJ) -ldl -lm -lpthread" \
	  CGO_LDFLAGS_ALLOW=".*" \
	  go build \
	    -ldflags="-s -w -extldflags=-static" \
	    -tags netgo \
	    -o $(OUT) $(PKG)
	@echo "Built: $(OUT)"

## ── Housekeeping ─────────────────────────────────────────────────────────────

clean:
	rm -f $(OUT)
