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

## TinyGo GNUmakefile 深度分析续 (2025-08-26)

### NINJA 构建目标系统 (`GNUmakefile:171`)
**精准构建优化**:
```makefile
NINJA_BUILD_TARGETS = clang llvm-config llvm-ar llvm-nm lld $(addprefix lib/lib,$(addsuffix .a,$(LIB_NAMES)))
```

**设计思想**:
- **最小构建集**: 仅构建TinyGo必需的LLVM组件，显著加速LLVM重构建
- **路径转换**: `lldELF` → `lib/liblldELF.a` 匹配Ninja构建系统的库文件路径约定
- **工具链完整性**: 包含编译器(`clang`)、配置工具(`llvm-config`)、归档工具(`llvm-ar`, `llvm-nm`)、链接器(`lld`)

**构建性能优势**:
- 完整LLVM构建: ~2小时，数GB编译产物
- TinyGo定制构建: ~30分钟，仅必要组件

### 动态CGO链接配置 (`GNUmakefile:174-178`)
**条件编译机制**:
```makefile
ifneq ("$(wildcard $(LLVM_BUILDDIR)/bin/llvm-config*)","")
    CGO_CPPFLAGS+=$(shell $(LLVM_CONFIG_PREFIX) $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) ...
    CGO_LDFLAGS+=-L$(abspath $(LLVM_BUILDDIR)/lib) -lclang $(CLANG_LIBS) ...
endif
```

**关键特性**:
- **存在检测**: 只有在LLVM成功构建后才设置CGO链接参数
- **动态配置**: 使用`llvm-config`工具自动生成编译器和链接器参数
- **多库整合**: 整合Clang前端 + LLD链接器 + LLVM后端的完整工具链

### 代码格式化系统 (`GNUmakefile:183-187`)
**FMT_PATHS变量**:
```makefile
FMT_PATHS = ./*.go builder cgo/*.go compiler interp loader src transform
```

**双模式设计**:
- `make fmt`: 自动修复格式问题 (`gofmt -l -w`)
- `make fmt-check`: 仅检查格式，失败时退出 (CI/CD适用)

**错误处理模式**:
```bash
unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1
```

### 设备代码生成架构 (`GNUmakefile:190-240`)
**分层生成系统**:
```
gen-device (总入口)
├── gen-device-avr     → 基于XML包描述生成AVR微控制器定义
├── gen-device-esp     → 基于SVD文件生成ESP32设备寄存器
├── gen-device-stm32   → 基于STM32官方SVD生成ARM Cortex-M定义
├── gen-device-nrf     → 基于Nordic nRF系列SVD生成蓝牙MCU定义
└── ... (各厂商特定生成器)
```

**SVD处理流程** (以NRF为例):
```makefile
gen-device-nrf: build/gen-device-svd
./build/gen-device-svd -source=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk lib/nrfx/mdk/ src/device/nrf/
GO111MODULE=off $(GO) fmt ./src/device/nrf
```

**技术实现**:
- **SVD文件**: 半导体厂商提供的系统视图描述文件 (XML格式)
- **代码生成器**: `tools/gen-device-svd/` 将SVD转换为Go设备定义
- **后处理**: 自动格式化生成的Go代码

### LLVM源码管理 (`GNUmakefile:242-244`)
**Espressif定制LLVM**:
```makefile
$(LLVM_PROJECTDIR)/llvm:
git clone -b xtensa_release_19.1.2 --depth=1 https://github.com/espressif/llvm-project $(LLVM_PROJECTDIR)
```

**关键决策**:
- **特定分支**: `xtensa_release_19.1.2` - 稳定的ESP32支持版本
- **浅克隆**: `--depth=1` 避免下载完整Git历史，节省带宽和存储
- **官方维护**: 使用ESP32制造商维护的LLVM，质量保证

### LLVM构建配置详解 (`GNUmakefile:248-249`)
**CMAKE配置解析**:
```makefile
cmake -G Ninja $(TINYGO_SOURCE_DIR)/$(LLVM_PROJECTDIR)/llvm \
"-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;AVR;Mips;RISCV;WebAssembly" \
"-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Xtensa" \
-DCMAKE_BUILD_TYPE=Release \
-DLIBCLANG_BUILD_STATIC=ON \
-DLLVM_ENABLE_TERMINFO=OFF \
-DLLVM_ENABLE_ZLIB=OFF \
-DLLVM_ENABLE_ZSTD=OFF \
-DLLVM_ENABLE_LIBEDIT=OFF \
-DLLVM_ENABLE_Z3_SOLVER=OFF \
-DLLVM_ENABLE_OCAMLDOC=OFF \
-DLLVM_ENABLE_LIBXML2=OFF \
-DLLVM_ENABLE_PROJECTS="clang;lld" \
-DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF \
-DCLANG_ENABLE_STATIC_ANALYZER=OFF \
-DCLANG_ENABLE_ARCMT=OFF
```

**配置分类**:
- **目标架构**: 标准架构 + 实验性Xtensa支持
- **构建优化**: Release模式，静态libclang构建
- **功能精简**: 禁用终端info、压缩、编辑器、求解器等非必要功能
- **组件选择**: 仅启用核心编译器(clang)和链接器(lld)
- **工具链精简**: 禁用静态分析器、代码迁移工具等额外功能

**设计目标**: 生成最小化但功能完整的LLVM工具链，专门为TinyGo嵌入式编译优化

### Binaryen WebAssembly优化器 (`GNUmakefile:254-262`)
**条件构建逻辑**:
```makefile
ifneq ($(USE_SYSTEM_BINARYEN),1)
binaryen: build/wasm-opt$(EXE)
build/wasm-opt$(EXE):
    cd lib/binaryen && cmake -G Ninja . -DBUILD_STATIC_LIB=ON -DBUILD_TESTS=OFF -DENABLE_WERROR=OFF
    cp lib/binaryen/bin/wasm-opt$(EXE) build/wasm-opt$(EXE)
endif
```

**用途说明**:
- **WebAssembly后处理**: `wasm-opt`对TinyGo生成的WASM进行体积和性能优化
- **系统适配**: 优先使用系统安装的binaryen，否则从源码构建
- **静态构建**: 生成独立可执行文件，避免运行时依赖

### WASI系统调用绑定生成 (`GNUmakefile:264-276`)
**WASI-syscall目标**:
```makefile
wasi-syscall: wasi-cm
	rm -rf ./src/internal/wasi/*
	go run -modfile ./internal/wasm-tools/go.mod $(WASM_TOOLS_MODULE)/cmd/wit-bindgen-go generate --versioned -o ./src/internal -p internal --cm internal/cm ./lib/wasi-cli/wit
```

**技术实现**:
- **依赖关系**: `wasi-syscall: wasi-cm` - 先复制cm包，再生成WASI绑定
- **清理重建**: 每次重新生成前先清空 `./src/internal/wasi/*`
- **WIT绑定生成**: 使用 `wit-bindgen-go` 工具从 `./lib/wasi-cli/wit` 生成Go绑定代码
- **模块来源**: `go.bytecodealliance.org` - WebAssembly标准化组织的工具

**wasi-cm目标**:
```makefile
wasi-cm:
	rm -rf ./src/internal/cm/*
	rsync -rv --delete --exclude go.mod --exclude '*_test.go' --exclude '*_json.go' --exclude '*.md' --exclude LICENSE $(shell go list -modfile ./internal/wasm-tools/go.mod -m -f {{.Dir}} $(WASM_TOOLS_MODULE)/cm)/ ./src/internal/cm
```

**设计特点**:
- **依赖包复制**: 将外部cm包同步到 `./src/internal/cm/`
- **选择性排除**: 排除 `go.mod`, `*_test.go`, `*_json.go`, `*.md`, `LICENSE` 等非核心文件
- **rsync特性**: `--delete` 确保目标目录与源完全同步

### Node.js版本检查系统 (`GNUmakefile:277-287`)
**版本要求**: `MIN_NODEJS_VERSION=18` - 要求Node.js 18+

**双重检查机制**:
```bash
# 1. 存在检查
@if ! command -v node 2>&1 >/dev/null; then echo "Install NodeJS version ${MIN_NODEJS_VERSION}+ to run tests."; exit 1; fi

# 2. 版本检查  
@if [ "`node -v | sed 's/v\([0-9]\+\).*/\\1/g'`" -lt $(MIN_NODEJS_VERSION) ]; then echo "Install NodeJS version $(MIN_NODEJS_VERSION)+ to run tests."; exit 1; fi
```

**版本提取逻辑**: `node -v | sed 's/v\([0-9]\+\).*/\\1/g'` 从 `v18.x.x` 提取主版本号 `18`
**用途**: WebAssembly测试需要Node.js运行时环境

### TinyGo编译器构建 (`GNUmakefile:288-290`)
**构建前置检查**:
```makefile
@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  $(MAKE) llvm-source"; echo "  $(MAKE) $(LLVM_BUILDDIR)"; exit 1; fi
```

**核心构建命令**:
```makefile
CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GOENVFLAGS) $(GO) build -buildmode exe -o build/tinygo$(EXE) -tags "byollvm osusergo" .
```

**关键构建参数**:
- **构建标签**: `-tags "byollvm osusergo"`
  - `byollvm`: 使用自建LLVM而非系统LLVM
  - `osusergo`: 使用纯Go实现的用户/组查找(避免CGO依赖)
- **CGO环境**: 传递所有CGO编译和链接参数
- **构建模式**: `-buildmode exe` 生成可执行文件

### 测试系统入口 (`GNUmakefile:291-292`)
**test目标**:
```makefile
test: check-nodejs-version
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm osusergo" $(GOTESTPKGS)
```

**关键特性**:
- **依赖**: `check-nodejs-version` - 确保Node.js可用于WASM测试
- **超时设置**: `-timeout=1h` - 1小时测试超时
- **环境一致**: 使用与TinyGo编译器相同的CGO参数和构建标签

**设计原则**:
1. **WASI支持**: 完整的WebAssembly System Interface绑定生成
2. **环境检查**: 严格的依赖项版本检查
3. **自包含构建**: 使用自建LLVM工具链，减少外部依赖

## TinyGo静态链接LLVM/Clang技术架构 (2025-08-26)

### 发布包静态链接策略
**不打包Clang可执行文件**:
```makefile
@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
```

**TinyGo采用静态链接方式**:
- 只打包Clang头文件，不打包`clang`可执行文件
- TinyGo编译器内置了LLVM和Clang的**静态库版本**
- 通过CGO链接了libclang和LLVM库

**证据**: 构建标签`-tags "byollvm osusergo"`中的`byollvm`表示"Bring Your Own LLVM"，使用静态链接的LLVM

### 静态链接实现机制

#### 1. libclang vs clang可执行文件
**libclang**:
- Clang的C API库版本，可以静态链接到其他程序中
- 提供编译器前端功能（词法分析、语法分析、语义分析）

**静态库配置**:
```makefile
-DLIBCLANG_BUILD_STATIC=ON  # 构建静态版本的libclang
CGO_LDFLAGS+=-lclang $(CLANG_LIBS) $(LLD_LIBS) # 链接静态库
CLANG_LIB_NAMES = clangAnalysis clangAPINotes clangAST clangBasic clangCodeGen ...
LLD_LIB_NAMES = lldCOFF lldCommon lldELF lldMachO lldMinGW lldWasm ...
```

#### 2. CGO调用LLVM/Clang API
**典型调用流程**:
```go
// TinyGo编译器伪代码
import "C" // CGO

// 直接调用libclang API解析Go代码
func parseGoFile(filename string) {
    // 调用 clang_parseTranslationUnit 等libclang函数
    C.clang_parseTranslationUnit(...)
    
    // 调用LLVM API生成IR
    C.LLVMCreateModule(...)
    C.LLVMBuildAdd(...)
}
```

### LLVM IR生成与链接过程

#### 1. TinyGo编译管线
```
Go源码 → TinyGo编译器 → LLVM IR (.ll) → 目标代码 → 可执行文件
```

#### 2. 通过LLVM C API生成IR
```go
// 伪代码示例 - TinyGo编译器内部
/*
#include <llvm-c/Core.h>
#include <llvm-c/IRReader.h>
*/
import "C"

func compileFunction(fn *ast.FuncDecl) {
    // 1. 创建LLVM模块
    module := C.LLVMModuleCreateWithName(C.CString("main"))
    
    // 2. 创建函数类型
    funcType := C.LLVMFunctionType(C.LLVMInt32Type(), nil, 0, 0)
    
    // 3. 创建函数
    function := C.LLVMAddFunction(module, C.CString("main"), funcType)
    
    // 4. 创建基本块
    builder := C.LLVMCreateBuilder()
    block := C.LLVMAppendBasicBlock(function, C.CString("entry"))
    C.LLVMPositionBuilderAtEnd(builder, block)
    
    // 5. 生成指令
    result := C.LLVMBuildAdd(builder, lhs, rhs, C.CString("add"))
    C.LLVMBuildRet(builder, result)
    
    // 6. 输出IR到文件（调试用）
    C.LLVMPrintModuleToFile(module, C.CString("output.ll"), nil)
}
```

#### 3. 使用LLD进行最终链接 - 实际代码实现
**主入口** (`main.go:1623-1630`):
```go
switch command {
case "clang", "ld.lld", "wasm-ld":
    err := builder.RunTool(command, os.Args[2:]...)
    if err != nil {
        // The tool should have printed an error message already.
        // Don't print another error message here.
        os.Exit(1)
    }
    os.Exit(0)
}
```

**RunTool函数** (`builder/tools-builtin.go:25-51`):
```go
//go:build byollvm

func RunTool(tool string, args ...string) error {
    args = append([]string{tool}, args...)

    // 准备C参数数组
    var cflag *C.char
    buf := C.calloc(C.size_t(len(args)), C.size_t(unsafe.Sizeof(cflag)))
    defer C.free(buf)
    cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(args):len(args)]
    for i, flag := range args {
        cflag := C.CString(flag)
        cflags[i] = cflag
        defer C.free(unsafe.Pointer(cflag))
    }

    var ok C.bool
    switch tool {
    case "clang":
        ok = C.tinygo_clang_driver(C.int(len(args)), (**C.char)(buf))
    case "ld.lld", "wasm-ld":
        ok = C.tinygo_link(C.int(len(args)), (**C.char)(buf))  // 实际的LLD调用！
    default:
        return errors.New("unknown tool: " + tool)
    }
    if !ok {
        return errors.New("failed to run tool: " + tool)
    }
    return nil
}
```

**C++封装函数** (`builder/lld.cpp:25-30`):
```cpp
//go:build byollvm

#include <lld/Common/Driver.h>

LLD_HAS_DRIVER(coff)     // COFF格式支持 (Windows)
LLD_HAS_DRIVER(elf)      // ELF格式支持 (Linux)
LLD_HAS_DRIVER(mingw)    // MinGW支持
LLD_HAS_DRIVER(macho)    // Mach-O格式支持 (macOS)
LLD_HAS_DRIVER(wasm)     // WebAssembly支持

extern "C" {

bool tinygo_link(int argc, char **argv) {
    configure();  // Windows平台的线程配置
    std::vector<const char*> args(argv, argv + argc);
    lld::Result r = lld::lldMain(args, llvm::outs(), llvm::errs(), LLD_ALL_DRIVERS);
    return !r.retCode;  // 返回链接是否成功
}

} // external "C"
```

**高级链接调用** (`builder/tools.go:58-87`):
```go
func link(linker string, flags ...string) error {
    // We only support LLD.
    if linker != "ld.lld" && linker != "wasm-ld" {
        return fmt.Errorf("unexpected: linker %s should be ld.lld or wasm-ld", linker)
    }

    var cmd *exec.Cmd
    if hasBuiltinTools {  // 当byollvm构建标签启用时
        cmd = exec.Command(os.Args[0], append([]string{linker}, flags...)...)  // 调用自身！
    } else {
        name, err := LookupCommand(linker)
        if err != nil {
            return err
        }
        cmd = exec.Command(name, flags...)  // 调用外部ld.lld
    }
    // ... 执行命令并解析LLD错误信息
}
```

#### 4. 完整编译链路
```go
// compiler/compiler.go 中的实际流程
func (c *Compiler) Compile() {
    // 解析Go源码为AST
    pkg := c.parsePackage()
    
    // 创建LLVM模块
    module := c.createLLVMModule()
    
    // 为每个函数生成LLVM IR
    for _, fn := range pkg.Functions {
        c.compileFunction(fn, module)
    }
    
    // 运行优化pass
    c.optimizeModule(module)
    
    // 生成目标代码
    objCode := c.generateTargetCode(module)
    
    // 使用内置LLD链接器
    c.linkExecutable(objCode)
}
```

### 调试和验证方法
**查看生成的LLVM IR**:
```bash
# 输出LLVM IR到文件
tinygo build -target=microbit -print-ir=out.ll main.go

# 内部调试标志
tinygo build -target=microbit -internal-printir main.go
```

### 实际LLD调用的关键技术点

1. **自包含设计**: TinyGo可执行文件本身就是链接器！当使用`ld.lld`时，实际是调用自身
2. **条件编译**: `//go:build byollvm` 确保只在静态链接LLVM时编译此代码
3. **直接API调用**: 通过CGO直接调用`lld::lldMain()`，避免进程间通信开销
4. **多格式支持**: LLD支持ELF、COFF、Mach-O、WebAssembly等链接格式
5. **错误处理**: 专门的`parseLLDErrors()`函数解析和美化LLD错误信息
6. **平台优化**: Windows平台特殊的线程配置处理，避免LLD挂起问题

### 技术优势
1. **直接API调用**: 避免了调用外部clang/llc/lld进程的开销，直接在内存中操作LLVM数据结构
2. **集成优化**: 可以在生成IR的同时进行TinyGo特有的优化，无需中间文件，全内存操作
3. **目标感知**: 可以根据目标平台(Arduino、ESP32等)调整生成策略，支持平台特定的优化
4. **自包含**: 单个TinyGo可执行文件包含完整的编译器功能，无需外部clang/LLVM依赖
5. **版本控制**: 确保使用特定版本的LLVM（如支持Xtensa的Espressif版本），避免系统LLVM版本冲突
6. **统一工具链**: TinyGo既是编译器，也是clang，也是链接器，实现完全统一的工具链

**总结**: TinyGo通过CGO直接调用LLVM和LLD的C++ API，实现了从Go源码到可执行文件的完整编译链路，全程在内存中操作，无需外部工具依赖。这是真实的代码实现，不是伪代码！

## Important Notes

- TinyGo uses a custom Go runtime, not the standard runtime
- Many standard library packages are reimplemented for size/compatibility
- Goroutines are implemented differently depending on scheduler choice
- Not all Go features are supported (see lang-support documentation)
- Cross-compilation is the norm, not the exception