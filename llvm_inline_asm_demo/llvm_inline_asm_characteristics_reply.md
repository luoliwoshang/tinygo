# LLVM 内联汇编调用特性分析

## 1. LLVM 内联汇编的基本结构

LLVM 内联汇编在生成 IR 时需要创建一个完整的函数调用：

```llvm
call ReturnType asm [sideeffect] [alignstack] "asm_template", "constraint_string" (ArgumentType1 %arg1, ArgumentType2 %arg2, ...)
     ^^^^^^^^^^     ^^^^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
     返回类型       副作用和对齐标志              汇编模板         约束字符串        参数类型和值列表
```

### 1.1 关键组成部分

**汇编模板 (asm_template)**：
- 包含实际的汇编指令
- 使用 `$0`, `$1`, `$2` 等占位符引用参数
- 例如：`"mov $0, $1"`

**约束字符串 (constraint_string)**：
- 描述每个参数的约束条件
- 用逗号分隔：`"r,r,m"` (寄存器,寄存器,内存)
- 例如：`"=r,r"` (输出到寄存器,输入从寄存器)

**参数列表**：
- 每个参数都有确定的 LLVM 类型
- 参数顺序与约束字符串一一对应
- 例如：`(i32 %value, i32* %ptr)`

## 2. 编译时确定性要求

### 2.1 类型必须在编译时确定

LLVM 内联汇编**不允许**运行时类型多态：

```llvm
// ✅ 正确：类型明确
call void asm "mov $0, $1", "r,r" (i32 42, i32 %var)

// ❌ 错误：类型不确定
call void asm "mov $0, $1", "r,r" (%unknown_type %val1, %unknown_type %val2)
```

### 2.2 约束字符串必须在编译时确定

```llvm
// ✅ 正确：约束明确
call void asm "mov $0, $1", "r,m" (i32 %reg_val, i32* %mem_ptr)

// ❌ 错误：约束不确定 
call void asm "mov $0, $1", %dynamic_constraint (...) 
```

### 2.3 参数数量必须在编译时确定

```llvm
// ✅ 正确：参数数量固定
call void asm "add $0, $1, $2", "r,r,r" (i32 %a, i32 %b, i32 %c)

// ❌ 错误：参数数量动态
call void asm %dynamic_template, %dynamic_constraints (%variable_args...)
```

## 3. TinyGo 中的实现挑战

### 3.1 从 Go map 到 LLVM 参数的转换

TinyGo 需要将这样的 Go 代码：

```go
regs := map[string]interface{}{
    "input":  42,
    "output": &result,
}
arm.AsmFull("mov {output}, {input}", regs)
```

转换为这样的 LLVM IR：

```llvm
call void asm "mov $0, $1", "r,r" (i32* %result_ptr, i32 42)
```

这个转换过程需要：
1. 解析汇编模板中的 `{name}` 占位符
2. 从 map 中查找对应的值和类型
3. 生成正确的约束字符串
4. 按正确顺序排列参数

### 3.2 编译时分析的限制

为了生成正确的 LLVM IR，TinyGo 编译器必须在**编译时**确定：

- **每个参数的确切类型**：`int32` → `"r"` 约束，`*int32` → `"m"` 约束
- **参数的完整列表**：哪些 `{name}` 会被使用
- **参数的顺序**：生成 `$0, $1, $2...` 的正确映射

这就是为什么需要基本块约束的根本原因。



#### 为什么移除了pointer的支持
https://github.com/tinygo-org/tinygo/commit/cad6a57077c7887025ba532a03f41d6ad78daa72
https://reviews.llvm.org/D116531 llvm相关的讨论