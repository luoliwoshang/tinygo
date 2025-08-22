#!/bin/bash

echo "=== LLVM 内联汇编编译时类型确定演示 ==="
echo ""

echo "1. 编译正确示例 (同基本块)..."
echo "命令: ./build/tinygo build -target=microbit -internal-printir good_example.go"
echo ""

# 尝试编译正确示例
if ../build/tinygo build -target=microbit -internal-printir good_example.go 2>&1 | head -20; then
    echo "✅ 正确示例编译成功 - LLVM IR 类型确定"
else
    echo "❌ 编译失败"
fi

echo ""
echo "2. 编译错误示例 (跨基本块)..."
echo "命令: ./build/tinygo build -target=microbit -internal-printir bad_example.go"
echo ""

# 尝试编译错误示例  
if ../build/tinygo build -target=microbit -internal-printir bad_example.go 2>&1; then
    echo "⚠️  意外成功 - 应该报错"
else
    echo "✅ 预期报错: register value map must be created in the same basic block"
fi

echo ""
echo "=== 结论 ==="
echo "LLVM 内联汇编要求编译时确定参数类型和约束，"
echo "跨基本块的 map 构造会导致类型不确定，因此 TinyGo 添加了基本块约束。"