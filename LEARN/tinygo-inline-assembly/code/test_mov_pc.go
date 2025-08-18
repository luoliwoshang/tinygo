package main

import "device"

func main() {
	result := device.AsmFull("MOV {}, PC", nil)
	println("PC value:", result)
}