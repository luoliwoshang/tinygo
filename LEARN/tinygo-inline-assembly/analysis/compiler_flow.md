# TinyGo 编译器内联汇编处理流程

## 🔄 完整编译流程

```
Go源码 → 词法分析 → 语法分析 → SSA → LLVM IR → 机器码
  ↓         ↓         ↓       ↓       ↓         ↓
device.   tokens   AST    特殊   inline   目标汇编
Asm()              tree   函数处理  asm      指令
```

## 📍 关键代码位置

### 1. 编译器主入口
**文件**: `compiler/compiler.go`  
**行号**: 1865
```go
case name == "device.Asm" || name == "device/arm.Asm" || name == "device/arm64.Asm" || name == "device/avr.Asm" || name == "device/riscv.Asm":
    return b.createInlineAsm(instr.Args)
```

### 2. 内联汇编实现
**文件**: `compiler/inlineasm.go`  
**行号**: 24-30
```go
func (b *builder) createInlineAsm(args []ssa.Value) (llvm.Value, error) {
    fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{}, false)
    asm := constant.StringVal(args[0].(*ssa.Const).Value)
    target := llvm.InlineAsm(fnType, asm, "", true, false, 0, false)
    return b.CreateCall(fnType, target, nil, ""), nil
}
```

### 3. 设备特定实现
**文件**: `src/device/arm/arm.go`  
**行号**: 44
```go
func Asm(asm string)
```

## 🔧 处理流程详解

### 阶段1: 函数识别
```go
// compiler/compiler.go:1860-1868
if fn := instr.StaticCallee(); fn != nil {
    name := fn.RelString(nil)
    switch {
    case name == "device.Asm" || name == "device/arm.Asm":
        return b.createInlineAsm(instr.Args)  // 🎯 内联汇编处理
    }
}
```

**识别的函数名**:
- `device.Asm`
- `device/arm.Asm`  
- `device/arm64.Asm`
- `device/avr.Asm`
- `device/riscv.Asm`

### 阶段2: 参数提取
```go
// compiler/inlineasm.go:27
asm := constant.StringVal(args[0].(*ssa.Const).Value)
```

**过程**:
1. 从SSA指令中提取参数
2. 验证第一个参数是字符串常量
3. 提取汇编指令字符串

### 阶段3: LLVM内联汇编生成
```go
// compiler/inlineasm.go:28-29
target := llvm.InlineAsm(fnType, asm, "", true, false, 0, false)
return b.CreateCall(fnType, target, nil, "")
```

**LLVM参数解析**:
- `fnType`: void function type
- `asm`: 汇编指令字符串
- `""`: 约束字符串 (空=无约束)
- `true`: hasSideEffects (防止优化)
- `false`: isAlignStack
- `0`: inlineAsmDialect  
- `false`: canThrow

## 📊 SSA到IR转换

### Go源码
```go
arm.Asm("wfi")
```

### SSA表示
```
# SSA中间表示
%0 = Const <string> {"wfi"}
%1 = Call <()> device/arm.Asm %0
```

### 编译器处理
```go
// 识别特殊函数调用
if name == "device/arm.Asm" {
    // 提取字符串常量 "wfi"
    asm := "wfi"
    
    // 生成LLVM内联汇编
    llvm.InlineAsm(voidType, "wfi", "", true, false, 0, false)
}
```

### LLVM IR输出
```llvm
call void asm sideeffect "wfi", ""()
```

## 🏗️ 不同架构的处理差异

### ARM架构
```go
// src/device/arm/arm.go
func Asm(asm string)

// 生成ARM汇编指令
call void asm sideeffect "wfi", ""()
```

### ARM64架构  
```go
// src/device/arm64/arm64.go
func Asm(asm string)

// 生成ARM64汇编指令
call void asm sideeffect "yield", ""()
```

### RISC-V架构
```go
// src/device/riscv/riscv.go  
func Asm(asm string)

// 生成RISC-V汇编指令
call void asm sideeffect "wfi", ""()
```

## 🔍 AsmFull高级功能

### 函数签名
```go
func AsmFull(asm string, regs map[string]interface{}) uintptr
```

### 使用示例
```go
arm.AsmFull(
    "str {value}, {result}",
    map[string]interface{}{
        "value":  1,
        "result": &dest,
    })
```

### 编译器处理
```go
// compiler/inlineasm.go:46
func (b *builder) createInlineAsmFull(instr *ssa.CallCommon) (llvm.Value, error) {
    asmString := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
    // 处理寄存器映射...
}
```

## ⚡ 性能优化特征

### 编译时优化
1. **零开销抽象**: 无函数调用开销
2. **常量折叠**: 编译时字符串处理
3. **直接内联**: 避免间接调用

### 运行时特征
1. **无分支开销**: 直接指令执行
2. **缓存友好**: 指令与代码连续布局
3. **原子性**: 单个机器指令执行

### 防优化机制
```go
// sideeffect = true 防止以下优化:
// 1. 死代码消除 (DCE)
// 2. 指令重排 
// 3. 循环优化
// 4. 内联消除
```

## 🐛 错误处理

### 编译时检查
```go
// 检查参数是否为常量
if !ssa.IsConst(args[0]) {
    return error("inline assembly requires constant string")
}
```

### 链接时处理
- 内联汇编函数声明但不定义
- 链接器会报告未定义符号
- 这是正常行为，因为已转换为内联汇编

### 运行时行为
- 内联汇编直接执行机器指令
- 无Go运行时检查
- 完全依赖硬件行为

## 📈 扩展点

### 新架构支持
1. 在`compiler/compiler.go`添加识别模式
2. 创建`src/device/newarch/`包
3. 实现`Asm`和`AsmFull`函数声明
4. 添加目标三元组支持

### 新约束类型
1. 扩展`createInlineAsmFull`函数
2. 支持更多LLVM约束字符
3. 添加类型检查和验证

这个流程展示了TinyGo如何将高级Go语法无缝转换为底层硬件指令的完整机制。