package main

import "device"

func main() {
	// 测试多个 {} 占位符
	result := device.AsmFull("mov {}, r1; mov {}, r2", nil)
	println("Result:", result)
}