package main

import (
	"device/arm"
	"unsafe"
)

// 正确示例：所有 map 操作在同一基本块
func correctUsage() {
	var result uint32
	
	// ✅ 所有 map 操作在同一基本块 - 编译时类型确定
	regs := map[string]interface{}{
		"input":  uint32(42),                           // 编译时确定：uint32 → "r" 约束
		"output": uintptr(unsafe.Pointer(&result)),     // 编译时确定：uintptr → "r" 约束  
	}
	
	// 生成的 LLVM IR 类型确定：
	// call void asm "mov ${0:w}, ${1:w}", "r,r" (i32 42, i32 %ptr_as_int)
	arm.AsmFull("mov {output}, {input}", regs)
	
	println("result:", result)
}

func main() {
	correctUsage()
}