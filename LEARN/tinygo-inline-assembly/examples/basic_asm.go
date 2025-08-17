// 基础内联汇编示例
// 展示不同架构的内联汇编调用方式

package main

import (
	"device"     // 通用设备包
	"device/arm" // ARM特定包
)

func main() {
	// 1. 通用内联汇编 (适用于所有架构)
	device.Asm("nop") // 空操作指令
	
	// 2. ARM特定内联汇编
	arm.Asm("wfi")     // 等待中断
	arm.Asm("wfe")     // 等待事件
	arm.Asm("sev")     // 发送事件
	arm.Asm("dmb")     // 数据内存屏障
	arm.Asm("dsb")     // 数据同步屏障
	arm.Asm("isb")     // 指令同步屏障
}

// 不同架构的示例
func architectureSpecific() {
	// ARM64
	// arm64.Asm("yield")
	
	// AVR
	// avr.Asm("sleep")
	
	// RISC-V  
	// riscv.Asm("wfi")
}

/*
编译器识别的函数名模式:
- device.Asm
- device/arm.Asm  
- device/arm64.Asm
- device/avr.Asm
- device/riscv.Asm

源码位置: compiler/compiler.go:1865
*/