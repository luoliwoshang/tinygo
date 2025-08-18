package main

import "device"

func main() {
	// 情况2：有输出和输入
	result := device.AsmFull("mov {}, {value}", map[string]interface{}{
		"value": uint32(42),
	})
	println("Case 2 result:", result)
}