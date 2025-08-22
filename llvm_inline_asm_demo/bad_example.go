package main

import (
	"device/arm"
	"unsafe"
)

// 错误示例：跨基本块的 map 构造
func problematicUsage() {
	var result uint32
	
	// ❌ 跨基本块构造 - 编译时类型不确定
	regs := make(map[string]interface{})
	
	// 根据条件在不同基本块设置值
	if true { // 这会创建新的基本块
		regs["input"] = uint32(42)                      // Block A: uint32 类型
		regs["output"] = uintptr(unsafe.Pointer(&result))
	} else {  
		regs["input"] = int32(42)                       // Block B: int32 类型 (不同类型！)
		regs["output"] = uintptr(unsafe.Pointer(&result))
	}
	
	// 编译器无法确定：
	// 1. input 是 uint32 还是 int32？
	// 2. 应该生成什么约束字符串？
	// 3. LLVM IR 的函数签名是什么？
	arm.AsmFull("mov {output}, {input}", regs)  // 编译错误！
	
	println("result:", result)
}

func main() {
	problematicUsage()
}