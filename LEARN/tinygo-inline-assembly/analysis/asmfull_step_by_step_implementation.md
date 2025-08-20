# TinyGo AsmFull 逐步实现解析

## 📖 概述

本文档详细分析 TinyGo AsmFull 从 Go 源码到最终 ARM 机器码的完整转换过程，基于实际代码逐步追踪编译器内部处理机制。

## 🎯 分析目标

以下面的 Go 代码为例，追踪完整的编译转换流程：

```go
package main

import "device"

func main() {
    result := device.AsmFull("mov {}, {value}", map[string]interface{}{
        "value": uint32(42),
    })
    println("Result is:", result)
}
```

## 🚀 逐步转换过程

### 步骤 1：Go 编译器 AST → SSA 转换

**SSA（Static Single Assignment）表示**：
```go
func main():
b0:
  t0 = const "mov {}, {value}"                   // 汇编字符串常量
  t1 = make map[string]interface{}              // 创建空 map
  t2 = const "value"                            // key 常量
  t3 = const 42:uint32                          // value 常量
  t4 = make interface{} <- uint32 (t3)          // 装箱为 interface{}
  map update t1[t2] = t4                        // map["value"] = 42
  t5 = call device.AsmFull(t0, t1)              // 调用 AsmFull
  t6 = call println("Result is:", t5)           // 打印结果
  return
```

**关键特征**：
- 汇编字符串 `t0` 是编译时常量
- Map 操作在同一个基本块内完成
- `t5` 是 TinyGo 编译器需要特殊处理的调用

### 步骤 2：TinyGo 编译器函数识别

**位置**：`compiler/compiler.go:1867`

```go
if fn := instr.StaticCallee(); fn != nil {
    name := fn.RelString(nil)  // 获取函数名："device.AsmFull"
    switch {
    case name == "device.AsmFull" || name == "device/arm.AsmFull" || ...：
        return b.createInlineAsmFull(instr)  // 🎯 特殊处理分支
    }
}
```

**编译器决策**：
- 识别这不是普通函数调用
- 而是编译器内建函数，需要特殊处理
- 调用 `createInlineAsmFull` 进行内联汇编生成

### 步骤 3：参数提取与解析

**位置**：`compiler/inlineasm.go:46-68`

#### 3.1 提取汇编字符串
```go
asmString := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
// asmString = "mov {}, {value}"
```

#### 3.2 解析参数 Map
```go
registers := map[string]llvm.Value{}
if registerMap, ok := instr.Args[1].(*ssa.MakeMap); ok {
    for _, r := range *registerMap.Referrers() {
        switch r := r.(type) {
        case *ssa.MapUpdate:
            key := constant.StringVal(r.Key.(*ssa.Const).Value)           // "value"
            registers[key] = b.getValue(r.Value.(*ssa.MakeInterface).X, ...) // <LLVM i32 42>
        }
    }
}
```

**解析结果**：
```go
asmString = "mov {}, {value}"
registers = map[string]llvm.Value{
    "value": <LLVM Value representing i32 42>
}
```

### 步骤 4：输出占位符处理

**位置**：`compiler/inlineasm.go:76-83`

```go
hasOutput := false
asmString = regexp.MustCompile(`\{\}`).ReplaceAllStringFunc(asmString, func(s string) string {
    hasOutput = true
    return "$0"
})
if hasOutput {
    constraints = append(constraints, "=&r")
    registerNumbers[""] = 0
}
```

**转换过程**：
```
原始：  "mov {}, {value}"
处理后："mov $0, {value}"
状态：  hasOutput = true
约束：  constraints = ["=&r"]
编号：  registerNumbers = {"": 0}
```

**约束解释**：
- `=`：输出操作数
- `&`：early clobber（提前破坏，避免寄存器冲突）
- `r`：通用寄存器

### 步骤 5：输入占位符处理

**位置**：`compiler/inlineasm.go:84-112`

```go
asmString = regexp.MustCompile(`\{[a-zA-Z]+\}`).ReplaceAllStringFunc(asmString, func(s string) string {
    name := s[1 : len(s)-1]  // 提取 "value"
    
    // 验证参数存在
    if _, ok := registers[name]; !ok {
        err = b.makeError(instr.Pos(), "unknown register name: "+name)
        return s
    }
    
    // 分配操作数编号
    if _, ok := registerNumbers[name]; !ok {
        registerNumbers[name] = len(registerNumbers)  // 分配编号 1
        argTypes = append(argTypes, registers[name].Type())  // 添加 i32 类型
        args = append(args, registers[name])                 // 添加值 42
        
        // 生成约束
        switch registers[name].Type().TypeKind() {
        case llvm.IntegerTypeKind:
            constraints = append(constraints, "r")  // 通用寄存器约束
        }
    }
    
    return fmt.Sprintf("${%v}", registerNumbers[name])  // 返回 "${1}"
})
```

**转换过程**：
```
输入：     "mov $0, {value}"
处理 {value}：
  - name = "value"
  - registerNumbers["value"] = 1
  - argTypes = [i32]
  - args = [<LLVM i32 42>]
  - constraints = ["=&r", "r"]
输出：     "mov $0, ${1}"
```

### 步骤 6：LLVM 内联汇编生成

**位置**：`compiler/inlineasm.go:116-130`

```go
// 确定返回类型
var outputType llvm.Type
if hasOutput {
    outputType = b.uintptrType    // ARM 上是 i32
} else {
    outputType = b.ctx.VoidType()
}

// 构造函数类型
fnType := llvm.FunctionType(outputType, argTypes, false)
// fnType = i32 (i32) - 返回 i32，接受一个 i32 参数

// 创建内联汇编
target := llvm.InlineAsm(
    fnType,                           // 函数类型
    asmString,                        // "mov $0, ${1}"
    strings.Join(constraints, ","),   // "=&r,r"
    true,                            // hasSideEffects
    false,                           // isAlignStack
    0,                               // dialect
    false                            // canThrow
)

// 生成调用
result := b.CreateCall(fnType, target, args, "")
```

**生成的 LLVM IR**：
```llvm
%0 = call i32 asm sideeffect "mov $0, ${1}", "=&r,r"(i32 42), !dbg !59091
```

### 步骤 7：LLVM 后端架构特定处理

#### 7.1 寄存器分配（ARM 目标）

LLVM ARM 后端根据约束字符串进行寄存器分配：

```
约束 "=&r,r" 在 ARM 上的映射：
- $0 (输出，=&r)：分配到 r0 寄存器
- $1 (输入，r)：  分配到 r1 寄存器
```

#### 7.2 汇编模板展开

```
LLVM 模板：    "mov $0, ${1}"
ARM 展开结果： "mov r0, r1"
```

#### 7.3 最终机器码生成

```arm
; 编译器生成的完整指令序列
mov r1, #42    ; 加载常量 42 到输入寄存器 r1
mov r0, r1     ; 执行用户的汇编指令：r1 内容拷贝到 r0
; r0 现在包含值 42，作为函数返回值
```

### 步骤 8：实际验证

使用 TinyGo 编译器验证转换结果：

```bash
./build/tinygo build -target=microbit -internal-printir simple_asmfull_demo.go 2>&1 | grep -A2 -B2 'asm.*mov.*\$0.*\$'
```

**实际输出**：
```llvm
%0 = call i32 asm sideeffect "mov $0, ${1}", "=&r,r"(i32 42), !dbg !59091
```

证实了我们的分析完全正确！

## 🔧 关键技术点深度解析

### 0. SSA 阶段和调用时处理理解

#### SSA 优化阶段

我们在 `createInlineAsmFull` 中看到的 SSA 是**经过 Lifting 优化后的 SSA**：

**Lifting 前（朴素 SSA）**：
```ssa
%regMap_addr = alloca map[string]interface{}  // 栈上分配
store %tmp1, %regMap_addr                     // 存储到栈
%tmp2 = load %regMap_addr                     // 从栈加载
store %tmp3, %regMap_addr                     // 存储回栈
```

**Lifting 后（优化 SSA）**：
```ssa
%0 = MakeMap(map[string]interface{})          // 直接 SSA 值
%1 = MapUpdate(%0, "r0", &result)             // 直接数据流
%2 = MapUpdate(%1, "r1", a)                   // 直接数据流  
%3 = Call device.AsmFull("...", %2)           // 直接使用
```

**Lifting 优化的好处**：
- ✅ 消除栈分配和 load/store 操作
- ✅ 寄存器化局部变量
- ✅ 插入 φ 节点处理控制流合并
- ✅ 清晰的数据依赖关系，便于编译器分析

#### 调用时处理 vs 函数定义时处理

**关键理解**：`createInlineAsmFull` 是在**调用时（Call-site）**执行，不是函数定义时：

```go
// ✅ 调用时 - 编译器能看到完整结构
device.AsmFull("add {r0}, {r1}, {r2}", map[string]interface{}{
    "r0": &result,  // ← SSA: MakeMap + MapUpdate 序列
    "r1": a,        // ← 编译器在此调用点能完全分析
    "r2": b,
})                   // ← 触发 createInlineAsmFull

// ❌ 不是函数定义时 - 因为这是 builtin 函数
func AsmFull(asm string, regs map[string]interface{}) uintptr {
    // 没有 Go 实现，编译器内建处理
}
```

**为什么能获得 MakeMap**：
```go
if registerMap, ok := instr.Args[1].(*ssa.MakeMap); ok {
    // instr 是函数调用的 SSA 表示
    // instr.Args[1] 是调用的第二个参数
    // 由于传入字面量 map，所以是 *ssa.MakeMap
```

**编译时确定性要求**：
- Map 必须是字面量（编译时常量）
- 不能是运行时变量（无法静态分析）
- 必须在同一基本块（避免控制流复杂性）

### 1. 操作数编号规律

| 情况 | `{}` 输出 | `{name}` 输入 | 编号规律 |
|------|-----------|---------------|----------|
| 无输出 | 无 | 有 | `{name}` → `$0`, `$1`, `$2`... |
| 有输出 | 有 | 有/无 | `{}` → `$0`, `{name}` → `$1`, `$2`... |

**设计原则**：输出操作数优先占据 `$0` 位置

### 2. 约束字符串构建

```go
// 输出约束
"=&r"  // = 输出，& early clobber，r 通用寄存器

// 输入约束  
"r"    // r 通用寄存器

// 组合约束
"=&r,r"  // 一个输出 + 一个输入
```

### 3. 类型系统集成

```go
switch registers[name].Type().TypeKind() {
case llvm.IntegerTypeKind:
    constraints = append(constraints, "r")
case llvm.PointerTypeKind:
    // TinyGo 0.23+ 不再支持
    err = b.makeError(instr.Pos(), "support for pointer operands was dropped")
}
```

**支持的类型**：
- ✅ 整数类型（int, uint, int32, uint32 等）
- ❌ 指针类型（在 TinyGo 0.23+ 中移除）
- ❌ 其他复合类型

### 4. 占位符替换和约束生成详解

#### 4.1 正则匹配和替换过程

```go
// 处理输出占位符 {}
asmString = regexp.MustCompile(`\{\}`).ReplaceAllStringFunc(asmString, func(s string) string {
    hasOutput = true
    return "$0"  // {} 总是映射到 $0
})

// 处理命名占位符 {name}
asmString = regexp.MustCompile(`\{[a-zA-Z]+\}`).ReplaceAllStringFunc(asmString, func(s string) string {
    name := s[1 : len(s)-1]  // 提取名称: {r0} → r0
    
    // 验证和去重
    if _, ok := registers[name]; !ok {
        err = b.makeError(instr.Pos(), "unknown register name: "+name)
        return s
    }
    
    // 分配编号（去重机制）
    if _, ok := registerNumbers[name]; !ok {
        registerNumbers[name] = len(registerNumbers)
        argTypes = append(argTypes, registers[name].Type())
        args = append(args, registers[name])
        constraints = append(constraints, "r")  // 添加约束
    }
    
    return fmt.Sprintf("${%v}", registerNumbers[name])  // {r0} → ${1}
})
```

#### 4.2 完整转换示例

**输入**：
```go
device.AsmFull("add {r0}, {r1}, {r2}", map[string]interface{}{
    "r0": &result,
    "r1": a,
    "r2": b,
})
```

**处理步骤**：
1. 解析 registers map：`{"r0": &result, "r1": a, "r2": b}`
2. 替换 `{r0}` → 分配编号 0 → `${0}`，约束 `["r"]`
3. 替换 `{r1}` → 分配编号 1 → `${1}`，约束 `["r", "r"]`  
4. 替换 `{r2}` → 分配编号 2 → `${2}`，约束 `["r", "r", "r"]`

**输出**：
```go
asmString = "add ${0}, ${1}, ${2}"
constraints = ["r", "r", "r"]
args = [&result, a, b]
```

**最终 LLVM**：
```llvm
call void asm sideeffect "add ${0}, ${1}, ${2}", "r,r,r"(ptr %result, i32 %a, i32 %b)
```

### 5. 错误处理机制

```go
// 编译时验证
if r.Block() != registerMap.Block() {
    return llvm.Value{}, b.makeError(instr.Pos(), 
        "register value map must be created in the same basic block")
}

// 参数检查
if _, ok := registers[name]; !ok {
    err = b.makeError(instr.Pos(), "unknown register name: "+name)
}
```

**安全保证**：
- 所有错误在编译时发现
- 提供精确的错误位置信息
- 防止运行时意外错误

## 🎯 性能优化分析

### 1. 编译时优化

- **零运行时开销**：所有处理在编译时完成
- **直接内联**：生成直接的 LLVM 内联汇编
- **无函数调用**：编译器内建处理，避免函数调用开销

### 2. 寄存器优化

- **Early Clobber**：`=&r` 避免输入输出寄存器冲突
- **LLVM 优化**：利用 LLVM 先进的寄存器分配算法
- **架构感知**：自动选择目标架构的最优寄存器

### 3. 内存优化

- **编译时展开**：map 参数在编译时完全展开
- **零分配**：运行时无额外内存分配
- **直接映射**：Go 值直接映射到 LLVM 值

## 📊 完整转换流程图

```
Go 源码
  ↓ Go 编译器
SSA 表示
  ↓ TinyGo 识别
createInlineAsmFull
  ↓ 参数解析
registers map + asmString
  ↓ 正则处理
"{}" → "$0", "{name}" → "${N}"
  ↓ LLVM 生成
asm sideeffect "...", "constraints"(args)
  ↓ LLVM 后端
架构特定寄存器分配
  ↓ 汇编生成
目标架构机器码
  ↓ 执行
硬件指令执行
```

## 🔍 与现有文档的补充关系

本文档与现有分析文档的关系：

- **[asmfull_uintptr_design.md](./asmfull_uintptr_design.md)**：专注于返回值类型设计
- **[compiler_flow.md](./compiler_flow.md)**：宏观编译器流程
- **本文档**：微观逐步实现细节，实际代码验证

三者形成完整的 AsmFull 机制分析体系，从设计理念到实现细节，从理论分析到实际验证。

---

**相关文档**：
- [AsmFull uintptr 设计分析](./asmfull_uintptr_design.md)
- [编译器处理流程](./compiler_flow.md) 
- [LLVM 内联汇编参考](./llvm_inline_asm_reference.md)
- [测试用例集合](../code/)