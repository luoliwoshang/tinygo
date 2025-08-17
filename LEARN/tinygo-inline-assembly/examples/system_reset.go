// 系统复位示例 - 展示内联汇编在实际系统功能中的应用
// 这是TinyGo源码中真实的SystemReset实现

package main

import "device/arm"

// SystemReset 执行硬件系统复位
// 这是从src/device/arm/scb.go中提取的实际代码
func SystemReset() {
	// 第一步：设置系统复位请求
	// 向ARM Cortex-M的系统控制块(SCB)写入复位命令
	// SCB.AIRCR.Set((0x5FA << SCB_AIRCR_VECTKEY_Pos) | SCB_AIRCR_SYSRESETREQ_Msk)
	
	// 第二步：等待复位生效
	// 复位需要几个时钟周期才能生效，在此期间让CPU进入低功耗状态
	for {
		arm.Asm("wfi") // Wait For Interrupt - 低功耗等待
	}
}

// 相关的低功耗函数
func waitForEvents() {
	// WFE - Wait For Event
	// 等待其他处理器核心或外设发送的事件信号
	arm.Asm("wfe")
}

func sendEvent() {
	// SEV - Send Event
	// 向所有处理器核心发送事件信号，唤醒等待WFE的核心
	arm.Asm("sev")
}

func memoryBarriers() {
	// 内存屏障指令确保内存操作的顺序
	arm.Asm("dmb") // Data Memory Barrier - 数据内存屏障
	arm.Asm("dsb") // Data Synchronization Barrier - 数据同步屏障  
	arm.Asm("isb") // Instruction Synchronization Barrier - 指令同步屏障
}

func main() {
	// 演示各种ARM指令
	memoryBarriers()
	sendEvent()
	waitForEvents()
	
	// 注意：实际使用中不要调用SystemReset()，它会重启系统！
	// SystemReset()
}

/*
LLVM IR输出分析:

1. WFI指令在无限循环中:
for.body:                                         ; preds = %for.body, %gep.next
  call void asm sideeffect "wfi", ""(), !dbg !21227
  br label %for.body, !dbg !21226

2. WFE指令独立调用:
define internal void @runtime.waitForEvents(ptr %context) #0 !dbg !57896 {
entry:
  call void asm sideeffect "wfe", ""(), !dbg !57898
  ret void, !dbg !57899
}

关键特征:
- sideeffect: 防止编译器优化掉
- 空约束 "": 无输入输出参数
- !dbg: 调试信息标记
*/