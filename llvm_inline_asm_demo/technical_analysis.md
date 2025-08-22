# LLVM 内联汇编编译时类型确定技术分析

## 核心发现

TinyGo 的基本块约束 **不是 SSA 理论限制**，而是 **LLVM 内联汇编的技术要求**。

## LLVM 内联汇编要求

LLVM 内联汇编在编译时需要生成完整的函数调用：

```llvm
call RetType asm sideeffect "asm_template", "constraints" (Type1 %arg1, Type2 %arg2, ...)
                              ^^^^^^^^^^^^^   ^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                              汇编模板        约束字符串     参数类型和值列表
```

**关键要求**：
- 参数**类型**必须编译时确定
- 约束**字符串**必须编译时确定  
- 函数**签名**必须编译时确定

## 跨基本块的问题

### 问题 1：类型不确定性
```go
regs := make(map[string]interface{})
if condition {
    regs["r0"] = uint32(42)    // uint32 → "r" 约束
} else {
    regs["r0"] = int32(42)     // int32 → "r" 约束 (相同约束，不同类型！)
}
arm.AsmFull("mov {r0}, #1", regs)
```

编译器无法确定：应该生成 `call void asm "mov $0, #1", "r" (i32 %val)` 还是 `call void asm "mov $0, #1", "r" (i32 %val)`？

### 问题 2：参数存在性不确定  
```go
regs := make(map[string]interface{})
if condition {
    regs["r0"] = 42
    regs["r1"] = 100    // 可能不存在
}
arm.AsmFull("add {r0}, {r1}", regs)  // r1 可能未定义
```

### 问题 3：约束字符串动态变化
```go
regs := make(map[string]interface{})
if condition {
    regs["val"] = 42            // 整数 → "r" 约束
} else {
    regs["val"] = &someVar      // 指针 → "*m" 约束 (不同约束！)
}
```

## 实验验证结果

### 正确示例 (同基本块)
- ✅ 编译成功
- ✅ 生成稳定的 LLVM IR
- ✅ 类型在编译时确定

### 错误示例 (跨基本块)  
- ❌ 编译失败
- ❌ 错误信息：`register value map must be created in the same basic block`
- ❌ 无法生成 LLVM IR

## 结论

**基本块约束的真正原因**：
1. **LLVM 技术限制** - 内联汇编需要编译时类型确定
2. **实现简化** - 避免复杂的跨基本块类型分析  
3. **错误预防** - 防止生成无效的 LLVM IR

这是一个**编译器工程决策**，而非编译器理论约束。

## 引申思考

其他编译器（如 GCC）如何处理类似问题：
- 更复杂的控制流分析
- 更精确的类型推导
- 运行时类型检查（性能代价）

TinyGo 选择了**编译时安全 + 实现简化**的保守方案。