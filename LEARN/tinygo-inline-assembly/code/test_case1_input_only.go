package main

import "device"

func main() {
	// 情况1：只有输入，无输出
	device.AsmFull("mov r0, {value}", map[string]interface{}{
		"value": uint32(42),
	})
	println("Case 1 completed")
}