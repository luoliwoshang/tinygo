# TinyGo AsmFull 测试用例代码

本目录包含用于验证 TinyGo AsmFull 内联汇编行为的测试用例。

## 📋 文件列表

### 核心验证用例

| 文件名 | 用途 | 特征 |
|--------|------|------|
| `test_case1_input_only.go` | 只有输入参数 | 无`{}`，有`{value}` |
| `test_case2_output_input.go` | 输出+输入参数 | 有`{}`，有`{value}` |
| `test_case3_output_only.go` | 只有输出参数 | 有`{}`，无输入 |

### 辅助测试用例

| 文件名 | 用途 | 说明 |
|--------|------|------|
| `test_mov_pc.go` | 获取程序计数器 | 基础输出示例 |
| `test_multiple_empty_braces.go` | 多个`{}`占位符 | 验证共享寄存器行为 |
| `test_no_braces.go` | 无占位符 | 纯执行，无返回值 |

## 🔧 使用方法

### 编译测试
```bash
# 基础编译
./build/tinygo build -target=microbit <文件名>

# 查看LLVM IR
./build/tinygo build -target=microbit -internal-printir <文件名>

# 查看生成的汇编
./build/tinygo build -target=microbit -internal-printasm <文件名>
```

### 运行示例
```bash
# 编译并运行（需要目标硬件或模拟器）
./build/tinygo run -target=microbit test_case1_input_only.go
```

## 📊 验证结果

这些测试用例验证了 TinyGo AsmFull 的以下行为：

1. **操作数编号规律**：
   - 无输出：`{name}` → `$0`, `$1`, `$2`...
   - 有输出：`{}` → `$0`, `{name}` → `$1`, `$2`...

2. **返回值机制**：
   - 有`{}`：返回 `i32`/`uintptr`
   - 无`{}`：返回 `void`

3. **多重占位符**：
   - 多个`{}`共享同一输出寄存器
   - 最后写入的指令决定返回值

## 🔗 相关文档

- [AsmFull uintptr 设计分析](../analysis/asmfull_uintptr_design.md)
- [编译器处理流程](../analysis/compiler_flow.md)
- [LLVM 内联汇编参考](../analysis/llvm_inline_asm_reference.md)