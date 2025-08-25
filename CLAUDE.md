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

## TinyGo ESP32/Xtensa 编译机制分析 (2025-08-25)

### 核心发现
TinyGo 通过使用 Espressif 官方维护的 LLVM 分支来支持 ESP32/Xtensa 架构编译，巧妙地解决了标准 LLVM/Clang 不支持 Xtensa 的问题。

### 关键文件和配置
- **LLVM 源码**: `GNUmakefile:244` - `git clone -b xtensa_release_19.1.2 https://github.com/espressif/llvm-project`
- **目标配置**: `targets/xtensa.json:2` - `"llvm-target": "xtensa"`
- **实验性目标**: `GNUmakefile:249` - `"-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Xtensa"`
- **ESP32 配置**: `targets/esp32.json:2` - `"inherits": ["xtensa"]`

### 构建流程
1. **获取源码**: `make llvm-source` → 下载 Espressif 的 LLVM 分支 (xtensa_release_19.1.2)
2. **构建 LLVM**: `make llvm-build` → 编译包含 Xtensa 后端的自定义 LLVM
3. **目标编译**: TinyGo 使用自建 LLVM 的 Xtensa 后端编译 ESP32 代码
4. **固件处理**: `builder/esp.go` 将 ELF 转换为 ESP32 专用固件格式

### 技术细节
- **不依赖外部工具链**: 无需安装 ESP-IDF 或 xtensa-esp32-elf-gcc
- **链接器**: 使用 LLVM 的 LLD (`"linker": "ld.lld"`) 直接支持 Xtensa
- **Go 编译伪装**: 编译时设置 `"goarch": "arm"` 让 Go 编译器正常工作
- **实验性支持**: Xtensa 在 LLVM 中标记为实验性目标，需单独启用

### 关键优势
- **官方质量保证**: 使用 Espressif (ESP32 制造商) 维护的稳定分支
- **统一工具链**: 与其他 TinyGo 目标使用相同的编译管线
- **简化部署**: 开发者无需配置复杂的交叉编译环境

### 构建系统分析
**为什么使用 GNUmakefile 而非纯 Go 构建**:
- **官方明确**: `BUILDING.md:33` - "The static build of TinyGo is driven by GNUmakefile"
- **复杂依赖**: 需要构建 C++ 库 (LLVM、libclang、LLD)
- **交叉编译**: 支持多架构的编译器工具链构建
- **GNU Make 特性**: 使用条件编译、并行构建等高级功能

**构建入口层次**:
```
GNUmakefile (官方构建入口)
├── make llvm-source  → 下载 Espressif LLVM 源码
├── make llvm-build   → 构建包含 Xtensa 后端的 LLVM
├── make tinygo       → 构建 TinyGo 编译器
└── make gen-device   → 生成微控制器设备定义
```

**文件系统证据**:
- 根目录仅有 `GNUmakefile`，无标准 `Makefile`
- 虽有 `go.mod` 但 TinyGo 非纯 Go 项目
- 需要构建自定义 LLVM 工具链，超出 Go 构建系统能力范围

## TinyGo GNUmakefile 详细分析 (2025-08-25)

### 工具链自动发现机制
**LLVM 工具优先级**:
```makefile
findLLVMTool = $(call detect,$(1),$(abspath llvm-build/bin/$(1)) $(foreach ver,$(LLVM_VERSIONS),$(call toolSearchPathsVersion,$(1),$(ver))) $(1))
```

**查找顺序** (以 clang 为例):
1. **自建 LLVM** - `llvm-build/bin/clang` (最高优先级，包含 Xtensa 后端)
2. **版本化工具** - `clang-19`, `clang-18`, ..., `clang-15` 
3. **系统默认** - `clang` (最低优先级)

**关键发现**:
- **Homebrew LLVM**: 支持 `AArch64 ARM AVR WebAssembly X86` 等 ❌ **无 Xtensa**
- **自建 LLVM**: 支持 `X86 ARM AArch64 AVR Mips RISCV WebAssembly Xtensa` ✅ **有 Xtensa**
- **工具来源**: `llvm-ar`, `llvm-nm` 等都来自 `make llvm-build`，非系统工具

### 构建优化配置
**ccache 编译缓存**:
- 自动检测系统 ccache: `LLVM_CCACHE_BUILD=$(if $(shell command -v ccache),ON,OFF)`
- 作用: 显著加速 LLVM 重复构建 (2小时 → 30分钟)

**调试和优化选项**:
- `ASSERT=1` → `LLVM_ENABLE_ASSERTIONS=ON` (开启 LLVM 内部断言)
- `ASAN=1` → `LLVM_USE_SANITIZER=Address` (内存错误检测)

**静态链接构建**:
- `STATIC=1` → 完全静态链接的 TinyGo 二进制
- **限制**: 仅支持 musl libc (Alpine Linux)，不支持 glibc (Ubuntu/CentOS)
- **用途**: Docker 发布构建，生成可移植的无依赖二进制
- **栈大小**: musl 下需设置 1MB 栈 (`-Wl,-z,stack-size=1048576`)

### Make 变量来源机制
**变量优先级** (高到低):
1. **命令行参数**: `make tinygo STATIC=1`
2. **环境变量**: `export STATIC=1; make tinygo`  
3. **Makefile 定义**: `STATIC ?= 0` (GNUmakefile 中未定义默认值)

**功能开关设计模式**:
- 默认行为: 动态链接构建 (适合开发)
- 可选行为: `STATIC=1` 静态链接构建 (适合分发)

### Make .PHONY 机制解析
**Make 的工作原理**:
- Make 通过**文件时间戳比较**来决定是否执行目标
- 如果目标文件存在且比依赖文件新，则跳过执行
- 如果目标文件不存在或比依赖文件旧，则重新执行

**问题场景**:
```makefile
clean:
    rm *.o
```

如果目录中存在名为 `clean` 的文件:
```bash
touch clean    # 意外创建了 clean 文件
make clean     # Make 检查发现 clean 文件已存在
                # 结果: rm *.o 不会执行！
```

**解决方案**:
```makefile
.PHONY: clean
clean:
    rm *.o
```

**`.PHONY` 声明的作用**:
- 告诉 Make 这些目标**不生成同名文件**
- 强制 Make **每次都执行**这些目标，忽略文件时间戳检查
- 确保 `clean`、`test`、`fmt` 等维护命令总是能正确执行

**TinyGo 中的伪目标** (`GNUmakefile:114`):
```makefile
.PHONY: all tinygo test $(LLVM_BUILDDIR) llvm-source clean fmt gen-device gen-device-nrf gen-device-nxp gen-device-avr gen-device-rp
```
包括构建目标、工具命令和各种微控制器代码生成任务

### LLVM 组件配置 (`GNUmakefile:116`)
**LLVM_COMPONENTS 变量定义**:
```makefile
LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines coverage debuginfodwarf debuginfopdb executionengine frontenddriver frontendhlsl frontendopenmp instrumentation interpreter ipo irreader libdriver linker lto mc mcjit objcarcopts option profiledata scalaropts support target windowsdriver windowsmanifest
```

**组件分类说明**:
- **核心组件**: `core`, `support`, `target` - LLVM基础功能
- **代码处理**: `asmparser`, `asmprinter`, `codegen`, `bitreader`, `bitwriter` - 汇编和代码生成
- **优化功能**: `ipo`, `scalaropts`, `lto` - 过程间优化、标量优化、链接时优化
- **架构支持**: `all-targets` - 支持所有目标架构（ARM、x86、RISC-V、Xtensa等）
- **调试支持**: `debuginfodwarf`, `debuginfopdb` - DWARF和PDB调试信息
- **执行引擎**: `executionengine`, `mcjit`, `interpreter` - JIT编译和解释执行

**用途**: 告诉LLVM构建系统需要编译哪些功能模块，减少构建时间和最终二进制大小

### 平台特定配置 (`GNUmakefile:118-146`)
**跨平台构建支持**:

**Windows 配置**:
```makefile
ifeq ($(OS),Windows_NT)
    EXE = .exe
    LLVM_OPTION += -DLLVM_ENABLE_PIC=OFF    # 禁用位置无关代码
    CGO_LDFLAGS += -static -static-libgcc -static-libstdc++
```
- 可执行文件需要 `.exe` 后缀
- 禁用PIC以兼容libclang
- 强制静态链接C++运行时

**macOS 配置**:
```makefile
else ifeq ($(uname),Darwin)
    MD5SUM ?= md5              # macOS使用md5而非md5sum
    CGO_LDFLAGS += -lxar       # 链接XAR归档库（Apple格式）
    USE_SYSTEM_BINARYEN ?= 1   # 优先使用系统WebAssembly工具
```

**FreeBSD/Linux 配置**:
```makefile
else
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group    # 处理静态库循环依赖
endif
```

**设计意图**: 确保TinyGo能在Windows、macOS、FreeBSD、Linux上正确构建，处理各平台的工具链差异

### 静态库链接配置 (`GNUmakefile:149-163`)

**MD5工具默认设置**:
```makefile
MD5SUM ?= md5sum    # 默认值，如果前面平台配置中未设置
```
- **macOS**: 已设置为 `MD5SUM = md5`
- **其他系统**: 使用默认的 `MD5SUM = md5sum`

**Clang 静态库配置**:
```makefile
CLANG_LIB_NAMES = clangAnalysis clangAPINotes clangAST ... (完整的Clang库列表)
CLANG_LIBS = $(START_GROUP) $(addprefix -l,$(CLANG_LIB_NAMES)) $(END_GROUP) -lstdc++
```

**Make函数解析**:
- `$(addprefix -l,$(CLANG_LIB_NAMES))`: 给每个库名添加 `-l` 前缀
  - `clangAnalysis` → `-lclangAnalysis`
  - `clangAST` → `-lclangAST`
- `$(START_GROUP)` 和 `$(END_GROUP)`: 处理循环依赖
  - 展开为 `-Wl,--start-group ... -Wl,--end-group`

**LLD 链接器库配置**:
```makefile
LLD_LIB_NAMES = lldCOFF lldCommon lldELF lldMachO lldMinGW lldWasm
LLD_LIBS = $(START_GROUP) $(addprefix -l,$(LLD_LIB_NAMES)) $(END_GROUP)
```

**额外依赖库**:
```makefile
EXTRA_LIB_NAMES = LLVMInterpreter LLVMMCA LLVMRISCVTargetMCA LLVMX86TargetMCA
LIB_NAMES = clang $(CLANG_LIB_NAMES) $(LLD_LIB_NAMES) $(EXTRA_LIB_NAMES)
```

**最终效果**: `CLANG_LIBS` 变量展开为完整的静态链接参数:
```
-Wl,--start-group -lclangAnalysis -lclangAPINotes -lclangAST ... -Wl,--end-group -lstdc++
```

## Important Notes

- TinyGo uses a custom Go runtime, not the standard runtime
- Many standard library packages are reimplemented for size/compatibility
- Goroutines are implemented differently depending on scheduler choice
- Not all Go features are supported (see lang-support documentation)
- Cross-compilation is the norm, not the exception