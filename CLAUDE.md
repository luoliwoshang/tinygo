# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

TinyGo is a Go compiler for microcontrollers, WebAssembly (WASM/WASI), and command-line tools. It uses LLVM for compilation and targets resource-constrained environments where the standard Go compiler would be too heavy.

## Key Architecture Components

### Core Directories
- **`compiler/`** - LLVM-based Go compiler implementation, handles Go AST to LLVM IR translation
- **`builder/`** - Build orchestration, manages the compilation pipeline and linking
- **`transform/`** - LLVM IR optimization passes specific to TinyGo's needs
- **`interp/`** - LLVM interpreter for compile-time evaluation of Go code
- **`cgo/`** - CGO implementation for C interoperability
- **`loader/`** - Go package loading and dependency resolution
- **`src/`** - TinyGo's standard library implementation (subset of Go's stdlib)
- **`targets/`** - Target configuration files for different platforms (JSON format)

### Target Support
TinyGo supports three main categories:
1. **Microcontrollers** - ARM Cortex-M, AVR, RISC-V, ESP32/ESP8266, RP2040/RP2350
2. **WebAssembly** - Browser (WASM) and server-side (WASI) targets
3. **Desktop OS** - Linux, macOS, Windows (for testing and development)

## Common Development Commands

### Building TinyGo
```bash
# Build TinyGo compiler (requires LLVM to be built first)
make tinygo

# Build LLVM dependencies (takes ~1 hour)
make llvm-source
make llvm-build
```

### Testing
```bash
# Run Go unit tests for TinyGo components
make test

# Test standard library packages on host platform
make tinygo-test-fast

# Test on WebAssembly targets
make tinygo-test-wasm
make tinygo-test-wasip1-fast

# Test on baremetal/microcontroller simulation
make tinygo-test-baremetal

# Smoke test - quick validation of basic functionality
make smoketest
```

### Code Quality
```bash
# Format Go source code
make fmt

# Lint source code
make lint

# Spell check documentation and comments
make spell
```

### Device Support Generation
```bash
# Generate device-specific code for all platforms
make gen-device

# Generate for specific platforms
make gen-device-avr      # AVR microcontrollers
make gen-device-stm32    # STM32 ARM microcontrollers
make gen-device-nrf      # Nordic nRF chips
make gen-device-esp      # ESP32/ESP8266
make gen-device-rp       # Raspberry Pi Pico/RP2040/RP2350
```

### Release Building
```bash
# Create release tarball
make release

# Clean build artifacts
make clean
```

## Testing Strategy

TinyGo has several test categories:
- **Unit tests** - Standard Go tests in `*_test.go` files
- **Integration tests** - Full compilation tests in `testdata/`
- **Corpus tests** - External package compatibility tests
- **Platform tests** - Target-specific compilation validation

When running tests, be aware that:
- WebAssembly tests require Node.js 18+
- Some standard library tests are platform-specific
- Baremetal tests run in QEMU simulation

## Build System Details

The build system uses GNU Make with these key concepts:
- **LLVM integration** - TinyGo statically links against LLVM libraries
- **Cross-compilation** - Single TinyGo binary can target multiple architectures
- **Submodules** - External libraries in `lib/` directory are Git submodules
- **Target files** - JSON configuration files define compiler flags and memory layouts

## Development Workflow

1. Make changes to Go source code
2. Run `make fmt` to format code
3. Run `make test` for unit tests
4. Run `make tinygo-test-fast` for integration tests
5. Run `make smoketest` for quick validation
6. Run `make lint` before committing

For device-specific changes, regenerate device files with `make gen-device-<platform>` after updating SVD files or device definitions.

## Important Notes

- TinyGo's standard library in `src/` is NOT the same as Go's stdlib - it's a custom implementation
- Target files in `targets/` define memory layouts, linker scripts, and compile flags
- The compiler produces LLVM IR, not Go assembly
- WebAssembly support includes both browser (WASM) and WASI targets
- Garbage collection can be configured per target (none, conservative, precise)