# TinyGo 内联汇编机制深度解析

## 📖 概述

本项目深度分析了TinyGo编译器如何处理内联汇编调用，从Go源代码到最终的LLVM IR的完整转换过程。

## 🗂️ 文件结构

```
tinygo-inline-assembly/
├── README.md              # 本文档
├── examples/              # 示例代码
│   ├── wfi_example.go     # WFI指令示例
│   ├── basic_asm.go       # 基础内联汇编
│   └── system_reset.go    # 系统复位示例
├── reproduction.md        # 完整复现流程
└── analysis/              # 分析结果
    ├── llvm_ir_output.txt # LLVM IR输出示例
    └── compiler_flow.md   # 编译器处理流程
```

## 🎯 核心发现

### 1. 内联汇编转换机制

TinyGo将 `device.Asm("instruction")` 转换为LLVM内联汇编：

```llvm
call void asm sideeffect "instruction", ""()
```

### 2. 关键组件

- **编译器入口**: `compiler/compiler.go:1865`
- **处理函数**: `compiler/inlineasm.go:24-30`
- **LLVM集成**: 使用 `llvm.InlineAsm()` 直接嵌入

### 3. 实际IR示例

```llvm
; ARM WFI指令
call void asm sideeffect "wfi", ""(), !dbg !21227

; ARM WFE指令  
call void asm sideeffect "wfe", ""(), !dbg !57898

; 带寄存器约束
call void asm sideeffect "", "r"(ptr %2), !dbg !57583
```

## 🔧 支持的架构

- `device.Asm` (通用)
- `device/arm.Asm` (ARM)
- `device/arm64.Asm` (ARM64)
- `device/avr.Asm` (AVR)
- `device/riscv.Asm` (RISC-V)

## 🚀 快速开始

1. 查看 [复现流程](reproduction.md)
2. 运行示例代码查看效果
3. 分析生成的LLVM IR

## 📚 相关资源

- TinyGo项目: https://tinygo.org/
- LLVM内联汇编文档: https://llvm.org/docs/LangRef.html#inline-assembler-expressions
- ARM指令集参考: https://developer.arm.com/documentation

## 🔍 技术要点

- **零开销**: 直接内联，无函数调用开销
- **防优化**: `sideeffect`标志防止被优化掉
- **类型安全**: 编译时检查和验证
- **架构感知**: 不同目标架构有对应实现

---
*本文档基于TinyGo源码分析，版本信息见具体提交记录*