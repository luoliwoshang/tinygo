package main

import "device"

func main() {
	// 不需要返回值的情况 - 没有{}
	device.AsmFull("nop", nil)
	println("NOP without {} executed")
	
	// 需要返回值的情况 - 有{}
	result := device.AsmFull("MOV {}, PC", nil)
	println("PC with {}:", result)
}