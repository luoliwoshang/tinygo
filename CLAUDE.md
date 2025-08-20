# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building TinyGo
- **Build TinyGo**: `make tinygo` or `make` - Build the TinyGo compiler
- **Build with LLVM**: First run `make llvm-source` then `make llvm-build` for static linking
- **Quick build**: `go build -tags byollvm .` - For development (requires system LLVM)

### Testing
- **Run Go tests**: `make test` - Run TinyGo's own test suite
- **Run TinyGo tests**: `make tinygo-test` - Test standard library packages with TinyGo
- **Fast tests**: `make tinygo-test-fast` - Run subset of standard library tests
- **WASI tests**: `make tinygo-test-wasi` - Test WebAssembly System Interface functionality
- **Smoke tests**: `make smoketest` - Basic functionality verification

### Code Quality
- **Format code**: `make fmt` - Reformat all Go source files
- **Check formatting**: `make fmt-check` - Verify code formatting
- **Lint code**: `make lint` - Run revive linter
- **Spell check**: `make spell` - Check spelling in source files

### Device Code Generation
- **Generate all devices**: `make gen-device` - Generate microcontroller-specific sources
- **Generate specific devices**: `make gen-device-avr`, `make gen-device-stm32`, etc.

### Building Examples and Testing Targets
- **Build for target**: `./build/tinygo build -target=<target> examples/<example>`
- **Flash to device**: `./build/tinygo flash -target=<target> examples/<example>`
- **Common targets**: `arduino`, `microbit`, `pico`, `wasm`, `wasip1`

## Repository Architecture

### Core Components

**Compiler (`/compiler/`)**
- `compiler.go` - Main compilation logic
- `inlineasm.go` - Inline assembly handling (critical for embedded targets)
- `interface.go` - Go interface implementation
- `gc.go` - Garbage collection integration
- `llvm.go` - LLVM IR generation and optimization

**Builder (`/builder/`)**
- `build.go` - Build orchestration and target-specific compilation
- `config.go` - Build configuration management
- `builtins.go` - Built-in function implementations

**Runtime (`/src/runtime/`)**
- Cross-platform runtime implementations for different targets
- `runtime_*.go` - Target-specific runtime code (ARM, AVR, WASM, etc.)
- `gc_*.go` - Garbage collector implementations (none, leaking, conservative, etc.)
- `scheduler_*.go` - Scheduler variants (none, tasks, threads)

**Machine Abstraction (`/src/machine/`)**
- Hardware abstraction layer for microcontrollers
- `machine_*.go` - MCU-specific GPIO, SPI, I2C, UART implementations
- `board_*.go` - Board-specific pin mappings and configurations

**Device Definitions (`/src/device/`)**
- Generated register definitions for microcontrollers
- Organized by architecture: `arm/`, `avr/`, `riscv/`, etc.
- Generated from manufacturer SVD files

### Target System
- **Target specs** (`/targets/`): JSON files defining compilation targets
- **Linker scripts**: `.ld` files for memory layout
- **Startup code**: Assembly files for MCU initialization

### Testing Structure
- Unit tests alongside source code (`*_test.go`)
- Integration tests in `/testdata/`
- Corpus tests for external package compatibility
- Target-specific smoke tests in Makefile

## Code Patterns and Conventions

### Target-Specific Code
- Use build tags for target-specific implementations
- Follow naming pattern: `filename_targetname.go`
- Example: `runtime_cortexm.go`, `machine_atmega.go`

### LLVM Integration
- TinyGo generates LLVM IR, not native assembly
- Heavy use of LLVM's optimization passes
- Custom passes for TinyGo-specific optimizations in `/transform/`

### Garbage Collection
- Multiple GC strategies depending on target constraints
- Baremetal targets often use conservative or no GC
- Desktop targets can use precise or Boehm GC

### Cross-Compilation
- Extensive cross-compilation support for embedded targets
- Uses custom GOROOT with TinyGo-specific runtime
- Target-specific toolchain integration (OpenOCD, debuggers, etc.)

## Development Tips

### Working with Inline Assembly
- TinyGo transforms `device.Asm("instruction")` to LLVM inline assembly
- Implementation in `compiler/inlineasm.go:24-30`
- ARM implementation in `src/device/arm/arm.go:44`

### Adding New Targets
1. Create target JSON spec in `/targets/`
2. Add board definition in `/src/machine/board_*.go`
3. Include linker script and startup code if needed
4. Test with smoke tests

### Debugging Build Issues
- Use `-internal-printir` flag to examine LLVM IR
- Use `-x` flag to see executed commands
- Check target JSON spec for configuration issues

### Performance Considerations
- TinyGo optimizes for size over speed by default
- Use `-opt` flag to control optimization level
- Stack usage analysis available with `-print-stacks`

## Important Notes

- TinyGo uses a custom Go runtime, not the standard runtime
- Many standard library packages are reimplemented for size/compatibility
- Goroutines are implemented differently depending on scheduler choice
- Not all Go features are supported (see lang-support documentation)
- Cross-compilation is the norm, not the exception