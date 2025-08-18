package main

import "device"

func main() {
	// 情况3：只有输出，无输入
	result := device.AsmFull("mov {}, r0", nil)
	println("Case 3 result:", result)
}