// WFI (Wait For Interrupt) 指令示例
// 展示如何在TinyGo中使用ARM的低功耗等待指令

package main

import "device/arm"

func main() {
	// WFI指令让CPU进入低功耗状态，等待中断唤醒
	// 这在嵌入式系统中用于节省电力
	arm.Asm("wfi")
}

// SystemReset 函数中的实际使用场景
func SystemReset() {
	// 设置系统复位请求
	// SCB.AIRCR.Set((0x5FA << SCB_AIRCR_VECTKEY_Pos) | SCB_AIRCR_SYSRESETREQ_Msk)
	
	// 等待复位生效期间进入低功耗状态
	for {
		arm.Asm("wfi")
	}
}

/*
编译命令:
./build/tinygo build -target=microbit -internal-printir wfi_example.go

生成的LLVM IR:
call void asm sideeffect "wfi", ""(), !dbg !21227
*/