# TinyGo 内联汇编基本块约束调查过程

## 调查背景

在分析 TinyGo 的 `createInlineAsmFull` 函数时，发现了一个基本块约束检查：
```go
if r.Block() != registerMap.Block() {
    return llvm.Value{}, b.makeError(instr.Pos(), "register value map must be created in the same basic block")
}
```

本次调查旨在理解：
1. 这个约束存在的技术原因
2. 为什么 TinyGo 移除了指针类型支持
3. LLVM 内联汇编的底层要求

## 调查过程

### 第一阶段：追溯历史提交

**方法**：使用 git 命令查找相关提交
```bash
git log --oneline --grep="basic block" --grep="AsmFull" -i
git log --oneline -S "register value map must be created in the same basic block"
```

**发现**：
- 基本块约束从 AsmFull 功能首次实现时就存在
- 提交：`392bba83` (2018年10月15日)
- 作者：Ayke van Laethem
- 标题：`compiler: add support for parameters to inline assembly`

**结论**：这不是后来的 bug 修复，而是原始设计决策。

### 第二阶段：分析 LLVM 内联汇编要求

**发现的技术要求**：
```llvm
call RetType asm [sideeffect] "asm_template", "constraint_string" (Type1 %arg1, Type2 %arg2)
     ^^^^^^^^                 ^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^
     返回类型                  汇编模板      约束字符串         参数类型和值列表
```

**关键特点**：
- 所有信息必须在**编译时确定**
- 参数类型、约束字符串、参数数量都不能动态变化
- 这与普通函数调用的类型系统要求相同

### 第三阶段：跨基本块问题分析

**问题场景**：
```go
regs := make(map[string]interface{})
if condition {
    regs["value"] = uint32(42)    // 类型1: uint32
} else {
    regs["value"] = int32(42)     // 类型2: int32  
}
arm.AsmFull("mov {value}, #1", regs)  // 编译器无法确定类型
```

**技术难题**：
- 编译器无法静态确定最终的参数类型
- 无法生成稳定的 LLVM IR 函数签名
- 跨基本块的控制流让类型分析变得复杂

### 第四阶段：创建验证演示

**创建了演示项目**：`llvm_inline_asm_demo/`
- `good_example.go` - 同基本块的正确用法 ✅
- `bad_example.go` - 跨基本块的错误用法 ❌  
- `analyze_output.sh` - 自动化测试脚本

**验证结果**：
- 正确示例：编译成功，生成稳定 LLVM IR
- 错误示例：编译失败，报错 "register value map must be created in the same basic block"

### 第五阶段：指针支持移除调查

**发现关键提交**：`cad6a570` (2022年4月20日)
**标题**：`compiler: remove support for memory references in AsmFull`

**官方原因**：
```
Memory references (`*m` in LLVM IR inline assembly) need a pointer type
starting in LLVM 14. This is a bit inconvenient and requires a new API
in the go-llvm package.
```

**LLVM 技术变更**：
- LLVM 14 引入 opaque pointers 准备
- 内存约束 `*m` 需要额外的 `elementtype` 属性
- 示例：`call void asm "str $0, [$1]", "r,*m"(i32 %val, i32* elementtype(i32) %ptr)`

### 第六阶段：查找 LLVM 官方文档

**找到的关键文档**：
- **LLVM 代码评审**: https://reviews.llvm.org/D116531
- **标题**: `[LangRef] Require elementtype attribute for indirect inline asm operands`

**核心解释**：
```
Indirect inline asm operands may require the materialization of a memory 
access according to the pointer element type. As this will no longer be 
available with opaque pointers, we require it to be explicitly annotated 
using the elementtype attribute.
```

## 调查结论

### 1. **基本块约束的原因**
- **LLVM 技术限制**：内联汇编需要编译时类型确定
- **静态分析要求**：编译器必须能确定所有参数信息
- **类型安全保障**：避免生成无效的 LLVM IR

### 2. **指针支持移除的原因**
- **LLVM 14 破坏性变更**：需要 elementtype 属性支持
- **实现复杂度**：需要在 go-llvm 中添加新 API
- **维护负担**：AsmFull 本身就难以正确使用

### 3. **设计哲学**
- 选择**编译时安全** + **实现简化**的保守方案
- 优先**功能稳定性**而非功能完整性
- 推荐使用 **CGo** 作为复杂内联汇编的替代方案

## 核心发现

这个约束**不是 SSA 理论限制**，而是**LLVM 编译时类型确定的实际需求**。TinyGo 的实现是一个合理的工程权衡，在功能性和可维护性之间找到了平衡点。

## 相关链接

- **TinyGo 基本块约束实现**: https://github.com/tinygo-org/tinygo/commit/392bba8394ece9e8b90ee2d6e82590b1492bc62d
- **TinyGo 移除指针支持**: https://github.com/tinygo-org/tinygo/commit/cad6a57077c7887025ba532a03f41d6ad78daa72
- **LLVM elementtype 要求**: https://reviews.llvm.org/D116531