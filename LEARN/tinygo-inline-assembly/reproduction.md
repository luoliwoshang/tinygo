# TinyGo 内联汇编完整复现流程

## 🎯 目标
复现和观察TinyGo如何将Go代码中的`device.Asm()`调用转换为LLVM内联汇编IR。

## 📋 环境要求

### 系统环境
- macOS/Linux/Windows
- Git
- Go 1.19+
- LLVM 工具链

### TinyGo源码
```bash
git clone https://github.com/tinygo-org/tinygo.git
cd tinygo
make build  # 编译TinyGo编译器
```

## 🔍 步骤1: 创建测试代码

### 基础WFI示例
```bash
cat > test_wfi.go << 'EOF'
package main

import "device/arm"

func main() {
	arm.Asm("wfi")
}
EOF
```

### 系统复位示例  
```bash
cat > test_reset.go << 'EOF'
package main

import "device/arm"

func SystemReset() {
	for {
		arm.Asm("wfi")
	}
}

func main() {
	// SystemReset() // 注释掉避免实际复位
}
EOF
```

## 🔍 步骤2: 编译并输出LLVM IR

### ARM目标 (推荐)
```bash
# 编译microbit目标并输出IR
./build/tinygo build -target=microbit -internal-printir test_wfi.go

# 保存IR到文件
./build/tinygo build -target=microbit -internal-printir test_wfi.go > wfi_ir.txt 2>&1
```

### WASM目标 (更简洁的IR)
```bash
./build/tinygo build -target=wasm -internal-printir test_wfi.go > wasm_ir.txt 2>&1
```

## 🔍 步骤3: 搜索关键IR模式

### 查找内联汇编
```bash
# 搜索asm sideeffect模式
grep -A5 -B5 "asm.*sideeffect.*wfi" wfi_ir.txt

# 搜索main函数
grep -A10 -B5 "define.*main.main" wfi_ir.txt
```

### 期望的输出模式
```llvm
; WFI指令的内联汇编
call void asm sideeffect "wfi", ""(), !dbg !21227

; 在循环中的使用
for.body:                                         
  call void asm sideeffect "wfi", ""()
  br label %for.body
```

## 🔍 步骤4: 分析编译器源码

### 查看编译器处理逻辑
```bash
# 查看内联汇编识别代码
grep -n "device.Asm\|device/arm.Asm" compiler/compiler.go

# 查看内联汇编实现
cat compiler/inlineasm.go
```

### 关键代码位置
- **识别**: `compiler/compiler.go:1865`
- **实现**: `compiler/inlineasm.go:24-30`  
- **ARM实现**: `src/device/arm/arm.go:44`

## 🔍 步骤5: 验证不同指令

### 测试多种ARM指令
```bash
cat > test_multiple.go << 'EOF'
package main

import "device/arm"

func main() {
	arm.Asm("nop")  // 空操作
	arm.Asm("wfi")  // 等待中断
	arm.Asm("wfe")  // 等待事件
	arm.Asm("sev")  // 发送事件
	arm.Asm("dmb")  // 数据内存屏障
}
EOF

./build/tinygo build -target=microbit -internal-printir test_multiple.go | grep "asm sideeffect"
```

## 🔍 步骤6: 对比不同架构

### ARM64
```bash
# 如果支持ARM64目标
./build/tinygo build -target=pico -internal-printir test_wfi.go | grep "asm sideeffect"
```

### RISC-V
```bash
# 创建RISC-V示例
cat > test_riscv.go << 'EOF'
package main

import "device/riscv"

func main() {
	riscv.Asm("wfi")
}
EOF
```

## 📊 预期结果

### 成功标志
1. **编译通过**: 无语法错误
2. **找到IR**: 看到 `call void asm sideeffect` 模式
3. **指令正确**: 汇编字符串与输入匹配
4. **标志存在**: 包含 `sideeffect` 防优化标志

### 常见IR模式
```llvm
; 单个指令
call void asm sideeffect "wfi", ""()

; 带约束的指令  
call void asm sideeffect "", "r"(ptr %2)

; 在函数中
define internal void @functionName() {
entry:
  call void asm sideeffect "instruction", ""()
  ret void
}
```

## 🐛 故障排除

### 编译错误
- 检查TinyGo是否正确编译
- 确认目标平台支持
- 验证导入路径正确

### 找不到IR
- 确认使用了 `-internal-printir` 标志
- 检查stderr输出: `2>&1`
- 尝试不同的搜索模式

### 链接错误
- 内联汇编函数声明但未定义是正常的
- 关注IR输出而不是最终可执行文件

## 📈 进阶分析

### 优化级别影响
```bash
# 不同优化级别的对比
./build/tinygo build -opt=0 -target=microbit -internal-printir test_wfi.go
./build/tinygo build -opt=2 -target=microbit -internal-printir test_wfi.go
```

### 调试信息
```bash
# 包含调试信息
./build/tinygo build -no-debug=false -target=microbit -internal-printir test_wfi.go | grep -A3 -B3 "!dbg"
```

## 🎯 验证清单

- [ ] TinyGo编译器构建成功
- [ ] 测试代码编译通过  
- [ ] 成功输出LLVM IR
- [ ] 找到 `asm sideeffect` 模式
- [ ] 指令字符串匹配输入
- [ ] 理解编译器处理流程
- [ ] 验证不同架构差异

完成所有步骤后，你将完全理解TinyGo内联汇编的工作机制！