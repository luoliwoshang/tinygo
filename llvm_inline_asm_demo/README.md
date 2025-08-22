# LLVM 内联汇编编译时类型确定演示

本目录演示了为什么 TinyGo 的 AsmFull 需要基本块约束 - 这是 LLVM 内联汇编的固有特性要求。

## 核心问题

LLVM 内联汇编需要在**编译时**确定：
1. 参数的确切类型
2. 约束字符串  
3. 函数签名

## 文件说明

- `good_example.go` - 正确的同基本块用法
- `bad_example.go` - 跨基本块的问题用法  
- `analyze_output.sh` - 分析生成的 LLVM IR

## 验证方法

```bash
# 编译正确示例并查看 LLVM IR
./build/tinygo build -target=microbit -internal-printir good_example.go

# 尝试编译错误示例（会报错）
./build/tinygo build -target=microbit -internal-printir bad_example.go
```

## 预期结果

- 正确示例：生成稳定的 LLVM IR
- 错误示例：编译时报错 "register value map must be created in the same basic block"